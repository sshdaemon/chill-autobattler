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
	return nil
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
