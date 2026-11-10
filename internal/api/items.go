package api

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zemiak/movies/internal/library"
)

// lookup finds an item by (client-formatted) id.
func (c *ctx) lookup(id string) *library.Item {
	return c.lib.Items[library.NormalizeID(id)]
}

// itemByID handles GET /Items/{id}: items, views, the root folder, people and genres.
func (s *Server) itemByID(c *ctx) {
	id := library.NormalizeID(c.param("id"))
	if it := c.lib.Items[id]; it != nil {
		c.json(s.item(c, it, nil))
		return
	}
	if v := c.lib.View(id); v != nil {
		c.json(s.viewDTO(c, v))
		return
	}
	if id == rootID {
		c.json(s.rootDTO(c))
		return
	}
	if p := c.lib.People[id]; p != nil {
		c.json(s.personDTO(c, p))
		return
	}
	if g, ok := c.lib.Genres[id]; ok {
		c.json(s.namedDTO(c, "Genre", "genre:", g))
		return
	}
	if st, ok := c.lib.Studios[id]; ok {
		c.json(s.namedDTO(c, "Studio", "studio:", st))
		return
	}
	c.notFound()
}

func (s *Server) personDTO(c *ctx, p *library.Person) map[string]any {
	d := map[string]any{
		"Name": p.Name, "ServerId": s.cfg.ServerID, "Id": p.ID, "Type": "Person", "IsFolder": false,
		"ImageTags": map[string]any{}, "BackdropImageTags": []any{}, "LocationType": "FileSystem",
		"MediaType": "Unknown", "ProviderIds": map[string]any{}, "Overview": "",
		"UserData": map[string]any{"PlaybackPositionTicks": 0, "PlayCount": 0, "IsFavorite": false, "Played": false, "Key": p.ID, "ItemId": p.ID},
	}
	if p.Thumb != "" {
		d["ImageTags"] = map[string]any{"Primary": library.ID(p.Thumb)[:16]}
		d["PrimaryImageTag"] = library.ID(p.Thumb)[:16]
	}
	return d
}

func (s *Server) namedDTO(c *ctx, typ, prefix, name string) map[string]any {
	id := library.ID(prefix + name)
	return map[string]any{
		"Name": name, "ServerId": s.cfg.ServerID, "Id": id, "Type": typ, "IsFolder": true,
		"SortName": library.SortName(name), "ImageTags": map[string]any{}, "BackdropImageTags": []any{},
		"LocationType": "FileSystem", "MediaType": "Unknown", "ProviderIds": map[string]any{},
		"UserData": map[string]any{"PlaybackPositionTicks": 0, "PlayCount": 0, "IsFavorite": false, "Played": false, "Key": id, "ItemId": id},
	}
}

// scope returns the items below ParentId (or everything), honoring Recursive.
func (s *Server) scope(c *ctx, parent string, recursive bool) []*library.Item {
	lib := c.lib
	if parent == "" || parent == rootID {
		if recursive {
			return lib.All()
		}
		return nil // the views; handled by the caller
	}
	if v := lib.View(parent); v != nil {
		if !recursive {
			return v.Children
		}
		var out []*library.Item
		for _, it := range v.Children {
			out = append(out, it)
			for _, se := range it.Children {
				out = append(out, se)
				out = append(out, se.Children...)
			}
		}
		return out
	}
	it := lib.Items[parent]
	if it == nil {
		return nil
	}
	switch it.Type {
	case library.TypeSeries:
		if !recursive {
			return it.Children
		}
		var out []*library.Item
		for _, se := range it.Children {
			out = append(out, se)
			out = append(out, se.Children...)
		}
		return out
	case library.TypeSeason:
		return it.Children
	}
	return nil
}

type itemFilter func(*library.Item) bool

