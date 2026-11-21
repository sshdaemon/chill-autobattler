package sim

type Unit struct {
	Index  int
	Team   int
	Class  Class
	Def    *UnitDef // roster entry
	Skills []Skill
	Stats  Stats       // with skill bonuses applied
	Temper Temperament // with abilities applied

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
	Group    int // its group in a tactic (0 left, 1 right; 0 without one)

	cd, busy, retarget, slow, fear int // tick counters
	held                           int // ticks a holding unit has waited for an enemy
	lineAhead                      Fx  // EngageLine: how far ahead of its allies' centre it stood at the start
	lineSet                        bool
	flankSide                      Fx   // EngageFlank: which side it goes round (+1/-1 in x; 0 not yet chosen)
	flanked                        bool // EngageFlank: it has gone round, or got close
	guarding                       bool // StayNear bodyguard: heading back to its ward
	keeping                        bool // Keep: stepping back from its target
	pathGoal                       Vec  // the goal u.path was found for
	fearFrom                       Vec  // where dread drives it away from
	mods                           mods // this tick's modifiers (mods.go)
	buffs                          []buff
	slowFactor                     Fx // speed factor while slow > 0
	kit                            loadout
	kiteTime, kiteRest, pathTimer  int

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
	return newUnitPlan(index, team, spec, pos, PlanInstinct)
}

// newUnitPlan is newUnit for a side fighting by a plan.
func newUnitPlan(index, team int, spec UnitSpec, pos Vec, plan Plan) *Unit {
	return newUnitGroup(index, team, spec, pos, plan, 0)
}

// newUnitGroup is newUnitPlan for a unit of a tactic's group g.
func newUnitGroup(index, team int, spec UnitSpec, pos Vec, plan Plan, g int) *Unit {
	def := RosterOf(spec.Class)
	p := PlanOf(plan)
	var extra []Effect
	if p != nil {
		extra = p.effectsFor(def, g)
	}
	st := foldStats(def, spec.Skills, extra...)
	var kit loadout
	kit.learn(def, spec.Skills, extra...)
	face := V(0, -One) // Blue faces up the field (-z), Red down
	if team == 1 {
		face = V(0, One)
	}
	return &Unit{
		Index: index, Team: team, Class: spec.Class, Def: def, Skills: append([]Skill(nil), spec.Skills...), Stats: st, Temper: temperOf(def, p, g, false, &kit),
		Pos: pos, Face: face, MaxHP: st.HP, HP: st.HP, Target: -1, kit: kit, Group: g,
	}
}

func (u *Unit) Has(s Skill) bool { return hasSkill(u.Skills, s) }

