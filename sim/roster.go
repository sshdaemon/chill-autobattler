package sim

import "fmt"

// The unit roster: every unit type a battle can field, as data. A unit is
// its base stats, traits, matchup tags and five learnable skills; a skill is
// a list of effects from a fixed set. Adding a unit means adding a roster
// file (unit_*.go); the battle code only knows effect kinds, never units or
// skills by name.

// Tag marks what a unit is, for matchup bonuses. Tags are bits.
type Tag uint8

const (
	TagMelee Tag = 1 << iota
	TagRanged
	TagCaster
)

// Tags lists every declared tag with its name; a roster entry may only use these.
var Tags = []struct {
	Tag  Tag
	Name string
}{{TagMelee, "melee"}, {TagRanged, "ranged"}, {TagCaster, "caster"}}

// TagBonus is a damage multiplier against victims with a tag.
type TagBonus struct {
	Tag  Tag
	Mult Fx
}

type Projectile uint8

const (
	NoProjectile Projectile = iota
	Arrow
	Bolt
)

// EffectKind ids are in ids.go.

type StatKind uint8

const (
	StatHP StatKind = iota + 1
	StatDamage
	StatRange
	StatSpeed
	StatAttackSpeed
	StatAuraRadius // the reach of the unit's auras, +Pct%
)

var statNames = [...]string{"", "hp", "damage", "range", "speed", "attackSpeed", "auraRadius"}

func (s StatKind) String() string {
	if int(s) < len(statNames) {
		return statNames[s]
	}
	return fmt.Sprintf("stat(%d)", s)
}

// Effect is one building block of a skill. Which fields matter depends on Kind.
type Effect struct {
	Kind      EffectKind
	Stat      StatKind // StatBonus
	Pct       int      // bonus %, chance % or splash %
	Value     Fx       // reach or distance, metres
	Radius    Fx       // metres
	Speed     Fx       // metres per second
	Factor    Fx       // speed or damage multiplier
	Ticks     int      // cooldown or duration
	Duration  int      // ticks an effect lasts, when Ticks is its cooldown
	Mod       ModKind  // auras, buffs: which modifier
	Tag       Tag      // which units it concerns (bonus against, guarded allies)
	Cond      Cond     // DamageTaken: when it applies
	Threshold int      // a % of health, or a count of units
	Foes      bool     // Aura: on enemies instead of allies
}

// SkillDef is a learnable skill: its number within the unit (1..MaxSkills, in
// order), a name and its effects. The first effect is the skill's headline
// (its Pct and Value are what the UI shows).
type SkillDef struct {
	Skill   Skill
	Name    string
	Effects []Effect
}

// Main is the headline effect.
func (d SkillDef) Main() Effect { return d.Effects[0] }

// UnitDef is one roster entry. Stats.Ranged is its trait (a unit that isn't
// ranged is melee); ranged units need a Projectile and Stats.ShotSpeed.
// Temper is how it fights on its own (temperament.go).
type UnitDef struct {
	ID         Class
	Name       string
	Stats      Stats
	Temper     Temperament
	Projectile Projectile
	Tags       Tag
	Bonus      []TagBonus // first bonus whose tag the victim has wins
	Skills     [MaxSkills]SkillDef
	// Traits: effects every unit of the type has, learned or not (a song, a cone attack).
	Traits []Effect
	// InRedPool: Red's generated armies draw from these, in roster order.
	// Changing the pool changes Red's armies (and sim.Version, through the roster fingerprint).
	InRedPool bool
}

// Roster is indexed by Class: Roster[c].ID == c. Each unit_<name>.go adds its
// entry with register (from an init func), so adding a unit edits no shared file.
var Roster []UnitDef

// register adds a unit to the roster, keeping it sorted by id. ValidateRoster
// (run by the tests) checks the ids are 0..n-1 with no gaps or duplicates.
func register(d UnitDef) {
	i := len(Roster)
	for i > 0 && Roster[i-1].ID > d.ID {
		i--
	}
	Roster = append(Roster, UnitDef{})
	copy(Roster[i+1:], Roster[i:])
	Roster[i] = d
	refreshVersion()
}

// RosterOf returns the entry for a class, or nil if there is none.
func RosterOf(c Class) *UnitDef {
	if int(c) < len(Roster) {
		return &Roster[c]
	}
	return nil
}

