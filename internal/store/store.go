// Package store keeps the server's mutable state: access tokens and per-user watch state
// (played, resume position, favorites). It is a JSON file, written shortly after each change.
package store

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// UserData is one user's state for one item, keyed by the item's Key.
type UserData struct {
	Played        bool       `json:"played,omitempty"`
	PlayCount     int        `json:"playCount,omitempty"`
	IsFavorite    bool       `json:"favorite,omitempty"`
	PositionTicks int64      `json:"position,omitempty"`
	LastPlayed    *time.Time `json:"lastPlayed,omitempty"`
	Updated       time.Time  `json:"updated"`
}

// Token is an access token handed out at login.
type Token struct {
	User     string    `json:"user"`
	Client   string    `json:"client,omitempty"`
	Device   string    `json:"device,omitempty"`
	DeviceID string    `json:"deviceId,omitempty"`
	Version  string    `json:"version,omitempty"`
	Created  time.Time `json:"created"`
	LastSeen time.Time `json:"lastSeen"`
}

type state struct {
	Tokens   map[string]*Token               `json:"tokens"`
	UserData map[string]map[string]*UserData `json:"userData"` // user -> item key -> data
}

// Store is safe for concurrent use.
type Store struct {
	path  string
	mu    sync.RWMutex
	st    state
	dirty chan struct{}
	done  chan struct{}
}

// Open loads the state file (a missing file is an empty state).
func Open(path string) (*Store, error) {
	s := &Store{path: path, dirty: make(chan struct{}, 1), done: make(chan struct{}),
		st: state{Tokens: map[string]*Token{}, UserData: map[string]map[string]*UserData{}}}
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, &s.st); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if s.st.Tokens == nil {
			s.st.Tokens = map[string]*Token{}
		}
		if s.st.UserData == nil {
			s.st.UserData = map[string]map[string]*UserData{}
		}
	case os.IsNotExist(err):
	default:
		return nil, err
	}
	go s.writer()
	return s, nil
}

func (s *Store) changed() {
	select {
	case s.dirty <- struct{}{}:
	default:
	}
}

// writer saves at most every 2 seconds.
func (s *Store) writer() {
	for {
		select {
		case <-s.dirty:
			time.Sleep(2 * time.Second)
			if err := s.Save(); err != nil {
				slog.Error("saving state", "err", err)
			}
		case <-s.done:
			return
		}
	}
}

// Close writes pending changes.
func (s *Store) Close() error {
	close(s.done)
	return s.Save()
}

// Save writes the state file now.
func (s *Store) Save() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(&s.st, "", " ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// ---------------------------------------------------------------- tokens

// NewToken creates and remembers a token.
func (s *Store) NewToken(t Token) string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	id := hex.EncodeToString(b)
	t.Created, t.LastSeen = time.Now(), time.Now()
	s.mu.Lock()
	// One token per user and device: logging in again replaces the old one.
	for k, old := range s.st.Tokens {
		if old.User == t.User && t.DeviceID != "" && old.DeviceID == t.DeviceID {
			delete(s.st.Tokens, k)
		}
	}
	s.st.Tokens[id] = &t
	s.mu.Unlock()
	s.changed()
	return id
}

// Token looks up a token; it also records when it was last used (saved lazily).
func (s *Store) Token(id string) (Token, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.st.Tokens[id]
	if !ok {
		return Token{}, false
	}
	if time.Since(t.LastSeen) > time.Hour {
		t.LastSeen = time.Now()
		s.changed()
	}
	return *t, true
}

// DeleteToken logs a token out.
func (s *Store) DeleteToken(id string) {
	s.mu.Lock()
	delete(s.st.Tokens, id)
	s.mu.Unlock()
	s.changed()
}

// ---------------------------------------------------------------- user data

// Get returns a copy of the user's data for an item (zero value if none).
func (s *Store) Get(user, key string) UserData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if d := s.st.UserData[user][key]; d != nil {
		return *d
	}
	return UserData{}
}

// Update changes the user's data for an item and returns the result.
func (s *Store) Update(user, key string, fn func(*UserData)) UserData {
	s.mu.Lock()
	m := s.st.UserData[user]
	if m == nil {
		m = map[string]*UserData{}
		s.st.UserData[user] = m
	}
	d := m[key]
	if d == nil {
		d = &UserData{}
	}
	fn(d)
	// Kept even when everything is cleared: Updated tells syncing clients (MinDateLastSavedForUser)
	// that e.g. "unplayed" changed.
	d.Updated = time.Now()
	m[key] = d
	out := *d
	s.mu.Unlock()
	s.changed()
	return out
}

// ---------------------------------------------------------------- import

// ImportRecord is one line of the Jellyfin export (tools/jellyfin-export.sh).
type ImportRecord struct {
	User          string     `json:"user"`
	Path          string     `json:"path"`
	Played        bool       `json:"played"`
	PlayCount     int        `json:"playCount"`
	IsFavorite    bool       `json:"favorite"`
	PositionTicks int64      `json:"position"`
	LastPlayed    *time.Time `json:"lastPlayed"`
}

// Import merges a JSON-lines export. keyFor maps a Jellyfin file path to an item key ("" = skip);
// userFor maps a Jellyfin user name to ours ("" = skip). Existing state wins where it is newer.
func (s *Store) Import(path string, keyFor func(string) string, userFor func(string) string) (imported, skipped int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var r ImportRecord
		if len(sc.Bytes()) == 0 {
			continue
		}
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return imported, skipped, fmt.Errorf("%s: %w", path, err)
		}
		key, user := keyFor(r.Path), userFor(r.User)
		if key == "" || user == "" {
			skipped++
			continue
		}
		s.Update(user, key, func(d *UserData) {
			if d.LastPlayed != nil && r.LastPlayed != nil && d.LastPlayed.After(*r.LastPlayed) {
				return
			}
			d.Played = d.Played || r.Played
			if r.PlayCount > d.PlayCount {
				d.PlayCount = r.PlayCount
			}
			d.IsFavorite = d.IsFavorite || r.IsFavorite
			if r.PositionTicks > 0 && !r.Played {
				d.PositionTicks = r.PositionTicks
			}
			if r.LastPlayed != nil {
				t := *r.LastPlayed
				d.LastPlayed = &t
			}
		})
		imported++
	}
	return imported, skipped, sc.Err()
}
