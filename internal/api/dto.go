package api

import (
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/zemiak/movies/internal/library"
	"github.com/zemiak/movies/internal/store"
)

// fieldSet is the "Fields" query parameter; nil means "everything" (single-item requests).
type fieldSet map[string]bool

func (f fieldSet) has(name string) bool { return f == nil || f[name] }

func (c *ctx) fields() fieldSet {
	f := fieldSet{}
	for _, n := range c.qList("fields") {
		f[strings.ToLower(n)] = true
	}
	return f
}

// playables returns the playable items of a top-level item / folder.
func playables(it *library.Item) []*library.Item {
	if it.Playable() {
		return []*library.Item{it}
	}
	return it.Episodes()
}

// userData builds the UserData object; folders aggregate their episodes.
func (s *Server) userData(user string, it *library.Item) map[string]any {
	own := s.store.Get(user, it.Key)
	d := map[string]any{
		"PlaybackPositionTicks": own.PositionTicks,
		"PlayCount":             own.PlayCount,
		"IsFavorite":            own.IsFavorite,
		"Played":                own.Played,
		"Key":                   it.ID,
		"ItemId":                it.ID,
	}
	if own.LastPlayed != nil {
		d["LastPlayedDate"] = jfTime(*own.LastPlayed)
	}
	if it.IsFolder() {
		eps := it.Episodes()
		unplayed := 0
		var last *time.Time
		for _, ep := range eps {
			ud := s.store.Get(user, ep.Key)
			if !ud.Played {
				unplayed++
			}
			if ud.LastPlayed != nil && (last == nil || ud.LastPlayed.After(*last)) {
				last = ud.LastPlayed
			}
		}
		d["UnplayedItemCount"] = unplayed
		d["Played"] = len(eps) > 0 && unplayed == 0
		d["PlaybackPositionTicks"] = 0
		if len(eps) > 0 {
			d["PlayedPercentage"] = float64(len(eps)-unplayed) * 100 / float64(len(eps))
		}
		if last != nil {
			d["LastPlayedDate"] = jfTime(*last)
		}
	} else if rt := it.Runtime(); rt > 0 && own.PositionTicks > 0 {
		d["PlayedPercentage"] = math.Min(100, float64(own.PositionTicks)*100/float64(rt))
	}
	return d
}

func imageTags(it *library.Item) (map[string]any, []string) {
	tags := map[string]any{}
	for typ, imgs := range it.Images {
		if typ != library.ImgBackdrop && len(imgs) > 0 {
			tags[typ] = imgs[0].Tag
		}
	}
	backdrops := []string{}
	for _, im := range it.Images[library.ImgBackdrop] {
		backdrops = append(backdrops, im.Tag)
	}
	return tags, backdrops
}

func firstTag(it *library.Item, typ string) string {
	if it == nil || len(it.Images[typ]) == 0 {
		return ""
	}
	return it.Images[typ][0].Tag
}

func etag(it *library.Item) string {
	return library.ID(fmt.Sprintf("%s|%d", it.Key, it.DateModified.UnixNano()))[:16]
}