func (c Class) String() string {
	if d := RosterOf(c); d != nil {
		return d.Name
	}
	return fmt.Sprintf("class(%d)", c)
}

// SkillDefOf returns the class's definition of a skill, ok=false if the class
// can't learn it.
func SkillDefOf(c Class, s Skill) (SkillDef, bool) {
	if d := RosterOf(c); d != nil {
		for _, k := range d.Skills {
			if k.Skill == s {
				return k, true
			}
		}
	}
	return SkillDef{}, false
}

// MatchupOf is the damage multiplier for att hitting v.
func MatchupOf(att, v *UnitDef) Fx {
	for _, b := range att.Bonus {
		if v.Tags&b.Tag != 0 {
			return b.Mult
		}
	}
	return One
}

// RedPool lists the unit types Red's generated armies can draw from: those
// in Red's pool from the playable factions, in roster order.
func RedPool() []Class {
	var out []Class
	for _, d := range Roster {
		if !d.InRedPool {
			continue
		}
		for _, f := range Factions {
			if f.Playable && f.Has(d.ID) {
				out = append(out, d.ID)
				break
			}
		}
	}
	return out
}

// ValidateRoster checks a roster against the unit-roster rules. Ids: a unit's
// id is its position, allocated in ids.go (ids are never reused); a unit's
// skills are numbered 1..MaxSkills in order. Neither needs a proto change.
func ValidateRoster(r []UnitDef) error {
	if err := validateEntries(r); err != nil {
		return err
	}
	return ValidatePlans(Plans, r) // every plan must suit every unit
}

func validateEntries(r []UnitDef) error {
	var declared Tag
	for _, t := range Tags {
		declared |= t.Tag
	}
	pool := 0
	for i, d := range r {
		fail := func(format string, args ...any) error {
			return fmt.Errorf("roster entry %d (%s): %s", i, d.Name, fmt.Sprintf(format, args...))
		}
		if int(d.ID) != i {
			return fail("id %d, want %d (ids must be unique and match the position)", d.ID, i)
		}
		if d.Name == "" {
			return fail("no name")
		}
		st := d.Stats
		if st.HP <= 0 || st.DmgLo <= 0 || st.DmgHi < st.DmgLo || st.Cooldown <= 0 || st.Range <= 0 || st.Speed <= 0 || st.Sight <= 0 || st.SwingTicks <= 0 {
			return fail("stats must be positive (and DmgHi >= DmgLo): %+v", st)
		}
		if st.Ranged != (d.Projectile != NoProjectile) || st.Ranged != (st.ShotSpeed > 0) {
			return fail("a ranged unit needs a projectile and a shot speed, a melee unit neither")
		}
		if err := checkTemperament(&d); err != nil {
			return fail("%v", err)
		}
		if d.Tags == 0 || d.Tags&^declared != 0 {
			return fail("tags %b: need at least one, all declared", d.Tags)
		}
		for _, b := range d.Bonus {
			if b.Tag&^declared != 0 || b.Tag == 0 || b.Mult <= 0 {
				return fail("bonus %+v: undeclared tag or multiplier <= 0", b)
			}
		}
		if d.InRedPool {
			pool++
		}
		var kinds [len(effectKinds)]bool
		for _, e := range d.Traits {
			if err := checkEffect(e, &d); err != nil {
				return fail("trait: %v", err)
			}
			if e.Kind != StatBonus && !effectKinds[e.Kind].stacks {
				kinds[e.Kind] = true
			}
		}
		for si, s := range d.Skills {
			if s.Name == "" || len(s.Effects) == 0 {
				return fail("skill %d: needs a name and at least one effect (five skills per unit)", si+1)
			}
			if s.Skill != Skill(si+1) {
				return fail("skill %s is numbered %d, want %d (skills are numbered 1..%d in order)", s.Name, s.Skill, si+1, MaxSkills)
			}
			for _, e := range s.Effects {
				if err := checkEffect(e, &d); err != nil {
					return fail("skill %s: %v", s.Name, err)
				}
				if e.Kind != StatBonus && !effectKinds[e.Kind].stacks {
					if kinds[e.Kind] {
						return fail("more than one %v effect", e.Kind)
					}
					kinds[e.Kind] = true
				}
			}
		}
	}
	if pool == 0 {
		return fmt.Errorf("no roster entry is in Red's pool")
	}
	return nil
}
