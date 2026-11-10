package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zemiak/movies/internal/library"
	"github.com/zemiak/movies/internal/store"
)

type testEnv struct {
	t     *testing.T
	ts    *httptest.Server
	token string
	srv   *Server
	lib   *library.Manager
}

func setup(t *testing.T) *testEnv {
	t.Helper()
	root := t.TempDir()
	w := func(p, c string) {
		p = filepath.Join(root, p)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	w("movies/Alien/Alien.m4v", strings.Repeat("x", 1000))
	w("movies/Alien/Alien.nfo", `<movie><title>Alien</title><year>1979</year><runtime>117</runtime></movie>`)
	w("tv/Show/Season 01/S01E01 One.m4v", "x")
	w("tv/Show/Season 01/S01E02 Two.m4v", "x")
	w("tv/Show/Season 01/S01E03 Three.m4v", "x")

	st, err := store.Open(filepath.Join(root, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	mgr, _ := library.NewManager(&library.Scanner{Libraries: []library.LibraryConfig{
		{Key: "movies", Name: "Movies", CollectionType: "movies", Path: filepath.Join(root, "movies")},
		{Key: "tv", Name: "TV", CollectionType: "tvshows", Path: filepath.Join(root, "tv")},
	}}, 0)
	srv := New(Config{ServerName: "Test", ServerID: "sid", Users: map[string]string{"me": "pw"},
		CacheDir: filepath.Join(root, "cache")}, mgr, st)
	e := &testEnv{t: t, ts: httptest.NewServer(srv), srv: srv, lib: mgr}
	t.Cleanup(e.ts.Close)

	var login struct{ AccessToken string }
	code := e.do("POST", "/Users/AuthenticateByName", `{"Username":"me","Pw":"pw"}`, &login)
	if code != 200 || login.AccessToken == "" {
		t.Fatalf("login: %d", code)
	}
	e.token = login.AccessToken
	return e
}

func (e *testEnv) do(method, path, body string, out any) int {
	e.t.Helper()
	req, _ := http.NewRequest(method, e.ts.URL+path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	auth := `MediaBrowser Client="Infuse-Direct", Device="Apple TV", DeviceId="dev1", Version="8.1"`
	if e.token != "" {
		auth += `, Token="` + e.token + `"`
	}
	req.Header.Set("X-Emby-Authorization", auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if out != nil && len(data) > 0 {
		if err := json.Unmarshal(data, out); err != nil {
			e.t.Fatalf("%s %s: %v: %s", method, path, err, data)
		}
	}
	return resp.StatusCode
}

type items struct {
	Items []struct {
		Name, Id, Type string
		UserData       struct {
			Played                bool
			PlaybackPositionTicks int64
			UnplayedItemCount     int
		}
	}
	TotalRecordCount int
}

func TestLoginAndAuth(t *testing.T) {
	e := setup(t)
	if code := e.do("POST", "/Users/AuthenticateByName", `{"Username":"me","Pw":"wrong"}`, nil); code != 401 {
		t.Errorf("wrong password: %d", code)
	}
	tok := e.token
	e.token = ""
	if code := e.do("GET", "/Items", "", nil); code != 401 {
		t.Errorf("no token: %d", code)
	}
	e.token = "bogus"
	if code := e.do("GET", "/Items", "", nil); code != 401 {
		t.Errorf("bad token: %d", code)
	}
	e.token = tok
	// Case-insensitive routes and the /emby prefix.
	var info map[string]any
	if code := e.do("GET", "/emby/system/info/public", "", &info); code != 200 || info["Version"] != Version {
		t.Errorf("public info: %d %v", code, info)
	}
}

func TestBrowseAndWatchState(t *testing.T) {
	e := setup(t)
	var views items
	e.do("GET", "/Users/x/Views", "", &views)
	if views.TotalRecordCount != 2 {
		t.Fatalf("views: %+v", views)
	}
	var movies items
	e.do("GET", "/Users/x/Items?Recursive=true&IncludeItemTypes=Movie&Fields=MediaSources", "", &movies)
	if movies.TotalRecordCount != 1 || movies.Items[0].Name != "Alien" {
		t.Fatalf("movies: %+v", movies)
	}
	alien := movies.Items[0].Id

	// Stop at 50% of a 117-minute movie: resumable.
	pos := int64(117 * 60 * 1e7 / 2)
	body, _ := json.Marshal(map[string]any{"ItemId": alien, "PositionTicks": pos})
	if code := e.do("POST", "/Sessions/Playing/Stopped", string(body), nil); code != 204 {
		t.Fatalf("stopped: %d", code)
	}
	var resume items
	e.do("GET", "/Users/x/Items/Resume", "", &resume)
	if resume.TotalRecordCount != 1 || resume.Items[0].UserData.PlaybackPositionTicks != pos {
		t.Errorf("resume: %+v", resume)
	}
	// Stop at 95%: played, position reset.
	body, _ = json.Marshal(map[string]any{"ItemId": alien, "PositionTicks": int64(117 * 60 * 1e7 * 0.95)})
	e.do("POST", "/Sessions/Playing/Stopped", string(body), nil)
	e.do("GET", "/Users/x/Items/Resume", "", &resume)
	if resume.TotalRecordCount != 0 {
		t.Errorf("resume after finishing: %+v", resume)
	}
	var played items
	e.do("GET", "/Items?Recursive=true&IsPlayed=true", "", &played)
	if played.TotalRecordCount != 1 || played.Items[0].Id != alien {
		t.Errorf("played: %+v", played)
	}

	// Series: mark episode 1 played -> next up is episode 2.
	var series items
	e.do("GET", "/Items?Recursive=true&IncludeItemTypes=Series", "", &series)
	var eps items
	e.do("GET", "/Shows/"+series.Items[0].Id+"/Episodes", "", &eps)
	if eps.TotalRecordCount != 3 {
		t.Fatalf("episodes: %+v", eps)
	}
	e.do("POST", "/Users/x/PlayedItems/"+eps.Items[0].Id, "", nil)
	var next items
	e.do("GET", "/Shows/NextUp", "", &next)
	if next.TotalRecordCount != 1 || next.Items[0].Name != "Two" {
		t.Errorf("next up: %+v", next)
	}
	e.do("GET", "/Items?Recursive=true&IncludeItemTypes=Series", "", &series)
	if series.Items[0].UserData.UnplayedItemCount != 2 {
		t.Errorf("unplayed count: %+v", series.Items[0].UserData)
	}
	// Marking the series unplayed resets all episodes, and syncing clients see the change.
	before := time.Now().UTC().Add(-time.Millisecond).Format(time.RFC3339Nano)
	e.do("DELETE", "/UserPlayedItems/"+series.Items[0].Id, "", nil)
	var changed items
	e.do("GET", "/Items?Recursive=true&IncludeItemTypes=Episode&MinDateLastSavedForUser="+before, "", &changed)
	if changed.TotalRecordCount != 3 {
		t.Errorf("changed since unplayed: %+v", changed)
	}
	e.do("GET", "/Shows/NextUp", "", &next)
	if next.TotalRecordCount != 0 {
		t.Errorf("next up after reset: %+v", next)
	}
}

func TestStream(t *testing.T) {
	e := setup(t)
	var movies items
	e.do("GET", "/Items?Recursive=true&IncludeItemTypes=Movie", "", &movies)
	req, _ := http.NewRequest("GET", e.ts.URL+"/Videos/"+movies.Items[0].Id+"/stream.mp4?static=true&api_key="+e.token, nil)
	req.Header.Set("Range", "bytes=10-19")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != 206 || len(data) != 10 || resp.Header.Get("Content-Range") != "bytes 10-19/1000" {
		t.Errorf("range: %d %d %q", resp.StatusCode, len(data), resp.Header.Get("Content-Range"))
	}
	var pi struct {
		MediaSources []struct {
			Id                 string
			SupportsDirectPlay bool
			Container          string
		}
	}
	e.do("POST", "/Items/"+movies.Items[0].Id+"/PlaybackInfo", `{}`, &pi)
	if len(pi.MediaSources) != 1 || !pi.MediaSources[0].SupportsDirectPlay || pi.MediaSources[0].Container != "mp4" {
		t.Errorf("playback info: %+v", pi)
	}
}

func TestImageTarget(t *testing.T) {
	mk := func(q string) *ctx {
		c := &ctx{q: map[string]string{}}
		for _, kv := range strings.Split(q, "&") {
			if k, v, ok := strings.Cut(kv, "="); ok {
				c.q[k] = v
			}
		}
		return c
	}
	if w, h, ok := target(2000, 3000, mk("maxwidth=400")); !ok || w != 400 || h != 600 {
		t.Errorf("maxwidth: %d %d %v", w, h, ok)
	}
	if w, h, ok := target(3840, 2160, mk("fillwidth=960&fillheight=540")); !ok || w != 960 || h != 540 {
		t.Errorf("fill: %d %d %v", w, h, ok)
	}
	if _, _, ok := target(300, 450, mk("maxwidth=400")); ok {
		t.Errorf("no upscaling")
	}
}

func TestOpenAccess(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "m", "A"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(root, "m", "A", "A.m4v"), []byte("x"), 0o644)
	st, _ := store.Open(filepath.Join(root, "state.json"))
	t.Cleanup(func() { st.Close() })
	mgr, _ := library.NewManager(&library.Scanner{Libraries: []library.LibraryConfig{
		{Key: "m", Name: "M", CollectionType: "movies", Path: filepath.Join(root, "m")}}}, 0)
	ts := httptest.NewServer(New(Config{ServerName: "T", ServerID: "s"}, mgr, st))
	t.Cleanup(ts.Close)
	e := &testEnv{t: t, ts: ts}
	var login struct {
		AccessToken string
		User        struct{ Name string }
	}
	if code := e.do("POST", "/Users/AuthenticateByName", `{"Username":"whoever","Pw":""}`, &login); code != 200 || login.User.Name != OpenUser {
		t.Fatalf("open login: %d %+v", code, login)
	}
	var movies items
	if code := e.do("GET", "/Items?Recursive=true", "", &movies); code != 200 || movies.TotalRecordCount != 1 {
		t.Errorf("no-token request: %d %+v", code, movies)
	}
}
