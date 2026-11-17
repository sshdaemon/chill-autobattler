package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"chill/server/store"
)

func TestMiddleware(t *testing.T) {
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "a.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	token, hash := NewToken()
	if err := db.CreatePlayer(context.Background(), store.CreatePlayerParams{ID: "p1", TokenHash: hash, CreatedAt: 1}); err != nil {
		t.Fatal(err)
	}
	var seen string
	h := Middleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = PlayerID(r.Context()) }))
	for header, want := range map[string]string{"Bearer " + token: "p1", "Bearer nope": "", "": "", token: ""} {
		req := httptest.NewRequest("POST", "/x", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		seen = "unset"
		h.ServeHTTP(httptest.NewRecorder(), req)
		if seen != want {
			t.Errorf("Authorization %q: player %q, want %q", header, seen, want)
		}
	}
	if string(Hash(token)) == token {
		t.Error("the token must be stored hashed")
	}
}
