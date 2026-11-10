package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zemiak/movies/internal/library"
	"github.com/zemiak/movies/internal/store"
)

func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// playbackInfo handles GET/POST /Items/{id}/PlaybackInfo. We only offer direct play.
func (s *Server) playbackInfo(c *ctx) {
	it := c.lookup(c.param("id"))
	if it == nil || !it.Playable() {
		c.notFound()
		return
	}
	var body struct {
		MediaSourceId string
	}
	if c.r.Method == http.MethodPost && !c.body(&body) {
		return
	}
	want := library.NormalizeID(c.q["mediasourceid"])
	if want == "" {
		want = library.NormalizeID(body.MediaSourceId)
	}
	srcs := []any{}
	for _, ms := range it.Sources {
		if want == "" || ms.ID == want {
			srcs = append(srcs, s.mediaSource(c, it, ms))
		}
	}
	if len(srcs) == 0 {
		for _, ms := range it.Sources {
			srcs = append(srcs, s.mediaSource(c, it, ms))
		}
	}
	c.json(map[string]any{"MediaSources": srcs, "PlaySessionId": newSessionID()})
}

func (c *ctx) source(it *library.Item) *library.MediaSource {
	want := library.NormalizeID(c.q["mediasourceid"])
	for _, ms := range it.Sources {
		if ms.ID == want {
			return ms
		}
	}
	return it.Sources[0]
}

var contentTypes = map[string]string{".m4v": "video/x-m4v", ".mp4": "video/mp4", ".mkv": "video/x-matroska",
	".avi": "video/x-msvideo", ".mov": "video/quicktime", ".ts": "video/mp2t", ".m2ts": "video/mp2t",
	".webm": "video/webm", ".wmv": "video/x-ms-wmv", ".mpg": "video/mpeg", ".mpeg": "video/mpeg",
	".srt": "application/x-subrip", ".vtt": "text/vtt", ".ass": "text/x-ssa", ".ssa": "text/x-ssa"}

func serveFile(c *ctx, path string, download bool) {
	f, err := os.Open(path)
	if err != nil {
		c.notFound()
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		c.notFound()
		return
	}
	ext := strings.ToLower(filepath.Ext(path))
	ct := contentTypes[ext]
	if ct == "" {
		ct = mime.TypeByExtension(ext)
	}
	if ct != "" {
		c.w.Header().Set("Content-Type", ct)
	}
	if download {
		c.w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filepath.Base(path)}))
	}
	http.ServeContent(c.w, c.r, "", fi.ModTime(), f)
}

// stream handles /Videos/{id}/stream[.ext]: the file as is, with Range support.
func (s *Server) stream(c *ctx) {
	file := strings.ToLower(c.param("file"))
	if file != "stream" && !strings.HasPrefix(file, "stream.") {
		c.notFound() // HLS (master.m3u8, ...) would need transcoding
		return
	}
	it := c.lookup(c.param("id"))
	if it == nil || !it.Playable() {
		c.notFound()
		return
	}
	serveFile(c, c.source(it).Path, false)
}

func (s *Server) download(c *ctx) {
	it := c.lookup(c.param("id"))
	if it == nil || !it.Playable() {
		c.notFound()
		return
	}
	serveFile(c, c.source(it).Path, true)
}

// subtitle serves an external subtitle file.
func (s *Server) subtitle(c *ctx) {
	it := c.lookup(c.param("id"))
	if it == nil || !it.Playable() {
		c.notFound()
		return
	}
	idx, err := strconv.Atoi(c.param("index"))
	if err != nil {
		c.notFound()
		return
	}
	msID := library.NormalizeID(c.param("msid"))
	for _, ms := range it.Sources {
		if ms.ID != msID {
			continue
		}
		maxIndex := -1
		if ms.Probe != nil {
			for _, st := range ms.Probe.Streams {
				maxIndex = max(maxIndex, st.Index)
			}
		}
		if i := idx - maxIndex - 1; i >= 0 && i < len(ms.Subtitles) {
			serveFile(c, ms.Subtitles[i].Path, false)
			return
		}
	}
	c.notFound()
}

// ---------------------------------------------------------------- watch state

// playState is the body of /Sessions/Playing{,/Progress,/Stopped}.
type playState struct {
	ItemId        string
	MediaSourceId string
	PositionTicks *int64
	IsPaused      bool
	PlaySessionId string
	Failed        bool
}

const (
	minResumePct      = 5
	maxResumePct      = 90
	minResumeDuration = 300 * 1e7 // shorter videos are "played" as soon as they are stopped
)

