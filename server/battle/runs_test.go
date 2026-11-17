package battle

import (
	"testing"
	"time"

	"connectrpc.com/connect/v2"

	chillv1 "chill/pb/chill/v1"
	"chill/server/store"
	"chill/sim"
)

// playMatch starts the run's next battle, locks blue, plays it honestly and
// verifies it. It returns the start response and the verify response.
func (e *env) playMatch(runID string, blue *chillv1.Army) (*chillv1.StartBattleResponse, *chillv1.VerifyBattleResponse) {
	st := e.mustStart(runID)
	lr, err := e.lock(st.BattleId, blue)
	if err != nil {
		e.t.Fatalf("lock: %v", err)
	}
	res, err := e.verify(st.BattleId, play(e.t, st, lr.CombatSeed, blue))
	if err != nil || !res.Fair {
		e.t.Fatalf("verify: %v %v", err, res)
	}
	return st, res
}

// winUntil plays fresh runs with strong armies until one wins its first match.
func (e *env) firstWin(blue *chillv1.Army) (*chillv1.Run, *chillv1.VerifyBattleResponse) {
	for i := 0; i < 40; i++ {
		r := e.startRun(e.ctx)
		_, res := e.playMatch(r.Id, blue)
		if res.Outcome.Result == chillv1.BattleResult_BATTLE_RESULT_BLUE_WIN {
			return res.Run, res
		}
	}
	e.t.Fatal("no win in 40 tries")
	return nil, nil
}

func TestNewRun(t *testing.T) {
	e := newEnv(t)
	r := e.startRun(e.ctx)
	if r.Name == "" || r.Match != 1 || r.Wins != 0 || r.Score != 0 || len(r.Veterans) != 0 || r.State != chillv1.RunState_RUN_STATE_READY {
		t.Errorf("new run %v", r)
	}
	if got := e.getRun(); got.Id != r.Id {
		t.Error("GetRun should return the active run (reload resumes)")
	}
	r2 := e.startRun(e.ctx)
	old, _ := e.runs.DB.GetRun(e.ctx, r.Id)
	if old.State != "over" || old.EndReason.String != "restarted" || e.getRun().Id != r2.Id {
		t.Errorf("restart should end the old run: %+v", old)
	}
	if _, err := e.runs.StartRun(t.Context(), &chillv1.StartRunRequest{}); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Errorf("no token: %v", err)
	}
}

func TestWinPicksAndNextMatch(t *testing.T) {
	e := newEnv(t)
	r, res := e.firstWin(mix())
	survivors := 0
	for _, s := range res.Outcome.Survivors {
		if s.Index < 5 {
			survivors++
		}
	}
	if r.Wins != 1 || r.Match != 2 || r.Score != uint32(100+10*survivors) || len(r.Veterans) != survivors {
		t.Fatalf("after a win with %d survivors: %v", survivors, r)
	}
	if r.State != chillv1.RunState_RUN_STATE_PICKING || len(r.OwedPicks) != survivors {
		t.Fatalf("every survivor owes a pick: %v", r)
	}
	if _, err := e.start(r.Id); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("battle before picks: %v", err)
	}

	// invalid picks change nothing
	v0 := r.Veterans[0]
	foreign := chillv1.Skill_SKILL_BLINK // mage only
	if v0.UnitClass == chillv1.UnitClass_UNIT_CLASS_MAGE {
		foreign = chillv1.Skill_SKILL_DASH
	}
	bad := [][]*chillv1.SkillPick{
		{{Veteran: 0, Skill: foreign}},
		nil, // too few
	}
	for _, picks := range bad {
		if _, err := e.runs.PickSkills(e.ctx, &chillv1.PickSkillsRequest{RunId: r.Id, Picks: picks}); code(err) != connect.CodeInvalidArgument {
			t.Errorf("bad picks %v: %v", picks, err)
		}
	}

	var picks []*chillv1.SkillPick
	for _, i := range r.OwedPicks {
		cls := sim.Class(r.Veterans[i].UnitClass - 1)
		picks = append(picks, &chillv1.SkillPick{Veteran: i, Skill: chillv1.Skill(sim.SkillsOf[cls][0].Skill)})
	}
	pr, err := e.runs.PickSkills(e.ctx, &chillv1.PickSkillsRequest{RunId: r.Id, Picks: picks})
	if err != nil || pr.Run.State != chillv1.RunState_RUN_STATE_READY || pr.Run.RedSkills != uint32(len(picks)) {
		t.Fatalf("picks: %v %v", err, pr)
	}
	if _, err := e.runs.PickSkills(e.ctx, &chillv1.PickSkillsRequest{RunId: r.Id, Picks: picks}); code(err) != connect.CodeFailedPrecondition {
		t.Errorf("picking twice: %v", err)
	}

	// match 2: Red gets as many skills as the veterans; veterans must lead the army
	st := e.mustStart(r.Id)
	if st.RedSkillCount != uint32(len(picks)) {
		t.Errorf("red skills %d, want %d", st.RedSkillCount, len(picks))
	}
	if _, err := e.lock(st.BattleId, mix()); code(err) != connect.CodeInvalidArgument {
		t.Errorf("army without the veterans: %v", err)
	}
	var recruits []chillv1.UnitClass
	for i := len(pr.Run.Veterans); i < 5; i++ {
		recruits = append(recruits, W)
	}
	blue := army(pr.Run.Veterans, recruits...)
	lr, err := e.lock(st.BattleId, blue)
	if err != nil {
		t.Fatalf("army with veterans: %v", err)
	}
	if got := e.getRun(); !got.Battle.Locked || got.Battle.CombatSeed != lr.CombatSeed || len(got.Battle.Blue.Units) != 5 {
		t.Errorf("GetRun must carry what a reload needs to replay: %v", got.Battle)
	}
}

