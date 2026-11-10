package library

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Manager keeps the current library snapshot, rescans periodically or on demand, and runs
// ffprobe on new files in the background.
type Manager struct {
	scanner  *Scanner
	lib      atomic.Pointer[Library]
	trigger  chan struct{}
	Interval time.Duration
	Workers  int
}

// NewManager scans once, synchronously, so the server starts with a full library.
func NewManager(s *Scanner, interval time.Duration) (*Manager, []PendingProbe) {
	m := &Manager{scanner: s, trigger: make(chan struct{}, 1), Interval: interval, Workers: 2}
	lib, pending := s.Scan()
	m.lib.Store(lib)
	return m, pending
}

// Get returns the current snapshot (never nil).
func (m *Manager) Get() *Library { return m.lib.Load() }

// Rescan asks for a new scan (non-blocking).
func (m *Manager) Rescan() {
	select {
	case m.trigger <- struct{}{}:
	default:
	}
}

// Run is the background loop; it returns when ctx is done.
func (m *Manager) Run(ctx context.Context, pending []PendingProbe) {
	ticker := time.NewTicker(m.Interval)
	defer ticker.Stop()
	for {
		if len(pending) > 0 && m.scanner.Probes != nil && m.scanner.Probes.Enabled() {
			m.probe(ctx, pending)
			if ctx.Err() != nil {
				return
			}
			lib, _ := m.scanner.Scan()
			m.lib.Store(lib)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-m.trigger:
		}
		var lib *Library
		lib, pending = m.scanner.Scan()
		m.lib.Store(lib)
		if m.scanner.Probes != nil {
			if err := m.scanner.Probes.Save(); err != nil {
				slog.Error("saving probe cache", "err", err)
			}
		}
	}
}

func (m *Manager) probe(ctx context.Context, pending []PendingProbe) {
	start := time.Now()
	slog.Info("probing media files", "count", len(pending))
	jobs := make(chan PendingProbe)
	var wg sync.WaitGroup
	var done atomic.Int64
	for i := 0; i < m.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				m.scanner.Probes.Run(ctx, p.Path, p.Size, p.ModTime)
				if n := done.Add(1); n%100 == 0 {
					slog.Info("probing media files", "done", n, "of", len(pending))
					if err := m.scanner.Probes.Save(); err != nil {
						slog.Error("saving probe cache", "err", err)
					}
				}
			}
		}()
	}
loop:
	for _, p := range pending {
		select {
		case jobs <- p:
		case <-ctx.Done():
			break loop
		}
	}
	close(jobs)
	wg.Wait()
	if err := m.scanner.Probes.Save(); err != nil {
		slog.Error("saving probe cache", "err", err)
	}
	slog.Info("probing finished", "count", done.Load(), "took", time.Since(start).Round(time.Second))
}
