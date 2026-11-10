package library

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	videoExts = map[string]bool{".m4v": true, ".mp4": true, ".mkv": true, ".avi": true, ".mov": true,
		".wmv": true, ".ts": true, ".m2ts": true, ".webm": true, ".mpg": true, ".mpeg": true, ".flv": true}
	imageExts    = []string{".jpg", ".jpeg", ".png", ".webp"}
	subtitleExts = map[string]string{".srt": "srt", ".ass": "ass", ".ssa": "ssa", ".vtt": "vtt", ".sub": "sub", ".smi": "smi"}
	extraDirs    = map[string]string{"extras": "Featurette", "featurettes": "Featurette", "behind the scenes": "BehindTheScenes",
		"deleted scenes": "DeletedScene", "interviews": "Interview", "scenes": "Scene", "shorts": "Short",
		"trailers": "Trailer", "other": "Clip", "samples": "", "sample": ""}

	reSeasonDir  = regexp.MustCompile(`(?i)^(?:season|series|staffel|serie|seria|sezona|saison|temporada)[ ._-]*(\d{1,4})(.*)$`)
	reSpecials   = regexp.MustCompile(`(?i)^(specials?|extras? season|season[ ._-]*specials?)$`)
	reSxxEyy     = regexp.MustCompile(`(?i)(?:^|[^a-z0-9])s(\d{1,4})[ ._-]*e(\d{1,4})(?:-?e(\d{1,4}))?`)
	reNxNN       = regexp.MustCompile(`(?:^|[^0-9])(\d{1,2})x(\d{1,3})(?:[^0-9]|$)`)
	reLeadingNum = regexp.MustCompile(`^(?:e|ep|episode)?[ ._-]*(\d{1,4})(?:[ ._-]|$)`)
	reYearSuffix = regexp.MustCompile(`^(.*?)[ ._]*[(\[](\d{4})[)\]]\s*$`)
)

// dirListing is one directory, read once.
type dirListing struct {
	path  string
	files map[string]os.DirEntry // lowercased name -> entry
	names []string               // file names, sorted
	dirs  []string               // subdirectory names, sorted
}

func listDir(path string) (*dirListing, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	l := &dirListing{path: path, files: map[string]os.DirEntry{}}
	for _, e := range entries {
		n := e.Name()
		if strings.HasPrefix(n, ".") || strings.HasPrefix(n, "@") || strings.HasPrefix(n, "#") {
			continue
		}
		if e.IsDir() {
			l.dirs = append(l.dirs, n)
			continue
		}
		if !e.Type().IsRegular() && e.Type()&os.ModeSymlink == 0 {
			continue
		}
		l.files[strings.ToLower(n)] = e
		l.names = append(l.names, n)
	}
	sort.Strings(l.names)
	sort.Strings(l.dirs)
	return l, nil
}

// videos returns the video files, without trailers and samples.
func (l *dirListing) videos() (videos, trailers []string) {
	for _, n := range l.names {
		ext := strings.ToLower(filepath.Ext(n))
		if !videoExts[ext] {
			continue
		}
		base := strings.ToLower(strings.TrimSuffix(n, filepath.Ext(n)))
		switch {
		case strings.HasSuffix(base, "-trailer") || strings.HasSuffix(base, ".trailer") || strings.HasSuffix(base, "_trailer"):
			trailers = append(trailers, n)
		case strings.HasSuffix(base, "-sample") || strings.HasSuffix(base, ".sample") || base == "sample":
		default:
			videos = append(videos, n)
		}
	}
	return
}

func (l *dirListing) has(name string) bool { _, ok := l.files[strings.ToLower(name)]; return ok }

func (l *dirListing) info(name string) (os.FileInfo, bool) {
	e, ok := l.files[strings.ToLower(name)]
	if !ok {
		return nil, false
	}
	fi, err := os.Stat(filepath.Join(l.path, e.Name()))
	if err != nil {
		return nil, false
	}
	return fi, true
}

