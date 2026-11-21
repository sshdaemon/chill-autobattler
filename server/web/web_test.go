package web

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func site(t *testing.T) (http.Handler, string) {
	dir := t.TempDir()
	write := func(p, body string) {
		f := filepath.Join(dir, filepath.FromSlash(p))
		os.MkdirAll(filepath.Dir(f), 0o755)
		if err := os.WriteFile(f, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("index.html", "<!doctype html><p>chill</p>")
	write("sw.js", "self.x = 1")
	write("assets/App-AbCd1234.js", "console.log('app')")
	wasm := string(bytes.Repeat([]byte("wasm"), 4096))
	write("sim.wasm", wasm)
	write("assets/nature/ground.jpg", "jpeg bytes")
	sum := sha256.Sum256([]byte(wasm))
	return Handler(dir), hex.EncodeToString(sum[:])[:10]
}

func get(h http.Handler, url string, gzipOK bool) *http.Response {
	r := httptest.NewRequest(http.MethodGet, url, nil)
	if gzipOK {
		r.Header.Set("Accept-Encoding", "gzip, br")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Result()
}

func TestCacheHeaders(t *testing.T) {
	h, wasmHash := site(t)
	for _, c := range []struct{ url, want string }{
		{"/", revalidate},
		{"/index.html", revalidate},
		{"/sw.js", revalidate},
		{"/assets/App-AbCd1234.js", immutable},
		{"/sim.wasm?v=" + wasmHash, immutable},
		{"/sim.wasm?v=0123456789", short}, // not this file's hash: don't pin it
		{"/sim.wasm", short},
		{"/assets/nature/ground.jpg", short},
	} {
		res := get(h, c.url, false)
		if res.StatusCode != http.StatusOK {
			t.Errorf("%s: status %d", c.url, res.StatusCode)
		}
		if got := res.Header.Get("Cache-Control"); got != c.want {
			t.Errorf("%s: Cache-Control %q, want %q", c.url, got, c.want)
		}
	}
	if res := get(h, "/nope.js", false); res.StatusCode != http.StatusNotFound {
		t.Errorf("missing file: %d", res.StatusCode)
	}
	post := httptest.NewRecorder()
	h.ServeHTTP(post, httptest.NewRequest(http.MethodPost, "/index.html", nil))
	if post.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST: %d", post.Code)
	}
	if res := get(h, "/../web.go", false); res.StatusCode != http.StatusNotFound {
		t.Errorf("path out of the directory: %d", res.StatusCode)
	}
}

func TestGzip(t *testing.T) {
	h, wasmHash := site(t)
	res := get(h, "/sim.wasm?v="+wasmHash, true)
	if res.Header.Get("Content-Encoding") != "gzip" || res.Header.Get("Content-Type") != "application/wasm" {
		t.Fatalf("wasm with gzip accepted: %v", res.Header)
	}
	zr, err := gzip.NewReader(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(zr)
	if len(body) != 4*4096 {
		t.Errorf("gunzipped %d bytes, want %d", len(body), 4*4096)
	}
	if res := get(h, "/sim.wasm", false); res.Header.Get("Content-Encoding") != "" {
		t.Error("gzip sent to a client that didn't ask")
	}
	if res := get(h, "/assets/nature/ground.jpg", true); res.Header.Get("Content-Encoding") != "" {
		t.Error("a JPEG was gzipped")
	}
	if res := get(h, "/", true); res.Header.Get("Content-Encoding") != "gzip" {
		t.Error("index.html not gzipped")
	}
}
