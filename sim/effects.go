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
	// stacks: a unit may have several effects of this kind (they add up);
	// other kinds at most one per unit.
	stacks bool
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
	// onHit runs after the unit's direct hit landed on its victim: a melee
	// blow, an arrow or bolt, or each victim of a cone (slow, splash, curse).
	onHit func(b *Battle, u, v *Unit, from Vec, a *active)
	// aura runs every tick for a living unit: it adds modifiers to the units
	// around it (songs, foretelling, mocking).
	aura func(b *Battle, src *Unit, a *active)
	// self runs every tick after auras: modifiers from the unit's own state (fury, phalanx).
	self func(b *Battle, u *Unit, a *active)
	// onHurt can change damage the unit takes, before the damage-taken
	// modifier (shield wall).
	onHurt func(b *Battle, v, att *Unit, from Vec, amount Fx, a *active) Fx
	// intercept: the unit (an ally of the victim) takes a share of a blow;
	// returns the share (guard).
	intercept func(b *Battle, guard, v, att *Unit, amount Fx, a *active) Fx
	// onDeath runs for living units when any unit dies (rally).
	onDeath func(b *Battle, holder, dead *Unit, a *active)
	// onCharged runs when an enemy dashes at the unit (brace).
	onCharged func(b *Battle, t, charger *Unit, a *active)
	// strike replaces the unit's blow when its swing lands (cone); true if it did.
	strike func(b *Battle, u, v *Unit, a *active) bool
	// kiteShoot: the unit keeps shooting while it backs away.
	kiteShoot bool
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
	kiteShoot bool
}

// learn adds the non-stat effects of the unit's skills to its loadout, then
// the extras (its side's plan).
func (kit *loadout) learn(def *UnitDef, skills []Skill, extra ...Effect) {
	for k := -1; k <= len(def.Skills); k++ { // the type's traits, its learned skills, then the extras
		var effects []Effect
		if k < 0 {
			effects = def.Traits
		} else if k == len(def.Skills) {
			effects = extra
		} else if hasSkill(skills, def.Skills[k].Skill) {
			effects = def.Skills[k].Effects
		}
		for i := range effects {
			e := &effects[i]
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
			kit.kiteShoot = kit.kiteShoot || ek.kiteShoot
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
