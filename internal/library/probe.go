package library

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// Probe is the subset of `ffprobe -show_format -show_streams` output we use.
type Probe struct {
	FromNFO bool `json:"fromNfo,omitempty"`
	Format  struct {
		FormatName string `json:"format_name"`
		Duration   string `json:"duration"`
		BitRate    string `json:"bit_rate"`
	} `json:"format"`
	Streams []ProbeStream `json:"streams"`
}

// ProbeStream is one ffprobe stream.
type ProbeStream struct {
	Index              int               `json:"index"`
	CodecName          string            `json:"codec_name"`
	CodecType          string            `json:"codec_type"`
	CodecTag           string            `json:"codec_tag_string,omitempty"`
	Profile            string            `json:"profile,omitempty"`
	Level              int               `json:"level,omitempty"`
	Width              int               `json:"width,omitempty"`
	Height             int               `json:"height,omitempty"`
	DisplayAspectRatio string            `json:"display_aspect_ratio,omitempty"`
	PixFmt             string            `json:"pix_fmt,omitempty"`
	RFrameRate         string            `json:"r_frame_rate,omitempty"`
	AvgFrameRate       string            `json:"avg_frame_rate,omitempty"`
	TimeBase           string            `json:"time_base,omitempty"`
	BitRate            string            `json:"bit_rate,omitempty"`
	BitsPerRawSample   string            `json:"bits_per_raw_sample,omitempty"`
	Channels           int               `json:"channels,omitempty"`
	ChannelLayout      string            `json:"channel_layout,omitempty"`
	SampleRate         string            `json:"sample_rate,omitempty"`
	FieldOrder         string            `json:"field_order,omitempty"`
	ColorTransfer      string            `json:"color_transfer,omitempty"`
	ColorPrimaries     string            `json:"color_primaries,omitempty"`
	ColorSpace         string            `json:"color_space,omitempty"`
	Disposition        map[string]int    `json:"disposition,omitempty"`
	Tags               map[string]string `json:"tags,omitempty"`
	SideData           []map[string]any  `json:"side_data_list,omitempty"`
}

// DurationTicks is the duration in 100ns ticks.
func (p *Probe) DurationTicks() int64 {
	f, err := strconv.ParseFloat(p.Format.Duration, 64)
	if err != nil {
		return 0
	}
	return int64(f * 1e7)
}

// Bitrate is the overall bitrate in bits/s.
func (p *Probe) Bitrate() int {
	b, _ := strconv.Atoi(p.Format.BitRate)
	return b
}

type probeEntry struct {
	Size  int64  `json:"size"`
	Mod   int64  `json:"mod"`
	Probe *Probe `json:"probe"` // nil: ffprobe failed (not retried until the file changes)
}

// ProbeCache runs ffprobe on media files and remembers the results on disk.
type ProbeCache struct {
	path    string
	ffprobe string // empty: ffprobe not installed
	mu      sync.Mutex
	entries map[string]probeEntry
	dirty   bool
}

// NewProbeCache loads the cache file (if any) and finds ffprobe.
func NewProbeCache(path string) *ProbeCache {
	c := &ProbeCache{path: path, entries: map[string]probeEntry{}}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &c.entries); err != nil {
			slog.Warn("probe cache unreadable, starting empty", "err", err)
		}
	}
	if p, err := exec.LookPath("ffprobe"); err == nil {
		c.ffprobe = p
	} else {
		slog.Warn("ffprobe not found: stream details only from NFO files")
	}
	return c
}

// Get returns a cached probe for the file, and whether the cache knows this file version.
func (c *ProbeCache) Get(path string, size int64, mod time.Time) (*Probe, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[path]
	if !ok || e.Size != size || e.Mod != mod.Unix() {
		return nil, false
	}
	return e.Probe, true
}

// Enabled reports whether ffprobe is available.
func (c *ProbeCache) Enabled() bool { return c.ffprobe != "" }

// Run probes a file and stores the result.
func (c *ProbeCache) Run(ctx context.Context, path string, size int64, mod time.Time) *Probe {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	out, err := exec.CommandContext(ctx, c.ffprobe, "-v", "quiet", "-print_format", "json",
		"-show_format", "-show_streams", path).Output()
	var p *Probe
	if err == nil {
		p = &Probe{}
		if json.Unmarshal(out, p) != nil {
			p = nil
		}
	}
	if p == nil {
		if ctx.Err() != nil && ctx.Err() != context.DeadlineExceeded {
			return nil // shutting down: don't remember as failed
		}
		slog.Warn("ffprobe failed", "path", path, "err", err)
	}
	c.mu.Lock()
	c.entries[path] = probeEntry{Size: size, Mod: mod.Unix(), Probe: p}
	c.dirty = true
	c.mu.Unlock()
	return p
}

// Prune drops entries for files no longer present.
func (c *ProbeCache) Prune(present map[string]bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for p := range c.entries {
		if !present[p] {
			delete(c.entries, p)
			c.dirty = true
		}
	}
}

// Save writes the cache if it changed.
func (c *ProbeCache) Save() error {
	c.mu.Lock()
	if !c.dirty {
		c.mu.Unlock()
		return nil
	}
	data, err := json.Marshal(c.entries)
	c.dirty = false
	c.mu.Unlock()
	if err != nil {
		return err
	}
	return writeFileAtomic(c.path, data)
}

func writeFileAtomic(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
