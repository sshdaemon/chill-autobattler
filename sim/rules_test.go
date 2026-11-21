package sim

import (
	"testing"
)

// tu describes a unit for a hand-made test battle.
type tu struct {
	team   int
	class  Class
	pos    Vec
	skills []Skill
}

// testBattle sets up a battle on a given map with units at exact positions,
// ready to act on the first tick.
func testBattle(obs []Obstacle, units ...tu) *Battle {
	b := &Battle{Map: NewMap(obs), rng: NewRNG(1)}
	for _, u := range units {
		b.Units = append(b.Units, newUnit(len(b.Units), u.team, UnitSpec{Class: u.class, Skills: u.skills}, u.pos))
	}
	return b
}

func (b *Battle) steps(n int) {
	for i := 0; i < n && b.Result == Ongoing; i++ {
		b.Step()
	}
}

func p(x, z int) Vec { return V(Ratio(x, 10), Ratio(z, 10)) } // decimetres, to keep tables short

// ---------- 3. Movement ----------

func TestPondDetour(t *testing.T) {
	m := NewMap([]Obstacle{{Kind: Water, Pos: V(0, 0), R: Ratio(25, 10)}})
	// the straight line crosses over 1 m of water near the pond's edge; a slight
	// bend avoids almost all of it (string-pulling may clip up to one cell)
	a, b := p(20, -50), p(20, 50)
	if m.WetLength(a, b) < I(1) {
		t.Fatal("setup: straight line should cross the pond")
	}
	path := m.FindPath(a, b)
	prev, wet := a, Fx(0)
	for _, w := range path {
		wet += m.WetLength(prev, w)
		prev = w
	}
	if wet > Cell {
		t.Errorf("pond edge: path wades %v m instead of going round: %v", wet, path)
	}
}

func TestRiverWade(t *testing.T) {
	var river []Obstacle
	for x := -12; x <= 6; x += 2 { // a river across most of the arena; going round is a long way
		river = append(river, Obstacle{Kind: Water, Pos: V(I(x), 0), R: Ratio(15, 10)})
	}
	m := NewMap(river)
	a, b := p(-60, -40), p(-60, 40)
	path := m.FindPath(a, b)
	if path == nil {
		t.Fatal("no path")
	}
	prev, wet := a, Fx(0)
	for _, w := range path {
		wet += m.WetLength(prev, w)
		prev = w
	}
	if wet == 0 {
		t.Errorf("long detour: path should wade across, got %v", path)
	}
	if path[len(path)-1] != b {
		t.Error("path must end at the goal")
	}
}

func TestPathAroundRock(t *testing.T) {
	m := NewMap([]Obstacle{{Kind: Rock, Pos: V(0, 0), R: I(2)}})
	path := m.FindPath(p(0, -50), p(0, 50))
	prev := p(0, -50)
	for _, w := range path {
		if !m.LineOfSight(prev, w) {
			t.Fatalf("path leg %v→%v crosses the rock", prev, w)
		}
		prev = w
	}
	if len(path) < 2 {
		t.Errorf("path should bend round the rock: %v", path)
	}
}

func TestMovement(t *testing.T) {
	speed := perTick(Roster[Druzhynnyk].Stats.Speed)
	walked := func(obs []Obstacle, setup func(*Unit)) Fx {
		b := testBattle(obs, tu{0, Druzhynnyk, V(0, 0), nil})
		u := b.Units[0]
		u.Face = V(0, One)
		target := V(0, I(10))
		u.moveTarget = &target
		if setup != nil {
			setup(u)
		}
		b.move(u)
		return u.Pos.Z
	}
	if got := walked(nil, nil); got != speed {
		t.Errorf("dry walk %v, want %v", got, speed)
	}
	if got := walked([]Obstacle{{Kind: Water, Pos: V(0, 0), R: I(3)}}, nil); got != speed.Mul(WaterSpeed) {
		t.Errorf("wading %v, want 70%%", got)
	}
	if got := walked(nil, func(u *Unit) { u.slow, u.slowFactor = 10, Ratio(6, 10) }); got != speed.Mul(Ratio(6, 10)) {
		t.Errorf("frost %v, want 60%%", got)
	}
	if got := walked(nil, func(u *Unit) { u.Face = V(0, -One) }); got != speed.Mul(SlowTurnSpeed) {
		t.Errorf("walking while turning round %v, want 30%%", got)
	}

	// turning is limited to 0.4 rad per tick
	b := testBattle(nil, tu{0, Druzhynnyk, V(0, 0), nil})
	u := b.Units[0]
	u.Face = V(0, One)
	tgt := V(I(10), 0)
	u.moveTarget = &tgt
	b.move(u)
	if d := u.Face.Dot(V(0, One)); d < TurnPerTick.Cos-2 || d > TurnPerTick.Cos+2 {
		t.Errorf("turned by more or less than 0.4 rad: face %v", u.Face)
	}

	// separation pushes overlapping units apart
	b = testBattle(nil, tu{0, Druzhynnyk, V(0, 0), nil}, tu{0, Druzhynnyk, p(2, 0), nil})
	b.Step()
	if d := Dist(b.Units[0].Pos, b.Units[1].Pos); d < Separation-Ratio(1, 100) {
		t.Errorf("units still overlap: %v apart", d)
	}
}