// image finds "<stem>.<ext>" for the image extensions.
func (l *dirListing) image(stem string) (Image, time.Time, bool) {
	for _, ext := range imageExts {
		e, ok := l.files[strings.ToLower(stem+ext)]
		if !ok {
			continue
		}
		p := filepath.Join(l.path, e.Name())
		fi, err := os.Stat(p)
		if err != nil || fi.Size() == 0 {
			continue
		}
		return Image{Path: p, Tag: tag(p, fi)}, fi.ModTime(), true
	}
	return Image{}, time.Time{}, false
}

func tag(path string, fi os.FileInfo) string {
	sum := md5.Sum([]byte(fmt.Sprintf("%s|%d|%d", path, fi.Size(), fi.ModTime().UnixNano())))
	return hex.EncodeToString(sum[:8])
}

// imageSpec lists the file stems to try for each image type, in order of preference.
type imageSpec map[string][]string

func (l *dirListing) images(it *Item, spec imageSpec) {
	if it.Images == nil {
		it.Images = map[string][]Image{}
	}
	for typ, stems := range spec {
		if len(it.Images[typ]) > 0 && typ != ImgBackdrop {
			continue
		}
		seen := map[string]bool{}
		for _, im := range it.Images[typ] {
			seen[im.Path] = true
		}
		for _, stem := range stems {
			img, mod, ok := l.image(stem)
			if !ok || seen[img.Path] {
				continue
			}
			seen[img.Path] = true
			it.Images[typ] = append(it.Images[typ], img)
			if mod.After(it.DateModified) {
				it.DateModified = mod
			}
			if typ != ImgBackdrop {
				break
			}
		}
	}
}

func movieImageSpec(base string, generic bool) imageSpec {
	s := imageSpec{
		ImgPrimary:  {base + "-poster", base},
		ImgBackdrop: {base + "-fanart", base + "-backdrop"},
		ImgLogo:     {base + "-clearlogo", base + "-logo"},
		ImgThumb:    {base + "-landscape", base + "-thumb"},
		ImgBanner:   {base + "-banner"},
		ImgArt:      {base + "-clearart"},
		ImgDisc:     {base + "-discart", base + "-disc"},
	}
	for i := 1; i <= 9; i++ {
		s[ImgBackdrop] = append(s[ImgBackdrop], fmt.Sprintf("%s-fanart%d", base, i))
	}
	if generic {
		for k, v := range folderImageSpec() {
			s[k] = append(s[k], v...)
		}
	}
	return s
}

func folderImageSpec() imageSpec {
	s := imageSpec{
		ImgPrimary:  {"poster", "folder", "cover", "movie", "show"},
		ImgBackdrop: {"fanart", "backdrop", "background", "art"},
		ImgLogo:     {"clearlogo", "logo"},
		ImgThumb:    {"landscape", "thumb"},
		ImgBanner:   {"banner"},
		ImgArt:      {"clearart"},
		ImgDisc:     {"discart", "disc"},
	}
	for i := 1; i <= 9; i++ {
		s[ImgBackdrop] = append(s[ImgBackdrop], fmt.Sprintf("fanart%d", i), fmt.Sprintf("backdrop%d", i))
	}
	return s
}

func seasonImageSpec(n int) imageSpec {
	prefixes := []string{fmt.Sprintf("season%02d", n), fmt.Sprintf("season%d", n)}
	if n == 0 {
		prefixes = append([]string{"season-specials", "season-special"}, prefixes...)
	}
	s := imageSpec{}
	for _, p := range prefixes {
		s[ImgPrimary] = append(s[ImgPrimary], p+"-poster", p)
		s[ImgBackdrop] = append(s[ImgBackdrop], p+"-fanart")
		s[ImgBanner] = append(s[ImgBanner], p+"-banner")
		s[ImgThumb] = append(s[ImgThumb], p+"-landscape", p+"-thumb")
	}
	return s
}

// Scanner walks the library folders.
type Scanner struct {
	Libraries []LibraryConfig
	Probes    *ProbeCache
}

