// Command movies is a small Jellyfin-compatible media server for Infuse: it serves Kodi-style
// folders (videos + tinyMediaManager NFO files + local artwork) over the Jellyfin API, with direct
// play only. No scraping, no transcoding, no web UI.
//
// Configuration (environment):
//
//	PORT           listen port (8096)
//	SERVER_NAME    name shown in clients ("Movies")
//	SERVER_ID      stable server id (default: derived from SERVER_NAME)
//	PUBLISHED_URL  external URL, e.g. https://movies.example.com (default: from the request)
//	LIBRARIES      "Name:type:path;..." with type movies or tvshows
//	               (default "Movies:movies:/data/movies;TV Shows:tvshows:/data/tvshows")
//	USERS          "name:password,..." (or USERS_FILE with one name:password per line); unset:
//	               no authentication, every client is the user "admin" (for a trusted home network)
//	DATA_DIR       state, ffprobe cache, resized images (/config)
//	SCAN_INTERVAL  how often to rescan the folders (30m)
//	LOG_LEVEL      debug, info, warn, error (info)
//	IMPORT_PATH_MAP  "from=to,..." prefix rewrites for paths in Jellyfin import files
//
// Watch state exported from Jellyfin (tools/jellyfin-export.sh) is imported at startup from
// $DATA_DIR/import/*.jsonl; imported files are renamed to *.done.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/zemiak/movies/internal/api"
	"github.com/zemiak/movies/internal/library"
	"github.com/zemiak/movies/internal/store"
)

func env(name, def string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return def
}

var reSlug = regexp.MustCompile(`[^a-z0-9]+`)

func parseLibraries(spec string) ([]library.LibraryConfig, error) {
	var out []library.LibraryConfig
	seen := map[string]bool{}
	for _, part := range strings.Split(spec, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		f := strings.SplitN(part, ":", 3)
		if len(f) != 3 || (f[1] != "movies" && f[1] != "tvshows") {
			return nil, fmt.Errorf("LIBRARIES: %q is not Name:movies|tvshows:/path", part)
		}
		key := strings.Trim(reSlug.ReplaceAllString(strings.ToLower(f[0]), "-"), "-")
		if seen[key] {
			return nil, fmt.Errorf("LIBRARIES: duplicate name %q", f[0])
		}
		seen[key] = true
		out = append(out, library.LibraryConfig{Key: key, Name: f[0], CollectionType: f[1], Path: filepath.Clean(f[2])})
	}
	if len(out) == 0 {
		return nil, errors.New("LIBRARIES: no libraries")
	}
	return out, nil
}

func parseUsers() (map[string]string, error) {
	spec := os.Getenv("USERS")
	if file := os.Getenv("USERS_FILE"); file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		spec += "\n" + string(data)
	}
	users := map[string]string{}
	for _, entry := range strings.FieldsFunc(spec, func(r rune) bool { return r == ',' || r == '\n' }) {
		entry = strings.TrimSpace(entry)
		if entry == "" || strings.HasPrefix(entry, "#") {
			continue
		}
		name, pw, ok := strings.Cut(entry, ":")
		if !ok || name == "" || pw == "" {
			return nil, fmt.Errorf("USERS: entry %q is not name:password", name)
		}
		users[name] = pw
	}
	return users, nil
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	var level slog.Level
	if err := level.UnmarshalText([]byte(env("LOG_LEVEL", "info"))); err != nil {
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})))

	libs, err := parseLibraries(env("LIBRARIES", "Movies:movies:/data/movies;TV Shows:tvshows:/data/tvshows"))
	if err != nil {
		return err
	}
	users, err := parseUsers()
	if err != nil {
		return err
	}
	interval, err := time.ParseDuration(env("SCAN_INTERVAL", "30m"))
	if err != nil || interval < time.Minute {
		return fmt.Errorf("SCAN_INTERVAL: must be a duration of at least 1m (%v)", err)
	}
	dataDir := env("DATA_DIR", "/config")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	name := env("SERVER_NAME", "Movies")

	st, err := store.Open(filepath.Join(dataDir, "state.json"))
	if err != nil {
		return err
	}
	probes := library.NewProbeCache(filepath.Join(dataDir, "probe-cache.json"))
	if len(users) == 0 {
		slog.Warn("no USERS: authentication is off, every client is user " + api.OpenUser)
	}
	mgr, pending := library.NewManager(&library.Scanner{Libraries: libs, Probes: probes}, interval)

	importWatchState(dataDir, libs, users, st)

	srv := api.New(api.Config{
		ServerName:   name,
		ServerID:     env("SERVER_ID", library.ID("server:"+name)),
		PublishedURL: os.Getenv("PUBLISHED_URL"),
		Users:        users,
		CacheDir:     filepath.Join(dataDir, "images"),
	}, mgr, st)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go mgr.Run(ctx, pending)

	hs := &http.Server{
		Addr:              ":" + env("PORT", "8096"),
		Handler:           srv,
		ReadHeaderTimeout: 30 * time.Second,
		IdleTimeout:       5 * time.Minute,
	}
	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", hs.Addr, "server", name, "libraries", len(libs), "users", len(users))
		errc <- hs.ListenAndServe()
	}()
	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		slog.Info("shutting down")
		sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = hs.Shutdown(sctx)
	}
	if err := probes.Save(); err != nil {
		slog.Error("saving probe cache", "err", err)
	}
	return st.Close()
}

// importWatchState imports $DATA_DIR/import/*.jsonl (Jellyfin exports).
func importWatchState(dataDir string, libs []library.LibraryConfig, users map[string]string, st *store.Store) {
	files, _ := filepath.Glob(filepath.Join(dataDir, "import", "*.jsonl"))
	if len(files) == 0 {
		return
	}
	var rewrites [][2]string
	for _, m := range strings.Split(os.Getenv("IMPORT_PATH_MAP"), ",") {
		if from, to, ok := strings.Cut(strings.TrimSpace(m), "="); ok {
			rewrites = append(rewrites, [2]string{from, to})
		}
	}
	keyFor := func(p string) string {
		for _, rw := range rewrites {
			if strings.HasPrefix(p, rw[0]) {
				p = rw[1] + strings.TrimPrefix(p, rw[0])
				break
			}
		}
		for _, l := range libs {
			if rel, ok := strings.CutPrefix(p, l.Path+"/"); ok {
				return l.Key + "/" + rel
			}
		}
		return ""
	}
	only := ""
	if len(users) == 0 {
		only = api.OpenUser
	}
	if len(users) == 1 {
		for n := range users {
			only = n
		}
	}
	userFor := func(u string) string {
		if _, ok := users[u]; ok {
			return u
		}
		return only
	}
	for _, f := range files {
		n, skipped, err := st.Import(f, keyFor, userFor)
		if err != nil {
			slog.Error("import failed", "file", f, "err", err)
			continue
		}
		slog.Info("imported watch state", "file", f, "records", n, "skipped", skipped)
		if err := os.Rename(f, f+".done"); err != nil {
			slog.Error("renaming imported file", "file", f, "err", err)
		}
	}
}
