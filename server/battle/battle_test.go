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
	"chill/server/store"
	"chill/sim"
)

type env struct {
	t    *testing.T
	svc  *Service
	path string
	now  time.Time
}

func newEnv(t *testing.T) *env {
	e := &env{t: t, path: filepath.Join(t.TempDir(), "b.db"), now: time.UnixMilli(1_800_000_000_000)}
	e.open()
	return e
}

func (e *env) open() {
	db, err := store.Open(context.Background(), e.path)
	if err != nil {
		e.t.Fatal(err)
	}
	e.t.Cleanup(func() { db.Close() })
	e.svc = NewService(db)
	e.svc.Now = func() time.Time { return e.now }
}

var ctx = context.Background()

func army() *chillv1.Army {
	w, a, m := chillv1.UnitClass_UNIT_CLASS_WARRIOR, chillv1.UnitClass_UNIT_CLASS_ARCHER, chillv1.UnitClass_UNIT_CLASS_MAGE
	return &chillv1.Army{Formation: chillv1.Formation_FORMATION_3_2,
		Units: []*chillv1.UnitSpec{{UnitClass: w}, {UnitClass: w}, {UnitClass: a}, {UnitClass: a}, {UnitClass: m}}}
}

func (e *env) start() *chillv1.StartBattleResponse {
	r, err := e.svc.StartBattle(ctx, &chillv1.StartBattleRequest{SimVersion: sim.Version, RedSkillCount: 2})
	if err != nil {
		e.t.Fatal(err)
	}
	return r
}

func (e *env) lock(id string, blue *chillv1.Army) (*chillv1.LockDeploymentResponse, error) {
	return e.svc.LockDeployment(ctx, &chillv1.LockDeploymentRequest{BattleId: id, Blue: blue})
}

// play is the honest client: it plays what it was given, as the browser does.
func play(t *testing.T, st *chillv1.StartBattleResponse, combat uint64, blue *chillv1.Army) *chillv1.BattleOutcome {
	in, err := convert.Input(&chillv1.BattleInput{SimVersion: sim.Version, Seed: st.Seed, CombatSeed: combat, RedSkillCount: st.RedSkillCount, Blue: blue})
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
	return e.svc.VerifyBattle(ctx, &chillv1.VerifyBattleRequest{BattleId: id, Claimed: claim})
}

func code(err error) connect.Code { return connect.CodeOf(err) }

// 4.1
func TestStartBattle(t *testing.T) {
	e := newEnv(t)
	a, b := e.start(), e.start()
	if a.BattleId == "" || a.BattleId == b.BattleId || a.Seed == b.Seed {
		t.Errorf("battles should differ: %v %v", a, b)
	}
	if a.DeployDeadlineMs != e.now.Add(2*time.Minute).UnixMilli() || a.RedSkillCount != 2 {
		t.Errorf("start response %v", a)
	}
	ra, _ := e.lock(a.BattleId, army())
	rb, _ := e.lock(b.BattleId, army())
	if ra.CombatSeed == rb.CombatSeed || ra.CombatSeed == a.Seed {
		t.Error("combat seeds should be fresh and differ from the map seed")
	}
	_, err := e.svc.StartBattle(ctx, &chillv1.StartBattleRequest{SimVersion: "0.0.1-old"})
	if code(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "0.0.1-old") || !strings.Contains(err.Error(), sim.Version) {
		t.Errorf("unknown version: %v", err)
	}
	if _, err := e.svc.StartBattle(ctx, &chillv1.StartBattleRequest{SimVersion: sim.Version, RedSkillCount: 99}); code(err) != connect.CodeInvalidArgument {
		t.Errorf("red skill count out of range: %v", err)
	}
}

