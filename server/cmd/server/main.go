// Fairness server: re-runs battles natively with the same sim the browser
// uses and checks the client's claimed result.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"chill/server/gen/chill/v1"
	"chill/server/gen/chill/v1/chillv1connect"
	"chill/sim"
)

type battleServer struct{}

func (battleServer) VerifyBattle(_ context.Context, req *chillv1.VerifyBattleRequest) (*chillv1.VerifyBattleResponse, error) {
	if req.SimVersion != sim.Version {
		return &chillv1.VerifyBattleResponse{Reason: fmt.Sprintf("unknown sim version %q (server has %q)", req.SimVersion, sim.Version)}, nil
	}
	r := sim.Run(req.Seed, int(req.MaxTicks))
	res := &chillv1.VerifyBattleResponse{Winner: int32(r.Winner), Ticks: uint32(r.Ticks), Checksum: r.Checksum}
	switch {
	case r.Checksum != req.ClaimedChecksum:
		res.Reason = "checksum mismatch: the battle did not play out as claimed"
	case int32(r.Winner) != req.ClaimedWinner:
		res.Reason = "winner mismatch"
	default:
		res.Fair = true
	}
	return res, nil
}

func main() {
	rpc := connect.NewServer()
	chillv1connect.RegisterBattleServiceHandler(rpc, battleServer{})
	mux := http.NewServeMux()
	connecthttp.Mount(mux, rpc)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprintln(w, "ok", sim.Version) })
	log.Println("server on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