func nameIDs(prefix string, names []string) []any {
	out := []any{}
	for _, n := range names {
		out = append(out, map[string]any{"Name": n, "Id": library.ID(prefix + n)})
	}
	return out
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func parentID(it *library.Item) string {
	if it.Parent != nil {
		return it.Parent.ID
	}
	return it.View.ID
}

// item builds a BaseItemDto.
func (s *Server) item(c *ctx, it *library.Item, f fieldSet) map[string]any {
	tags, backdrops := imageTags(it)
	d := map[string]any{
		"Name":              it.Name,
		"ServerId":          s.cfg.ServerID,
		"Id":                it.ID,
		"Etag":              etag(it),
		"DateCreated":       jfTime(it.DateCreated),
		"DateLastSaved":     jfTime(it.DateModified),
		"CanDelete":         false,
		"CanDownload":       it.Playable(),
		"SortName":          it.SortName,
		"Type":              it.Type,
		"IsFolder":          it.IsFolder(),
		"ParentId":          parentID(it),
		"LocationType":      "FileSystem",
		"ImageTags":         tags,
		"BackdropImageTags": backdrops,
		"ImageBlurHashes":   map[string]any{},
		"UserData":          s.userData(c.user, it),
		"PlayAccess":        "Full",
		"ChannelId":         nil,
	}
	if it.OriginalTitle != "" {
		d["OriginalTitle"] = it.OriginalTitle
	}
	if it.Year > 0 {
		d["ProductionYear"] = it.Year
	}
	if !it.Premiere.IsZero() {
		d["PremiereDate"] = jfTime(time.Date(it.Premiere.Year(), it.Premiere.Month(), it.Premiere.Day(), 0, 0, 0, 0, time.UTC))
	}
	if it.CommunityRating > 0 {
		d["CommunityRating"] = math.Round(it.CommunityRating*10) / 10
	}
	if it.OfficialRating != "" {
		d["OfficialRating"] = it.OfficialRating
	}
	if rt := it.Runtime(); rt > 0 {
		d["RunTimeTicks"] = rt
	}
	// Cheap fields are always sent, whatever "Fields" asks for: some clients don't ask but use them.
	d["Overview"] = it.Overview
	d["Path"] = it.Path
	d["Taglines"] = []string{}
	if it.Tagline != "" {
		d["Taglines"] = []string{it.Tagline}
	}
	d["Genres"] = orEmpty(it.Genres)
	d["GenreItems"] = nameIDs("genre:", it.Genres)
	d["Studios"] = nameIDs("studio:", it.Studios)
	d["Tags"] = orEmpty(it.Tags)
	d["ProductionLocations"] = orEmpty(it.Countries)
	ids := map[string]string{}
	for k, v := range it.ProviderIDs {
		ids[k] = v
	}
	d["ProviderIds"] = ids
	d["ExternalUrls"] = externalURLs(it)
	d["RemoteTrailers"] = []any{}
	if f.has("people") {
		d["People"] = s.people(it)
	}
	if it.ExtraType != "" {
		d["ExtraType"] = it.ExtraType
	}

	switch it.Type {
	case library.TypeMovie, library.TypeEpisode, library.TypeVideo:
		s.addMediaFields(c, d, it, f)
	case library.TypeSeries:
		d["Status"] = it.Status
		d["ChildCount"] = len(it.Children)
		d["RecursiveItemCount"] = len(it.Episodes())
		d["AirDays"] = []any{}
		d["DisplayOrder"] = ""
		d["DateLastMediaAdded"] = jfTime(it.DateModified)
		d["MediaType"] = "Unknown"
		d["SpecialFeatureCount"] = len(it.Extras)
	case library.TypeSeason:
		se := it.Series()
		d["IndexNumber"] = it.Index
		d["SeriesName"] = se.Name
		d["SeriesId"] = se.ID
		d["ChildCount"] = len(it.Children)
		d["RecursiveItemCount"] = len(it.Children)
		d["MediaType"] = "Unknown"
		d["SpecialFeatureCount"] = len(it.Extras)
		s.addSeriesImages(d, se)
	}
	if it.Type == library.TypeEpisode {
		season := it.Parent
		se := season.Parent
		d["IndexNumber"] = it.Index
		d["ParentIndexNumber"] = it.ParentIndex
		d["SeriesName"] = se.Name
		d["SeriesId"] = se.ID
		d["SeasonId"] = season.ID
		d["SeasonName"] = season.Name
		s.addSeriesImages(d, se)
		if t := firstTag(season, library.ImgPrimary); t != "" {
			d["ParentPrimaryImageItemId"] = season.ID
			d["ParentPrimaryImageTag"] = t
		}
	}
	if it.Type == library.TypeMovie {
		d["LocalTrailerCount"] = countExtras(it, "Trailer")
		d["SpecialFeatureCount"] = len(it.Extras) - countExtras(it, "Trailer")
		if len(it.Parts) > 0 {
			d["PartCount"] = 1 + len(it.Parts)
		}
		if it.SetName != "" {
			d["CollectionName"] = it.SetName // not Jellyfin; harmless for clients
		}
	}
	return d
}

func countExtras(it *library.Item, typ string) int {
	n := 0
	for _, e := range it.Extras {
		if e.ExtraType == typ {
			n++
		}
	}
	return n
}

func (s *Server) addSeriesImages(d map[string]any, se *library.Item) {
	if t := firstTag(se, library.ImgPrimary); t != "" {
		d["SeriesPrimaryImageTag"] = t
	}
	if bs := se.Images[library.ImgBackdrop]; len(bs) > 0 {
		d["ParentBackdropItemId"] = se.ID
		tags := []string{}
		for _, b := range bs {
			tags = append(tags, b.Tag)
		}
		d["ParentBackdropImageTags"] = tags
	}
	if t := firstTag(se, library.ImgLogo); t != "" {
		d["ParentLogoItemId"] = se.ID
		d["ParentLogoImageTag"] = t
	}
	if t := firstTag(se, library.ImgThumb); t != "" {
		d["ParentThumbItemId"] = se.ID
		d["ParentThumbImageTag"] = t
		d["SeriesThumbImageTag"] = t
	}
	if t := firstTag(se, library.ImgArt); t != "" {
		d["ParentArtItemId"] = se.ID
		d["ParentArtImageTag"] = t
	}
}

func externalURLs(it *library.Item) []any {
	out := []any{}
	if id := it.ProviderIDs["Imdb"]; id != "" {
		out = append(out, map[string]any{"Name": "IMDb", "Url": "https://www.imdb.com/title/" + id})
	}
	if id := it.ProviderIDs["Tmdb"]; id != "" {
		kind := "movie"
		if it.Type == library.TypeSeries {
			kind = "tv"
		}
		if it.Type == library.TypeMovie || it.Type == library.TypeSeries {
			out = append(out, map[string]any{"Name": "TheMovieDb", "Url": "https://www.themoviedb.org/" + kind + "/" + id})
		}
	}
	return out
}

func (s *Server) people(it *library.Item) []any {
	out := []any{}
	for _, p := range it.People {
		d := map[string]any{"Name": p.Name, "Id": p.ID, "Type": p.Type}
		if p.Role != "" {
			d["Role"] = p.Role
		}
		if p.Thumb != "" {
			d["PrimaryImageTag"] = library.ID(p.Thumb)[:16]
		}
		out = append(out, d)
	}
	return out
}

func (s *Server) addMediaFields(c *ctx, d map[string]any, it *library.Item, f fieldSet) {
	ms := it.Sources[0]
	d["MediaType"] = "Video"
	d["VideoType"] = "VideoFile"
	d["Container"] = ms.Container
	d["IsHD"] = false
	d["HasSubtitles"] = len(ms.Subtitles) > 0
	if ms.Probe != nil {
		for _, st := range ms.Probe.Streams {
			switch {
			case st.CodecType == "video" && st.Disposition["attached_pic"] == 0 && d["Width"] == nil:
				d["Width"], d["Height"] = st.Width, st.Height
				d["IsHD"] = st.Width >= 1200 || st.Height >= 700
			case st.CodecType == "subtitle":
				d["HasSubtitles"] = true
			}
		}
	}
	if len(it.Sources) > 1 {
		d["MediaSourceCount"] = len(it.Sources)
	}
	d["LocationType"] = "FileSystem"
	if f.has("mediasources") || f.has("mediastreams") {
		srcs := []any{}
		for _, m := range it.Sources {
			srcs = append(srcs, s.mediaSource(c, it, m))
		}
		if f.has("mediasources") {
			d["MediaSources"] = srcs
		}
		if f.has("mediastreams") {
			d["MediaStreams"] = srcs[0].(map[string]any)["MediaStreams"]
		}
	}
	if f.has("chapters") {
		d["Chapters"] = []any{}
	}
}

// mediaSource builds a MediaSourceInfo for direct play.
func (s *Server) mediaSource(c *ctx, it *library.Item, ms *library.MediaSource) map[string]any {
	streams, defAudio, defSub := s.mediaStreams(c, it, ms)
	rt := it.RuntimeTicks
	bitrate := 0
	if ms.Probe != nil {
		if t := ms.Probe.DurationTicks(); t > 0 {
			rt = t
		}
		bitrate = ms.Probe.Bitrate()
	}
	if bitrate == 0 && rt > 0 {
		bitrate = int(float64(ms.Size*8) / (float64(rt) / 1e7))
	}
	d := map[string]any{
		"Protocol":                            "File",
		"Id":                                  ms.ID,
		"Path":                                ms.Path,
		"Type":                                "Default",
		"Container":                           ms.Container,
		"Size":                                ms.Size,
		"Name":                                sourceName(it, ms),
		"IsRemote":                            false,
		"ETag":                                library.ID(fmt.Sprintf("%s|%d|%d", ms.Path, ms.Size, ms.ModTime.UnixNano()))[:16],
		"ReadAtNativeFramerate":               false,
		"IgnoreDts":                           false,
		"IgnoreIndex":                         false,
		"GenPtsInput":                         false,
		"SupportsTranscoding":                 false,
		"SupportsDirectStream":                true,
		"SupportsDirectPlay":                  true,
		"IsInfiniteStream":                    false,
		"UseMostCompatibleTranscodingProfile": false,
		"RequiresOpening":                     false,
		"RequiresClosing":                     false,
		"RequiresLooping":                     false,
		"SupportsProbing":                     true,
		"VideoType":                           "VideoFile",
		"MediaStreams":                        streams,
		"MediaAttachments":                    []any{},
		"Formats":                             []any{},
		"RequiredHttpHeaders":                 map[string]any{},
		"TranscodingSubProtocol":              "http",
		"HasSegments":                         false,
		"DirectStreamUrl": fmt.Sprintf("/Videos/%s/stream.%s?Static=true&MediaSourceId=%s&api_key=%s",
			it.ID, ms.Container, ms.ID, url.QueryEscape(c.token)),
	}
	if rt > 0 {
		d["RunTimeTicks"] = rt
	}
	if bitrate > 0 {
		d["Bitrate"] = bitrate
	}
	if defAudio >= 0 {
		d["DefaultAudioStreamIndex"] = defAudio
	}
	if defSub >= 0 {
		d["DefaultSubtitleStreamIndex"] = defSub
	}
	return d
}

var languageNames = map[string]string{
	"eng": "English", "en": "English", "ces": "Czech", "cze": "Czech", "cs": "Czech", "slk": "Slovak", "slo": "Slovak",
	"sk": "Slovak", "deu": "German", "ger": "German", "de": "German", "fra": "French", "fre": "French", "fr": "French",
	"spa": "Spanish", "es": "Spanish", "ita": "Italian", "it": "Italian", "rus": "Russian", "ru": "Russian",
	"pol": "Polish", "pl": "Polish", "hun": "Hungarian", "hu": "Hungarian", "jpn": "Japanese", "ja": "Japanese",
	"por": "Portuguese", "pt": "Portuguese", "nld": "Dutch", "dut": "Dutch", "nl": "Dutch", "swe": "Swedish",
	"sv": "Swedish", "nor": "Norwegian", "no": "Norwegian", "dan": "Danish", "da": "Danish", "fin": "Finnish",
	"fi": "Finnish", "ukr": "Ukrainian", "uk": "Ukrainian", "hrv": "Croatian", "hr": "Croatian", "chi": "Chinese",
	"zho": "Chinese", "zh": "Chinese", "kor": "Korean", "ko": "Korean", "tur": "Turkish", "tr": "Turkish",
	"gre": "Greek", "ell": "Greek", "el": "Greek", "heb": "Hebrew", "he": "Hebrew", "ara": "Arabic", "ar": "Arabic",
}

// iso3 maps 2-letter codes (from subtitle file names) to the 3-letter codes Jellyfin reports.
var iso3 = map[string]string{"en": "eng", "cs": "cze", "sk": "slo", "de": "ger", "fr": "fre", "es": "spa",
	"it": "ita", "ru": "rus", "pl": "pol", "hu": "hun", "ja": "jpn", "pt": "por", "nl": "dut", "sv": "swe",
	"no": "nor", "da": "dan", "fi": "fin", "uk": "ukr", "hr": "hrv", "zh": "chi", "ko": "kor", "tr": "tur"}

func langName(code string) string {
	if n, ok := languageNames[strings.ToLower(code)]; ok {
		return n
	}
	return code
}

var textSubCodecs = map[string]bool{"subrip": true, "srt": true, "ass": true, "ssa": true, "mov_text": true,
	"webvtt": true, "vtt": true, "text": true, "sub": false, "smi": true}

func frameRate(s string) float64 {
	a, b, ok := strings.Cut(s, "/")
	if !ok {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}
	n, _ := strconv.ParseFloat(a, 64)
	m, _ := strconv.ParseFloat(b, 64)
	if m == 0 {
		return 0
	}
	return math.Round(n/m*1000) / 1000
}

func resolutionLabel(w, h int) string {
	switch {
	case w >= 3800 || h >= 2100:
		return "4K"
	case w >= 2500 || h >= 1400:
		return "1440p"
	case w >= 1900 || h >= 1000:
		return "1080p"
	case w >= 1260 || h >= 700:
		return "720p"
	case h >= 560:
		return "576p"
	case h >= 470:
		return "480p"
	case h > 0:
		return fmt.Sprintf("%dp", h)
	}
	return ""
}

func channelLabel(ch int, layout string) string {
	switch ch {
	case 1:
		return "Mono"
	case 2:
		return "Stereo"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	}
	if layout != "" {
		return layout
	}
	if ch > 0 {
		return fmt.Sprintf("%d ch", ch)
	}
	return ""
}

func joinNonEmpty(sep string, parts ...string) string {
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, sep)
}