func TestLosingEndsTheRun(t *testing.T) {
	e := newEnv(t)
	for i := 0; i < 40; i++ {
		r := e.startRun(e.ctx)
		_, res := e.playMatch(r.Id, army(nil, M, M, M, M, M))
		if res.Outcome.Result == chillv1.BattleResult_BATTLE_RESULT_BLUE_WIN {
			continue
		}
		if res.Run.State != chillv1.RunState_RUN_STATE_OVER || (res.Run.EndReason != "defeat" && res.Run.EndReason != "draw") {
			t.Fatalf("a lost match must end the run: %v", res.Run)
		}
		if _, err := e.start(r.Id); code(err) != connect.CodeFailedPrecondition {
			t.Errorf("battle in a finished run: %v", err)
		}
		return
	}
	t.Fatal("no loss in 40 tries")
}

func TestAbandonedBattleIsALoss(t *testing.T) {
	e := newEnv(t)
	// locked, then the tab closed: past the verify deadline
	r := e.startRun(e.ctx)
	st := e.mustStart(r.Id)
	e.lock(st.BattleId, mix())
	e.now = e.now.Add(11 * time.Minute)
	if got := e.getRun(); got.State != chillv1.RunState_RUN_STATE_OVER || got.EndReason != "abandoned" {
		t.Errorf("locked and never verified: %v", got)
	}
	// started, never locked: past the deploy deadline
	r = e.startRun(e.ctx)
	e.mustStart(r.Id)
	e.now = e.now.Add(3 * time.Minute)
	if got := e.getRun(); got.State != chillv1.RunState_RUN_STATE_OVER || got.EndReason != "abandoned" {
		t.Errorf("never locked: %v", got)
	}
}

func TestScoreboard(t *testing.T) {
	e := newEnv(t)
	// another player's finished run with a known high score
	other := e.other()
	or := e.startRun(other)
	if err := e.runs.DB.WinMatch(e.ctx, store.WinMatchParams{ID: or.Id, State: "ready", Points: 990}); err != nil {
		t.Fatal(err)
	}
	e.startRun(other) // restart finishes it
	// an abandoned run of ours must reach the board through the sweep
	r := e.startRun(e.ctx)
	e.mustStart(r.Id)
	e.now = e.now.Add(3 * time.Minute)
	sb, err := e.runs.GetScoreboard(e.ctx, &chillv1.GetScoreboardRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sb.Top) < 2 || sb.Top[0].Score != 990 || sb.Top[0].Mine {
		t.Fatalf("top: %v", sb.Top)
	}
	mine := 0
	for _, row := range sb.Top {
		if row.Mine {
			mine++
		}
	}
	if mine != 1 || sb.MyBestRank != 2 {
		t.Errorf("our abandoned run should be listed and ranked 2nd: %v rank %d", sb.Top, sb.MyBestRank)
	}
}
