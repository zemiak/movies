package api

import (
	"net/http"
	"sort"

	"github.com/zemiak/movies/internal/library"
)

func (s *Server) registerRoutes() {
	const (
		get     = "GET"
		post    = "POST"
		del     = "DELETE"
		getPost = "GET,POST"
	)
	text := func(body string) handlerFunc {
		return func(c *ctx) {
			c.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = c.w.Write([]byte(body))
		}
	}

	// --- anonymous
	s.handleAnon(get, "/", text("movies: a Jellyfin-compatible media server\n"))
	s.handleAnon(get, "/web/index.html", text("movies: no web UI\n"))
	s.handleAnon(get, "/health", text("ok\n"))
	s.handleAnon(get, "/System/Info/Public", func(c *ctx) { c.json(s.publicInfo(c)) })
	s.handleAnon(getPost, "/System/Ping", text(`"Jellyfin Server"`))
	s.handleAnon(get, "/Branding/Configuration", func(c *ctx) {
		c.json(map[string]any{"LoginDisclaimer": "", "CustomCss": "", "SplashscreenEnabled": false})
	})
	s.handleAnon(get, "/Branding/Css", text(""))
	s.handleAnon(get, "/Branding/Css.css", text(""))
	s.handleAnon(get, "/QuickConnect/Enabled", func(c *ctx) { c.json(false) })
	s.handleAnon(get, "/Users/Public", func(c *ctx) {
		if s.cfg.open { // lets clients offer the user without asking for a password
			c.json([]any{s.userDTO(OpenUser)})
			return
		}
		c.json([]any{})
	})
	s.handleAnon(post, "/Users/AuthenticateByName", s.login)
	s.handleAnon(get, "/Startup/Configuration", func(c *ctx) { c.notFound() })
	// Images are anonymous in Jellyfin too.
	s.handleAnon(get, "/Items/{id}/Images/{type}", s.image)
	s.handleAnon(get, "/Items/{id}/Images/{type}/{index}", s.image)

	// --- system, users, sessions
	s.handle(get, "/System/Info", s.systemInfo)
	s.handle(get, "/System/Endpoint", func(c *ctx) { c.json(map[string]any{"IsLocal": true, "IsInNetwork": true}) })
	s.handle(get, "/System/Configuration", func(c *ctx) {
		c.json(map[string]any{"EnableMetrics": false, "ServerName": s.cfg.ServerName, "MinResumePct": minResumePct,
			"MaxResumePct": maxResumePct, "MinResumeDurationSeconds": minResumeDuration / 1e7, "UICulture": "en-US"})
	})
	s.handle(get, "/Users/Me", func(c *ctx) { c.json(s.userDTO(c.user)) })
	s.handle(get, "/Users", func(c *ctx) {
		out := []any{}
		names := make([]string, 0, len(s.cfg.Users))
		for n := range s.cfg.Users {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			out = append(out, s.userDTO(n))
		}
		c.json(out)
	})
	s.handle(get, "/Users/{uid}", func(c *ctx) { c.json(s.userDTO(c.user)) })
	s.handle(post, "/Sessions/Logout", func(c *ctx) { s.store.DeleteToken(c.token); c.noContent() })
	s.handle(post, "/Sessions/Capabilities", func(c *ctx) { c.noContent() })
	s.handle(post, "/Sessions/Capabilities/Full", func(c *ctx) { c.noContent() })
	s.handle(get, "/Sessions", func(c *ctx) { c.json([]any{s.sessionInfo(c, c.user)}) })
	s.handle(get, "/DisplayPreferences/{id}", s.displayPreferences)
	s.handle(post, "/DisplayPreferences/{id}", func(c *ctx) { c.noContent() })
	s.handle(get, "/Plugins", emptyArray)
	s.handle(get, "/Packages", emptyArray)
	s.handle(get, "/ScheduledTasks", emptyArray)
	s.handle(get, "/Localization/Cultures", emptyArray)
	s.handle(get, "/Localization/Countries", emptyArray)
	s.handle(get, "/Localization/ParentalRatings", emptyArray)
	s.handle(get, "/Localization/Options", emptyArray)

	// --- libraries
	s.handle(get, "/Users/{uid}/Views", s.views)
	s.handle(get, "/UserViews", s.views)
	s.handle(get, "/Users/{uid}/GroupingOptions", emptyArray)
	s.handle(get, "/UserViews/GroupingOptions", emptyArray)
	s.handle(get, "/Library/MediaFolders", s.views)
	s.handle(get, "/Library/VirtualFolders", s.virtualFolders)
	s.handle(post, "/Library/Refresh", func(c *ctx) { s.lib.Rescan(); c.noContent() })
	s.handle(post, "/Items/{id}/Refresh", func(c *ctx) { s.lib.Rescan(); c.noContent() })

	// --- items
	s.handle(get, "/Items", s.items)
	s.handle(get, "/Users/{uid}/Items", s.items)
	s.handle(get, "/Users/{uid}/Items/Root", func(c *ctx) { c.json(s.rootDTO(c)) })
	s.handle(get, "/Items/Root", func(c *ctx) { c.json(s.rootDTO(c)) })
	s.handle(get, "/Users/{uid}/Items/Resume", s.resume)
	s.handle(get, "/UserItems/Resume", s.resume)
	s.handle(get, "/Users/{uid}/Items/Latest", s.latest)
	s.handle(get, "/Items/Latest", s.latest)
	s.handle(get, "/Items/Filters", s.filtersInfo(true))
	s.handle(get, "/Items/Filters2", s.filtersInfo(false))
	s.handle(get, "/Items/Counts", s.counts)
	s.handle(get, "/Users/{uid}/Items/{id}", s.itemByID)
	s.handle(get, "/Items/{id}", s.itemByID)
	s.handle(get, "/Users/{uid}/Items/{id}/Intros", emptyList)
	s.handle(get, "/Items/{id}/Intros", emptyList)
	s.handle(get, "/Users/{uid}/Items/{id}/SpecialFeatures", s.extras(false))
	s.handle(get, "/Items/{id}/SpecialFeatures", s.extras(false))
	s.handle(get, "/Users/{uid}/Items/{id}/LocalTrailers", s.extras(true))
	s.handle(get, "/Items/{id}/LocalTrailers", s.extras(true))
	s.handle(get, "/Items/{id}/Similar", emptyList)
	s.handle(get, "/Movies/{id}/Similar", emptyList)
	s.handle(get, "/Shows/{id}/Similar", emptyList)
	s.handle(get, "/Movies/Recommendations", emptyArray)
	s.handle(get, "/Items/{id}/ThemeMedia", s.themeMedia)
	s.handle(get, "/Items/{id}/ThemeSongs", emptyList)
	s.handle(get, "/Items/{id}/ThemeVideos", emptyList)
	s.handle(get, "/Items/{id}/Ancestors", s.ancestors)
	s.handle(get, "/Items/{id}/CriticReviews", emptyList)
	s.handle(get, "/Items/{id}/ExternalIdInfos", emptyArray)
	s.handle(get, "/Items/{id}/Images", s.imagesList)
	s.handle(get, "/MediaSegments/{id}", emptyList)
	s.handle(get, "/Videos/{id}/AdditionalParts", s.additionalParts)

	// --- TV
	s.handle(get, "/Shows/NextUp", s.nextUp)
	s.handle(get, "/Shows/Upcoming", emptyList)
	s.handle(get, "/Shows/{id}/Seasons", s.seasons)
	s.handle(get, "/Shows/{id}/Episodes", s.episodes)

	// --- names, search
	s.handle(get, "/Genres", s.names("Genre", "genre:", func(it *itemT) []string { return it.Genres }))
	s.handle(get, "/Genres/{name}", s.namedByName("Genre", "genre:"))
	s.handle(get, "/Studios", s.names("Studio", "studio:", func(it *itemT) []string { return it.Studios }))
	s.handle(get, "/Studios/{name}", s.namedByName("Studio", "studio:"))
	s.handle(get, "/Persons", s.persons)
	s.handle(get, "/Persons/{name}", s.personByName)
	s.handle(get, "/Search/Hints", s.searchHints)

	// --- playback
	s.handle(getPost, "/Items/{id}/PlaybackInfo", s.playbackInfo)
	s.handle(get, "/Videos/{id}/{file}", s.stream)
	s.handle(get, "/Items/{id}/Download", s.download)
	s.handle(get, "/Items/{id}/File", s.download)
	s.handle(get, "/Videos/{id}/{msid}/Subtitles/{index}/{file}", s.subtitle)
	s.handle(get, "/Videos/{id}/{msid}/Subtitles/{index}/{start}/{file}", s.subtitle)
	s.handle(post, "/Sessions/Playing", s.playing("start"))
	s.handle(post, "/Sessions/Playing/Progress", s.playing("progress"))
	s.handle(post, "/Sessions/Playing/Stopped", s.playing("stopped"))
	s.handle(post, "/Sessions/Playing/Ping", func(c *ctx) { c.noContent() })
	s.handle(post, "/Users/{uid}/PlayingItems/{id}", s.playing("start"))
	s.handle(post, "/Users/{uid}/PlayingItems/{id}/Progress", s.playing("progress"))
	s.handle(del, "/Users/{uid}/PlayingItems/{id}", s.playing("stopped"))
	s.handle(post, "/PlayingItems/{id}", s.playing("start"))
	s.handle(post, "/PlayingItems/{id}/Progress", s.playing("progress"))
	s.handle(del, "/PlayingItems/{id}", s.playing("stopped"))

	// --- watch state
	s.handle(post, "/Users/{uid}/PlayedItems/{id}", s.markPlayed(true))
	s.handle(del, "/Users/{uid}/PlayedItems/{id}", s.markPlayed(false))
	s.handle(post, "/UserPlayedItems/{id}", s.markPlayed(true))
	s.handle(del, "/UserPlayedItems/{id}", s.markPlayed(false))
	s.handle(post, "/Users/{uid}/FavoriteItems/{id}", s.markFavorite(true))
	s.handle(del, "/Users/{uid}/FavoriteItems/{id}", s.markFavorite(false))
	s.handle(post, "/UserFavoriteItems/{id}", s.markFavorite(true))
	s.handle(del, "/UserFavoriteItems/{id}", s.markFavorite(false))
	s.handle(get, "/Users/{uid}/Items/{id}/UserData", s.getUserData)
	s.handle(get, "/UserItems/{id}/UserData", s.getUserData)
	s.handle(post, "/Users/{uid}/Items/{id}/UserData", s.updateUserData)
	s.handle(post, "/UserItems/{id}/UserData", s.updateUserData)

	// Not supported: say so cleanly instead of 404-ing on every poll.
	s.handle(get, "/LiveTv/Info", func(c *ctx) { c.json(map[string]any{"Services": []any{}, "IsEnabled": false, "EnabledUsers": []any{}}) })
	s.handle(get, "/Playlists/{id}/Items", emptyList)
	s.handle(get, "/socket", func(c *ctx) { http.Error(c.w, "websocket not supported", http.StatusNotImplemented) })
}

type itemT = library.Item
