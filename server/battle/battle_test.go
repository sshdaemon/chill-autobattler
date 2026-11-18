package battle

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"google.golang.org/protobuf/proto"

	chillv1 "chill/pb/chill/v1"
	"chill/pb/convert"
	"chill/server/auth"
	"chill/server/run"
	"chill/server/store"
	"chill/sim"
)

// env is a server with a temp DB, an injected clock and one logged-in player.
type env struct {
	t      *testing.T
	svc    *Service
	runs   *run.Service
	path   string
	now    time.Time
	ctx    context.Context // carries the player
	player string
}

func newEnv(t *testing.T) *env {
	e := &env{t: t, path: filepath.Join(t.TempDir(), "b.db"), now: time.UnixMilli(1_800_000_000_000)}
	e.open()
	tok, err := e.runs.CreatePlayer(context.Background(), &chillv1.CreatePlayerRequest{})
	if err != nil {
		t.Fatal(err)
	}
	e.player, _ = e.svc.DB.PlayerByTokenHash(context.Background(), auth.Hash(tok.PlayerToken))
	e.ctx = auth.WithPlayer(context.Background(), e.player)
	return e
}

func (e *env) open() {
	db, err := store.Open(context.Background(), e.path)
	if err != nil {
		e.t.Fatal(err)
	}
	e.t.Cleanup(func() { db.Close() })
	clock := func() time.Time { return e.now }
	e.runs = run.NewService(db)
	e.runs.Now = clock
	e.svc = NewService(db, e.runs)
	e.svc.Now = clock
}

// as returns a context for another, new player.
func (e *env) other() context.Context {
	tok, _ := e.runs.CreatePlayer(context.Background(), &chillv1.CreatePlayerRequest{})
	id, _ := e.svc.DB.PlayerByTokenHash(context.Background(), auth.Hash(tok.PlayerToken))
	return auth.WithPlayer(context.Background(), id)
}

func (e *env) startRun(ctx context.Context) *chillv1.Run {
	r, err := e.runs.StartRun(ctx, &chillv1.StartRunRequest{})
	if err != nil {
		e.t.Fatal(err)
	}
	return r.Run
}

func (e *env) getRun() *chillv1.Run {
	r, err := e.runs.GetRun(e.ctx, &chillv1.GetRunRequest{})
	if err != nil {
		e.t.Fatal(err)
	}
	return r.Run
}

func (e *env) start(runID string) (*chillv1.StartBattleResponse, error) {
	return e.svc.StartBattle(e.ctx, &chillv1.StartBattleRequest{SimVersion: sim.Version, RunId: runID})
}

func (e *env) mustStart(runID string) *chillv1.StartBattleResponse {
	st, err := e.start(runID)
	if err != nil {
		e.t.Fatal(err)
	}
	return st
}

// army: the run's veterans first, then recruits (level 1: their first ability).
func army(vets []*chillv1.UnitSpec, recruits ...chillv1.UnitClass) *chillv1.Army {
	a := &chillv1.Army{Formation: chillv1.Formation_FORMATION_3_2}
	for _, v := range vets {
		a.Units = append(a.Units, &chillv1.UnitSpec{UnitClass: v.UnitClass, Skills: v.Skills})
	}
	for _, c := range recruits {
		a.Units = append(a.Units, &chillv1.UnitSpec{UnitClass: c, Skills: []uint32{1}})
	}
	return a
}

const (
	W = chillv1.UnitClass_UNIT_CLASS_WARRIOR
	A = chillv1.UnitClass_UNIT_CLASS_ARCHER
	M = chillv1.UnitClass_UNIT_CLASS_MAGE
)

func mix() *chillv1.Army { return army(nil, W, W, A, A, M) }

func (e *env) lock(id string, blue *chillv1.Army) (*chillv1.LockDeploymentResponse, error) {
	return e.svc.LockDeployment(e.ctx, &chillv1.LockDeploymentRequest{BattleId: id, Blue: blue})
}

