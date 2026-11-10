package api

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // decoder

	"github.com/zemiak/movies/internal/library"
)

// imageCache resizes images on request and keeps the results on disk.
type imageCache struct {
	dir      string
	sem      chan struct{} // limits concurrent decodes (big fanart is ~30 MB decoded)
	mu       sync.Mutex
	inflight map[string]*sync.WaitGroup
}

func newImageCache(dir string) *imageCache {
	return &imageCache{dir: dir, sem: make(chan struct{}, 3), inflight: map[string]*sync.WaitGroup{}}
}

func normalizeImageType(t string) string {
	for _, k := range []string{library.ImgPrimary, library.ImgBackdrop, library.ImgLogo, library.ImgThumb,
		library.ImgBanner, library.ImgArt, library.ImgDisc} {
		if strings.EqualFold(k, t) {
			return k
		}
	}
	return t
}

// image handles /Items/{id}/Images/{type}[/{index}].
func (s *Server) image(c *ctx) {
	id := library.NormalizeID(c.param("id"))
	typ := normalizeImageType(c.param("type"))
	idx, _ := strconv.Atoi(c.param("index"))
	if p := c.lib.People[id]; p != nil {
		if p.Thumb == "" || typ != library.ImgPrimary {
			c.notFound()
			return
		}
		// Actor images are TMDB URLs in the NFO files; let the client fetch them.
		http.Redirect(c.w, c.r, p.Thumb, http.StatusFound)
		return
	}
	it := c.lib.Items[id]
	if it == nil || idx < 0 || idx >= len(it.Images[typ]) {
		c.notFound()
		return
	}
	img := it.Images[typ][idx]
	path := img.Path
	if resized, ok := s.images.resized(img, c); ok {
		path = resized
	}
	f, err := os.Open(path)
	if err != nil {
		c.notFound()
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		c.notFound()
		return
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		c.w.Header().Set("Content-Type", "image/png")
	case ".webp":
		c.w.Header().Set("Content-Type", "image/webp")
	default:
		c.w.Header().Set("Content-Type", "image/jpeg")
	}
	if c.q["tag"] != "" {
		c.w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		c.w.Header().Set("Cache-Control", "public, max-age=86400")
	}
	c.w.Header().Set("ETag", `"`+img.Tag+`"`)
	http.ServeContent(c.w, c.r, "", fi.ModTime(), f)
}

// imagesList handles GET /Items/{id}/Images.
func (s *Server) imagesList(c *ctx) {
	it := c.lookup(c.param("id"))
	out := []any{}
	if it != nil {
		for typ, imgs := range it.Images {
			for i, im := range imgs {
				d := map[string]any{"ImageType": typ, "ImageTag": im.Tag, "Path": im.Path}
				if typ == library.ImgBackdrop {
					d["ImageIndex"] = i
				}
				if fi, err := os.Stat(im.Path); err == nil {
					d["Size"] = fi.Size()
				}
				out = append(out, d)
			}
		}
	}
	c.json(out)
}

// target computes the output size from Jellyfin's sizing parameters; ok=false means "original".
func target(w, h int, c *ctx) (int, int, bool) {
	maxW, maxH := c.qInt("maxwidth", 0), c.qInt("maxheight", 0)
	if v := c.qInt("width", 0); v > 0 && (maxW == 0 || v < maxW) {
		maxW = v
	}
	if v := c.qInt("height", 0); v > 0 && (maxH == 0 || v < maxH) {
		maxH = v
	}
	fillW, fillH := c.qInt("fillwidth", 0), c.qInt("fillheight", 0)
	scale := 1.0
	if fillW > 0 || fillH > 0 {
		// Smallest size covering fillWidth x fillHeight.
		sw, sh := 0.0, 0.0
		if fillW > 0 {
			sw = float64(fillW) / float64(w)
		}
		if fillH > 0 {
			sh = float64(fillH) / float64(h)
		}
		scale = max(sw, sh)
	}
	if maxW > 0 {
		scale = min(scale, float64(maxW)/float64(w))
	}
	if maxH > 0 {
		scale = min(scale, float64(maxH)/float64(h))
	}
	if scale >= 0.95 {
		return w, h, false
	}
	nw, nh := max(1, int(float64(w)*scale+0.5)), max(1, int(float64(h)*scale+0.5))
	return nw, nh, true
}

// resized returns the path of a resized copy, creating it if needed.
func (ic *imageCache) resized(img library.Image, c *ctx) (string, bool) {
	if ic.dir == "" {
		return "", false
	}
	params := []string{"maxwidth", "maxheight", "width", "height", "fillwidth", "fillheight"}
	any := false
	for _, p := range params {
		if c.q[p] != "" {
			any = true
		}
	}
	if !any {
		return "", false
	}
	quality := c.qInt("quality", 90)
	if quality < 30 || quality > 100 {
		quality = 90
	}
	ext := strings.ToLower(filepath.Ext(img.Path))
	outExt := ".jpg"
	if ext == ".png" {
		outExt = ".png"
	}
	var keyParts []string
	for _, p := range params {
		keyParts = append(keyParts, c.q[p])
	}
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%s|%d", img.Path, img.Tag, strings.Join(keyParts, ","), quality)))
	key := hex.EncodeToString(sum[:])
	out := filepath.Join(ic.dir, key[:2], key+outExt)
	if _, err := os.Stat(out); err == nil {
		return out, true
	}

	// One resize per key at a time.
	ic.mu.Lock()
	if wg, ok := ic.inflight[key]; ok {
		ic.mu.Unlock()
		wg.Wait()
		_, err := os.Stat(out)
		return out, err == nil
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ic.inflight[key] = wg
	ic.mu.Unlock()
	defer func() {
		ic.mu.Lock()
		delete(ic.inflight, key)
		ic.mu.Unlock()
		wg.Done()
	}()

	ic.sem <- struct{}{}
	defer func() { <-ic.sem }()

	f, err := os.Open(img.Path)
	if err != nil {
		return "", false
	}
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		f.Close()
		return "", false
	}
	nw, nh, need := target(cfg.Width, cfg.Height, c)
	if !need {
		f.Close()
		return "", false
	}
	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		return "", false
	}
	src, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		slog.Debug("image decode failed", "path", img.Path, "err", err)
		return "", false
	}
	var dst draw.Image
	if outExt == ".png" {
		dst = image.NewNRGBA(image.Rect(0, 0, nw, nh))
	} else {
		dst = image.NewRGBA(image.Rect(0, 0, nw, nh))
	}
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return "", false
	}
	tmp, err := os.CreateTemp(filepath.Dir(out), "tmp-*")
	if err != nil {
		return "", false
	}
	if outExt == ".png" {
		err = (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(tmp, dst)
	} else {
		err = jpeg.Encode(tmp, dst, &jpeg.Options{Quality: quality})
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(tmp.Name())
		return "", false
	}
	if err := os.Rename(tmp.Name(), out); err != nil {
		os.Remove(tmp.Name())
		return "", false
	}
	return out, true
}
