// Fairness server: issues battles (seeds), takes the player's army before
// revealing the combat seed, and re-runs finished battles natively with the
// same sim the browser uses to check the claimed outcome.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"chill/pb/chill/v1/chillv1connect"
	"chill/server/auth"
	"chill/server/battle"
	"chill/server/run"
	"chill/server/store"
	"chill/server/web"
	"chill/sim"
)

func main() {
	path := os.Getenv("CHILL_DB")
	if path == "" {
		path = "chill.db"
	}
	db, err := store.Open(context.Background(), path)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	runs := run.NewService(db)
	rpc := connect.NewServer()
	chillv1connect.RegisterRunServiceHandler(rpc, runs)
	chillv1connect.RegisterBattleServiceHandler(rpc, battle.NewService(db, runs))
	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpc)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "ok", sim.Version) })
	// the built client (pnpm build), when there is one: CHILL_WEB, default ../web/dist
	dist := os.Getenv("CHILL_WEB")
	if dist == "" {
		dist = "../web/dist"
	}
	if st, err := os.Stat(dist); err == nil && st.IsDir() {
		mux.Handle("/", web.Handler(dist)) // not "GET /": it would clash with the RPC routes' patterns
		log.Printf("serving the client from %s", dist)
	}
	log.Printf("server on :8080 (db %s)", path)
	log.Fatal(http.ListenAndServe(":8080", auth.Middleware(db, mux)))
}
