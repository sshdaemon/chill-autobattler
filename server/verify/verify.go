// Package verify re-runs a battle natively and checks a client's claimed
// outcome against it.
package verify

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect/v2"

	chillv1 "chill/pb/chill/v1"
	"chill/pb/convert"
	"chill/sim"
)

type Server struct{}

func (Server) VerifyBattle(_ context.Context, req *chillv1.VerifyBattleRequest) (*chillv1.VerifyBattleResponse, error) {
	if v := req.GetInput().GetSimVersion(); v != sim.Version {
		return &chillv1.VerifyBattleResponse{Reason: fmt.Sprintf("unknown sim version %q (server has %q)", v, sim.Version)}, nil
	}
	in, err := convert.Input(req.GetInput())
	if err != nil {
		if errors.Is(err, sim.ErrInvalid) {
			return nil, connect.Errorf(connect.CodeInvalidArgument, "%v", err)
		}
		return nil, err
	}
	o, err := sim.Run(in)
	if err != nil {
		return nil, connect.Errorf(connect.CodeInvalidArgument, "%v", err)
	}
	res := &chillv1.VerifyBattleResponse{Outcome: convert.Outcome(o)}
	claimed := req.GetClaimed()
	switch {
	case claimed.GetResult() != res.Outcome.Result:
		res.Reason = "result mismatch: the battle did not end as claimed"
	case claimed.GetTicks() != res.Outcome.Ticks:
		res.Reason = "tick count mismatch"
	case claimed.GetChecksum() != res.Outcome.Checksum:
		res.Reason = "checksum mismatch: the battle did not play out as claimed"
	default:
		res.Fair = true
	}
	return res, nil
}