// filters builds the item filters from the query parameters shared by the item endpoints.
func (s *Server) filters(c *ctx) []itemFilter {
	var fs []itemFilter
	if types := c.qList("includeitemtypes"); len(types) > 0 {
		set := map[string]bool{}
		for _, t := range types {
			set[strings.ToLower(t)] = true
		}
		fs = append(fs, func(it *library.Item) bool { return set[strings.ToLower(it.Type)] })
	}
	if types := c.qList("excludeitemtypes"); len(types) > 0 {
		set := map[string]bool{}
		for _, t := range types {
			set[strings.ToLower(t)] = true
		}
		fs = append(fs, func(it *library.Item) bool { return !set[strings.ToLower(it.Type)] })
	}
	if mt := c.qList("mediatypes"); len(mt) > 0 {
		video := false
		for _, m := range mt {
			video = video || strings.EqualFold(m, "Video")
		}
		fs = append(fs, func(it *library.Item) bool { return video && it.Playable() })
	}
	if ex := c.qList("excludeitemids"); len(ex) > 0 {
		set := map[string]bool{}
		for _, id := range ex {
			set[library.NormalizeID(id)] = true
		}
		fs = append(fs, func(it *library.Item) bool { return !set[it.ID] })
	}
	if v, ok := c.qBool("isfolder"); ok {
		fs = append(fs, func(it *library.Item) bool { return it.IsFolder() == v })
	}
	if v, ok := c.qBool("isplayed"); ok {
		fs = append(fs, func(it *library.Item) bool { return s.played(c.user, it) == v })
	}
	if v, ok := c.qBool("isfavorite"); ok {
		fs = append(fs, func(it *library.Item) bool { return s.store.Get(c.user, it.Key).IsFavorite == v })
	}
	if v, ok := c.qBool("isresumable"); ok {
		fs = append(fs, func(it *library.Item) bool { return s.resumable(c.user, it) == v })
	}
	for _, f := range c.qList("filters") {
		switch strings.ToLower(f) {
		case "isfolder":
			fs = append(fs, func(it *library.Item) bool { return it.IsFolder() })
		case "isnotfolder":
			fs = append(fs, func(it *library.Item) bool { return !it.IsFolder() })
		case "isplayed":
			fs = append(fs, func(it *library.Item) bool { return s.played(c.user, it) })
		case "isunplayed":
			fs = append(fs, func(it *library.Item) bool { return !s.played(c.user, it) })
		case "isfavorite", "isfavoriteorlikes", "likes":
			fs = append(fs, func(it *library.Item) bool { return s.store.Get(c.user, it.Key).IsFavorite })
		case "isresumable":
			fs = append(fs, func(it *library.Item) bool { return s.resumable(c.user, it) })
		case "dislikes":
			fs = append(fs, func(*library.Item) bool { return false })
		}
	}
	if term := c.q["searchterm"]; term != "" {
		t := library.Fold(strings.TrimSpace(term))
		fs = append(fs, func(it *library.Item) bool {
			return strings.Contains(library.Fold(it.Name), t) || strings.Contains(library.Fold(it.OriginalTitle), t)
		})
	}
	if p := c.q["namestartswith"]; p != "" {
		pf := library.Fold(p)
		fs = append(fs, func(it *library.Item) bool { return strings.HasPrefix(it.SortName, pf) })
	}
	if p := c.q["namestartswithorgreater"]; p != "" {
		pf := library.Fold(p)
		fs = append(fs, func(it *library.Item) bool { return it.SortName >= pf })
	}
	if p := c.q["namelessthan"]; p != "" {
		pf := library.Fold(p)
		fs = append(fs, func(it *library.Item) bool { return it.SortName < pf })
	}
	if gs := c.qList("genres"); len(gs) > 0 {
		fs = append(fs, anyOf(gs, func(it *library.Item) []string { return it.Genres }))
	}
	if ids := c.qList("genreids"); len(ids) > 0 {
		var names []string
		for _, id := range ids {
			names = append(names, c.lib.Genres[library.NormalizeID(id)])
		}
		fs = append(fs, anyOf(names, func(it *library.Item) []string { return it.Genres }))
	}
	if st := c.qList("studios"); len(st) > 0 {
		fs = append(fs, anyOf(st, func(it *library.Item) []string { return it.Studios }))
	}
	if ids := c.qList("studioids"); len(ids) > 0 {
		var names []string
		for _, id := range ids {
			names = append(names, c.lib.Studios[library.NormalizeID(id)])
		}
		fs = append(fs, anyOf(names, func(it *library.Item) []string { return it.Studios }))
	}
	if tags := c.qList("tags"); len(tags) > 0 {
		fs = append(fs, anyOf(tags, func(it *library.Item) []string { return it.Tags }))
	}
	if ids := c.qList("personids"); len(ids) > 0 {
		set := map[string]bool{}
		for _, id := range ids {
			set[library.NormalizeID(id)] = true
		}
		types := map[string]bool{}
		for _, t := range c.qList("persontypes") {
			types[strings.ToLower(t)] = true
		}
		fs = append(fs, func(it *library.Item) bool {
			for _, p := range it.People {
				if set[p.ID] && (len(types) == 0 || types[strings.ToLower(p.Type)]) {
					return true
				}
			}
			return false
		})
	}
	if ys := c.qList("years"); len(ys) > 0 {
		set := map[int]bool{}
		for _, y := range ys {
			n, _ := strconv.Atoi(y)
			set[n] = true
		}
		fs = append(fs, func(it *library.Item) bool { return set[it.Year] })
	}
	if v := c.q["parentindexnumber"]; v != "" {
		n, _ := strconv.Atoi(v)
		fs = append(fs, func(it *library.Item) bool { return it.Type == library.TypeEpisode && it.ParentIndex == n })
	}
	if v := c.q["indexnumber"]; v != "" {
		n, _ := strconv.Atoi(v)
		fs = append(fs, func(it *library.Item) bool { return it.HasIndex && it.Index == n })
	}
	if st := c.qList("seriesstatus"); len(st) > 0 {
		fs = append(fs, anyOf(st, func(it *library.Item) []string { return []string{it.Status} }))
	}
	if v := c.q["mincommunityrating"]; v != "" {
		min, _ := strconv.ParseFloat(v, 64)
		fs = append(fs, func(it *library.Item) bool { return it.CommunityRating >= min })
	}
	for _, p := range []struct{ name, provider string }{{"hastmdbid", "Tmdb"}, {"hasimdbid", "Imdb"}, {"hastvdbid", "Tvdb"}} {
		if v, ok := c.qBool(p.name); ok {
			prov := p.provider
			fs = append(fs, func(it *library.Item) bool { return (it.ProviderIDs[prov] != "") == v })
		}
	}
	if v, ok := c.qBool("hasoverview"); ok {
		fs = append(fs, func(it *library.Item) bool { return (it.Overview != "") == v })
	}
	if t, ok := parseTime(c.q["mindatelastsaved"]); ok {
		fs = append(fs, func(it *library.Item) bool { return it.DateModified.After(t) })
	}
	if t, ok := parseTime(c.q["mindatelastsavedforuser"]); ok {
		fs = append(fs, func(it *library.Item) bool {
			for _, p := range append([]*library.Item{it}, it.Episodes()...) {
				if s.store.Get(c.user, p.Key).Updated.After(t) {
					return true
				}
			}
			return false
		})
	}
	if t, ok := parseTime(c.q["minpremieredate"]); ok {
		fs = append(fs, func(it *library.Item) bool { return !it.Premiere.Before(t) })
	}
	if t, ok := parseTime(c.q["maxpremieredate"]); ok {
		fs = append(fs, func(it *library.Item) bool { return !it.Premiere.IsZero() && !it.Premiere.After(t) })
	}
	if t, ok := parseTime(c.q["mindatecreated"]); ok {
		fs = append(fs, func(it *library.Item) bool { return !it.DateCreated.Before(t) })
	}
	if lt := c.qList("locationtypes"); len(lt) > 0 {
		ok := false
		for _, l := range lt {
			ok = ok || strings.EqualFold(l, "FileSystem")
		}
		fs = append(fs, func(*library.Item) bool { return ok })
	}
	if v, ok := c.qBool("ismissing"); ok && v {
		fs = append(fs, func(*library.Item) bool { return false })
	}
	if v, ok := c.qBool("isunaired"); ok && v {
		fs = append(fs, func(*library.Item) bool { return false })
	}
	return fs
}

