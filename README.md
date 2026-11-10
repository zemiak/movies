# movies

A small Jellyfin-compatible media server for [Infuse](https://firecore.com/infuse). It serves
Kodi-style folders (video files, tinyMediaManager `.nfo` files and local artwork) over the parts of
the Jellyfin API that Infuse uses: browsing, metadata, artwork, direct play and watch-state sync.

It does not scrape metadata, transcode, or provide a web UI. It is a single Go binary;
`ffprobe` is optional.

## What it reads

```
Movies/
  300/
    300.m4v                 movie
    300 - Cesky.m4v         second version of the same movie ("Name - Label")
    300.nfo                 metadata (title, plot, ratings, cast, ids, ...)
    300-poster.jpg  300-fanart.jpg  300-clearlogo.png  300-landscape.jpg  300-banner.jpg ...
    300.cs.srt              external subtitles (<name>.<lang>[.forced].srt)
    extras/*.m4v            special features (also featurettes/, behind the scenes/, trailers/ ...)
  The Two Towers/
    The Two Towers - Disc1.m4v, - Disc2.m4v   one movie in two parts
TVShows/
  Firefly/
    tvshow.nfo  poster.jpg  fanart.jpg  season01-poster.jpg  season-specials-poster.jpg
    Season 01/
      Episode S01E01_Serenity.m4v  (+ .nfo, -thumb.jpg)
      extras/*.m4v
```

Episode numbers come from the episode NFO, then from `SxxEyy` / `1x02` in the file name. Season
numbers come from the `Season NN` folder. Folder and file names are the fallback when there is no
NFO.

## Running

```sh
devbox run build        # build/movies
USERS=admin:secret \
LIBRARIES="Movies:movies:/data/movies;TV Shows:tvshows:/data/tvshows" \
DATA_DIR=./data build/movies
```

| Variable | Default | |
|---|---|---|
| `USERS` | (required) | `name:password,...` (or `USERS_FILE`, one per line) |
| `LIBRARIES` | `Movies:movies:/data/movies;TV Shows:tvshows:/data/tvshows` | `Name:movies\|tvshows:/path;...` |
| `DATA_DIR` | `/config` | `state.json` (watch state, tokens), `probe-cache.json`, `images/` |
| `PORT` | `8096` | |
| `SERVER_NAME` | `Movies` | the name shown in Infuse |
| `SERVER_ID` | derived from the name | |
| `PUBLISHED_URL` | from the request | |
| `SCAN_INTERVAL` | `30m` | full rescan (it takes under a second for ~12k files) |
| `LOG_LEVEL` | `info` | `debug` also logs image and progress requests |
| `IMPORT_PATH_MAP` | | `from=to,...` path prefix rewrites for imported Jellyfin paths |

In Infuse, add a **Jellyfin** share with the server address, port and a user from `USERS`.

Deployment on k3s: `apps/movies` in the media-rpi repo.

## Migrating watch state from Jellyfin

`tools/jellyfin-export.sh` reads Jellyfin's `library.db` and `jellyfin.db` and prints JSON lines
with played state, play count, resume position, favorite and last played date, keyed by file path.
It reads either a directory with copies of the databases or, with `--kubectl <namespace>`, the
running pod. Put the output into `$DATA_DIR/import/*.jsonl` and start the server: the file is
imported and renamed to `*.done`.

## Behaviour

- **Item ids** are hashes of the file paths, so they survive restarts and rescans. Renaming a file
  makes it a new item.
- **Playback**: `PlaybackInfo` offers each file as a direct-play source. `/Videos/{id}/stream` serves
  the file with HTTP Range support. Stream details (codecs, audio languages, resolution, HDR) come
  from `ffprobe`, cached per file, or from the NFO's `<fileinfo>` until `ffprobe` has run.
- **Watch state** follows Jellyfin's rules. When playback stops past 90%, the item is played. Before
  5%, the resume position is cleared. In between, the position is kept. Videos under 5 minutes count
  as played once they are stopped.
- **Images** are resized on request (`maxWidth`, `fillHeight`, ...) and cached on disk. Actor photos
  redirect to the TMDB URLs in the NFO files.
- Unknown API calls get a 404 and are logged at WARN (`known=false`). Check those first if a client
  misbehaves.

## Development

```sh
devbox run test     # go vet + go test
```
