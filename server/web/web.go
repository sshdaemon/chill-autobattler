// Package web serves the built client (web/dist) next to the RPCs, with the
// cache headers the client's versioned assets rely on (openspec change
// asset-loading):
//   - a public asset asked for as `path?v=<hash>` whose hash matches the file,
//     and Vite's hashed chunks (/assets/<name>-<hash>.js|css): cached for a
//     year as immutable;
//   - index.html and sw.js: revalidated on every load;
//   - anything else: five minutes.
//
// Compressible files (WASM, JS, JSON, glTF and its buffers, SVG) are sent
// gzipped to clients that accept it, compressed once per file version.
package web

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	immutable  = "public, max-age=31536000, immutable"
	revalidate = "no-cache"
	short      = "public, max-age=300"
)

var (
	hashedChunk  = regexp.MustCompile(`^/assets/[^/]+-[A-Za-z0-9_-]{8}\.(js|css)$`)
	compressible = map[string]bool{".wasm": true, ".js": true, ".mjs": true, ".json": true, ".gltf": true, ".bin": true, ".glb": true, ".svg": true, ".css": true, ".html": true}
	types        = map[string]string{".wasm": "application/wasm", ".gltf": "model/gltf+json", ".glb": "model/gltf-binary", ".bin": "application/octet-stream", ".js": "text/javascript; charset=utf-8", ".mjs": "text/javascript; charset=utf-8"}
)

// Handler serves the files under dir.
func Handler(dir string) http.Handler {
	return &handler{dir: dir, files: http.FileServer(http.Dir(dir)), cache: map[string]*entry{}}
}

type handler struct {
	dir   string
	files http.Handler
	mu    sync.Mutex
	cache map[string]*entry // per file: its hash and gzipped bytes, for its modification time
}

type entry struct {
	mod  time.Time
	hash string // first 10 hex of sha256, as the client's manifest has it
	gz   []byte // nil until asked for, or not compressible
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	p := path.Clean("/" + r.URL.Path)
	if p == "/" {
		p = "/index.html"
	}
	file := filepath.Join(h.dir, filepath.FromSlash(p))
	st, err := os.Stat(file)
	if err != nil || st.IsDir() {
		http.NotFound(w, r)
		return
	}
	e, err := h.entry(file, st)
	if err != nil {
		http.Error(w, "read failed", http.StatusInternalServerError)
		return
	}
	switch v := r.URL.Query().Get("v"); {
	case v != "" && v == e.hash, v == "" && hashedChunk.MatchString(p):
		w.Header().Set("Cache-Control", immutable)
	case p == "/index.html" || p == "/sw.js":
		w.Header().Set("Cache-Control", revalidate)
	default:
		w.Header().Set("Cache-Control", short) // includes a ?v= that isn't this file's: don't pin it
	}
	ext := strings.ToLower(filepath.Ext(p))
	if t, ok := types[ext]; ok {
		w.Header().Set("Content-Type", t)
	} else if t := mime.TypeByExtension(ext); t != "" {
		w.Header().Set("Content-Type", t)
	}
	w.Header().Add("Vary", "Accept-Encoding")
	if compressible[ext] && r.Header.Get("Range") == "" && acceptsGzip(r) {
		gz, err := h.gzipped(file, e)
		if err == nil {
			w.Header().Set("Content-Encoding", "gzip")
			http.ServeContent(w, r, p, st.ModTime(), bytes.NewReader(gz))
			return
		}
	}
	if p == "/index.html" { // FileServer and ServeFile redirect /index.html to /
		f, err := os.Open(file)
		if err != nil {
			http.Error(w, "read failed", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		http.ServeContent(w, r, p, st.ModTime(), f)
		return
	}
	h.files.ServeHTTP(w, r)
}

func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(strings.SplitN(part, ";", 2)[0]) == "gzip" {
			return true
		}
	}
	return false
}

// entry is a file's hash, recomputed when the file changes.
func (h *handler) entry(file string, st os.FileInfo) (*entry, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if e, ok := h.cache[file]; ok && e.mod.Equal(st.ModTime()) {
		return e, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	e := &entry{mod: st.ModTime(), hash: hex.EncodeToString(sum[:])[:10]}
	h.cache[file] = e
	return e, nil
}

func (h *handler) gzipped(file string, e *entry) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if e.gz != nil {
		return e.gz, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	zw.Write(data)
	zw.Close()
	e.gz = buf.Bytes()
	return e.gz, nil
}