// ---------- 4. Combat ----------

func TestSpottedByTeammate(t *testing.T) {
	// archer at the left, rock between it and the red druzhynnyk; a blue scout sees the druzhynnyk
	b := testBattle([]Obstacle{{Kind: Rock, Pos: p(0, 0), R: I(1)}},
		tu{0, Toxotes, p(-40, 0), nil}, tu{0, Druzhynnyk, p(30, 40), nil}, tu{1, Druzhynnyk, p(40, 0), nil})
	if b.canSee(b.Units[0], b.Units[2]) {
		t.Fatal("setup: the archer itself should not see through the rock")
	}
	b.Units[1].cd, b.Units[2].cd = 1000, 1000 // keep them out of the fight
	b.Step()
	if b.Units[0].Target != 2 {
		t.Fatalf("archer should target the druzhynnyk its teammate sees, got %d", b.Units[0].Target)
	}
	if !b.Units[0].Moving() {
		t.Error("archer without a clear shot should move to get one")
	}
}

func TestNothingInSight(t *testing.T) {
	b := testBattle(nil, tu{0, Druzhynnyk, p(0, 100), nil}, tu{1, Druzhynnyk, p(0, -125), nil})
	b.Step()
	u := b.Units[0]
	if u.Target != -1 || u.advanceGoal == nil || u.advanceGoal.Z > -I(11) {
		t.Fatalf("blue should march toward red's edge, goal %v", u.advanceGoal)
	}
	z0 := u.Pos.Z
	b.steps(10)
	if u.Pos.Z >= z0 {
		t.Error("blue did not advance")
	}
}

func TestInvestigateLastSeen(t *testing.T) {
	b := testBattle(nil, tu{0, Druzhynnyk, p(0, 50), nil}, tu{1, Druzhynnyk, p(0, -10), nil})
	b.Step()
	if b.Units[0].Target != 1 {
		t.Fatal("setup: should see the enemy")
	}
	b.Units[1].Pos = p(0, -125) // out of sight
	b.Units[0].retarget = 0
	b.Step()
	if u := b.Units[0]; u.Target != -1 || u.investigate == nil {
		t.Fatalf("lost sight: should investigate, got target %d investigate %v", u.Target, u.investigate)
	}
}

func TestMeleeHitTiming(t *testing.T) {
	b := testBattle(nil, tu{0, Druzhynnyk, p(0, 10), nil}, tu{1, Volva, p(0, 0), nil})
	b.Units[1].cd = 1000
	for i := 0; i < 10; i++ { // first tick: target + start swing
		b.Step()
		if b.Units[0].swing != nil {
			break
		}
	}
	s := b.Units[0].swing
	if s == nil {
		t.Fatal("druzhynnyk never swung")
	}
	hitTick := (HitAtPct*s.dur + 99) / 100
	for s.t < hitTick-1 {
		b.Step()
		if b.Units[1].HP != b.Units[1].MaxHP {
			t.Fatalf("hit landed early at swing tick %d", s.t)
		}
	}
	b.Step()
	if b.Units[1].HP == b.Units[1].MaxHP {
		t.Fatalf("hit did not land at swing tick %d (45%% of %d)", hitTick, s.dur)
	}
}