// 4.2
func TestLockDeployment(t *testing.T) {
	e := newEnv(t)
	st := e.start()
	r, err := e.lock(st.BattleId, army())
	if err != nil || r.CombatSeed == 0 {
		t.Fatalf("lock: %v %v", err, r)
	}
	other := army()
	other.Formation = chillv1.Formation_FORMATION_SPREAD
	if _, err := e.lock(st.BattleId, other); code(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "already locked") {
		t.Errorf("second lock: %v", err)
	}
	// the stored army is still the first one: the honest first-army claim verifies
	if res, err := e.verify(st.BattleId, play(t, st, r.CombatSeed, army())); err != nil || !res.Fair {
		t.Errorf("stored army changed? %v %v", err, res)
	}

	late := e.start()
	e.now = e.now.Add(2*time.Minute + time.Millisecond)
	if _, err := e.lock(late.BattleId, army()); code(err) != connect.CodeFailedPrecondition || !strings.Contains(err.Error(), "expired") {
		t.Errorf("too late: %v", err)
	}

	bad := e.start()
	six := army()
	six.Units = append(six.Units, six.Units[0])
	if _, err := e.lock(bad.BattleId, six); code(err) != connect.CodeInvalidArgument {
		t.Errorf("six units: %v", err)
	}
	outside := army()
	outside.Formation = chillv1.Formation_FORMATION_CUSTOM
	for _, u := range outside.Units {
		u.Position = &chillv1.Point{ZRaw: int64(sim.I(-10))}
	}
	if _, err := e.lock(bad.BattleId, outside); code(err) != connect.CodeInvalidArgument {
		t.Errorf("outside zone: %v", err)
	}
	if _, err := e.lock(bad.BattleId, army()); err != nil {
		t.Errorf("battle should still be unlocked after invalid armies: %v", err)
	}
	if _, err := e.lock("nope", army()); code(err) != connect.CodeNotFound {
		t.Errorf("unknown battle: %v", err)
	}
}

// 4.3
func TestVerifyBattle(t *testing.T) {
	e := newEnv(t)
	st := e.start()
	r, _ := e.lock(st.BattleId, army())
	honest := play(t, st, r.CombatSeed, army())
	res, err := e.verify(st.BattleId, honest)
	if err != nil || !res.Fair || !proto.Equal(res.Outcome, honest) {
		t.Fatalf("honest claim: %v %v", err, res)
	}

	// re-rolled: played with a different combat seed or army than stored
	for name, claim := range map[string]func(*chillv1.StartBattleResponse, uint64) *chillv1.BattleOutcome{
		"other combat seed": func(s *chillv1.StartBattleResponse, c uint64) *chillv1.BattleOutcome { return play(t, s, c+1, army()) },
		"other army": func(s *chillv1.StartBattleResponse, c uint64) *chillv1.BattleOutcome {
			a := army()
			a.Formation = chillv1.Formation_FORMATION_5_0
			return play(t, s, c, a)
		},
	} {
		s := e.start()
		lr, _ := e.lock(s.BattleId, army())
		res, err := e.verify(s.BattleId, claim(s, lr.CombatSeed))
		if err != nil || res.Fair {
			t.Errorf("%s: should not be fair: %v %v", name, err, res)
		}
	}

	if _, err := e.verify("nope", honest); code(err) != connect.CodeNotFound {
		t.Errorf("unknown battle: %v", err)
	}
	unlocked := e.start()
	if _, err := e.verify(unlocked.BattleId, honest); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("unlocked battle: %v", err)
	}
}

func TestVerifyOnce(t *testing.T) {
	e := newEnv(t)
	st := e.start()
	r, _ := e.lock(st.BattleId, army())
	honest := play(t, st, r.CombatSeed, army())
	wrong := proto.Clone(honest).(*chillv1.BattleOutcome)
	wrong.Checksum++
	first, _ := e.verify(st.BattleId, wrong)
	second, err := e.verify(st.BattleId, honest)
	if err != nil || first.Fair || second.Fair || second.Reason != first.Reason || !proto.Equal(second.Outcome, honest) {
		t.Errorf("first verdict must be final: %v / %v %v", first, second, err)
	}
}

func TestVerifyExpires(t *testing.T) {
	e := newEnv(t)
	st := e.start()
	r, _ := e.lock(st.BattleId, army())
	e.now = e.now.Add(10 * time.Minute)
	if _, err := e.verify(st.BattleId, play(t, st, r.CombatSeed, army())); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("expired verify: %v", err)
	}
}

func TestRestartBetweenLockAndVerify(t *testing.T) {
	e := newEnv(t)
	st := e.start()
	r, _ := e.lock(st.BattleId, army())
	e.svc.DB.Close()
	e.open() // a fresh service on the same file, as after a restart
	res, err := e.verify(st.BattleId, play(t, st, r.CombatSeed, army()))
	if err != nil || !res.Fair {
		t.Errorf("after restart: %v %v", err, res)
	}
}
