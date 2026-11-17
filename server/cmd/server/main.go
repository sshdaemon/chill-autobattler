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
	"chill/server/battle"
	"chill/server/store"
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

	rpc := connect.NewServer()
	chillv1connect.RegisterBattleServiceHandler(rpc, battle.NewService(db))
	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpc)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "ok", sim.Version) })
	log.Printf("server on :8080 (db %s)", path)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
