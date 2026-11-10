package library

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// nfo covers the <movie>, <tvshow> and <episodedetails> documents written by tinyMediaManager
// (and Kodi / Jellyfin, which use the same format).
type nfo struct {
	XMLName       xml.Name
	Title         string        `xml:"title"`
	OriginalTitle string        `xml:"originaltitle"`
	SortTitle     string        `xml:"sorttitle"`
	ShowTitle     string        `xml:"showtitle"`
	Year          string        `xml:"year"`
	Ratings       []nfoRating   `xml:"ratings>rating"`
	Rating        string        `xml:"rating"`
	Plot          string        `xml:"plot"`
	Outline       string        `xml:"outline"`
	Tagline       string        `xml:"tagline"`
	Runtime       string        `xml:"runtime"`
	MPAA          string        `xml:"mpaa"`
	Certification string        `xml:"certification"`
	ID            string        `xml:"id"`
	TmdbID        string        `xml:"tmdbid"`
	ImdbID        string        `xml:"imdbid"`
	TvdbID        string        `xml:"tvdbid"`
	UniqueIDs     []nfoUniqueID `xml:"uniqueid"`
	Countries     []string      `xml:"country"`
	Premiered     string        `xml:"premiered"`
	Aired         string        `xml:"aired"`
	DateAdded     string        `xml:"dateadded"`
	Genres        []string      `xml:"genre"`
	Studios       []string      `xml:"studio"`
	Tags          []string      `xml:"tag"`
	Credits       []string      `xml:"credits"`
	Directors     []string      `xml:"director"`
	Actors        []nfoActor    `xml:"actor"`
	Set           nfoSet        `xml:"set"`
	Status        string        `xml:"status"`
	Season        string        `xml:"season"`
	Episode       string        `xml:"episode"`
	NamedSeasons  []nfoNamed    `xml:"namedseason"`
	Streams       nfoStreams    `xml:"fileinfo>streamdetails"`
}

type nfoRating struct {
	Default string `xml:"default,attr"`
	Max     string `xml:"max,attr"`
	Name    string `xml:"name,attr"`
	Value   string `xml:"value"`
}

type nfoUniqueID struct {
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr"`
	Value   string `xml:",chardata"`
}

type nfoActor struct {
	Name  string `xml:"name"`
	Role  string `xml:"role"`
	Thumb string `xml:"thumb"`
	Type  string `xml:"type"`
}

type nfoSet struct {
	Name string `xml:"name"`
	Text string `xml:",chardata"`
}

type nfoNamed struct {
	Number string `xml:"number,attr"`
	Name   string `xml:",chardata"`
}

type nfoStreams struct {
	Video []struct {
		Codec    string `xml:"codec"`
		Aspect   string `xml:"aspect"`
		Width    int    `xml:"width"`
		Height   int    `xml:"height"`
		Duration string `xml:"durationinseconds"`
	} `xml:"video"`
	Audio []struct {
		Codec    string `xml:"codec"`
		Language string `xml:"language"`
		Channels int    `xml:"channels"`
	} `xml:"audio"`
	Subtitle []struct {
		Language string `xml:"language"`
	} `xml:"subtitle"`
}

// readNFO parses an NFO file; a missing or broken file gives nil.
func readNFO(path string) *nfo {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	// Some NFOs have a URL after the XML; Kodi allows that.
	if i := bytes.LastIndexByte(data, '>'); i >= 0 {
		data = data[:i+1]
	}
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity
	d.CharsetReader = func(_ string, r io.Reader) (io.Reader, error) { return r, nil }
	var n nfo
	if err := d.Decode(&n); err != nil {
		return nil
	}
	n.trim()
	return &n
}

func (n *nfo) trim() {
	for _, p := range []*string{&n.Title, &n.OriginalTitle, &n.SortTitle, &n.ShowTitle, &n.Year, &n.Rating,
		&n.Plot, &n.Outline, &n.Tagline, &n.Runtime, &n.MPAA, &n.Certification, &n.ID, &n.TmdbID,
		&n.ImdbID, &n.TvdbID, &n.Premiered, &n.Aired, &n.DateAdded, &n.Status, &n.Season, &n.Episode,
		&n.Set.Name, &n.Set.Text} {
		*p = strings.TrimSpace(*p)
	}
}

func (n *nfo) year() int {
	y, _ := strconv.Atoi(n.Year)
	return y
}

func (n *nfo) overview() string {
	if n.Plot != "" {
		return n.Plot
	}
	return n.Outline
}