func anyOf(want []string, get func(*library.Item) []string) itemFilter {
	set := map[string]bool{}
	for _, w := range want {
		set[strings.ToLower(w)] = true
	}
	return func(it *library.Item) bool {
		for _, v := range get(it) {
			if set[strings.ToLower(v)] {
				return true
			}
		}
		return false
	}
}

func (s *Server) played(user string, it *library.Item) bool {
	if it.IsFolder() {
		eps := it.Episodes()
		for _, ep := range eps {
			if !s.store.Get(user, ep.Key).Played {
				return false
			}
		}
		return len(eps) > 0
	}
	return s.store.Get(user, it.Key).Played
}

func (s *Server) resumable(user string, it *library.Item) bool {
	if !it.Playable() {
		return false
	}
	d := s.store.Get(user, it.Key)
	return d.PositionTicks > 0
}

func apply(items []*library.Item, fs []itemFilter) []*library.Item {
	if len(fs) == 0 {
		return items
	}
	out := make([]*library.Item, 0, len(items))
next:
	for _, it := range items {
		for _, f := range fs {
			if !f(it) {
				continue next
			}
		}
		out = append(out, it)
	}
	return out
}

// sortItems implements SortBy / SortOrder.
func (s *Server) sortItems(c *ctx, items []*library.Item) {
	keys := c.qList("sortby")
	if len(keys) == 0 {
		return
	}
	orders := c.qList("sortorder")
	lastPlayed := func(it *library.Item) time.Time {
		var t time.Time
		for _, p := range playables(it) {
			if lp := s.store.Get(c.user, p.Key).LastPlayed; lp != nil && lp.After(t) {
				t = *lp
			}
		}
		return t
	}
	type cmp func(a, b *library.Item) int
	var cmps []cmp
	for i, k := range keys {
		desc := false
		if i < len(orders) {
			desc = strings.EqualFold(orders[i], "Descending")
		} else if len(orders) > 0 {
			desc = strings.EqualFold(orders[len(orders)-1], "Descending")
		}
		var f cmp
		switch strings.ToLower(k) {
		case "sortname", "name":
			f = func(a, b *library.Item) int { return strings.Compare(a.SortName, b.SortName) }
		case "seriessortname":
			f = func(a, b *library.Item) int { return strings.Compare(seriesSort(a), seriesSort(b)) }
		case "datecreated":
			f = func(a, b *library.Item) int { return a.DateCreated.Compare(b.DateCreated) }
		case "datelastcontentadded", "datelastmediaadded":
			f = func(a, b *library.Item) int { return a.DateModified.Compare(b.DateModified) }
		case "dateplayed":
			f = func(a, b *library.Item) int { return lastPlayed(a).Compare(lastPlayed(b)) }
		case "premieredate":
			f = func(a, b *library.Item) int { return a.Premiere.Compare(b.Premiere) }
		case "productionyear":
			f = func(a, b *library.Item) int { return a.Year - b.Year }
		case "communityrating", "criticrating":
			f = func(a, b *library.Item) int { return cmpFloat(a.CommunityRating, b.CommunityRating) }
		case "runtime":
			f = func(a, b *library.Item) int { return cmpInt64(a.Runtime(), b.Runtime()) }
		case "officialrating":
			f = func(a, b *library.Item) int { return strings.Compare(a.OfficialRating, b.OfficialRating) }
		case "playcount":
			f = func(a, b *library.Item) int {
				return s.store.Get(c.user, a.Key).PlayCount - s.store.Get(c.user, b.Key).PlayCount
			}
		case "isplayed", "isunplayed":
			f = func(a, b *library.Item) int { return boolInt(s.played(c.user, a)) - boolInt(s.played(c.user, b)) }
		case "isfolder":
			f = func(a, b *library.Item) int { return boolInt(a.IsFolder()) - boolInt(b.IsFolder()) }
		case "indexnumber":
			f = func(a, b *library.Item) int { return a.Index - b.Index }
		case "parentindexnumber", "airedepisodeorder", "airtime":
			f = func(a, b *library.Item) int {
				if a.ParentIndex != b.ParentIndex {
					return a.ParentIndex - b.ParentIndex
				}
				return a.Index - b.Index
			}
		case "random":
			rand.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
			return
		default:
			continue
		}
		if desc {
			g := f
			f = func(a, b *library.Item) int { return -g(a, b) }
		}
		cmps = append(cmps, f)
	}
	sort.SliceStable(items, func(i, j int) bool {
		for _, f := range cmps {
			if r := f(items[i], items[j]); r != 0 {
				return r < 0
			}
		}
		return false
	})
}