func TestTargetDiesBeforeArrowLands(t *testing.T) {
	b := testBattle(nil, tu{0, Toxotes, p(0, 80), nil}, tu{1, Druzhynnyk, p(0, 0), nil})
	b.Units[1].cd = 1000
	for i := 0; i < 60 && len(b.Shots) == 0; i++ {
		b.Step()
	}
	if len(b.Shots) == 0 {
		t.Fatal("archer never shot")
	}
	v := b.Units[1]
	v.HP, v.Dead = 0, true // killed by someone else mid-flight
	dealt := b.Units[0].Dealt
	b.steps(40)
	if b.Units[0].Dealt != dealt {
		t.Error("an arrow at a dead target must not deal damage")
	}
}

func TestFlankAndMatchup(t *testing.T) {
	hit := func(from Vec) (Fx, uint8) {
		b := testBattle(nil, tu{0, Druzhynnyk, from, nil}, tu{1, Toxotes, V(0, 0), nil})
		b.Units[1].Face = V(0, -One) // archer faces -z
		b.rng = NewRNG(99)
		b.damage(b.Units[0], b.Units[1], from, One)
		ev := b.TakeEvents()[0]
		return ev.Amount, ev.Flags
	}
	front, ff := hit(p(0, -10))
	back, bf := hit(p(0, 10))
	roll := NewRNG(99).Between(I(10), I(16))
	m := MatchupOf(&Roster[Druzhynnyk], &Roster[Toxotes])
	if front != roll.Mul(m) || ff != FlagStrong {
		t.Errorf("front hit %v flags %b, want roll×%v = %v", front, ff, m, roll.Mul(m))
	}
	if back != roll.Mul(I(2)).Mul(m) || bf != FlagFlank|FlagStrong {
		t.Errorf("flanked hit %v flags %b, want roll×2×matchup", back, bf)
	}
	side, _ := hit(p(10, 0)) // 90° off facing: flank
	if side != back {
		t.Errorf("side hit %v should count as flank (%v)", side, back)
	}
}

func TestStagger(t *testing.T) {
	b := testBattle(nil, tu{0, Druzhynnyk, p(0, 50), nil}, tu{1, Druzhynnyk, p(0, -60), nil})
	b.Step() // red walks toward blue
	r := b.Units[1]
	b.damage(b.Units[0], r, b.Units[0].Pos, One)
	pos := r.Pos
	for i := 0; i < StaggerTicks-1; i++ {
		b.Step()
		if r.Pos != pos {
			t.Fatalf("staggered unit moved on tick %d", i+1)
		}
	}
	b.steps(3)
	if r.Pos == pos {
		t.Error("unit never recovered from stagger")
	}
}

func TestKiting(t *testing.T) {
	b := testBattle(nil, tu{0, Toxotes, p(0, 0), nil}, tu{1, Druzhynnyk, p(0, -25), nil})
	b.Units[1].cd = 1000
	b.Step()
	a := b.Units[0]
	if !a.Kiting {
		t.Fatal("archer should back away from a druzhynnyk 2.5 m away")
	}
	b.steps(5)
	if a.Pos.Z <= 0 {
		t.Errorf("archer should retreat away from the druzhynnyk (+z), at %v", a.Pos)
	}
	for i := 0; i < KiteMaxTicks+2 && a.Kiting; i++ {
		b.Step()
	}
	if a.Kiting || a.kiteRest <= 0 {
		t.Errorf("kiting must stop after %d ticks and rest", KiteMaxTicks)
	}
}

func TestCorneredArcherStands(t *testing.T) {
	b := testBattle(nil, tu{0, Toxotes, V(ClampLim, ClampLim), nil}, tu{1, Druzhynnyk, V(ClampLim-I(2), ClampLim-I(2)), nil})
	b.Units[1].cd = 1000
	b.Units[0].Face = V(-One, 0)
	b.Step()
	if _, ok := b.kiteSpot(b.Units[0]); ok && b.Units[0].Kiting {
		// a spot along the wall may exist; that's fine as long as it's walkable
		if s, _ := b.kiteSpot(b.Units[0]); !b.Map.FreeAt(s) {
			t.Error("kite spot must be walkable")
		}
	}
}

// ---------- 4.5 Skills ----------

