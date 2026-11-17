// Package auth identifies anonymous players: a secret bearer token per
// player, stored on the server only as its SHA-256 hash.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"chill/server/store"
)

type ctxKey struct{}

// PlayerID returns the authenticated player, or "" if the request had none.
func PlayerID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// WithPlayer is for tests and internal calls.
func WithPlayer(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// NewToken returns a fresh secret token and the hash to store.
func NewToken() (token string, hash []byte) {
	var b [32]byte
	rand.Read(b[:])
	token = hex.EncodeToString(b[:])
	return token, Hash(token)
}

func Hash(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

// Middleware resolves "Authorization: Bearer <token>" to a player and puts it
// in the context. Unknown tokens are dropped (not rejected) here; RPCs that
// need a player answer Unauthenticated via Require.
func Middleware(db *store.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok && token != "" {
			id, err := db.PlayerByTokenHash(r.Context(), Hash(token))
			if err == nil {
				r = r.WithContext(WithPlayer(r.Context(), id))
			} else if !errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "auth lookup failed", http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