// play is the honest client: it plays what it was given, as the browser does.
func play(t *testing.T, st *chillv1.StartBattleResponse, combat uint64, blue *chillv1.Army) *chillv1.BattleOutcome {
	in, err := convert.Input(&chillv1.BattleInput{SimVersion: sim.Version, Seed: st.Seed, CombatSeed: combat, RedSkillCount: st.RedSkillCount, RedHealthPct: st.RedHealthPct, Blue: blue})
	if err != nil {
		t.Fatal(err)
	}
	o, err := sim.Run(in)
	if err != nil {
		t.Fatal(err)
	}
	return convert.Outcome(o)
}

func (e *env) verify(id string, claim *chillv1.BattleOutcome) (*chillv1.VerifyBattleResponse, error) {
	return e.svc.VerifyBattle(e.ctx, &chillv1.VerifyBattleRequest{BattleId: id, Claimed: claim})
}

func code(err error) connect.Code { return connect.CodeOf(err) }

// ---------- battle-sessions ----------

func TestStartBattle(t *testing.T) {
	e := newEnv(t)
	r1 := e.startRun(e.ctx)
	a := e.mustStart(r1.Id)
	if a.BattleId == "" || a.DeployDeadlineMs != e.now.Add(2*time.Minute).UnixMilli() || a.RedSkillCount != 0 {
		t.Errorf("start response %v", a)
	}
	if _, err := e.start(r1.Id); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("second battle while one is in progress: %v", err)
	}
	r2 := e.startRun(e.ctx) // restart: a fresh run
	b := e.mustStart(r2.Id)
	if a.BattleId == b.BattleId || a.Seed == b.Seed {
		t.Error("battles should differ")
	}
	ra, _ := e.lock(b.BattleId, mix())
	if ra == nil || ra.CombatSeed == b.Seed {
		t.Error("combat seed should be fresh")
	}
	_, err := e.svc.StartBattle(e.ctx, &chillv1.StartBattleRequest{SimVersion: "0.0.1-old", RunId: r2.Id})
	if code(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "0.0.1-old") || !strings.Contains(err.Error(), sim.Version) {
		t.Errorf("unknown version: %v", err)
	}
	if _, err := e.svc.StartBattle(context.Background(), &chillv1.StartBattleRequest{SimVersion: sim.Version, RunId: r2.Id}); code(err) != connect.CodeUnauthenticated {
		t.Errorf("no player: %v", err)
	}
	if _, err := e.svc.StartBattle(e.other(), &chillv1.StartBattleRequest{SimVersion: sim.Version, RunId: r2.Id}); code(err) != connect.CodeNotFound {
		t.Errorf("someone else's run: %v", err)
	}
}