func (n *nfo) communityRating() float64 {
	var best float64
	for _, r := range n.Ratings {
		v, err := strconv.ParseFloat(strings.TrimSpace(r.Value), 64)
		if err != nil || v <= 0 {
			continue
		}
		if m, err := strconv.ParseFloat(r.Max, 64); err == nil && m > 0 && m != 10 {
			v = v * 10 / m
		}
		if r.Default == "true" {
			return v
		}
		if best == 0 {
			best = v
		}
	}
	if best == 0 {
		best, _ = strconv.ParseFloat(n.Rating, 64)
	}
	return best
}

// officialRating turns "US:R / US:Rated R" into "R".
func (n *nfo) officialRating() string {
	s := n.MPAA
	if s == "" {
		s = n.Certification
	}
	if i := strings.Index(s, "/"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ":"); i >= 0 && i <= 3 {
		s = s[i+1:]
	}
	return strings.TrimPrefix(strings.TrimSpace(s), "Rated ")
}

func (n *nfo) runtimeTicks() int64 {
	for _, v := range n.Streams.Video {
		if s, err := strconv.ParseFloat(strings.TrimSpace(v.Duration), 64); err == nil && s > 0 {
			return int64(s * 1e7)
		}
	}
	if m, err := strconv.Atoi(n.Runtime); err == nil && m > 0 {
		return int64(m) * 60 * 1e7
	}
	return 0
}

func (n *nfo) providerIDs() map[string]string {
	ids := map[string]string{}
	set := func(k, v string) {
		if v = strings.TrimSpace(v); v != "" && v != "0" {
			ids[k] = v
		}
	}
	set("Tmdb", n.TmdbID)
	set("Imdb", n.ImdbID)
	set("Tvdb", n.TvdbID)
	if strings.HasPrefix(n.ID, "tt") {
		set("Imdb", n.ID)
	}
	for _, u := range n.UniqueIDs {
		switch strings.ToLower(u.Type) {
		case "tmdb":
			set("Tmdb", u.Value)
		case "imdb":
			set("Imdb", u.Value)
		case "tvdb":
			set("Tvdb", u.Value)
		case "tmdbset":
			set("TmdbCollection", u.Value)
		case "tvmaze":
			set("TvMaze", u.Value)
		}
	}
	return ids
}

func (n *nfo) setName() string {
	if n.Set.Name != "" {
		return n.Set.Name
	}
	return n.Set.Text
}

func (n *nfo) people() []Person {
	var out []Person
	seen := map[string]bool{}
	add := func(name, role, typ, thumb string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[typ+"|"+name] {
			return
		}
		seen[typ+"|"+name] = true
		out = append(out, Person{ID: ID("person:" + name), Name: name, Role: strings.TrimSpace(role), Type: typ, Thumb: strings.TrimSpace(thumb)})
	}
	for _, a := range n.Actors {
		t := "Actor"
		if strings.EqualFold(strings.TrimSpace(a.Type), "GuestStar") {
			t = "GuestStar"
		}
		add(a.Name, a.Role, t, a.Thumb)
	}
	for _, d := range n.Directors {
		add(d, "", "Director", "")
	}
	for _, c := range n.Credits {
		add(c, "", "Writer", "")
	}
	return out
}

// probe builds stream info from the NFO's <fileinfo>, used until ffprobe has run.
func (n *nfo) probe() *Probe {
	s := n.Streams
	if len(s.Video) == 0 && len(s.Audio) == 0 {
		return nil
	}
	p := &Probe{FromNFO: true}
	idx := 0
	for _, v := range s.Video {
		p.Streams = append(p.Streams, ProbeStream{Index: idx, CodecType: "video", CodecName: strings.ToLower(v.Codec),
			Width: v.Width, Height: v.Height, Disposition: map[string]int{"default": 1}})
		if p.Format.Duration == "" {
			p.Format.Duration = strings.TrimSpace(v.Duration)
		}
		idx++
	}
	for i, a := range s.Audio {
		d := 0
		if i == 0 {
			d = 1
		}
		p.Streams = append(p.Streams, ProbeStream{Index: idx, CodecType: "audio", CodecName: strings.ToLower(a.Codec),
			Channels: a.Channels, Tags: map[string]string{"language": strings.TrimSpace(a.Language)}, Disposition: map[string]int{"default": d}})
		idx++
	}
	for _, sub := range s.Subtitle {
		p.Streams = append(p.Streams, ProbeStream{Index: idx, CodecType: "subtitle", CodecName: "mov_text",
			Tags: map[string]string{"language": strings.TrimSpace(sub.Language)}})
		idx++
	}
	return p
}

// parseDate accepts the date formats found in NFO files.
func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02", time.RFC3339} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}
