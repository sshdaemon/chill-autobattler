package sim

type Unit struct {
	Index  int
	Team   int
	Class  Class
	Skills []Skill
	Stats  Stats // with skill bonuses applied

	Pos      Vec
	Face     Vec // unit vector
	MaxHP    Fx
	HP       Fx
	Dead     bool
	Target   int // index of the unit it's fighting, or -1
	Kills    int
	Dealt    Fx
	Wet      bool
	Kiting   bool
	Slowed   bool
	Swinging bool

	cd, busy, retarget, dashCd, blinkCd, slow int // tick counters
	kiteTime, kiteRest, pathTimer             int

	swing       *swing
	dash        *dash
	moveTarget  *Vec
	path        []Vec
	investigate *Vec
	advanceGoal *Vec
}

type swing struct {
	t, dur int
	hit    bool
	victim int
}

type dash struct {
	from, to Vec
	t, dur   int
}

func newUnit(index, team int, spec UnitSpec, pos Vec) *Unit {
	st := ClassStats[spec.Class]
	pct := func(s Skill) int { // 100 + the skill's bonus, if the unit has it
		if !hasSkill(spec.Skills, s) {
			return 100
		}
		d, _ := SkillDefOf(spec.Class, s)
		return 100 + d.Pct
	}
	st.HP = st.HP.MulDiv(pct(SkillHealth), 100)
	st.DmgLo = st.DmgLo.MulDiv(pct(SkillDamage), 100)
	st.DmgHi = st.DmgHi.MulDiv(pct(SkillDamage), 100)
	st.Range = st.Range.MulDiv(pct(SkillRange), 100)
	st.Speed = st.Speed.MulDiv(pct(SkillSpeed), 100)
	face := V(0, -One) // Blue faces up the field (-z), Red down
	if team == 1 {
		face = V(0, One)
	}
	return &Unit{
		Index: index, Team: team, Class: spec.Class, Skills: append([]Skill(nil), spec.Skills...), Stats: st,
		Pos: pos, Face: face, MaxHP: st.HP, HP: st.HP, Target: -1,
	}
}

func (u *Unit) Has(s Skill) bool { return hasSkill(u.Skills, s) }

func (u *Unit) skill(s Skill) SkillDef {
	d, _ := SkillDefOf(u.Class, s)
	return d
}