func seriesSort(it *library.Item) string {
	if se := it.Series(); se != nil {
		return se.SortName
	}
	return it.SortName
}

func cmpFloat(a, b float64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func cmpInt64(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// page applies StartIndex / Limit and converts to DTOs.
func (s *Server) page(c *ctx, items []*library.Item) map[string]any {
	total := len(items)
	start := c.qInt("startindex", 0)
	if start > total {
		start = total
	}
	if start < 0 {
		start = 0
	}
	items = items[start:]
	if limit := c.qInt("limit", -1); limit >= 0 && limit < len(items) {
		items = items[:limit]
	}
	f := c.fields()
	out := make([]any, 0, len(items))
	for _, it := range items {
		out = append(out, s.item(c, it, f))
	}
	return wrap(out, total, start)
}

// items handles GET /Items and /Users/{userId}/Items.
func (s *Server) items(c *ctx) {
	var items []*library.Item
	if ids := c.qList("ids"); len(ids) > 0 {
		for _, id := range ids {
			if it := c.lookup(id); it != nil {
				items = append(items, it)
			}
		}
	} else {
		parent := library.NormalizeID(c.q["parentid"])
		recursive, _ := c.qBool("recursive")
		if (parent == "" || parent == rootID) && !recursive {
			s.views(c)
			return
		}
		items = s.scope(c, parent, recursive)
	}
	items = apply(items, s.filters(c))
	s.sortItems(c, items)
	c.json(s.page(c, items))
}

// resume handles GET /Users/{userId}/Items/Resume.
func (s *Server) resume(c *ctx) {
	parent := library.NormalizeID(c.q["parentid"])
	pool := c.lib.All()
	if parent != "" {
		pool = s.scope(c, parent, true)
	}
	var items []*library.Item
	for _, it := range pool {
		if s.resumable(c.user, it) && !s.store.Get(c.user, it.Key).Played {
			items = append(items, it)
		}
	}
	items = apply(items, s.filters(c))
	sort.SliceStable(items, func(i, j int) bool {
		a, b := s.store.Get(c.user, items[i].Key).LastPlayed, s.store.Get(c.user, items[j].Key).LastPlayed
		return a != nil && (b == nil || a.After(*b))
	})
	c.json(s.page(c, items))
}

// latest handles GET /Users/{userId}/Items/Latest (a plain array, not wrapped).
func (s *Server) latest(c *ctx) {
	parent := library.NormalizeID(c.q["parentid"])
	var pool []*library.Item
	if parent == "" {
		pool = c.lib.All()
	} else {
		pool = s.scope(c, parent, true)
	}
	var playable []*library.Item
	for _, it := range pool {
		if it.Playable() {
			playable = append(playable, it)
		}
	}
	types := map[string]bool{}
	for _, t := range c.qList("includeitemtypes") {
		types[strings.ToLower(t)] = true
	}
	if played, ok := c.qBool("isplayed"); ok {
		playable = apply(playable, []itemFilter{func(it *library.Item) bool { return s.played(c.user, it) == played }})
	}
	sort.SliceStable(playable, func(i, j int) bool { return playable[i].DateCreated.After(playable[j].DateCreated) })
	group := true
	if v, ok := c.qBool("groupitems"); ok {
		group = v
	}
	limit := c.qInt("limit", 20)
	seen := map[string]bool{}
	var out []*library.Item
	for _, it := range playable {
		if len(out) >= limit {
			break
		}
		cand := it
		if it.Type == library.TypeEpisode && group {
			cand = it.Series()
		}
		if len(types) > 0 && !types[strings.ToLower(cand.Type)] && !(cand.Type == library.TypeSeries && types["episode"]) {
			continue
		}
		if seen[cand.ID] {
			continue
		}
		seen[cand.ID] = true
		out = append(out, cand)
	}
	f := c.fields()
	res := make([]any, 0, len(out))
	for _, it := range out {
		res = append(res, s.item(c, it, f))
	}
	c.json(res)
}

// nextUp handles GET /Shows/NextUp: per series, the episode after the last one watched.
func (s *Server) nextUp(c *ctx) {
	seriesFilter := library.NormalizeID(c.q["seriesid"])
	parent := library.NormalizeID(c.q["parentid"])
	type cand struct {
		ep   *library.Item
		last time.Time
	}
	var cands []cand
	for _, v := range c.lib.Views {
		if parent != "" && parent != v.ID {
			continue
		}
		for _, se := range v.Children {
			if se.Type != library.TypeSeries || (seriesFilter != "" && se.ID != seriesFilter) {
				continue
			}
			var eps []*library.Item
			for _, season := range se.Children {
				if season.Index != 0 {
					eps = append(eps, season.Children...)
				}
			}
			lastIdx := -1
			var lastTime time.Time
			for i, ep := range eps {
				ud := s.store.Get(c.user, ep.Key)
				if ud.Played {
					lastIdx = i
					if ud.LastPlayed != nil && ud.LastPlayed.After(lastTime) {
						lastTime = *ud.LastPlayed
					}
				}
			}
			if lastIdx < 0 {
				continue
			}
			for _, ep := range eps[lastIdx+1:] {
				if !s.store.Get(c.user, ep.Key).Played {
					cands = append(cands, cand{ep, lastTime})
					break
				}
			}
		}
	}
	sort.SliceStable(cands, func(i, j int) bool { return cands[i].last.After(cands[j].last) })
	items := make([]*library.Item, len(cands))
	for i, x := range cands {
		items[i] = x.ep
	}
	c.json(s.page(c, items))
}

// seasons handles GET /Shows/{id}/Seasons.
func (s *Server) seasons(c *ctx) {
	se := c.lookup(c.param("id"))
	if se == nil || se.Type != library.TypeSeries {
		c.notFound()
		return
	}
	items := apply(append([]*library.Item(nil), se.Children...), s.filters(c))
	c.json(s.page(c, items))
}

// episodes handles GET /Shows/{id}/Episodes.
func (s *Server) episodes(c *ctx) {
	se := c.lookup(c.param("id"))
	if se == nil || se.Type != library.TypeSeries {
		c.notFound()
		return
	}
	var items []*library.Item
	seasonID := library.NormalizeID(c.q["seasonid"])
	seasonNum, hasNum := -1, false
	if v, err := strconv.Atoi(c.q["season"]); err == nil {
		seasonNum, hasNum = v, true
	}
	for _, season := range se.Children {
		if seasonID != "" && season.ID != seasonID {
			continue
		}
		if hasNum && season.Index != seasonNum {
			continue
		}
		items = append(items, season.Children...)
	}
	if start := library.NormalizeID(c.q["startitemid"]); start != "" {
		for i, ep := range items {
			if ep.ID == start {
				items = items[i:]
				break
			}
		}
	}
	if adj := library.NormalizeID(c.q["adjacentto"]); adj != "" {
		for i, ep := range items {
			if ep.ID == adj {
				lo, hi := max(0, i-1), min(len(items), i+2)
				items = items[lo:hi]
				break
			}
		}
	}
	items = apply(items, s.filters(c))
	s.sortItems(c, items)
	c.json(s.page(c, items))
}

// extras handles /Items/{id}/SpecialFeatures (everything but trailers) and LocalTrailers.
func (s *Server) extras(trailers bool) handlerFunc {
	return func(c *ctx) {
		it := c.lookup(c.param("id"))
		out := []any{}
		if it != nil {
			for _, e := range it.Extras {
				if (e.ExtraType == "Trailer") == trailers {
					out = append(out, s.item(c, e, nil))
				}
			}
		}
		c.json(out)
	}
}

func (s *Server) ancestors(c *ctx) {
	it := c.lookup(c.param("id"))
	if it == nil {
		c.notFound()
		return
	}
	out := []any{}
	for p := it.Parent; p != nil; p = p.Parent {
		out = append(out, s.item(c, p, fieldSet{}))
	}
	out = append(out, s.viewDTO(c, it.View), s.rootDTO(c))
	c.json(out)
}

// names handles /Genres and /Studios.
func (s *Server) names(typ, prefix string, get func(*library.Item) []string) handlerFunc {
	return func(c *ctx) {
		parent := library.NormalizeID(c.q["parentid"])
		var pool []*library.Item
		if parent == "" {
			pool = c.lib.Top()
		} else {
			pool = s.scope(c, parent, false)
		}
		pool = apply(pool, s.filters(c))
		counts := map[string]int{}
		for _, it := range pool {
			for _, n := range get(it) {
				counts[n]++
			}
		}
		names := make([]string, 0, len(counts))
		for n := range counts {
			names = append(names, n)
		}
		sort.Slice(names, func(i, j int) bool { return library.SortName(names[i]) < library.SortName(names[j]) })
		total := len(names)
		start := min(max(c.qInt("startindex", 0), 0), total)
		names = names[start:]
		if l := c.qInt("limit", -1); l >= 0 && l < len(names) {
			names = names[:l]
		}
		out := []any{}
		for _, n := range names {
			d := s.namedDTO(c, typ, prefix, n)
			d["ChildCount"] = counts[n]
			out = append(out, d)
		}
		c.json(wrap(out, total, start))
	}
}

func (s *Server) namedByName(typ, prefix string) handlerFunc {
	return func(c *ctx) {
		c.json(s.namedDTO(c, typ, prefix, c.param("name")))
	}
}

func (s *Server) persons(c *ctx) {
	term := library.Fold(c.q["searchterm"])
	var ps []*library.Person
	for _, p := range c.lib.People {
		if term == "" || strings.Contains(library.Fold(p.Name), term) {
			ps = append(ps, p)
		}
	}
	sort.Slice(ps, func(i, j int) bool { return ps[i].Name < ps[j].Name })
	total := len(ps)
	start := min(max(c.qInt("startindex", 0), 0), total)
	ps = ps[start:]
	if l := c.qInt("limit", -1); l >= 0 && l < len(ps) {
		ps = ps[:l]
	}
	out := []any{}
	for _, p := range ps {
		out = append(out, s.personDTO(c, p))
	}
	c.json(wrap(out, total, start))
}

func (s *Server) personByName(c *ctx) {
	if p := c.lib.People[library.ID("person:"+c.param("name"))]; p != nil {
		c.json(s.personDTO(c, p))
		return
	}
	c.notFound()
}

func (s *Server) searchHints(c *ctx) {
	term := library.Fold(strings.TrimSpace(c.q["searchterm"]))
	if term == "" {
		c.json(map[string]any{"SearchHints": []any{}, "TotalRecordCount": 0})
		return
	}
	items := apply(c.lib.All(), append(s.filters(c), func(it *library.Item) bool {
		return strings.Contains(library.Fold(it.Name), term)
	}))
	total := len(items)
	if l := c.qInt("limit", 50); l < len(items) {
		items = items[:l]
	}
	out := []any{}
	for _, it := range items {
		h := map[string]any{"ItemId": it.ID, "Id": it.ID, "Name": it.Name, "Type": it.Type, "IsFolder": it.IsFolder(),
			"MatchedTerm": c.q["searchterm"], "MediaType": "Video"}
		if it.Year > 0 {
			h["ProductionYear"] = it.Year
		}
		if t := firstTag(it, library.ImgPrimary); t != "" {
			h["PrimaryImageTag"] = t
		}
		if t := firstTag(it, library.ImgThumb); t != "" {
			h["ThumbImageTag"], h["ThumbImageItemId"] = t, it.ID
		}
		if bs := it.Images[library.ImgBackdrop]; len(bs) > 0 {
			h["BackdropImageTag"], h["BackdropImageItemId"] = bs[0].Tag, it.ID
		}
		if rt := it.Runtime(); rt > 0 {
			h["RunTimeTicks"] = rt
		}
		if se := it.Series(); se != nil && it.Type != library.TypeSeries {
			h["Series"] = se.Name
		}
		if it.HasIndex {
			h["IndexNumber"] = it.Index
			if it.Type == library.TypeEpisode {
				h["ParentIndexNumber"] = it.ParentIndex
			}
		}
		out = append(out, h)
	}
	c.json(map[string]any{"SearchHints": out, "TotalRecordCount": total})
}

func (s *Server) filtersInfo(legacy bool) handlerFunc {
	return func(c *ctx) {
		parent := library.NormalizeID(c.q["parentid"])
		pool := c.lib.Top()
		if parent != "" {
			pool = s.scope(c, parent, false)
		}
		pool = apply(pool, s.filters(c))
		genres, tags, ratings, years := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[int]bool{}
		for _, it := range pool {
			for _, g := range it.Genres {
				genres[g] = true
			}
			for _, t := range it.Tags {
				tags[t] = true
			}
			if it.OfficialRating != "" {
				ratings[it.OfficialRating] = true
			}
			if it.Year > 0 {
				years[it.Year] = true
			}
		}
		keys := func(m map[string]bool) []string {
			out := []string{}
			for k := range m {
				out = append(out, k)
			}
			sort.Strings(out)
			return out
		}
		if legacy {
			ys := []int{}
			for y := range years {
				ys = append(ys, y)
			}
			sort.Ints(ys)
			c.json(map[string]any{"Genres": keys(genres), "Tags": keys(tags), "OfficialRatings": keys(ratings), "Years": ys})
			return
		}
		gs := []any{}
		for _, g := range keys(genres) {
			gs = append(gs, map[string]any{"Name": g, "Id": library.ID("genre:" + g)})
		}
		c.json(map[string]any{"Genres": gs, "Tags": keys(tags)})
	}
}

func (s *Server) counts(c *ctx) {
	movies, series, episodes := 0, 0, 0
	for _, it := range c.lib.All() {
		switch it.Type {
		case library.TypeMovie:
			movies++
		case library.TypeSeries:
			series++
		case library.TypeEpisode:
			episodes++
		}
	}
	c.json(map[string]any{"MovieCount": movies, "SeriesCount": series, "EpisodeCount": episodes, "ArtistCount": 0,
		"ProgramCount": 0, "TrailerCount": 0, "SongCount": 0, "AlbumCount": 0, "MusicVideoCount": 0, "BoxSetCount": 0,
		"BookCount": 0, "ItemCount": movies + series + episodes})
}

func (s *Server) themeMedia(c *ctx) {
	id := library.NormalizeID(c.param("id"))
	empty := map[string]any{"Items": []any{}, "TotalRecordCount": 0, "StartIndex": 0, "OwnerId": id}
	c.json(map[string]any{"ThemeVideosResult": empty, "ThemeSongsResult": empty, "SoundtrackSongsResult": empty})
}

// additionalParts handles GET /Videos/{id}/AdditionalParts (parts 2..n of a multi-part movie).
func (s *Server) additionalParts(c *ctx) {
	var parts []*library.Item
	if it := c.lookup(c.param("id")); it != nil {
		parts = it.Parts
	}
	c.json(s.page(c, parts))
}