// PendingProbe is a media file that ffprobe hasn't seen (in this version) yet.
type PendingProbe struct {
	Path    string
	Size    int64
	ModTime time.Time
}

type scan struct {
	*Scanner
	lib     *Library
	pending []PendingProbe
	present map[string]bool
}

// Scan builds a fresh library snapshot.
func (s *Scanner) Scan() (*Library, []PendingProbe) {
	start := time.Now()
	sc := &scan{Scanner: s, present: map[string]bool{}, lib: &Library{
		Items: map[string]*Item{}, People: map[string]*Person{}, Genres: map[string]string{},
		Studios: map[string]string{}, Scanned: start,
	}}
	for _, cfg := range s.Libraries {
		v := &View{LibraryConfig: cfg, ID: ID("view:" + cfg.Key)}
		if fi, err := os.Stat(cfg.Path); err == nil {
			v.Created = fi.ModTime()
		} else {
			slog.Error("library folder missing", "library", cfg.Name, "path", cfg.Path, "err", err)
		}
		switch cfg.CollectionType {
		case "tvshows":
			sc.scanShows(v)
		default:
			sc.scanMovies(v, cfg.Path, 0)
		}
		sortItems(v.Children)
		sc.lib.Views = append(sc.lib.Views, v)
	}
	sc.index()
	if s.Probes != nil {
		s.Probes.Prune(sc.present)
	}
	movies, episodes := 0, 0
	for _, it := range sc.lib.Items {
		switch it.Type {
		case TypeMovie:
			if !it.IsPart() {
				movies++
			}
		case TypeEpisode:
			episodes++
		}
	}
	slog.Info("library scanned", "movies", movies, "episodes", episodes, "items", len(sc.lib.Items),
		"pendingProbes", len(sc.pending), "took", time.Since(start).Round(time.Millisecond))
	return sc.lib, sc.pending
}

func (sc *scan) key(v *View, path string) string {
	rel, err := filepath.Rel(v.Path, path)
	if err != nil {
		rel = path
	}
	return v.Key + "/" + filepath.ToSlash(rel)
}

// source builds a MediaSource for a video file in a listing.
func (sc *scan) source(v *View, l *dirListing, name string, n *nfo) *MediaSource {
	p := filepath.Join(l.path, name)
	fi, ok := l.info(name)
	if !ok {
		return nil
	}
	key := sc.key(v, p)
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(name), "."))
	container := ext
	switch ext {
	case "m4v", "mp4", "mov":
		container = "mp4"
	case "mkv":
		container = "mkv"
	}
	ms := &MediaSource{ID: ID(key), Key: key, Name: strings.TrimSuffix(name, filepath.Ext(name)), Path: p,
		Size: fi.Size(), ModTime: fi.ModTime(), Container: container}
	sc.present[p] = true
	if sc.Probes != nil {
		if pr, known := sc.Probes.Get(p, fi.Size(), fi.ModTime()); known {
			ms.Probe = pr
		} else {
			sc.pending = append(sc.pending, PendingProbe{Path: p, Size: fi.Size(), ModTime: fi.ModTime()})
		}
	}
	if ms.Probe == nil && n != nil {
		ms.Probe = n.probe()
	}
	ms.Subtitles = l.subtitles(ms.Name)
	return ms
}

func (l *dirListing) subtitles(base string) []Subtitle {
	var out []Subtitle
	lb := strings.ToLower(base)
	for _, n := range l.names {
		ext := strings.ToLower(filepath.Ext(n))
		codec, ok := subtitleExts[ext]
		if !ok {
			continue
		}
		stem := strings.ToLower(strings.TrimSuffix(n, filepath.Ext(n)))
		if stem != lb && !strings.HasPrefix(stem, lb+".") {
			continue
		}
		sub := Subtitle{Path: filepath.Join(l.path, n), Codec: codec}
		for _, tok := range strings.Split(strings.TrimPrefix(stem[len(lb):], "."), ".") {
			switch {
			case tok == "forced":
				sub.Forced = true
			case len(tok) == 2 || len(tok) == 3:
				sub.Language = tok
			}
		}
		out = append(out, sub)
	}
	return out
}