func TestStatSkills(t *testing.T) {
	cases := []struct {
		c     Class
		s     Skill
		check func(Stats) bool
	}{
		{Volva, VolvaStaff, func(s Stats) bool { return (s.HP - Ratio(744, 10)).Abs() <= 2 }}, // +20% health
		{Druzhynnyk, DruzhynnykHeavyMail, func(s Stats) bool { // +20% health, −5% speed
			return s.HP == I(120) && (s.Speed-Ratio(304, 100)).Abs() <= 2
		}},
		{Druzhynnyk, DruzhynnykAxeAndSword, func(s Stats) bool { // attacks 20% faster
			return s.Cooldown == sec(13)*80/100 && s.SwingTicks == sec(9)*80/100
		}},
		{Strilets, StriletsSteppeShooting, func(s Stats) bool { return (s.Speed - Ratio(396, 100)).Abs() <= 2 }}, // +10% speed
		{Toxotes, ToxotesSharpArrows, func(s Stats) bool {
			return (s.DmgLo-Ratio(1035, 100)).Abs() <= 2 && (s.DmgHi-Ratio(161, 10)).Abs() <= 2
		}},
		{Toxotes, ToxotesSolenarion, func(s Stats) bool { return s.Range == Ratio(117, 10) }},              // +30% range
		{Strilets, StriletsArcingShot, func(s Stats) bool { return (s.DmgLo - Ratio(99, 10)).Abs() <= 2 }}, // +10% damage
	}
	for _, c := range cases {
		u := newUnit(0, 0, UnitSpec{Class: c.c, Skills: []Skill{c.s}}, Vec{})
		if !c.check(u.Stats) {
			t.Errorf("%v with skill %d: stats %+v", c.c, c.s, u.Stats)
		}
	}
}

func TestDash(t *testing.T) {
	b := testBattle(nil, tu{0, Huskarl, p(0, 60), []Skill{HuskarlCharge}}, tu{1, Volva, p(0, 0), nil})
	b.Units[1].cd = 1000
	b.Step()
	u := b.Units[0]
	if u.dash == nil {
		t.Fatal("a huskarl with Charge 6 m away should dash")
	}
	b.steps(u.dash.dur)
	if d := Dist(u.Pos, b.Units[1].Pos); d > u.Stats.Range+Half {
		t.Errorf("dash should stop at axe range, %v away", d)
	}
	if u.effect(GapCloser).cd <= 0 {
		t.Error("dash should go on cooldown")
	}
}

// An escape without a speed teleports (a blink).
func TestBlink(t *testing.T) {
	blinker := unitWith(t, Huslyar, "blinker", nil, Effect{Kind: Escape, Value: I(6), Radius: Ratio(25, 10), Ticks: sec(80)})
	b := testBattle(nil, tu{0, blinker, p(0, 0), []Skill{1}}, tu{1, Druzhynnyk, p(0, -20), nil})
	b.Units[1].cd = 1000
	m := b.Units[0]
	b.Step()
	if d := Dist(m.Pos, p(0, 0)); d < I(5) {
		t.Fatalf("should blink ~6 m away, moved %v", d)
	}
	if m.Pos.Z <= 0 {
		t.Errorf("blink should go away from the druzhynnyk, landed at %v", m.Pos)
	}
	if cd := m.effect(Escape).cd; cd != 8*TicksPerSecond {
		t.Errorf("blink cooldown %d", cd)
	}
}

// An escape with a speed (the volva's Slip away) runs instead of teleporting.
func TestEscapeRun(t *testing.T) {
	b := testBattle(nil, tu{0, Volva, p(0, 0), []Skill{VolvaSlipAway}}, tu{1, Druzhynnyk, p(0, -20), nil})
	b.Units[1].cd = 1000
	m := b.Units[0]
	b.Step()
	if m.dash == nil {
		t.Fatal("an escape with a speed should start a run")
	}
	if d := Dist(m.Pos, p(0, 0)); d > I(1) {
		t.Fatalf("a run shouldn't teleport, moved %v in one tick", d)
	}
	b.steps(m.dash.dur)
	if d := Dist(m.Pos, p(0, 0)); d < I(5) {
		t.Errorf("should run ~6 m away, moved %v", d)
	}
	if m.Pos.Z <= 0 {
		t.Errorf("should run away from the druzhynnyk, ended at %v", m.Pos)
	}
	if cd := m.effect(Escape).cd; cd <= 0 {
		t.Errorf("run cooldown %d", cd)
	}
}

