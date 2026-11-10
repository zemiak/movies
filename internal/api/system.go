package api

import (
	"crypto/subtle"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/zemiak/movies/internal/library"
	"github.com/zemiak/movies/internal/store"
)

func urlUnescape(s string) (string, error) { return url.QueryUnescape(s) }

func (s *Server) userID(name string) string { return library.ID("user:" + name) }

func (s *Server) publicInfo(c *ctx) map[string]any {
	return map[string]any{
		"LocalAddress":           c.baseURL(),
		"ServerName":             s.cfg.ServerName,
		"Version":                Version,
		"ProductName":            "Jellyfin Server",
		"OperatingSystem":        "Linux",
		"Id":                     s.cfg.ServerID,
		"StartupWizardCompleted": true,
	}
}

func (s *Server) systemInfo(c *ctx) {
	info := s.publicInfo(c)
	for k, v := range map[string]any{
		"OperatingSystemDisplayName": "Linux",
		"PackageName":                "movies",
		"HasPendingRestart":          false,
		"IsShuttingDown":             false,
		"SupportsLibraryMonitor":     false,
		"WebSocketPortNumber":        8096,
		"CompletedInstallations":     []any{},
		"CanSelfRestart":             false,
		"CanLaunchWebBrowser":        false,
		"ProgramDataPath":            "/data",
		"WebPath":                    "",
		"ItemsByNamePath":            "",
		"CachePath":                  s.cfg.CacheDir,
		"LogPath":                    "",
		"InternalMetadataPath":       "",
		"TranscodingTempPath":        "",
		"CastReceiverApplications":   []any{},
		"HasUpdateAvailable":         false,
		"EncoderLocation":            "System",
		"SystemArchitecture":         strings.ToUpper(runtime.GOARCH),
	} {
		info[k] = v
	}
	c.json(info)
}

func (s *Server) userDTO(name string) map[string]any {
	return map[string]any{
		"Name":                      name,
		"ServerId":                  s.cfg.ServerID,
		"Id":                        s.userID(name),
		"HasPassword":               !s.cfg.open,
		"HasConfiguredPassword":     !s.cfg.open,
		"HasConfiguredEasyPassword": false,
		"EnableAutoLogin":           false,
		"LastLoginDate":             jfTime(time.Now()),
		"LastActivityDate":          jfTime(time.Now()),
		"Configuration": map[string]any{
			"PlayDefaultAudioTrack":      true,
			"SubtitleLanguagePreference": "",
			"DisplayMissingEpisodes":     false,
			"GroupedFolders":             []any{},
			"SubtitleMode":               "Default",
			"DisplayCollectionsView":     false,
			"EnableLocalPassword":        false,
			"OrderedViews":               []any{},
			"LatestItemsExcludes":        []any{},
			"MyMediaExcludes":            []any{},
			"HidePlayedInLatest":         true,
			"RememberAudioSelections":    true,
			"RememberSubtitleSelections": true,
			"EnableNextEpisodeAutoPlay":  true,
			"CastReceiverId":             "",
		},
		"Policy": map[string]any{
			"IsAdministrator":                  true,
			"IsHidden":                         true,
			"EnableCollectionManagement":       false,
			"EnableSubtitleManagement":         false,
			"EnableLyricManagement":            false,
			"IsDisabled":                       false,
			"BlockedTags":                      []any{},
			"AllowedTags":                      []any{},
			"EnableUserPreferenceAccess":       true,
			"AccessSchedules":                  []any{},
			"BlockUnratedItems":                []any{},
			"EnableRemoteControlOfOtherUsers":  false,
			"EnableSharedDeviceControl":        false,
			"EnableRemoteAccess":               true,
			"EnableLiveTvManagement":           false,
			"EnableLiveTvAccess":               false,
			"EnableMediaPlayback":              true,
			"EnableAudioPlaybackTranscoding":   false,
			"EnableVideoPlaybackTranscoding":   false,
			"EnablePlaybackRemuxing":           false,
			"ForceRemoteSourceTranscoding":     false,
			"EnableContentDeletion":            false,
			"EnableContentDeletionFromFolders": []any{},
			"EnableContentDownloading":         true,
			"EnableSyncTranscoding":            false,
			"EnableMediaConversion":            false,
			"EnabledDevices":                   []any{},
			"EnableAllDevices":                 true,
			"EnabledChannels":                  []any{},
			"EnableAllChannels":                true,
			"EnabledFolders":                   []any{},
			"EnableAllFolders":                 true,
			"InvalidLoginAttemptCount":         0,
			"LoginAttemptsBeforeLockout":       -1,
			"MaxActiveSessions":                0,
			"EnablePublicSharing":              false,
			"BlockedMediaFolders":              []any{},
			"BlockedChannels":                  []any{},
			"RemoteClientBitrateLimit":         0,
			"AuthenticationProviderId":         "Jellyfin.Server.Implementations.Users.DefaultAuthenticationProvider",
			"PasswordResetProviderId":          "Jellyfin.Server.Implementations.Users.DefaultPasswordResetProvider",
			"SyncPlayAccess":                   "None",
		},
	}
}

