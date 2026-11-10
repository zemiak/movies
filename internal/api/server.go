// Package api implements the subset of the Jellyfin HTTP API that Infuse (and similar clients)
// use to browse a library, show metadata, play files directly and sync watch state.
package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zemiak/movies/internal/library"
	"github.com/zemiak/movies/internal/store"
)

// Version is the Jellyfin version we claim to be; clients check it for API compatibility.
const Version = "10.10.6"

// Config configures the API server.
type Config struct {
	ServerName   string
	ServerID     string
	PublishedURL string            // e.g. https://movies.example.com; empty: derived from the request
	Users        map[string]string // name -> password; empty: no authentication (see OpenUser)
	CacheDir     string            // resized images
	open         bool              // no users configured: authentication off
}

// Server is the HTTP handler.
type Server struct {
	cfg    Config
	lib    *library.Manager
	store  *store.Store
	routes []route
	images *imageCache
	start  time.Time
}

type handlerFunc func(c *ctx)

type route struct {
	methods []string
	parts   []string
	h       handlerFunc
	anon    bool
}

// ctx is one request.
type ctx struct {
	w      http.ResponseWriter
	r      *http.Request
	s      *Server
	params map[string]string
	q      map[string]string // query parameters, lowercased names
	user   string            // authenticated user name ("" for anonymous routes)
	token  string
	auth   map[string]string // fields of the (X-Emby-)Authorization header
	lib    *library.Library
}

// OpenUser is the only user when authentication is off.
const OpenUser = "admin"

// New creates the server. Without users, authentication is off: every request is OpenUser's.
func New(cfg Config, lib *library.Manager, st *store.Store) *Server {
	if len(cfg.Users) == 0 {
		cfg.open = true
		cfg.Users = map[string]string{OpenUser: ""}
	}
	s := &Server{cfg: cfg, lib: lib, store: st, images: newImageCache(cfg.CacheDir), start: time.Now()}
	s.registerRoutes()
	return s
}

func (s *Server) handle(methods string, pattern string, h handlerFunc) {
	s.routes = append(s.routes, route{methods: strings.Split(methods, ","), parts: splitPath(strings.ToLower(pattern)), h: h})
}

func (s *Server) handleAnon(methods string, pattern string, h handlerFunc) {
	s.handle(methods, pattern, h)
	s.routes[len(s.routes)-1].anon = true
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int64
}

func (w *statusWriter) WriteHeader(code int) {
	if w.status == 0 {
		w.status = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += int64(n)
	return n, err
}

// ServeHTTP routes a request.
func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	w := &statusWriter{ResponseWriter: rw}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	segs := splitPath(r.URL.Path)
	// Jellyfin accepts an optional /emby (or /mediabrowser) prefix and is case-insensitive.
	if len(segs) > 0 {
		switch strings.ToLower(segs[0]) {
		case "emby", "mediabrowser", "jellyfin":
			segs = segs[1:]
		}
	}
	c := &ctx{w: w, r: r, s: s, q: map[string]string{}, lib: s.lib.Get()}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			c.q[strings.ToLower(k)] = v[0]
		}
	}
	c.parseAuth()

	rt, params, methodMismatch := s.match(r.Method, segs)
	known := rt != nil
	switch {
	case rt == nil && methodMismatch:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case rt == nil:
		http.NotFound(w, r)
	default:
		c.params = params
		if !rt.anon && !c.authenticate() {
			w.WriteHeader(http.StatusUnauthorized)
			break
		}
		rt.h(c)
	}

	level := slog.LevelInfo
	lp := strings.ToLower(r.URL.Path)
	if strings.Contains(lp, "/images/") || strings.HasSuffix(lp, "/playing/progress") || strings.HasSuffix(lp, "/playing/ping") {
		level = slog.LevelDebug
	}
	if !known {
		level = slog.LevelWarn
	}
	slog.Log(r.Context(), level, "request", "method", r.Method, "path", r.URL.Path, "query", r.URL.RawQuery,
		"status", w.status, "bytes", w.bytes, "took", time.Since(start).Round(time.Millisecond),
		"client", c.auth["client"], "device", c.auth["device"], "user", c.user, "known", known)
}