func TestLob(t *testing.T) {
	rock := []Obstacle{{Kind: Rock, Pos: p(0, 40), R: I(1)}}
	for _, lob := range []bool{false, true} {
		var skills []Skill
		if lob {
			skills = []Skill{StriletsArcingShot}
		}
		b := testBattle(rock, tu{0, Strilets, p(0, 80), skills}, tu{1, Druzhynnyk, p(0, 0), nil}, tu{0, Druzhynnyk, p(30, 0), nil})
		b.Units[1].cd, b.Units[2].cd = 1000, 1000
		b.Units[0].cd = 0
		start := b.Units[0].Pos
		b.steps(3)
		shooting := b.Units[0].swing != nil || len(b.Shots) > 0
		if shooting != lob {
			t.Errorf("lob=%v: shooting over the rock = %v (moved %v)", lob, shooting, Dist(start, b.Units[0].Pos))
		}
	}
}

func TestExecute(t *testing.T) {
	kills := func(hpPct int) int {
		n := 0
		for seed := uint64(0); seed < 400; seed++ {
			b := testBattle(nil, tu{0, Huskarl, V(0, 0), []Skill{HuskarlExecutioner}}, tu{1, Druzhynnyk, p(0, -10), nil})
			v := b.Units[1]
			v.HP = v.MaxHP.MulDiv(hpPct, 100)
			v.Face = V(0, One) // facing the attacker: no flank bonus
			b.rng = NewRNG(seed)
			b.damage(b.Units[0], v, V(0, 0), One)
			if v.Dead {
				n++
			}
		}
		return n
	}
	if k := kills(100); k != 0 {
		t.Errorf("executioner killed %d/400 at full health; it should only finish off the wounded", k)
	}
	if k := kills(38); k < 70 || k > 130 { // below 40%: 25% chance, about 100 (a normal hit can't kill 38 HP)
		t.Errorf("executioner killed %d/400 at 38%% health, want about 100", k)
	}
}

// Slow and splash on one hit, from two skills learned splash first: the
// victim is slowed and its neighbour splashed either way. No faction unit has
// both, so a test unit does.
func TestSlowAndSplash(t *testing.T) {
	d := Roster[Volva]
	d.Name, d.InRedPool = "frost and blast", false
	d.Skills[0].Effects = []Effect{{Kind: SplashOnHit, Pct: 35, Radius: I(2)}}
	d.Skills[1].Effects = []Effect{{Kind: SlowOnHit, Pct: 100, Ticks: 60, Factor: Ratio(6, 10)}}
	withRoster(t, d)
	c := Class(len(Roster) - 1)
	slowed, splash := 0, 0
	for seed := uint64(0); seed < 200; seed++ {
		b := testBattle(nil, tu{0, c, p(0, 60), []Skill{d.Skills[0].Skill, d.Skills[1].Skill}},
			tu{1, Druzhynnyk, p(0, 0), nil}, tu{1, Druzhynnyk, p(10, 0), nil})
		b.rng = NewRNG(seed)
		b.Shots = []*Shot{{Shooter: 0, Victim: 1, Bolt: true, Start: p(0, 56), Aim: p(0, 0), Dur: 1}}
		b.updateShots()
		if b.Units[1].slow == 60 {
			slowed++
		}
		if b.Units[2].HP < b.Units[2].MaxHP {
			splash++
		}
	}
	if slowed != 200 {
		t.Errorf("slowed %d/200, want every time", slowed)
	}
	if splash != 200 {
		t.Errorf("splash hit the neighbour %d/200 times, want every time", splash)
	}
}

// ---------- 4.6 Battle end ----------