func (s *Server) sessionInfo(c *ctx, user string) map[string]any {
	return map[string]any{
		"PlayState":                map[string]any{"CanSeek": false, "IsPaused": false, "IsMuted": false, "RepeatMode": "RepeatNone"},
		"AdditionalUsers":          []any{},
		"Capabilities":             map[string]any{"PlayableMediaTypes": []any{}, "SupportedCommands": []any{}, "SupportsMediaControl": false},
		"RemoteEndPoint":           c.r.RemoteAddr,
		"PlayableMediaTypes":       []any{},
		"Id":                       library.ID("session:" + c.auth["deviceid"] + user),
		"UserId":                   s.userID(user),
		"UserName":                 user,
		"Client":                   c.auth["client"],
		"LastActivityDate":         jfTime(time.Now()),
		"LastPlaybackCheckIn":      jfTime(time.Time{}.Add(time.Second)),
		"DeviceName":               c.auth["device"],
		"DeviceId":                 c.auth["deviceid"],
		"ApplicationVersion":       c.auth["version"],
		"IsActive":                 true,
		"SupportsMediaControl":     false,
		"SupportsRemoteControl":    false,
		"NowPlayingQueue":          []any{},
		"NowPlayingQueueFullItems": []any{},
		"HasCustomDeviceName":      false,
		"ServerId":                 s.cfg.ServerID,
		"SupportedCommands":        []any{},
	}
}

func (s *Server) login(c *ctx) {
	var req struct {
		Username string
		Pw       string
		Password string
	}
	if !c.body(&req) {
		return
	}
	pw := req.Pw
	if pw == "" {
		pw = req.Password
	}
	var user string
	if s.cfg.open {
		user = OpenUser // any name and password
	}
	for name, want := range s.cfg.Users {
		if s.cfg.open {
			break
		}
		if strings.EqualFold(name, strings.TrimSpace(req.Username)) &&
			subtle.ConstantTimeCompare([]byte(want), []byte(pw)) == 1 {
			user = name
		}
	}
	if user == "" {
		time.Sleep(500 * time.Millisecond)
		http.Error(c.w, "Invalid username or password entered.", http.StatusUnauthorized)
		return
	}
	token := s.store.NewToken(store.Token{User: user, Client: c.auth["client"], Device: c.auth["device"],
		DeviceID: c.auth["deviceid"], Version: c.auth["version"]})
	c.user = user
	c.json(map[string]any{
		"User":        s.userDTO(user),
		"SessionInfo": s.sessionInfo(c, user),
		"AccessToken": token,
		"ServerId":    s.cfg.ServerID,
	})
}

func (s *Server) displayPreferences(c *ctx) {
	c.json(map[string]any{
		"Id":                 c.param("id"),
		"SortBy":             "SortName",
		"RememberIndexing":   false,
		"PrimaryImageHeight": 250,
		"PrimaryImageWidth":  250,
		"CustomPrefs":        map[string]any{},
		"ScrollDirection":    "Horizontal",
		"ShowBackdrop":       true,
		"RememberSorting":    false,
		"SortOrder":          "Ascending",
		"ShowSidebar":        false,
		"Client":             c.q["client"],
	})
}

