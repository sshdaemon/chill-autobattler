// Fairness server: re-runs battles natively with the same sim the browser
// uses and checks the client's claimed result.
package main

import (
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"chill/pb/chill/v1/chillv1connect"
	"chill/server/verify"
	"chill/sim"
)

func main() {
	rpc := connect.NewServer()
	chillv1connect.RegisterBattleServiceHandler(rpc, verify.Server{})
	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpc)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "ok", sim.Version) })
	log.Println("server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