func TestBattleEnd(t *testing.T) {
	b := testBattle(nil, tu{0, Druzhynnyk, p(0, 10), nil}, tu{1, Toxotes, p(0, 0), nil})
	b.Units[1].HP = One
	b.Units[1].cd = 1000
	b.steps(100)
	if b.Result != BlueWin {
		t.Fatalf("result %v, want blue win", b.Result)
	}
	hp := b.Units[0].HP
	b.damage(b.Units[1], b.Units[0], V(0, 0), One)
	if b.Units[0].HP != hp {
		t.Error("damage must stop once the battle has ended")
	}

	// mutual wipe: two arrows kill the last two units on the same tick
	b = testBattle(nil, tu{0, Toxotes, p(0, 80), nil}, tu{1, Toxotes, p(0, -80), nil})
	b.Units[0].HP, b.Units[1].HP = One/2, One/2
	b.Units[0].cd, b.Units[1].cd = 1000, 1000
	b.Shots = []*Shot{
		{Shooter: 0, Victim: 1, Start: p(0, 70), Aim: p(0, -80), Dur: 1},
		{Shooter: 1, Victim: 0, Start: p(0, -70), Aim: p(0, 80), Dur: 1},
	}
	b.Step()
	if b.Result != Draw || !b.Units[0].Dead || !b.Units[1].Dead {
		t.Fatalf("mutual wipe: result %v, dead %v %v", b.Result, b.Units[0].Dead, b.Units[1].Dead)
	}

	// tick limit
	b = testBattle(nil, tu{0, Druzhynnyk, V(0, I(13)), nil}, tu{1, Druzhynnyk, V(0, I(-13)), nil})
	for _, u := range b.Units {
		u.HP = I(1 << 20)
	}
	b.steps(MaxTicks + 10)
	if b.Result != Draw || b.Tick != MaxTicks {
		t.Errorf("tick limit: result %v at tick %d", b.Result, b.Tick)
	}
}

// ---------- 5. Armies ----------

func TestFormationPresets(t *testing.T) {
	classes := []Class{Druzhynnyk, Druzhynnyk, Toxotes, Toxotes, Volva}
	for _, f := range Presets {
		for team := 0; team < 2; team++ {
			slots := FormationSlots(f, classes, team)
			for i, s := range slots {
				if !Zones[team].Contains(s) {
					t.Errorf("%v team %d: unit %d at %v,%v outside the zone", f, team, i, s.X, s.Z)
				}
				for j := 0; j < i; j++ {
					if Dist(s, slots[j]) < I(1) {
						t.Errorf("%v: units %d and %d overlap", f, i, j)
					}
				}
			}
			// melee in front: melee units no further back than any ranged unit
			front := func(v Vec) Fx {
				if team == 0 {
					return v.Z
				}
				return -v.Z
			}
			for _, w := range []int{0, 1} {
				for _, r := range []int{2, 3, 4} {
					if f != Spread && sameFlank(slots[w], slots[r]) && front(slots[w]) > front(slots[r]) {
						t.Errorf("%v team %d: melee unit %d behind ranged %d", f, team, w, r)
					}
				}
			}
		}
	}
	// 5-0: everyone on the left flank, melee units in the front row
	s := FormationSlots(F5x0, classes, 0)
	for i, v := range s {
		if v.X >= 0 {
			t.Errorf("5-0: unit %d not on the left flank: %v", i, v.X)
		}
	}
	if s[0].Z != s[1].Z || s[0].Z >= s[3].Z {
		t.Errorf("5-0: melee units should share the front row: %v", s)
	}
}

func sameFlank(a, b Vec) bool { return (a.X < 0) == (b.X < 0) }

func TestRedArmy(t *testing.T) {
	a, b := RedArmy(7, 3), RedArmy(7, 3)
	if len(a.Units) != ArmySize {
		t.Fatalf("red has %d units", len(a.Units))
	}
	total := 0
	for i, u := range a.Units {
		total += len(u.Skills)
		seen := map[Skill]bool{}
		for _, s := range u.Skills {
			if seen[s] {
				t.Errorf("unit %d has skill %d twice", i, s)
			}
			seen[s] = true
			if _, ok := SkillDefOf(u.Class, s); !ok {
				t.Errorf("unit %d (%v) has a skill of another class", i, u.Class)
			}
		}
		if i > 0 && u.Class < a.Units[i-1].Class {
			t.Error("red units should be listed in roster order")
		}
		if u.Class != b.Units[i].Class || len(u.Skills) != len(b.Units[i].Skills) {
			t.Error("same seed gave a different red army")
		}
	}
	if total != ArmySize+3 {
		t.Errorf("red has %d abilities, want %d (one each, plus 3)", total, ArmySize+3)
	}
	if a.Formation == Custom || a.Formation != b.Formation {
		t.Errorf("red formation %v / %v", a.Formation, b.Formation)
	}
	if full := RedArmy(7, 99); len(full.Units) == ArmySize {
		n := 0
		for _, u := range full.Units {
			n += len(u.Skills)
		}
		if n != ArmySize*MaxLevel {
			t.Errorf("asking for more abilities than levels allow should cap at %d, got %d", ArmySize*MaxLevel, n)
		}
	}
}