// think runs one tick of a living unit's decisions (prototype's step loop).
func (b *Battle) think(u *Unit) {
	u.cd--
	u.busy--
	u.retarget--
	u.dashCd--
	u.blinkCd--

	if u.Target < 0 || b.Units[u.Target].Dead || u.retarget <= 0 {
		prev := u.Target
		u.Target = b.nearestEnemy(u)
		u.retarget = RetargetTicks
		if u.Target != prev {
			u.path = nil
		}
		if u.Target < 0 && prev >= 0 && !b.Units[prev].Dead { // lost sight: check where it was
			p := b.Units[prev].Pos
			u.investigate = &p
		}
		if u.Target >= 0 {
			u.investigate = nil
		}
	}

	if u.dash != nil { // a quick straight rush; nothing else happens meanwhile
		u.dash.t++
		u.Pos = Lerp(u.dash.from, u.dash.to, u.dash.t, u.dash.dur)
		u.moveTarget = nil
		if u.dash.t >= u.dash.dur {
			u.dash = nil
		}
		return
	}

	if u.Has(SkillBlink) && u.blinkCd <= 0 && b.tryBlink(u) {
		return
	}

	if u.swing != nil {
		s := u.swing
		s.t++
		if !s.hit && s.t*100 >= HitAtPct*s.dur {
			s.hit = true
			if v := b.Units[s.victim]; !v.Dead {
				if u.Stats.Ranged {
					b.shoot(u, v)
				} else if Dist(u.Pos, v.Pos) < u.Stats.Range+MeleeSlack {
					b.damage(u, v, u.Pos, One)
				}
			}
		}
		if s.t >= s.dur {
			u.swing = nil
		}
		return
	}
	if u.busy > 0 { // staggered
		u.moveTarget = nil
		return
	}
	if u.Target < 0 {
		u.Kiting = false
		b.advance(u)
		return
	}
	t := b.Units[u.Target]

	if u.Stats.Kite && b.kite(u) {
		return
	}

	d := Dist(u.Pos, t.Pos)
	if u.Has(SkillDash) && u.dashCd <= 0 && d > u.Stats.Range+DashMinGap && d <= u.skill(SkillDash).Value && b.Map.LineOfSight(u.Pos, t.Pos) {
		dir := t.Pos.Sub(u.Pos).Norm()
		reach := d - u.Stats.Range.MulDiv(8, 10)
		dur := reach.Div(perTick(DashSpeed)).Ceil()
		if dur < 1 {
			dur = 1
		}
		u.dash = &dash{from: u.Pos, to: u.Pos.Add(dir.Scale(reach)), dur: dur}
		u.Face = dir
		u.dashCd = DashCooldown
		u.path = nil
		b.emit(Event{Kind: EvDash, Unit: u.Index, Other: -1})
		return
	}

	// ranged units need a clear line of fire (unless they lob over); otherwise they walk until they have one
	canHit := d <= u.Stats.Range && (!u.Stats.Ranged || u.Has(SkillLob) || b.Map.LineOfFire(u.Pos, t.Pos))
	if !canHit {
		b.moveToward(u, t.Pos)
		return
	}
	u.moveTarget, u.path = nil, nil
	want := t.Pos.Sub(u.Pos).Norm()
	aimed := u.Face.Dot(want) > CosAim
	u.Face = TurnToward(u.Face, want, TurnPerTick)
	if u.cd <= 0 && aimed {
		u.swing = &swing{dur: u.Stats.SwingTicks, victim: t.Index}
		jitter := 100 - CooldownJitter + b.rng.IntN(2*CooldownJitter+1)
		u.cd = u.Stats.Cooldown * jitter / 100
	}
}

// tryBlink teleports a mage away from nearby melee. It reports whether the
// mage blinked (which ends its turn).
func (b *Battle) tryBlink(u *Unit) bool {
	var sum Vec
	n := 0
	for _, e := range b.Units {
		if !e.Dead && e.Team != u.Team && !e.Stats.Ranged && Within(u.Pos, e.Pos, BlinkPanic) {
			sum = sum.Add(e.Pos)
			n++
		}
	}
	if n == 0 {
		return false
	}
	centre := V(sum.X/Fx(n), sum.Z/Fx(n))
	base := u.Pos.Sub(centre).Norm()
	if base.IsZero() {
		base = V(0, One)
	}
	for _, turn := range BlinkTurns {
		p := u.Pos.Add(base.Rot(turn).Scale(BlinkDist))
		if p.X.Abs() > EdgeLimit || p.Z.Abs() > EdgeLimit || !b.Map.FreeAt(p) {
			continue
		}
		b.emit(Event{Kind: EvBlink, Unit: u.Index, Other: -1, Amount: 0})
		u.Pos = p
		u.moveTarget, u.path, u.swing = nil, nil, nil
		u.blinkCd = u.skill(SkillBlink).Value.Floor() * TicksPerSecond
		return true
	}
	return false
}

// kite makes an archer back away from nearby melee, in bursts. It reports
// whether the archer is retreating this tick.
func (b *Battle) kite(u *Unit) bool {
	u.kiteRest--
	if u.Kiting {
		u.kiteTime++
		if u.kiteTime > KiteMaxTicks {
			u.Kiting, u.kiteTime, u.kiteRest = false, 0, KiteRestTicks
			u.moveTarget = nil
		}
	}
	if u.kiteRest <= 0 {
		if spot, ok := b.kiteSpot(u); ok {
			u.Kiting = true
			u.path = nil
			u.moveTarget = &spot
			return true
		}
	}
	if u.Kiting {
		u.Kiting = false
		u.moveTarget = nil
	}
	return false
}