func (sc *scan) applyNFO(it *Item, n *nfo) {
	if n == nil {
		return
	}
	if n.Title != "" {
		it.Name = n.Title
	}
	it.OriginalTitle = n.OriginalTitle
	if n.SortTitle != "" {
		it.SortName = SortName(n.SortTitle)
	}
	it.Overview = n.overview()
	it.Tagline = n.Tagline
	if y := n.year(); y > 0 {
		it.Year = y
	}
	if t := parseDate(n.Premiered); !t.IsZero() {
		it.Premiere = t
	} else if t := parseDate(n.Aired); !t.IsZero() {
		it.Premiere = t
	}
	if it.Year == 0 && !it.Premiere.IsZero() {
		it.Year = it.Premiere.Year()
	}
	if t := parseDate(n.DateAdded); !t.IsZero() {
		it.DateCreated = t
	}
	it.CommunityRating = n.communityRating()
	it.OfficialRating = n.officialRating()
	it.Genres = nonEmpty(n.Genres)
	it.Studios = nonEmpty(n.Studios)
	it.Tags = nonEmpty(n.Tags)
	it.Countries = nonEmpty(n.Countries)
	it.People = n.people()
	it.ProviderIDs = n.providerIDs()
	it.Status = n.Status
	it.SetName = n.setName()
	it.RuntimeTicks = n.runtimeTicks()
}