func (s *Server) playing(kind string) handlerFunc {
	return func(c *ctx) {
		var ps playState
		if !c.body(&ps) {
			return
		}
		if ps.ItemId == "" {
			ps.ItemId = c.param("id")
		}
		if ps.PositionTicks == nil {
			if v, err := strconv.ParseInt(c.q["positionticks"], 10, 64); err == nil {
				ps.PositionTicks = &v
			}
		}
		it := c.lookup(ps.ItemId)
		if it == nil || !it.Playable() {
			c.noContent()
			return
		}
		switch kind {
		case "start", "progress":
			s.store.Update(c.user, it.Key, func(d *store.UserData) {
				now := time.Now()
				d.LastPlayed = &now
				if ps.PositionTicks != nil && *ps.PositionTicks > 0 {
					d.PositionTicks = *ps.PositionTicks
				}
			})
		case "stopped":
			if ps.Failed || ps.PositionTicks == nil {
				break
			}
			s.stopped(c.user, it, *ps.PositionTicks)
		}
		c.noContent()
	}
}

// stopped applies Jellyfin's resume rules when playback stops.
func (s *Server) stopped(user string, it *library.Item, pos int64) {
	rt := it.Runtime()
	// Finishing the last part of a multi-part movie finishes the movie.
	if it.IsPart() && rt > 0 && float64(pos)*100/float64(rt) >= maxResumePct {
		if owner := it.Parent; owner.Parts[len(owner.Parts)-1] == it {
			s.store.Update(user, owner.Key, func(d *store.UserData) {
				now := time.Now()
				d.LastPlayed = &now
				d.Played = true
				d.PlayCount++
				d.PositionTicks = 0
			})
		}
	}
	s.store.Update(user, it.Key, func(d *store.UserData) {
		now := time.Now()
		d.LastPlayed = &now
		if rt <= 0 {
			if pos > 0 {
				d.PositionTicks = pos
			}
			return
		}
		pct := float64(pos) * 100 / float64(rt)
		switch {
		case pct >= maxResumePct || (rt < minResumeDuration && pos > 0):
			d.Played = true
			d.PlayCount++
			d.PositionTicks = 0
		case pct < minResumePct:
			d.PositionTicks = 0
		default:
			d.PositionTicks = pos
		}
	})
}

// targets is the item itself, or all episodes for a series/season.
func targets(it *library.Item) []*library.Item {
	if it.IsFolder() {
		return it.Episodes()
	}
	return []*library.Item{it}
}

func (s *Server) markPlayed(played bool) handlerFunc {
	return func(c *ctx) {
		it := c.lookup(c.param("id"))
		if it == nil {
			c.notFound()
			return
		}
		when := time.Now()
		if t, ok := parseTime(c.q["dateplayed"]); ok {
			when = t
		}
		for _, t := range targets(it) {
			s.store.Update(c.user, t.Key, func(d *store.UserData) {
				d.PositionTicks = 0
				if played {
					if !d.Played {
						d.PlayCount++
					}
					d.Played = true
					w := when
					d.LastPlayed = &w
				} else {
					d.Played = false
				}
			})
		}
		c.json(s.userData(c.user, it))
	}
}

func (s *Server) markFavorite(fav bool) handlerFunc {
	return func(c *ctx) {
		it := c.lookup(c.param("id"))
		if it == nil {
			c.notFound()
			return
		}
		s.store.Update(c.user, it.Key, func(d *store.UserData) { d.IsFavorite = fav })
		c.json(s.userData(c.user, it))
	}
}

func (s *Server) getUserData(c *ctx) {
	it := c.lookup(c.param("id"))
	if it == nil {
		c.notFound()
		return
	}
	c.json(s.userData(c.user, it))
}

// updateUserData handles POST /UserItems/{id}/UserData (partial updates).
func (s *Server) updateUserData(c *ctx) {
	it := c.lookup(c.param("id"))
	if it == nil {
		c.notFound()
		return
	}
	var body map[string]json.RawMessage
	if !c.body(&body) {
		return
	}
	get := func(name string, v any) bool {
		for k, raw := range body {
			if strings.EqualFold(k, name) && string(raw) != "null" {
				return json.Unmarshal(raw, v) == nil
			}
		}
		return false
	}
	var played, fav bool
	var pos int64
	var count int
	var last string
	hasPlayed, hasFav, hasPos, hasCount, hasLast := get("Played", &played), get("IsFavorite", &fav),
		get("PlaybackPositionTicks", &pos), get("PlayCount", &count), get("LastPlayedDate", &last)
	for _, t := range targets(it) {
		if t != it && !hasPlayed {
			continue
		}
		s.store.Update(c.user, t.Key, func(d *store.UserData) {
			if hasPlayed {
				d.Played = played
				if played {
					d.PositionTicks = 0
				}
			}
			if hasPos && t == it {
				d.PositionTicks = pos
			}
			if hasCount {
				d.PlayCount = count
			}
			if hasLast {
				if lt, ok := parseTime(last); ok {
					d.LastPlayed = &lt
				}
			}
		})
	}
	if hasFav {
		s.store.Update(c.user, it.Key, func(d *store.UserData) { d.IsFavorite = fav })
	}
	c.json(s.userData(c.user, it))
}
