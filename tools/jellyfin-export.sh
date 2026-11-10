#!/usr/bin/env bash
# Exports watch state (played, play count, resume position, favorites, last played) from a
# Jellyfin 10.10 server as JSON lines, for the movies server to import.
#
#   tools/jellyfin-export.sh <jellyfin-data-dir> > jellyfin-watch-state.jsonl
#       reads <dir>/library.db and <dir>/jellyfin.db (copies, not a running server's files)
#
#   tools/jellyfin-export.sh --kubectl <namespace> > jellyfin-watch-state.jsonl
#       copies the databases out of the running Jellyfin pod first (deployment "jellyfin")
#
# Import: put the file into the movies server's $DATA_DIR/import/ and restart it, e.g.
#   kubectl -n movies cp jellyfin-watch-state.jsonl <pod>:/config/import/jellyfin.jsonl
#   kubectl -n movies rollout restart deployment/movies
#
# Paths are Jellyfin's paths (e.g. /data/movies/...); they are matched against the movies
# server's library folders (set IMPORT_PATH_MAP=from=to there if they differ).
# Requires: sqlite3 (and kubectl for --kubectl).
set -euo pipefail

if [ "${1:-}" = "--kubectl" ]; then
    ns=${2:?namespace}
    dir=$(mktemp -d)
    trap 'rm -rf "$dir"' EXIT
    pod=$(kubectl -n "$ns" get pods -l app=jellyfin -o jsonpath='{.items[0].metadata.name}')
    # tar keeps the WAL with the database; a plain copy of a live SQLite file may be incomplete.
    kubectl -n "$ns" exec "$pod" -- sh -c 'cd /config/data/data && tar czf - library.db library.db-wal library.db-shm jellyfin.db 2>/dev/null' \
        | tar xzf - -C "$dir"
else
    dir=${1:?usage: $0 <jellyfin-data-dir> | --kubectl <namespace>}
fi

[ -f "$dir/library.db" ] && [ -f "$dir/jellyfin.db" ] || { echo "library.db / jellyfin.db not found in $dir" >&2; exit 1; }

# A user data row's key is one of: the item's UserDataKey, a provider id of a movie (Tmdb/Imdb),
# or the item's GUID (stored as a .NET GUID: the first three groups byte-swapped).
sqlite3 "$dir/library.db" <<SQL
ATTACH '$dir/jellyfin.db' AS jf;
.mode list
.headers off
WITH ud AS (
    SELECT u.*, lower(
        substr(key,7,2)||substr(key,5,2)||substr(key,3,2)||substr(key,1,2)||
        substr(key,12,2)||substr(key,10,2)||substr(key,17,2)||substr(key,15,2)||
        substr(key,20,4)||substr(key,25,12)) AS guidhex
    FROM UserDatas u
),
items AS (
    SELECT guid, type, Path, UserDataKey, ProviderIds FROM TypedBaseItems
    WHERE Path LIKE '/%' AND IsVirtualItem = 0
      AND type IN ('MediaBrowser.Controller.Entities.Movies.Movie',
                   'MediaBrowser.Controller.Entities.TV.Episode',
                   'MediaBrowser.Controller.Entities.TV.Series',
                   'MediaBrowser.Controller.Entities.Video')
),
matched AS (
    SELECT ud.*, i.Path AS path FROM ud JOIN items i ON i.UserDataKey = ud.key
    UNION
    SELECT ud.*, i.Path FROM ud JOIN items i ON lower(hex(i.guid)) = ud.guidhex
    UNION
    SELECT ud.*, i.Path FROM ud JOIN items i
        ON i.type = 'MediaBrowser.Controller.Entities.Movies.Movie'
       AND ('|' || i.ProviderIds || '|') LIKE '%=' || ud.key || '|%'
),
merged AS (
    SELECT m.path, usr.Username AS user,
           max(m.played) AS played, max(m.playCount) AS playCount, max(m.isFavorite) AS favorite,
           max(m.playbackPositionTicks) AS position, max(m.lastPlayedDate) AS lastPlayed
    FROM matched m JOIN jf.Users usr ON usr.InternalId = m.userId
    GROUP BY m.path, usr.Username
)
SELECT json_object(
    'user', user, 'path', path,
    'played', json(CASE WHEN played THEN 'true' ELSE 'false' END),
    'playCount', playCount,
    'favorite', json(CASE WHEN favorite THEN 'true' ELSE 'false' END),
    'position', position,
    'lastPlayed', CASE WHEN lastPlayed IS NULL THEN NULL ELSE replace(lastPlayed, ' ', 'T') END)
FROM merged
WHERE played OR favorite OR position > 0
ORDER BY path;
SQL