// kiteSpot sums a flee direction from every melee enemy within panic range,
// then picks the first walkable step along it (or angled alternatives).
// A cornered archer gets no spot and stands and shoots.
func (b *Battle) kiteSpot(u *Unit) (Vec, bool) {
	var flee Vec
	for _, e := range b.Units {
		if !e.Dead && e.Team != u.Team && !e.Stats.Ranged && Within(u.Pos, e.Pos, KitePanic) {
			flee = flee.Add(u.Pos.Sub(e.Pos).Norm())
		}
	}
	if flee.Len() < Ratio(1, 1000) {
		return Vec{}, false
	}
	base := flee.Norm()
	for _, turn := range KiteTurns {
		p := u.Pos.Add(base.Rot(turn).Scale(KiteStep))
		if b.Map.FreeAt(p) && b.Map.LineOfSight(u.Pos, p) {
			return p, true
		}
	}
	return Vec{}, false
}

// advance: nothing in sight. Check the last place an enemy was seen,
// otherwise march at the enemy's edge; once there, sweep random spots in
// their half.
func (b *Battle) advance(u *Unit) {
	edge := EnemyEdge
	if u.Team == 0 {
		edge = -EnemyEdge
	}
	if u.investigate == nil && u.advanceGoal == nil {
		g := b.Map.FreeSpot(V(u.Pos.X, edge))
		u.advanceGoal = &g
	}
	goal := u.advanceGoal
	if u.investigate != nil {
		goal = u.investigate
	}
	if Within(*goal, u.Pos, AdvanceReach) {
		if u.investigate != nil {
			u.investigate = nil
		} else {
			x := (b.rng.Frac() - Half).Mul(I(22))
			g := b.Map.FreeSpot(V(x, edge.Mul(b.rng.Frac())))
			u.advanceGoal = &g
		}
		u.path = nil
		return
	}
	b.moveToward(u, *goal)
}

// moveToward heads for a point, via a path when the straight line is blocked
// or wades through water it could go round.
func (b *Battle) moveToward(u *Unit, goal Vec) {
	u.pathTimer--
	if u.path == nil || u.pathTimer <= 0 {
		if b.Map.LineOfSight(u.Pos, goal) && (b.Map.InWater(goal) || b.Map.WetLength(u.Pos, goal) == 0) {
			u.path = []Vec{goal}
		} else {
			u.path = b.Map.FindPath(u.Pos, goal)
		}
		u.pathTimer = RetargetTicks
	}
	if len(u.path) == 0 {
		return
	}
	u.path[len(u.path)-1] = goal // the last leg follows a moving goal
	if len(u.path) > 1 && Within(u.path[0], u.Pos, WaypointReach) {
		u.path = u.path[1:]
	}
	wp := u.path[0]
	u.moveTarget = &wp
}

var arriveDist = Ratio(5, 100)

// move walks a unit toward its move target: turning at 8 rad/s, slower while
// turning sharply, slowed by water and frost.
func (b *Battle) move(u *Unit) {
	u.Wet = b.Map.InWater(u.Pos)
	if u.slow > 0 {
		u.slow--
	}
	u.Slowed = u.slow > 0
	u.Swinging = u.swing != nil
	if u.moveTarget == nil {
		return
	}
	d := u.moveTarget.Sub(u.Pos)
	dist := d.Len()
	if dist < arriveDist {
		u.moveTarget = nil
		return
	}
	want := d.Norm()
	sharp := u.Face.Dot(want) < CosSlowTurn
	u.Face = TurnToward(u.Face, want, TurnPerTick)
	speed := perTick(u.Stats.Speed)
	if u.Wet {
		speed = speed.Mul(WaterSpeed)
	}
	if u.Slowed {
		speed = speed.Mul(FrostSpeed)
	}
	if sharp {
		speed = speed.Mul(SlowTurnSpeed)
	}
	u.Pos = u.Pos.Add(want.Scale(Min(dist, speed)))
}

// Moving reports whether the unit is walking this tick (for animation).
func (u *Unit) Moving() bool { return u.moveTarget != nil || u.dash != nil }

func (u *Unit) Dashing() bool { return u.dash != nil }

func (u *Unit) Staggered() bool { return u.busy > 0 }