// think runs one tick of a living unit's decisions (prototype's step loop).
func (b *Battle) think(u *Unit) {
	u.cd--
	u.busy--
	u.retarget--
	for i := range u.kit.active {
		u.kit.active[i].cd--
	}

	if u.Target < 0 || b.Units[u.Target].Dead || u.retarget <= 0 {
		prev := u.Target
		u.Target = targetRules[u.Temper.Target].pick(b, u)
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

	if u.fear > 0 && u.dash == nil { // dread: run away, nothing else
		u.swing = nil
		if p, ok := b.stepAway(u, u.fearFrom); ok {
			u.path, u.moveTarget = nil, &p
		}
		return
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

	for i := range u.kit.active {
		if a := &u.kit.active[i]; a.kind.beforeAct != nil && a.kind.beforeAct(b, u, a) {
			return
		}
	}

	if u.swing != nil {
		s := u.swing
		s.t++
		if !s.hit && s.t*100 >= HitAtPct*s.dur {
			s.hit = true
			if v := b.Units[s.victim]; !v.Dead {
				if u.Stats.Ranged {
					b.shoot(u, v)
				} else if !b.strike(u, v) && Dist(u.Pos, v.Pos) < u.Stats.Range+MeleeSlack {
					b.damage(u, v, u.Pos, One)
					b.landed(u, v, u.Pos)
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
		b.idle(u)
		return
	}
	t := b.Units[u.Target]
	if b.behave(u, t) {
		return
	}

	if u.Temper.Kite && b.kite(u) {
		if u.kit.kiteShoot && u.cd <= 0 && Within(u.Pos, t.Pos, u.Stats.Range) && (u.kit.overCover || b.Map.LineOfFire(u.Pos, t.Pos)) {
			b.shoot(u, t) // looses an arrow while backing away
			u.cd = cooldownOf(u.Stats.Cooldown, u.mods[ModAttackSpeed])
		}
		return
	}

	d := Dist(u.Pos, t.Pos)
	for i := range u.kit.active {
		if a := &u.kit.active[i]; a.kind.approach != nil && a.kind.approach(b, u, t, d, a) {
			return
		}
	}

	// ranged units need a clear line of fire (unless they lob over); otherwise they walk until they have one
	canHit := d <= u.Stats.Range && (!u.Stats.Ranged || u.kit.overCover || b.Map.LineOfFire(u.Pos, t.Pos))
	if !canHit {
		if r := engageRules[u.Temper.Engage]; r.chase == nil || !r.chase(b, u, t) {
			b.moveToward(u, t.Pos)
		}
		return
	}
	u.moveTarget, u.path = nil, nil
	want := t.Pos.Sub(u.Pos).Norm()
	aimed := u.Face.Dot(want) > CosAim
	u.Face = TurnToward(u.Face, want, TurnPerTick)
	if u.cd <= 0 && aimed {
		u.swing = &swing{dur: u.Stats.SwingTicks, victim: t.Index}
		jitter := 100 - CooldownJitter + b.rng.IntN(2*CooldownJitter+1)
		u.cd = cooldownOf(u.Stats.Cooldown*jitter/100, u.mods[ModAttackSpeed])
	}
}

// strike lets an effect replace the unit's blow (a cone attack); false if none did.
func (b *Battle) strike(u, v *Unit) bool {
	for i := range u.kit.active {
		if a := &u.kit.active[i]; a.kind.strike != nil && a.kind.strike(b, u, v, a) {
			return true
		}
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
		g := b.freeSpot(u.Team, V(u.Pos.X, edge))
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
			g := b.freeSpot(u.Team, V(x, edge.Mul(b.rng.Frac())))
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
	goal = b.reachable(u.Team, goal)
	u.pathTimer--
	if u.path == nil || u.pathTimer <= 0 || !Within(goal, u.pathGoal, PathGoalSlack) {
		if b.Map.LineOfSight(u.Pos, goal) && (b.Map.InWater(goal) || b.Map.WetLength(u.Pos, goal) == 0) {
			u.path = []Vec{goal}
		} else {
			u.path = b.findPath(u.Team, u.Pos, goal)
		}
		u.pathTimer, u.pathGoal = RetargetTicks, goal
	}
	if len(u.path) == 0 {
		return
	}
	u.path[len(u.path)-1] = goal // the last leg follows a goal that drifts a little
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
	speed := pctOf(perTick(u.Stats.Speed), u.mods[ModSpeed])
	if u.Kiting { // archers back away facing their attackers, a little slower
		speed = speed.Mul(BackpedalSpeed)
	} else {
		if u.Face.Dot(want) < CosSlowTurn {
			speed = speed.Mul(SlowTurnSpeed)
		}
		u.Face = TurnToward(u.Face, want, TurnPerTick)
	}
	if u.Wet {
		speed = speed.Mul(WaterSpeed)
	}
	if u.Slowed {
		speed = speed.Mul(u.slowFactor)
	}
	u.Pos = u.Pos.Add(want.Scale(Min(dist, speed)))
}

// Moving reports whether the unit is walking this tick (for animation).
func (u *Unit) Moving() bool { return u.moveTarget != nil || u.dash != nil }

func (u *Unit) Dashing() bool { return u.dash != nil }

func (u *Unit) Staggered() bool { return u.busy > 0 }