func TestLockDeployment(t *testing.T) {
	e := newEnv(t)
	st := e.mustStart(e.startRun(e.ctx).Id)
	r, err := e.lock(st.BattleId, mix())
	if err != nil || r.CombatSeed == 0 {
		t.Fatalf("lock: %v %v", err, r)
	}
	other := mix()
	other.Formation = chillv1.Formation_FORMATION_SPREAD
	if _, err := e.lock(st.BattleId, other); code(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "already locked") {
		t.Errorf("second lock: %v", err)
	}
	if res, err := e.verify(st.BattleId, play(t, st, r.CombatSeed, mix())); err != nil || !res.Fair {
		t.Errorf("stored army changed? %v %v", err, res)
	}

	late := e.mustStart(e.startRun(e.ctx).Id)
	e.now = e.now.Add(2*time.Minute + time.Millisecond)
	if _, err := e.lock(late.BattleId, mix()); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("too late: %v", err)
	}

	bad := e.mustStart(e.startRun(e.ctx).Id)
	six := mix()
	six.Units = append(six.Units, six.Units[0])
	outside := mix()
	outside.Formation = chillv1.Formation_FORMATION_CUSTOM
	for _, u := range outside.Units {
		u.Position = &chillv1.Point{ZRaw: int64(sim.I(-10))}
	}
	skilled := mix()
	skilled.Units[0].Skills = []uint32{1, uint32(sim.WarriorDash)}
	bare := mix()
	bare.Units[0].Skills = nil
	for name, a := range map[string]*chillv1.Army{"six units": six, "outside zone": outside, "recruit at level 2": skilled, "recruit with no ability": bare, "four units": army(nil, W, W, A, A)} {
		if _, err := e.lock(bad.BattleId, a); code(err) != connect.CodeInvalidArgument {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := e.lock(bad.BattleId, mix()); err != nil {
		t.Errorf("battle should still be unlocked after invalid armies: %v", err)
	}
	if _, err := e.lock("nope", mix()); code(err) != connect.CodeNotFound {
		t.Errorf("unknown battle: %v", err)
	}
}

func TestVerifyBattle(t *testing.T) {
	e := newEnv(t)
	for name, claim := range map[string]func(*chillv1.StartBattleResponse, uint64) *chillv1.BattleOutcome{
		"other combat seed": func(s *chillv1.StartBattleResponse, c uint64) *chillv1.BattleOutcome { return play(t, s, c+1, mix()) },
		"other army": func(s *chillv1.StartBattleResponse, c uint64) *chillv1.BattleOutcome {
			a := mix()
			a.Formation = chillv1.Formation_FORMATION_5_0
			return play(t, s, c, a)
		},
	} {
		s := e.mustStart(e.startRun(e.ctx).Id)
		lr, _ := e.lock(s.BattleId, mix())
		res, err := e.verify(s.BattleId, claim(s, lr.CombatSeed))
		if err != nil || res.Fair {
			t.Errorf("%s: should not be fair: %v %v", name, err, res)
		}
		if res.GetRun().GetState() != chillv1.RunState_RUN_STATE_OVER || res.Run.EndReason != "not fair" {
			t.Errorf("%s: an unfair claim must end the run: %v", name, res.GetRun())
		}
	}
	if _, err := e.verify("nope", &chillv1.BattleOutcome{}); code(err) != connect.CodeNotFound {
		t.Errorf("unknown battle: %v", err)
	}
	unlocked := e.mustStart(e.startRun(e.ctx).Id)
	if _, err := e.verify(unlocked.BattleId, &chillv1.BattleOutcome{}); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("unlocked battle: %v", err)
	}
	if _, err := e.svc.VerifyBattle(e.other(), &chillv1.VerifyBattleRequest{BattleId: unlocked.BattleId}); code(err) != connect.CodeNotFound {
		t.Errorf("someone else's battle: %v", err)
	}
}

func TestVerifyOnce(t *testing.T) {
	e := newEnv(t)
	st := e.mustStart(e.startRun(e.ctx).Id)
	r, _ := e.lock(st.BattleId, mix())
	honest := play(t, st, r.CombatSeed, mix())
	wrong := proto.Clone(honest).(*chillv1.BattleOutcome)
	wrong.Checksum++
	first, _ := e.verify(st.BattleId, wrong)
	second, err := e.verify(st.BattleId, honest)
	if err != nil || first.Fair || second.Fair || second.Reason != first.Reason || !proto.Equal(second.Outcome, honest) {
		t.Errorf("first verdict must be final: %v / %v %v", first, second, err)
	}
}

func TestRestartBetweenLockAndVerify(t *testing.T) {
	e := newEnv(t)
	st := e.mustStart(e.startRun(e.ctx).Id)
	r, _ := e.lock(st.BattleId, mix())
	e.svc.DB.Close()
	e.open() // a fresh server on the same file
	res, err := e.verify(st.BattleId, play(t, st, r.CombatSeed, mix()))
	if err != nil || !res.Fair {
		t.Errorf("after restart: %v %v", err, res)
	}
}