func nonEmpty(in []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

// cleanTitle derives a title (and year) from a file or folder name.
func cleanTitle(name string) (string, int) {
	if m := reYearSuffix.FindStringSubmatch(name); m != nil {
		y, _ := strconv.Atoi(m[2])
		return strings.TrimSpace(m[1]), y
	}
	return strings.TrimSpace(strings.ReplaceAll(name, "_", " ")), 0
}

func modMax(t time.Time, ts ...time.Time) time.Time {
	for _, x := range ts {
		if x.After(t) {
			t = x
		}
	}
	return t
}

// ---------------------------------------------------------------- movies

func (sc *scan) scanMovies(v *View, dir string, depth int) {
	l, err := listDir(dir)
	if err != nil {
		slog.Warn("cannot read folder", "path", dir, "err", err)
		return
	}
	videos, trailers := l.videos()
	var movies []*Item
	if len(videos) > 0 {
		movies = sc.movieGroups(v, l, videos, depth > 0)
	}
	for _, d := range l.dirs {
		sub := filepath.Join(dir, d)
		if et, ok := extraDirs[strings.ToLower(d)]; ok {
			if len(movies) > 0 && et != "" {
				sc.addExtras(v, movies[0], sub, et)
			}
			continue
		}
		if depth < 4 {
			sc.scanMovies(v, sub, depth+1)
		}
	}
	if len(movies) > 0 {
		for _, t := range trailers {
			sc.addExtra(v, movies[0], l, t, "Trailer")
		}
	}
	v.Children = append(v.Children, movies...)
}

// video is one file, or a stack of parts ("Movie - Disc1.m4v", "Movie - Disc2.m4v").
type video struct {
	base  string   // file name without extension, or the stack's common name
	files []string // part 1 first
}

var reStack = regexp.MustCompile(`(?i)^(.*?)[ _.-]*(?:cd|dvd|p(?:ar)?t|dis[ck])[ _.-]*([0-9]{1,2})$`)

// stacks groups multi-part files the way Jellyfin does; a lone "Disc1" stays a plain file.
func stacks(files []string) []video {
	type stack struct {
		base  string
		parts map[int]string
	}
	byBase := map[string]*stack{}
	var order []string
	var out []video
	for _, f := range files {
		base := strings.TrimSuffix(f, filepath.Ext(f))
		if m := reStack.FindStringSubmatch(base); m != nil && strings.TrimSpace(m[1]) != "" {
			n, _ := strconv.Atoi(m[2])
			key := strings.ToLower(m[1]) + "|" + strings.ToLower(filepath.Ext(f))
			st := byBase[key]
			if st == nil {
				st = &stack{base: strings.TrimSpace(m[1]), parts: map[int]string{}}
				byBase[key] = st
				order = append(order, key)
			}
			if _, dup := st.parts[n]; !dup {
				st.parts[n] = f
				continue
			}
		}
		out = append(out, video{base: base, files: []string{f}})
	}
	for _, key := range order {
		st := byBase[key]
		nums := make([]int, 0, len(st.parts))
		for n := range st.parts {
			nums = append(nums, n)
		}
		sort.Ints(nums)
		if len(nums) == 1 {
			f := st.parts[nums[0]]
			out = append(out, video{base: strings.TrimSuffix(f, filepath.Ext(f)), files: []string{f}})
			continue
		}
		v := video{base: st.base}
		for _, n := range nums {
			v.files = append(v.files, st.parts[n])
		}
		out = append(out, v)
	}
	return out
}

// movieGroups turns the videos of a folder into movies. "Name - Label.ext" is a version of
// "Name.ext" (Jellyfin's naming for multiple versions), anything else is a separate movie.
func (sc *scan) movieGroups(v *View, l *dirListing, videos []string, inOwnFolder bool) []*Item {
	type group struct {
		base     string
		versions []video // primary first
	}
	byLen := stacks(videos)
	sort.SliceStable(byLen, func(i, j int) bool {
		if len(byLen[i].base) != len(byLen[j].base) {
			return len(byLen[i].base) < len(byLen[j].base)
		}
		return byLen[i].base < byLen[j].base
	})
	var groups []*group
	for _, vid := range byLen {
		var g *group
		lb := strings.ToLower(vid.base)
		for _, cand := range groups {
			lc := strings.ToLower(cand.base)
			if strings.HasPrefix(lb, lc+" - ") || strings.HasPrefix(lb, lc+" [") || strings.HasPrefix(lb, lc+" (") {
				g = cand
				break
			}
		}
		if g == nil {
			groups = append(groups, &group{base: vid.base, versions: []video{vid}})
		} else {
			g.versions = append(g.versions, vid)
		}
	}
	folder := filepath.Base(l.path)
	var out []*Item
	for _, g := range groups {
		generic := inOwnFolder && (len(groups) == 1 || strings.EqualFold(g.base, folder))
		n := readNFO(filepath.Join(l.path, g.base+".nfo"))
		if n == nil && generic && l.has("movie.nfo") {
			n = readNFO(filepath.Join(l.path, "movie.nfo"))
		}
		if n == nil {
			for _, ver := range g.versions[1:] {
				if n = readNFO(filepath.Join(l.path, ver.base+".nfo")); n != nil {
					break
				}
			}
		}
		primary := g.versions[0]
		it := &Item{Type: TypeMovie, View: v, Path: filepath.Join(l.path, primary.files[0])}
		it.Key = sc.key(v, it.Path)
		it.ID = ID(it.Key)
		it.Name, it.Year = cleanTitle(g.base)
		sc.applyNFO(it, n)
		for i, ver := range g.versions {
			ms := sc.source(v, l, ver.files[0], n)
			if ms == nil {
				continue
			}
			if i == 0 {
				ms.ID = it.ID // Jellyfin: the primary source has the item's id
			} else {
				// "300 - Cesky" -> "Cesky", "Movie [Director's Cut]" -> "Director's Cut"
				label := strings.TrimPrefix(strings.TrimSpace(ver.base[len(g.base):]), "-")
				ms.Name = strings.TrimSpace(strings.Trim(strings.TrimSpace(label), "[]()"))
			}
			it.Sources = append(it.Sources, ms)
			it.DateModified = modMax(it.DateModified, ms.ModTime)
		}
		if len(it.Sources) == 0 || it.Sources[0].ID != it.ID {
			continue
		}
		if it.DateCreated.IsZero() {
			it.DateCreated = it.Sources[0].ModTime
		}
		if fi, ok := l.info(g.base + ".nfo"); ok {
			it.DateModified = modMax(it.DateModified, fi.ModTime())
		}
		l.images(it, movieImageSpec(g.base, generic))
		for _, f := range primary.files {
			l.images(it, movieImageSpec(strings.TrimSuffix(f, filepath.Ext(f)), false))
		}
		for _, ver := range g.versions[1:] {
			l.images(it, movieImageSpec(ver.base, false))
		}
		if it.SortName == "" {
			it.SortName = SortName(it.Name)
		}
		// Parts 2..n of a stack: separate playable items (Jellyfin's "additional parts").
		for i, f := range primary.files[1:] {
			ms := sc.source(v, l, f, nil)
			if ms == nil {
				continue
			}
			name := fmt.Sprintf("%s (Part %d)", it.Name, i+2)
			it.Parts = append(it.Parts, &Item{Type: TypeMovie, View: v, Parent: it, Path: ms.Path, Key: ms.Key,
				ID: ms.ID, Name: name, SortName: SortName(name), Year: it.Year, Overview: it.Overview,
				DateCreated: it.DateCreated, DateModified: ms.ModTime, Images: it.Images,
				ProviderIDs: it.ProviderIDs, Index: i + 2, Sources: []*MediaSource{ms}})
		}
		out = append(out, it)
	}
	return out
}

// ---------------------------------------------------------------- extras

func (sc *scan) addExtras(v *View, owner *Item, dir, extraType string) {
	l, err := listDir(dir)
	if err != nil {
		return
	}
	videos, trailers := l.videos()
	for _, f := range append(videos, trailers...) {
		sc.addExtra(v, owner, l, f, extraType)
	}
}

func (sc *scan) addExtra(v *View, owner *Item, l *dirListing, file, extraType string) {
	ms := sc.source(v, l, file, nil)
	if ms == nil {
		return
	}
	base := strings.TrimSuffix(file, filepath.Ext(file))
	name, _ := cleanTitle(base)
	if extraType == "Trailer" && strings.HasPrefix(strings.ToLower(name), strings.ToLower(owner.Name)) {
		name = "Trailer"
	}
	it := &Item{Type: TypeVideo, View: v, Parent: owner, Path: ms.Path, Key: ms.Key, ID: ms.ID, Name: name,
		SortName: SortName(name), ExtraType: extraType, DateCreated: ms.ModTime, DateModified: ms.ModTime,
		Sources: []*MediaSource{ms}}
	l.images(it, imageSpec{ImgPrimary: {base + "-thumb", base}})
	owner.Extras = append(owner.Extras, it)
}

// ---------------------------------------------------------------- TV shows

func (sc *scan) scanShows(v *View) {
	l, err := listDir(v.Path)
	if err != nil {
		slog.Warn("cannot read folder", "path", v.Path, "err", err)
		return
	}
	for _, d := range l.dirs {
		if s := sc.scanSeries(v, filepath.Join(v.Path, d)); s != nil {
			v.Children = append(v.Children, s)
		}
	}
}

func (sc *scan) scanSeries(v *View, dir string) *Item {
	l, err := listDir(dir)
	if err != nil {
		slog.Warn("cannot read folder", "path", dir, "err", err)
		return nil
	}
	s := &Item{Type: TypeSeries, View: v, Path: dir}
	s.Key = sc.key(v, dir)
	s.ID = ID(s.Key)
	s.Name, s.Year = cleanTitle(filepath.Base(dir))
	n := readNFO(filepath.Join(dir, "tvshow.nfo"))
	sc.applyNFO(s, n)
	s.RuntimeTicks = 0
	if fi, err := os.Stat(dir); err == nil && s.DateCreated.IsZero() {
		s.DateCreated = fi.ModTime()
	}
	l.images(s, folderImageSpec())
	named := map[int]string{}
	if n != nil {
		for _, ns := range n.NamedSeasons {
			if num, err := strconv.Atoi(strings.TrimSpace(ns.Number)); err == nil && strings.TrimSpace(ns.Name) != "" {
				named[num] = strings.TrimSpace(ns.Name)
			}
		}
	}

	seasons := map[int]*Item{}
	season := func(num int, folder string) *Item {
		if se, ok := seasons[num]; ok {
			if se.Path == dir && folder != "" {
				se.Path = folder
			}
			return se
		}
		se := &Item{Type: TypeSeason, View: v, Parent: s, Path: dir, Index: num, HasIndex: true}
		if folder != "" {
			se.Path = folder
		}
		se.Key = fmt.Sprintf("%s/#season-%d", s.Key, num)
		se.ID = ID(se.Key)
		switch {
		case named[num] != "":
			se.Name = named[num]
		case num == 0:
			se.Name = "Specials"
		default:
			se.Name = fmt.Sprintf("Season %d", num)
		}
		se.SortName = fmt.Sprintf("%04d", num)
		l.images(se, seasonImageSpec(num))
		seasons[num] = se
		return se
	}

	// Episodes directly in the show folder.
	sc.addEpisodes(v, s, l, -1, season)
	for _, d := range l.dirs {
		sub := filepath.Join(dir, d)
		if et, ok := extraDirs[strings.ToLower(d)]; ok {
			if et != "" {
				sc.addExtras(v, s, sub, et)
			}
			continue
		}
		num := -1
		if m := reSeasonDir.FindStringSubmatch(d); m != nil {
			num, _ = strconv.Atoi(m[1])
			if extra := strings.TrimSpace(strings.Trim(m[2], " ._-")); extra != "" && named[num] == "" {
				named[num] = d
			}
		} else if reSpecials.MatchString(d) {
			num = 0
		}
		sl, err := listDir(sub)
		if err != nil {
			continue
		}
		if num >= 0 {
			se := season(num, sub)
			sl.images(se, folderImageSpec())
		}
		sc.addEpisodes(v, s, sl, num, season)
		for _, sd := range sl.dirs {
			if et, ok := extraDirs[strings.ToLower(sd)]; ok && et != "" {
				owner := s
				if num >= 0 {
					owner = seasons[num]
				}
				sc.addExtras(v, owner, filepath.Join(sub, sd), et)
			}
		}
	}

	nums := make([]int, 0, len(seasons))
	for num := range seasons {
		nums = append(nums, num)
	}
	sort.Ints(nums)
	for _, num := range nums {
		se := seasons[num]
		if len(se.Children) == 0 {
			continue
		}
		sort.SliceStable(se.Children, func(i, j int) bool {
			a, b := se.Children[i], se.Children[j]
			if a.Index != b.Index {
				return a.Index < b.Index
			}
			return a.Key < b.Key
		})
		for _, ep := range se.Children {
			if se.DateCreated.IsZero() || ep.DateCreated.Before(se.DateCreated) {
				se.DateCreated = ep.DateCreated
			}
			se.DateModified = modMax(se.DateModified, ep.DateModified)
			if se.Premiere.IsZero() && !ep.Premiere.IsZero() {
				se.Premiere = ep.Premiere
				se.Year = ep.Premiere.Year()
			}
		}
		s.Children = append(s.Children, se)
		s.DateModified = modMax(s.DateModified, se.DateModified)
	}
	if len(s.Children) == 0 {
		return nil
	}
	if s.SortName == "" {
		s.SortName = SortName(s.Name)
	}
	return s
}

// addEpisodes adds the videos of a folder as episodes. seasonNum < 0: take it from the file name.
func (sc *scan) addEpisodes(v *View, s *Item, l *dirListing, seasonNum int, season func(int, string) *Item) {
	videos, _ := l.videos()
	next := map[int]int{}
	for _, f := range videos {
		base := strings.TrimSuffix(f, filepath.Ext(f))
		n := readNFO(filepath.Join(l.path, base+".nfo"))
		sNum, eNum, title := parseEpisodeName(base)
		if n != nil {
			if x, err := strconv.Atoi(n.Season); err == nil && x >= 0 && seasonNum < 0 {
				sNum = x
			}
			if x, err := strconv.Atoi(n.Episode); err == nil && x > 0 {
				eNum = x
			}
		}
		if seasonNum >= 0 {
			sNum = seasonNum
		}
		if sNum < 0 {
			sNum = 1
		}
		se := season(sNum, "")
		if eNum <= 0 {
			next[sNum]++
			eNum = len(se.Children) + next[sNum]
		}
		ep := &Item{Type: TypeEpisode, View: v, Parent: se, Path: filepath.Join(l.path, f), Index: eNum,
			ParentIndex: sNum, HasIndex: true}
		ep.Key = sc.key(v, ep.Path)
		ep.ID = ID(ep.Key)
		if title == "" {
			title = fmt.Sprintf("Episode %d", eNum)
		}
		ep.Name = title
		sc.applyNFO(ep, n)
		ep.Status, ep.SetName = "", ""
		ms := sc.source(v, l, f, n)
		if ms == nil {
			continue
		}
		ms.ID = ep.ID
		ep.Sources = []*MediaSource{ms}
		ep.DateModified = modMax(ep.DateModified, ms.ModTime)
		if ep.DateCreated.IsZero() {
			ep.DateCreated = ms.ModTime
		}
		if fi, ok := l.info(base + ".nfo"); ok {
			ep.DateModified = modMax(ep.DateModified, fi.ModTime())
		}
		l.images(ep, imageSpec{ImgPrimary: {base + "-thumb", base, base + "-landscape"}})
		ep.SortName = fmt.Sprintf("%04d-%04d", sNum, eNum)
		se.Children = append(se.Children, ep)
	}
}

// parseEpisodeName gets season, episode and title from names like "Episode S01E03 Ireland",
// "Firefly S01E01_Serenity" or "1x03 - Name". Unknown numbers are -1.
func parseEpisodeName(base string) (season, episode int, title string) {
	season, episode = -1, -1
	rest := ""
	if m := reSxxEyy.FindStringSubmatchIndex(base); m != nil {
		season, _ = strconv.Atoi(base[m[2]:m[3]])
		episode, _ = strconv.Atoi(base[m[4]:m[5]])
		rest = base[m[1]:]
	} else if m := reNxNN.FindStringSubmatchIndex(base); m != nil {
		season, _ = strconv.Atoi(base[m[2]:m[3]])
		episode, _ = strconv.Atoi(base[m[4]:m[5]])
		rest = base[m[5]:]
	} else if m := reLeadingNum.FindStringSubmatchIndex(strings.ToLower(base)); m != nil {
		episode, _ = strconv.Atoi(base[m[2]:m[3]])
		rest = base[m[3]:]
	} else {
		return season, episode, strings.TrimSpace(strings.ReplaceAll(base, "_", " "))
	}
	title = strings.ReplaceAll(rest, "_", " ")
	title = strings.TrimSpace(strings.Trim(title, " .-_"))
	return
}

// ---------------------------------------------------------------- index

func (sc *scan) index() {
	lib := sc.lib
	var add func(it *Item)
	add = func(it *Item) {
		lib.Items[it.ID] = it
		for _, p := range it.People {
			if ex, ok := lib.People[p.ID]; ok {
				if ex.Thumb == "" {
					ex.Thumb = p.Thumb
				}
				continue
			}
			pp := p
			pp.Role, pp.Type = "", ""
			lib.People[p.ID] = &pp
		}
		for _, g := range it.Genres {
			lib.Genres[ID("genre:"+g)] = g
		}
		for _, s := range it.Studios {
			lib.Studios[ID("studio:"+s)] = s
		}
		for _, c := range it.Children {
			add(c)
		}
		for _, e := range it.Extras {
			add(e)
		}
		for _, p := range it.Parts {
			add(p)
		}
	}
	for _, v := range lib.Views {
		for _, it := range v.Children {
			add(it)
		}
	}
}
