package sim

import "fmt"

// Ability effects are plug-ins. Each kind of effect lives in its own file
// (effect_<name>.go): it registers itself under its id from ids.go with a
// check for its numbers and the hooks it needs. The battle calls the hooks of
// a unit's learned effects in kind-id order, so a new kind never changes how
// existing ones play, and adding one edits no shared code.

// effectKind is one kind of effect. Every hook is optional.
type effectKind struct {
	name string
	// check validates an effect's numbers for a roster entry.
	check func(e Effect, d *UnitDef) error
	// overCover: the unit's shots need no clear line of fire.
	overCover bool
	// temper changes the unit's temperament at spawn (e.g. a hunter targets ranged units first).
	temper func(t *Temperament, e Effect)
	// beforeAct runs before the unit swings or moves; true ends its turn (escape).
	beforeAct func(b *Battle, u *Unit, a *active) bool
	// approach runs when the unit has a target at distance d; true ends its turn (gap-closer).
	approach func(b *Battle, u, t *Unit, d Fx, a *active) bool
	// onBlow can change a hit's damage before it lands (finisher); scale is
	// the hit's share (One for a direct hit, less for splash).
	onBlow func(b *Battle, att, v *Unit, scale, amount Fx, a *active) Fx
	// onShotLands runs after the unit's projectile hit its victim (slow, splash).
	onShotLands func(b *Battle, u, v *Unit, s *Shot, a *active)
}

// active is one learned effect on a unit, with its state.
type active struct {
	kind *effectKind
	id   EffectKind
	e    *Effect
	cd   int // ticks until it can act again (counts down every tick)
}

var effectKinds [256]*effectKind

func registerEffect(id EffectKind, k *effectKind) {
	if effectKinds[id] != nil {
		panic(fmt.Sprintf("effect kind %d registered twice (%s, %s): allocate ids in ids.go", id, effectKinds[id].name, k.name))
	}
	effectKinds[id] = k
}

func (k EffectKind) String() string {
	if ek := effectKinds[k]; ek != nil {
		return ek.name
	}
	return fmt.Sprintf("effect(%d)", k)
}

// checkEffect validates one effect of a roster entry.
func checkEffect(e Effect, d *UnitDef) error {
	ek := effectKinds[e.Kind]
	if ek == nil {
		return fmt.Errorf("unknown effect kind %d", e.Kind)
	}
	if ek.check == nil {
		return nil
	}
	return ek.check(e, d)
}

// pctIn checks an effect's percentage.
func pctIn(e Effect, lo, hi int) error {
	if e.Pct < lo || e.Pct > hi {
		return fmt.Errorf("%v: %d%% outside %d..%d", e.Kind, e.Pct, lo, hi)
	}
	return nil
}

// needsRanged is the check for effects that act on projectiles.
func needsRanged(e Effect, d *UnitDef) error {
	if !d.Stats.Ranged {
		return fmt.Errorf("%v needs a ranged unit (it acts on projectiles)", e.Kind)
	}
	return nil
}

// loadout is a unit's learned non-stat effects, folded once at spawn so the
// battle walks a short list instead of scanning skills every tick.
type loadout struct {
	active    []active // in kind-id order (at most one of each kind)
	overCover bool
}

// learn adds the non-stat effects of the unit's skills to its loadout.
func (kit *loadout) learn(def *UnitDef, skills []Skill) {
	for _, sd := range def.Skills {
		if !hasSkill(skills, sd.Skill) {
			continue
		}
		for i := range sd.Effects {
			e := &sd.Effects[i]
			if e.Kind == StatBonus {
				continue
			}
			ek := effectKinds[e.Kind]
			at := len(kit.active)
			for at > 0 && kit.active[at-1].id > e.Kind {
				at--
			}
			kit.active = append(kit.active, active{})
			copy(kit.active[at+1:], kit.active[at:])
			kit.active[at] = active{kind: ek, id: e.Kind, e: e}
			kit.overCover = kit.overCover || ek.overCover
		}
	}
}

// effect returns the unit's learned effect of a kind, if it has one.
func (u *Unit) effect(k EffectKind) *active {
	for i := range u.kit.active {
		if u.kit.active[i].id == k {
			return &u.kit.active[i]
		}
	}
	return nil
}