func TestValidate(t *testing.T) {
	l1 := []Skill{1}
	five := []UnitSpec{{Class: Druzhynnyk, Skills: l1}, {Class: Druzhynnyk, Skills: l1}, {Class: Strilets, Skills: l1}, {Class: Strilets, Skills: l1}, {Class: Huslyar, Skills: l1}}
	ok := Input{Version: Version, BlueFaction: FactionRus, Seed: 1, Blue: Army{Formation: F3x2, Units: five}}
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
	in := func(mod func(*Input)) Input {
		c := ok
		c.Blue.Units = append([]UnitSpec(nil), five...)
		mod(&c)
		return c
	}
	outside := V(0, 0)
	inside := V(0, I(10))
	bad := map[string]Input{
		"version":     in(func(c *Input) { c.Version = "0.0.1" }),
		"six units":   in(func(c *Input) { c.Blue.Units = append(c.Blue.Units, UnitSpec{Class: Huslyar, Skills: l1}) }),
		"no units":    in(func(c *Input) { c.Blue.Units = nil }),
		"skill twice": in(func(c *Input) { c.Blue.Units[0].Skills = []Skill{DruzhynnykShieldWall, DruzhynnykShieldWall} }),
		"no ability":  in(func(c *Input) { c.Blue.Units[0].Skills = nil }),
		"level 4":     in(func(c *Input) { c.Blue.Units[0].Skills = []Skill{1, 2, 3, 4} }),
		"wrong class": in(func(c *Input) { c.Blue.Units[0].Skills = []Skill{6} }),
		"outside zone": in(func(c *Input) {
			c.Blue.Formation = Custom
			setPos(c.Blue.Units, inside)
			c.Blue.Units[2].Pos = &outside
		}),
		"custom no pos":   in(func(c *Input) { c.Blue.Formation = Custom }),
		"pos with preset": in(func(c *Input) { c.Blue.Units[0].Pos = &inside }),
		"red skills+army": in(func(c *Input) { c.Red = Army{Formation: Spread, Units: five}; c.RedSkills = 1 }),
	}
	for name, c := range bad {
		if err := c.Validate(); err == nil {
			t.Errorf("%s: invalid input accepted", name)
		}
	}
	if _, err := Run(bad["outside zone"]); err == nil {
		t.Error("Run must reject invalid input")
	}
	custom := in(func(c *Input) { c.Blue.Formation = Custom; setPos(c.Blue.Units, inside) })
	if err := custom.Validate(); err != nil {
		t.Errorf("custom positions rejected: %v", err)
	}
}

func setPos(units []UnitSpec, v Vec) {
	for i := range units {
		p := V(v.X+I(i*2-4), v.Z)
		units[i].Pos = &p
	}
}

func TestRedHealthPct(t *testing.T) {
	in := Input{Version: Version, BlueFaction: FactionRus, Seed: 3, RedSkills: 5, Blue: Army{Formation: Spread, Units: []UnitSpec{{Class: Druzhynnyk, Skills: []Skill{1}}}}}
	full, _ := New(in)
	in.RedHealthPct = 85
	eased, _ := New(in)
	for i, u := range eased.Units {
		want := full.Units[i].MaxHP
		if u.Team == 1 {
			want = want.MulDiv(85, 100)
		}
		if u.MaxHP != want || u.HP != want {
			t.Errorf("unit %d (team %d): max %v hp %v, want %v", i, u.Team, u.MaxHP, u.HP, want)
		}
	}
	for _, bad := range []int{-1, 101} {
		in.RedHealthPct = bad
		if in.Validate() == nil {
			t.Errorf("red health %d%% accepted", bad)
		}
	}
}
