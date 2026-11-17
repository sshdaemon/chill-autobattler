// Package battle issues battles and verifies them: StartBattle hands out a
// map seed, LockDeployment stores Blue's army and only then reveals the combat
// seed, VerifyBattle re-runs the stored battle and judges the client's claim.
// The server never steps a live battle.
package battle

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/proto"

	chillv1 "chill/pb/chill/v1"
	"chill/pb/convert"
	"chill/server/run"
	"chill/server/store"
	"chill/sim"
)

type Service struct {
	DB           *store.DB
	Runs         *run.Service
	Now          func() time.Time
	DeployWindow time.Duration // start → lock
	VerifyWindow time.Duration // lock → verify
}

func NewService(db *store.DB, runs *run.Service) *Service {
	return &Service{DB: db, Runs: runs, Now: time.Now, DeployWindow: 2 * time.Minute, VerifyWindow: 10 * time.Minute}
}

func (s *Service) nowMs() int64 { return s.Now().UnixMilli() }

func random64() uint64 {
	var b [8]byte
	rand.Read(b[:]) // never fails (crypto/rand panics on a broken system RNG)
	return binary.LittleEndian.Uint64(b[:])
}

func newID() string {
	var b [16]byte
	rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func wrongVersion(v string) error {
	return connect.Errorf(connect.CodeFailedPrecondition, "unknown sim version %q (server has %q)", v, sim.Version)
}

// StartBattle creates the next battle of a ready run. Red gets as many skills
// as the run's veterans have.
func (s *Service) StartBattle(ctx context.Context, req *chillv1.StartBattleRequest) (*chillv1.StartBattleResponse, error) {
	if req.SimVersion != sim.Version {
		return nil, wrongVersion(req.SimVersion)
	}
	player, err := run.Player(ctx)
	if err != nil {
		return nil, err
	}
	r, err := s.Runs.Get(ctx, player, req.RunId)
	if err != nil {
		return nil, err
	}
	if r.State != run.Ready || r.CurrentBattleID.Valid {
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "run %s can't start a battle now (%s)", r.ID, describe(r))
	}
	vets, err := run.Veterans(r)
	if err != nil {
		return nil, err
	}
	now := s.nowMs()
	p := store.CreateRunBattleParams{
		ID: newID(), SimVersion: sim.Version, MapSeed: int64(random64()), CombatSeed: int64(random64()),
		RedSkillCount: int64(run.SkillTotal(vets)), CreatedAt: now, DeployDeadline: now + s.DeployWindow.Milliseconds(),
		RunID: sql.NullString{String: r.ID, Valid: true}, MatchNo: sql.NullInt64{Int64: r.MatchNo, Valid: true},
	}
	err = s.DB.Tx(ctx, func(q *store.Queries) error {
		if err := q.CreateRunBattle(ctx, p); err != nil {
			return err
		}
		n, err := q.SetCurrentBattle(ctx, store.SetCurrentBattleParams{ID: r.ID, BattleID: sql.NullString{String: p.ID, Valid: true}})
		if err == nil && n == 0 {
			err = connect.Errorf(connect.CodeFailedPrecondition, "run %s already has a battle", r.ID)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return &chillv1.StartBattleResponse{BattleId: p.ID, Seed: uint64(p.MapSeed), RedSkillCount: uint32(p.RedSkillCount), DeployDeadlineMs: p.DeployDeadline}, nil
}

func describe(r store.Run) string {
	switch {
	case r.State == run.Over:
		return "it is over"
	case r.State == run.Picking:
		return "skill picks are owed"
	case r.CurrentBattleID.Valid:
		return "a battle is in progress"
	}
	return r.State
}

// owned loads a battle and its run, both owned by the calling player
// (NotFound otherwise, so ids can't be probed). The run is settled.
func (s *Service) owned(ctx context.Context, id string) (store.Battle, store.Run, error) {
	player, err := run.Player(ctx)
	if err != nil {
		return store.Battle{}, store.Run{}, err
	}
	b, err := s.get(ctx, id)
	if err != nil {
		return b, store.Run{}, err
	}
	if !b.RunID.Valid {
		return b, store.Run{}, connect.Errorf(connect.CodeNotFound, "no battle %q", id)
	}
	r, err := s.Runs.Get(ctx, player, b.RunID.String)
	if connect.CodeOf(err) == connect.CodeNotFound {
		return b, r, connect.Errorf(connect.CodeNotFound, "no battle %q", id)
	}
	return b, r, err
}

func (s *Service) get(ctx context.Context, id string) (store.Battle, error) {
	b, err := s.DB.GetBattle(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return b, connect.Errorf(connect.CodeNotFound, "no battle %q", id)
	}
	return b, err
}

// input rebuilds the battle input from stored columns and a Blue army.
func input(b store.Battle, blue *chillv1.Army) *chillv1.BattleInput {
	return &chillv1.BattleInput{
		SimVersion: b.SimVersion, Seed: uint64(b.MapSeed), CombatSeed: uint64(b.CombatSeed),
		RedSkillCount: uint32(b.RedSkillCount), Blue: blue,
	}
}

func (s *Service) LockDeployment(ctx context.Context, req *chillv1.LockDeploymentRequest) (*chillv1.LockDeploymentResponse, error) {
	b, r, err := s.owned(ctx, req.BattleId)
	if err != nil {
		return nil, err
	}
	if b.SimVersion != sim.Version {
		return nil, wrongVersion(b.SimVersion)
	}
	if r.State == run.Over || r.CurrentBattleID.String != b.ID {
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "battle %s is no longer the run's current battle (%s)", b.ID, describe(r))
	}
	vets, err := run.Veterans(r)
	if err != nil {
		return nil, err
	}
	if err := run.CheckArmy(vets, req.Blue); err != nil {
		return nil, connect.Errorf(connect.CodeInvalidArgument, "%v", err)
	}
	if _, err := convert.Input(input(b, req.Blue)); err != nil { // validate before storing
		return nil, connect.Errorf(connect.CodeInvalidArgument, "%v", err)
	}
	blue, err := proto.MarshalOptions{Deterministic: true}.Marshal(req.Blue)
	if err != nil {
		return nil, err
	}
	now := s.nowMs()
	verifyBy := now + s.VerifyWindow.Milliseconds()
	n, err := s.DB.LockBattle(ctx, store.LockBattleParams{
		ID: b.ID, BlueArmy: blue, Now: sql.NullInt64{Int64: now, Valid: true}, VerifyDeadline: sql.NullInt64{Int64: verifyBy, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	if n == 0 {
		if b, err = s.get(ctx, req.BattleId); err != nil {
			return nil, err
		}
		if b.LockedAt.Valid {
			return nil, connect.Errorf(connect.CodeFailedPrecondition, "battle %s is already locked", b.ID)
		}
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "battle %s expired: the deploy deadline has passed", b.ID)
	}
	return &chillv1.LockDeploymentResponse{CombatSeed: uint64(b.CombatSeed), VerifyDeadlineMs: verifyBy}, nil
}

func (s *Service) withRun(ctx context.Context, res *chillv1.VerifyBattleResponse, runID string) (*chillv1.VerifyBattleResponse, error) {
	r, err := s.DB.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	res.Run, err = s.Runs.Proto(ctx, r)
	return res, err
}

func stored(b store.Battle) (*chillv1.VerifyBattleResponse, error) {
	res := &chillv1.VerifyBattleResponse{Fair: b.Fair.Int64 == 1, Reason: b.Reason.String, Outcome: &chillv1.BattleOutcome{}}
	if err := proto.Unmarshal(b.Outcome, res.Outcome); err != nil {
		return nil, fmt.Errorf("stored outcome of %s: %w", b.ID, err)
	}
	return res, nil
}

func (s *Service) VerifyBattle(ctx context.Context, req *chillv1.VerifyBattleRequest) (*chillv1.VerifyBattleResponse, error) {
	b, r, err := s.owned(ctx, req.BattleId)
	if err != nil {
		return nil, err
	}
	if b.VerifiedAt.Valid { // the first verdict is final
		res, err := stored(b)
		if err != nil {
			return nil, err
		}
		return s.withRun(ctx, res, r.ID)
	}
	if !b.LockedAt.Valid {
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "battle %s has no locked deployment", b.ID)
	}
	if s.nowMs() >= b.VerifyDeadline.Int64 {
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "battle %s expired: the verify deadline has passed", b.ID)
	}
	if b.SimVersion != sim.Version {
		return nil, wrongVersion(b.SimVersion)
	}
	var blue chillv1.Army
	if err := proto.Unmarshal(b.BlueArmy, &blue); err != nil {
		return nil, fmt.Errorf("stored army of %s: %w", b.ID, err)
	}
	in, err := convert.Input(input(b, &blue))
	if err != nil {
		return nil, fmt.Errorf("stored battle %s no longer valid: %w", b.ID, err)
	}
	o, err := sim.Run(in)
	if err != nil {
		return nil, err
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
	outcome, _ := proto.MarshalOptions{Deterministic: true}.Marshal(res.Outcome)
	fair := int64(0)
	if res.Fair {
		fair = 1
	}
	now := s.nowMs()
	// The verdict and its effect on the run land together or not at all.
	var recorded int64
	err = s.DB.Tx(ctx, func(q *store.Queries) error {
		recorded, err = q.RecordVerdict(ctx, store.RecordVerdictParams{
			ID: b.ID, Now: sql.NullInt64{Int64: now, Valid: true}, Fair: sql.NullInt64{Int64: fair, Valid: true},
			Reason: sql.NullString{String: res.Reason, Valid: true}, Outcome: outcome,
		})
		if err != nil || recorded == 0 {
			return err
		}
		if r.State == run.Over || r.CurrentBattleID.String != b.ID {
			return nil // the run already moved on (e.g. restarted); the verdict still stands
		}
		return run.ApplyResult(ctx, q, r, res.Fair, &blue, o, now)
	})
	if err != nil {
		return nil, err
	}
	if recorded == 0 { // lost a race with another verify, or the deadline passed meanwhile
		if b, err = s.get(ctx, req.BattleId); err != nil {
			return nil, err
		}
		if b.VerifiedAt.Valid {
			res, err := stored(b)
			if err != nil {
				return nil, err
			}
			return s.withRun(ctx, res, r.ID)
		}
		return nil, connect.Errorf(connect.CodeFailedPrecondition, "battle %s expired: the verify deadline has passed", b.ID)
	}
	return s.withRun(ctx, res, r.ID)
}