func (s *Server) match(method string, segs []string) (*route, map[string]string, bool) {
	mismatch := false
	for i := range s.routes {
		rt := &s.routes[i]
		if len(rt.parts) != len(segs) {
			continue
		}
		params := map[string]string{}
		ok := true
		for j, p := range rt.parts {
			if strings.HasPrefix(p, "{") {
				params[p[1:len(p)-1]] = segs[j]
				continue
			}
			if !strings.EqualFold(p, segs[j]) {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		for _, m := range rt.methods {
			if m == method || (m == "GET" && method == http.MethodHead) {
				return rt, params, false
			}
		}
		mismatch = true
	}
	return nil, nil, mismatch
}

// ---------------------------------------------------------------- auth

func (c *ctx) parseAuth() {
	c.auth = map[string]string{}
	h := c.r.Header.Get("X-Emby-Authorization")
	if h == "" {
		h = c.r.Header.Get("Authorization")
	}
	if i := strings.IndexByte(h, ' '); i >= 0 {
		scheme := strings.ToLower(h[:i])
		if scheme == "mediabrowser" || scheme == "emby" {
			for _, part := range splitAuthFields(h[i+1:]) {
				k, v, ok := strings.Cut(part, "=")
				if !ok {
					continue
				}
				v = strings.Trim(strings.TrimSpace(v), `"`)
				if uv, err := urlUnescape(v); err == nil {
					v = uv
				}
				c.auth[strings.ToLower(strings.TrimSpace(k))] = v
			}
		}
	}
	for _, t := range []string{c.auth["token"], c.r.Header.Get("X-Emby-Token"), c.r.Header.Get("X-MediaBrowser-Token"),
		c.q["api_key"], c.q["apikey"]} {
		if t != "" {
			c.token = t
			break
		}
	}
	if c.token == "" {
		if h := c.r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
			c.token = strings.TrimPrefix(h, "Bearer ")
		}
	}
}

func splitAuthFields(s string) []string {
	var out []string
	var cur strings.Builder
	inQ := false
	for _, r := range s {
		switch {
		case r == '"':
			inQ = !inQ
			cur.WriteRune(r)
		case r == ',' && !inQ:
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

func (c *ctx) authenticate() bool {
	if c.s.cfg.open {
		c.user = OpenUser
		return true
	}
	if c.token == "" {
		return false
	}
	t, ok := c.s.store.Token(c.token)
	if !ok {
		return false
	}
	if _, exists := c.s.cfg.Users[t.User]; !exists {
		return false
	}
	c.user = t.User
	return true
}

// ---------------------------------------------------------------- helpers

func (c *ctx) json(v any) {
	c.w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(c.w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		slog.Debug("writing response", "err", err)
	}
}

func (c *ctx) noContent() { c.w.WriteHeader(http.StatusNoContent) }

func (c *ctx) notFound() { http.Error(c.w, "Not found", http.StatusNotFound) }

func (c *ctx) bad(msg string) { http.Error(c.w, msg, http.StatusBadRequest) }

// body decodes a JSON request body (an empty body is fine).
func (c *ctx) body(v any) bool {
	data, err := io.ReadAll(io.LimitReader(c.r.Body, 1<<20))
	if err != nil {
		c.bad("cannot read body")
		return false
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return true
	}
	if err := json.Unmarshal(data, v); err != nil {
		c.bad("invalid JSON: " + err.Error())
		return false
	}
	return true
}

func (c *ctx) param(name string) string { return c.params[name] }

func (c *ctx) qInt(name string, def int) int {
	if v, err := strconv.Atoi(c.q[name]); err == nil {
		return v
	}
	return def
}

func (c *ctx) qBool(name string) (bool, bool) {
	switch strings.ToLower(c.q[name]) {
	case "true", "1":
		return true, true
	case "false", "0":
		return false, true
	}
	return false, false
}

func (c *ctx) qList(name string) []string {
	v := c.q[name]
	if v == "" {
		return nil
	}
	var out []string
	for _, p := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == '|' }) {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// baseURL is the server's address as the client sees it.
func (c *ctx) baseURL() string {
	if c.s.cfg.PublishedURL != "" {
		return strings.TrimRight(c.s.cfg.PublishedURL, "/")
	}
	scheme := "http"
	if c.r.TLS != nil || strings.EqualFold(c.r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := c.r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = c.r.Host
	}
	return scheme + "://" + host
}

func jfTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format("2006-01-02T15:04:05.0000000Z")
}

func parseTime(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.0000000Z", "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