// ---------------------------------------------------------------- views

func (s *Server) viewDTO(c *ctx, v *library.View) map[string]any {
	unplayed := 0
	for _, it := range v.Children {
		for _, p := range playables(it) {
			if !s.store.Get(c.user, p.Key).Played {
				unplayed++
			}
		}
	}
	return map[string]any{
		"Name":                     v.Name,
		"ServerId":                 s.cfg.ServerID,
		"Id":                       v.ID,
		"Etag":                     v.ID[:16],
		"DateCreated":              jfTime(v.Created),
		"CanDelete":                false,
		"CanDownload":              false,
		"SortName":                 library.SortName(v.Name),
		"ExternalUrls":             []any{},
		"Path":                     v.Path,
		"EnableMediaSourceDisplay": true,
		"ChannelId":                nil,
		"Taglines":                 []any{},
		"Genres":                   []any{},
		"PlayAccess":               "Full",
		"RemoteTrailers":           []any{},
		"ProviderIds":              map[string]any{},
		"IsFolder":                 true,
		"ParentId":                 rootID,
		"Type":                     "CollectionFolder",
		"People":                   []any{},
		"Studios":                  []any{},
		"GenreItems":               []any{},
		"LocalTrailerCount":        0,
		"UserData": map[string]any{"PlaybackPositionTicks": 0, "PlayCount": 0, "IsFavorite": false,
			"Played": unplayed == 0, "Key": v.ID, "ItemId": v.ID, "UnplayedItemCount": unplayed},
		"ChildCount":           len(v.Children),
		"SpecialFeatureCount":  0,
		"DisplayPreferencesId": v.ID,
		"Tags":                 []any{},
		"CollectionType":       v.CollectionType,
		"ImageTags":            map[string]any{},
		"BackdropImageTags":    []any{},
		"ImageBlurHashes":      map[string]any{},
		"LocationType":         "FileSystem",
		"MediaType":            "Unknown",
		"LockedFields":         []any{},
		"LockData":             false,
	}
}

var rootID = library.ID("root")

func (s *Server) rootDTO(c *ctx) map[string]any {
	return map[string]any{
		"Name":              "Media Folders",
		"ServerId":          s.cfg.ServerID,
		"Id":                rootID,
		"DateCreated":       jfTime(s.start),
		"SortName":          "media folders",
		"IsFolder":          true,
		"Type":              "UserRootFolder",
		"ChildCount":        len(c.lib.Views),
		"ImageTags":         map[string]any{},
		"BackdropImageTags": []any{},
		"LocationType":      "FileSystem",
		"MediaType":         "Unknown",
		"UserData":          map[string]any{"PlaybackPositionTicks": 0, "PlayCount": 0, "IsFavorite": false, "Played": false, "Key": rootID, "ItemId": rootID},
	}
}

func (s *Server) views(c *ctx) {
	out := []any{}
	for _, v := range c.lib.Views {
		out = append(out, s.viewDTO(c, v))
	}
	c.json(wrap(out, len(out), 0))
}

func (s *Server) virtualFolders(c *ctx) {
	out := []any{}
	for _, v := range c.lib.Views {
		out = append(out, map[string]any{"Name": v.Name, "Locations": []string{v.Path}, "CollectionType": v.CollectionType,
			"ItemId": v.ID, "RefreshStatus": "Idle", "LibraryOptions": map[string]any{}})
	}
	c.json(out)
}

func wrap(items []any, total, start int) map[string]any {
	if items == nil {
		items = []any{}
	}
	return map[string]any{"Items": items, "TotalRecordCount": total, "StartIndex": start}
}

func emptyList(c *ctx)  { c.json(wrap(nil, 0, 0)) }
func emptyArray(c *ctx) { c.json([]any{}) }

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
