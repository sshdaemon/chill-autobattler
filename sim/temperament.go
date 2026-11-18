package sim

import "fmt"

// Temperament is how a unit fights on its own: whom it picks as a target,
// what it does while it has none, and whether it backs away from melee. It
// is data in the unit's roster entry, and learned abilities may change it
// (effectKind.temper, applied at spawn in kind-id order). Target and engage
// rules are plug-ins like effects: each lives in its own file (target_<name>.go,
// engage_<name>.go) and registers itself under its id from ids.go.
type Temperament struct {
	Target TargetRule
	Engage EngageRule
	Kite   bool // backs away from nearby melee in bursts (ranged units only)
	// The fields below are off at zero (today's behaviour).
	Trigger  Fx  // EngageHold: an enemy this close wakes it
	Leash    Fx  // doesn't chase further than this from its allies' centre
	Keep     Fx  // steps back from its target when it comes closer than this
	FallBack int // below this % of its health it retreats behind its allies
	StayNear Tag // with no target, keeps near allies with this tag (ranged: bodyguard; melee: behind the front line)
}

type targetRule struct {
	name string
	// pick returns the index of the enemy to fight, or -1.
	pick func(b *Battle, u *Unit) int
}

type engageRule struct {
	name string
	// idle moves a unit that has no target.
	idle func(b *Battle, u *Unit)
	// engage reports whether the unit goes for the target it has (nil: always).
	engage func(b *Battle, u, t *Unit) bool
}

var (
	targetRules [256]*targetRule
	engageRules [256]*engageRule
)

func registerTarget(id TargetRule, r *targetRule) {
	if targetRules[id] != nil {
		panic(fmt.Sprintf("target rule %d registered twice: allocate ids in ids.go", id))
	}
	targetRules[id] = r
}

func registerEngage(id EngageRule, r *engageRule) {
	if engageRules[id] != nil {
		panic(fmt.Sprintf("engage rule %d registered twice: allocate ids in ids.go", id))
	}
	engageRules[id] = r
}

func (r TargetRule) String() string {
	if t := targetRules[r]; t != nil {
		return t.name
	}
	return fmt.Sprintf("target(%d)", r)
}

func (r EngageRule) String() string {
	if e := engageRules[r]; e != nil {
		return e.name
	}
	return fmt.Sprintf("engage(%d)", r)
}

// checkTemperament validates a roster entry's temperament.
func checkTemperament(d *UnitDef) error {
	t := d.Temper
	if targetRules[t.Target] == nil {
		return fmt.Errorf("unknown target rule %d", t.Target)
	}
	if engageRules[t.Engage] == nil {
		return fmt.Errorf("unknown engage rule %d", t.Engage)
	}
	if t.Kite && !d.Stats.Ranged {
		return fmt.Errorf("only ranged units kite")
	}
	if t.Trigger < 0 || t.Leash < 0 || t.Keep < 0 || t.FallBack < 0 || t.FallBack > 90 {
		return fmt.Errorf("temperament %+v: distances must be >= 0 and fall back 0..90%%", t)
	}
	if t.Engage == EngageHold && t.Trigger <= 0 {
		return fmt.Errorf("a holding unit needs a trigger distance")
	}
	var declared Tag
	for _, x := range Tags {
		declared |= x.Tag
	}
	if t.StayNear&^declared != 0 {
		return fmt.Errorf("stay near an undeclared tag %b", t.StayNear)
	}
	return nil
}

// Behaviour helpers shared by the rules and the think loop.

// allyCentre is the mean position of a unit's living allies (not itself); ok
// is false when it has none.
func (b *Battle) allyCentre(u *Unit) (Vec, bool) { return b.centreOf(u, false) }

// centreOf is allyCentre, without the allies that follow the front line
// (StayNear melee) when leaders is set: those wait for the front, so the front
// mustn't wait for them.
func (b *Battle) centreOf(u *Unit, leaders bool) (Vec, bool) {
	var sum Vec
	n := 0
	for _, a := range b.Units {
		if leaders && a.Temper.StayNear&TagMelee != 0 {
			continue
		}
		if a != u && !a.Dead && a.Team == u.Team {
			sum = sum.Add(a.Pos)
			n++
		}
	}
	if n == 0 {
		return Vec{}, false
	}
	return V(sum.X/Fx(n), sum.Z/Fx(n)), true
}

// teamFighting reports whether any living ally of u is within reach of its target.
func (b *Battle) teamFighting(u *Unit) bool {
	for _, a := range b.Units {
		if a != u && !a.Dead && a.Team == u.Team && a.Target >= 0 && Within(a.Pos, b.Units[a.Target].Pos, a.Stats.Range+One) {
			return true
		}
	}
	return false
}