// mediaStreams converts ffprobe streams (and external subtitles) to Jellyfin MediaStreams.
func (s *Server) mediaStreams(c *ctx, it *library.Item, ms *library.MediaSource) ([]any, int, int) {
	out := []any{}
	defAudio, defSub, firstAudio := -1, -1, -1
	maxIndex := -1
	if p := ms.Probe; p != nil {
		for _, st := range p.Streams {
			if st.Index > maxIndex {
				maxIndex = st.Index
			}
			lang := st.Tags["language"]
			if lang == "und" {
				lang = ""
			}
			title := st.Tags["title"]
			isDefault := st.Disposition["default"] == 1
			d := map[string]any{
				"Codec":                  st.CodecName,
				"Index":                  st.Index,
				"IsDefault":              isDefault,
				"IsForced":               st.Disposition["forced"] == 1,
				"IsExternal":             false,
				"IsInterlaced":           false,
				"IsTextSubtitleStream":   false,
				"SupportsExternalStream": false,
				"Level":                  st.Level,
				"IsHearingImpaired":      st.Disposition["hearing_impaired"] == 1,
			}
			if lang != "" {
				d["Language"] = lang
			}
			if title != "" {
				d["Title"] = title
			}
			if st.CodecTag != "" && !strings.HasPrefix(st.CodecTag, "[") {
				d["CodecTag"] = st.CodecTag
			}
			if st.Profile != "" {
				d["Profile"] = st.Profile
			}
			if st.TimeBase != "" {
				d["TimeBase"] = st.TimeBase
			}
			if br, err := strconv.Atoi(st.BitRate); err == nil && br > 0 {
				d["BitRate"] = br
			}
			switch st.CodecType {
			case "video":
				if st.Disposition["attached_pic"] == 1 {
					continue
				}
				d["Type"] = "Video"
				d["Width"], d["Height"] = st.Width, st.Height
				if st.DisplayAspectRatio != "" && st.DisplayAspectRatio != "0:1" {
					d["AspectRatio"] = st.DisplayAspectRatio
				}
				if st.PixFmt != "" {
					d["PixelFormat"] = st.PixFmt
				}
				depth := 8
				if b, err := strconv.Atoi(st.BitsPerRawSample); err == nil && b > 0 {
					depth = b
				} else if strings.Contains(st.PixFmt, "10") {
					depth = 10
				} else if strings.Contains(st.PixFmt, "12") {
					depth = 12
				}
				d["BitDepth"] = depth
				if r := frameRate(st.RFrameRate); r > 0 {
					d["RealFrameRate"] = r
				}
				if r := frameRate(st.AvgFrameRate); r > 0 {
					d["AverageFrameRate"] = r
				}
				d["IsInterlaced"] = st.FieldOrder != "" && st.FieldOrder != "progressive" && st.FieldOrder != "unknown"
				d["IsAVC"] = st.CodecName == "h264"
				rng, rngType := "SDR", "SDR"
				switch st.ColorTransfer {
				case "smpte2084":
					rng, rngType = "HDR", "HDR10"
				case "arib-std-b67":
					rng, rngType = "HDR", "HLG"
				}
				for _, sd := range st.SideData {
					if t, _ := sd["side_data_type"].(string); strings.Contains(t, "DOVI") {
						rng, rngType = "HDR", "DOVI"
					}
				}
				d["VideoRange"], d["VideoRangeType"] = rng, rngType
				for k, v := range map[string]string{"ColorTransfer": st.ColorTransfer, "ColorPrimaries": st.ColorPrimaries, "ColorSpace": st.ColorSpace} {
					if v != "" {
						d[k] = v
					}
				}
				d["DisplayTitle"] = joinNonEmpty(" ", resolutionLabel(st.Width, st.Height), strings.ToUpper(st.CodecName), rng)
			case "audio":
				d["Type"] = "Audio"
				if st.Channels > 0 {
					d["Channels"] = st.Channels
				}
				if st.ChannelLayout != "" {
					d["ChannelLayout"] = st.ChannelLayout
				}
				if sr, err := strconv.Atoi(st.SampleRate); err == nil {
					d["SampleRate"] = sr
				}
				def := ""
				if isDefault {
					def = "Default"
				}
				d["DisplayTitle"] = joinNonEmpty(" - ", title, langName(lang), strings.ToUpper(st.CodecName),
					channelLabel(st.Channels, st.ChannelLayout), def)
				if firstAudio < 0 {
					firstAudio = st.Index
				}
				if isDefault && defAudio < 0 {
					defAudio = st.Index
				}
			case "subtitle":
				d["Type"] = "Subtitle"
				d["IsTextSubtitleStream"] = textSubCodecs[st.CodecName]
				forced := ""
				if st.Disposition["forced"] == 1 {
					forced = "Forced"
				}
				d["DisplayTitle"] = joinNonEmpty(" - ", title, langName(lang), strings.ToUpper(st.CodecName), forced)
				if (isDefault || st.Disposition["forced"] == 1) && defSub < 0 {
					defSub = st.Index
				}
			default:
				continue
			}
			out = append(out, d)
		}
	}
	for i, sub := range ms.Subtitles {
		idx := maxIndex + 1 + i
		lang := sub.Language
		if l3, ok := iso3[lang]; ok {
			lang = l3
		}
		forced := ""
		if sub.Forced {
			forced = "Forced"
		}
		d := map[string]any{
			"Codec":                  sub.Codec,
			"Index":                  idx,
			"Type":                   "Subtitle",
			"IsDefault":              false,
			"IsForced":               sub.Forced,
			"IsExternal":             true,
			"IsInterlaced":           false,
			"IsTextSubtitleStream":   textSubCodecs[sub.Codec],
			"SupportsExternalStream": true,
			"Path":                   sub.Path,
			"DeliveryMethod":         "External",
			"DeliveryUrl": fmt.Sprintf("/Videos/%s/%s/Subtitles/%d/0/Stream.%s?api_key=%s", it.ID, ms.ID, idx, sub.Codec,
				url.QueryEscape(c.token)),
			"DisplayTitle": joinNonEmpty(" - ", langName(lang), strings.ToUpper(sub.Codec), forced, "External"),
		}
		if lang != "" {
			d["Language"] = lang
		}
		out = append(out, d)
	}
	if defAudio < 0 {
		defAudio = firstAudio
	}
	return out, defAudio, defSub
}

// userDataDTO wraps a store.UserData for the played/favorite endpoints.
func (s *Server) userDataDTO(c *ctx, it *library.Item, _ store.UserData) map[string]any {
	return s.userData(c.user, it)
}

// sourceName labels the primary version by its audio language when there are several versions
// ("English" next to "Cesky" reads better than "300" next to "Cesky").
func sourceName(it *library.Item, ms *library.MediaSource) string {
	if len(it.Sources) < 2 || ms != it.Sources[0] || ms.Probe == nil {
		return ms.Name
	}
	for _, st := range ms.Probe.Streams {
		if st.CodecType == "audio" {
			if l := st.Tags["language"]; l != "" && l != "und" {
				return langName(l)
			}
			break
		}
	}
	return ms.Name
}
