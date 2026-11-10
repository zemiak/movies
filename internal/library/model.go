// Package library builds an in-memory media library from Kodi-style folders: video files,
// tinyMediaManager .nfo files and local artwork. Nothing is scraped from the internet.
package library

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Item types, named as in the Jellyfin API.
const (
	TypeMovie   = "Movie"
	TypeSeries  = "Series"
	TypeSeason  = "Season"
	TypeEpisode = "Episode"
	TypeVideo   = "Video" // extras (featurettes, trailers, ...)
)

// Image types, named as in the Jellyfin API.
const (
	ImgPrimary  = "Primary"
	ImgBackdrop = "Backdrop"
	ImgLogo     = "Logo"
	ImgThumb    = "Thumb"
	ImgBanner   = "Banner"
	ImgArt      = "Art"
	ImgDisc     = "Disc"
)

// LibraryConfig is one configured library (a Jellyfin "view").
type LibraryConfig struct {
	Key            string // stable key, prefix of all item keys ("movies")
	Name           string // display name ("Movies")
	CollectionType string // "movies" or "tvshows"
	Path           string // root folder
}

// View is a scanned library.
type View struct {
	LibraryConfig
	ID       string
	Children []*Item // movies or series
	Created  time.Time
}

// Image is a local image file.
type Image struct {
	Path string
	Tag  string // changes when the file changes (cache buster)
}

// Person is an actor, director or writer from the NFO files.
type Person struct {
	ID    string
	Name  string
	Role  string // character, for actors
	Type  string // Actor, Director, Writer, GuestStar
	Thumb string // remote URL from the NFO (served as a redirect)
}

// MediaSource is one playable file (a movie can have several versions).
type MediaSource struct {
	ID        string
	Key       string
	Name      string
	Path      string
	Size      int64
	ModTime   time.Time
	Container string
	Probe     *Probe // nil until probed (and no NFO stream details)
	Subtitles []Subtitle
}

// Subtitle is an external subtitle file next to the video.
type Subtitle struct {
	Path     string
	Codec    string // srt, ass, vtt, ...
	Language string
	Forced   bool
}

// Item is a movie, series, season, episode or extra.
type Item struct {
	ID       string
	Key      string // stable key: "<library>/<relative path>"; user data is stored under it
	Type     string
	View     *View
	Parent   *Item // season for episodes, series for seasons, owner for extras; nil for top level
	Path     string
	Children []*Item // seasons of a series, episodes of a season
	Extras   []*Item
	Parts    []*Item // parts 2..n of a multi-part movie (Parent is the movie)

	Name, OriginalTitle, SortName string
	Overview, Tagline             string
	Year                          int
	Premiere                      time.Time
	EndDate                       time.Time
	DateCreated                   time.Time
	DateModified                  time.Time // newest file belonging to the item
	CommunityRating               float64
	OfficialRating                string
	Genres, Studios, Tags         []string
	Countries                     []string
	People                        []Person
	ProviderIDs                   map[string]string
	Index, ParentIndex            int // episode/season numbers; 0 = unknown for episodes
	HasIndex                      bool
	Status                        string // series: Continuing / Ended
	ExtraType                     string
	SetName                       string
	RuntimeTicks                  int64 // from NFO; MediaSource probe wins

	Images  map[string][]Image
	Sources []*MediaSource
}

// Series returns the series of an episode or season.
func (it *Item) Series() *Item {
	switch it.Type {
	case TypeSeries:
		return it
	case TypeSeason:
		return it.Parent
	case TypeEpisode:
		if it.Parent != nil {
			return it.Parent.Parent
		}
	}
	return nil
}

// IsFolder reports whether the item has children in the Jellyfin sense.
func (it *Item) IsFolder() bool { return it.Type == TypeSeries || it.Type == TypeSeason }

// IsPart reports whether the item is part 2..n of a multi-part movie.
func (it *Item) IsPart() bool { return it.Type == TypeMovie && it.Parent != nil }

// Playable reports whether the item has a media file.
func (it *Item) Playable() bool { return len(it.Sources) > 0 }

// Runtime is the item's duration in ticks (100ns), from the probe if possible.
func (it *Item) Runtime() int64 {
	if len(it.Sources) > 0 {
		if p := it.Sources[0].Probe; p != nil && p.DurationTicks() > 0 {
			return p.DurationTicks()
		}
	}
	return it.RuntimeTicks
}

// Episodes returns all episodes below a series or season, in order.
func (it *Item) Episodes() []*Item {
	switch it.Type {
	case TypeEpisode:
		return []*Item{it}
	case TypeSeason:
		return it.Children
	case TypeSeries:
		var out []*Item
		for _, s := range it.Children {
			out = append(out, s.Children...)
		}
		return out
	}
	return nil
}

// Library is an immutable snapshot of all views.
type Library struct {
	Views   []*View
	Items   map[string]*Item // every item, extras included
	People  map[string]*Person
	Genres  map[string]string // id -> name
	Studios map[string]string // id -> name
	Scanned time.Time
}

// Top returns the top-level items of all views (movies and series).
func (l *Library) Top() []*Item {
	var out []*Item
	for _, v := range l.Views {
		out = append(out, v.Children...)
	}
	return out
}

// All returns every non-extra item, in a stable order.
func (l *Library) All() []*Item {
	var out []*Item
	for _, v := range l.Views {
		for _, it := range v.Children {
			out = append(out, it)
			for _, s := range it.Children {
				out = append(out, s)
				out = append(out, s.Children...)
			}
		}
	}
	return out
}

// View returns the view with the given ID.
func (l *Library) View(id string) *View {
	for _, v := range l.Views {
		if v.ID == id {
			return v
		}
	}
	return nil
}

// ID derives a Jellyfin-style id (32 lowercase hex chars) from a key.
func ID(key string) string {
	sum := md5.Sum([]byte(key))
	return hex.EncodeToString(sum[:])
}

// NormalizeID turns a client-supplied id (maybe dashed or uppercase GUID) into our format.
func NormalizeID(id string) string {
	return strings.ToLower(strings.ReplaceAll(id, "-", ""))
}

// Fold lowercases and strips diacritics, for sorting and searching.
func Fold(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// SortName builds a sort name the way Jellyfin does: folded, leading articles removed.
func SortName(name string) string {
	s := strings.TrimSpace(Fold(name))
	for _, a := range []string{"the ", "a ", "an "} {
		if strings.HasPrefix(s, a) && len(s) > len(a) {
			s = s[len(a):]
			break
		}
	}
	return s
}

func sortItems(items []*Item) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].SortName != items[j].SortName {
			return items[i].SortName < items[j].SortName
		}
		return items[i].Key < items[j].Key
	})
}