// back is the direction toward the unit's own side of the field.
func back(u *Unit) Vec {
	if u.Team == 0 {
		return V(0, One)
	}
	return V(0, -One)
}

// nearestSpotted returns the closest enemy u's team sees that passes keep, or -1.
func (b *Battle) nearestSpotted(u *Unit, keep func(e *Unit) bool) int {
	best, bd := -1, Fx(0)
	for _, i := range b.spotted[u.Team] {
		e := b.Units[i]
		if keep != nil && !keep(e) {
			continue
		}
		if d := Dist(u.Pos, e.Pos); best < 0 || d < bd {
			best, bd = i, d
		}
	}
	return best
}

// stepAway is the first free spot KiteStep away from `from`, or false.
func (b *Battle) stepAway(u *Unit, from Vec) (Vec, bool) {
	dir := u.Pos.Sub(from)
	if dir.Len() < Ratio(1, 1000) {
		dir = back(u)
	}
	base := dir.Norm()
	for _, turn := range KiteTurns {
		p := u.Pos.Add(base.Rot(turn).Scale(KiteStep))
		if b.Map.FreeAt(p) && b.Map.LineOfSight(u.Pos, p) {
			return p, true
		}
	}
	return Vec{}, false
}

// behave applies the temperament's fields before the unit fights its target
// t: falling back, keeping its distance, holding, and the leash. It reports
// whether it took the unit's turn.
func (b *Battle) behave(u, t *Unit) bool {
	tp := u.Temper
	if tp.FallBack > 0 && u.HP.MulDiv(100, tp.FallBack) < u.MaxHP {
		if c, ok := b.allyCentre(u); ok && Within(u.Pos, t.Pos, u.Stats.Sight) {
			away := c.Sub(t.Pos)
			if away.Len() > Ratio(1, 1000) {
				c = c.Add(away.Norm().Scale(I(3)))
			}
			b.moveToward(u, c)
			return true
		}
	}
	if tp.Keep > 0 && u.cd > 0 && Dist(u.Pos, t.Pos) < tp.Keep-Half {
		if p, ok := b.stepAway(u, t.Pos); ok {
			u.path, u.moveTarget = nil, &p
			return true
		}
	}
	if r := engageRules[tp.Engage]; r.engage != nil && !r.engage(b, u, t) {
		u.moveTarget, u.path = nil, nil
		u.Face = TurnToward(u.Face, t.Pos.Sub(u.Pos).Norm(), TurnPerTick)
		return true
	}
	if tp.Leash > 0 {
		if c, ok := b.allyCentre(u); ok && !Within(u.Pos, c, tp.Leash) && !Within(t.Pos, c, tp.Leash) && !Within(u.Pos, t.Pos, u.Stats.Range) {
			b.moveToward(u, c)
			return true
		}
	}
	return false
}

// idle moves a unit without a target as its engage rule says, kept near its
// StayNear allies: a unit that stays behind the front line (melee) follows its
// nearest melee ally, 3 m behind it; a bodyguard (any other tag) advances
// normally but turns back when it gets further than StayReach from its ward.
// Wards never wait for their guards, so an army always moves.
func (b *Battle) idle(u *Unit) {
	if tag := u.Temper.StayNear; tag != 0 {
		best, bd := (*Unit)(nil), Fx(0)
		for _, a := range b.Units {
			if a != u && !a.Dead && a.Team == u.Team && a.Def.Tags&tag != 0 {
				if d := Dist(u.Pos, a.Pos); best == nil || d < bd {
					best, bd = a, d
				}
			}
		}
		if best != nil && tag&TagMelee != 0 { // behind the front line: the melee leads
			if goal := best.Pos.Add(back(u).Scale(I(3))); !Within(u.Pos, goal, Ratio(15, 10)) {
				b.moveToward(u, goal)
			} else {
				u.moveTarget, u.path = nil, nil
			}
			return
		}
		if best != nil && !Within(u.Pos, best.Pos, StayReach) { // a bodyguard too far from its ward
			b.moveToward(u, best.Pos)
			return
		}
	}
	engageRules[u.Temper.Engage].idle(b, u)
}

// temperOf is the unit's temperament with its learned abilities applied.
func temperOf(def *UnitDef, kit *loadout) Temperament {
	t := def.Temper
	for i := range kit.active {
		if a := &kit.active[i]; a.kind.temper != nil {
			a.kind.temper(&t, *a.e)
		}
	}
	return t
}
