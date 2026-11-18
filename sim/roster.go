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

type EffectKind uint8

const (
	StatBonus   EffectKind = iota + 1 // Stat +Pct% (AttackSpeed: cooldown and swing ×Pct%)
	Finisher                          // Pct% chance to kill a victim below ExecuteBelowPct of max health
	SlowOnHit                         // Pct% chance per landed projectile: victim moves ×Factor for Ticks
	SplashOnHit                       // victim's allies within Radius take Pct% of the hit
	GapCloser                         // rush at Speed to a target up to Value away; Ticks cooldown
	Escape                            // teleport Value away from melee within Radius; Ticks cooldown
	OverCover                         // shots need no clear line of fire
)

var effectNames = [...]string{"", "statBonus", "finisher", "slowOnHit", "splashOnHit", "gapCloser", "escape", "overCover"}

func (k EffectKind) String() string {
	if int(k) < len(effectNames) {
		return effectNames[k]
	}
	return fmt.Sprintf("effect(%d)", k)
}

type StatKind uint8

const (
	StatHP StatKind = iota + 1
	StatDamage
	StatRange
	StatSpeed
	StatAttackSpeed
)

var statNames = [...]string{"", "hp", "damage", "range", "speed", "attackSpeed"}

func (s StatKind) String() string {
	if int(s) < len(statNames) {
		return statNames[s]
	}
	return fmt.Sprintf("stat(%d)", s)
}

// Effect is one building block of a skill. Which fields matter depends on Kind.
type Effect struct {
	Kind   EffectKind
	Stat   StatKind // StatBonus
	Pct    int      // bonus %, chance % or splash %
	Value  Fx       // reach or distance, metres
	Radius Fx       // metres
	Speed  Fx       // metres per second
	Factor Fx       // speed multiplier
	Ticks  int      // cooldown or duration
}

// SkillDef is a learnable skill: a wire id, a name and its effects. The first
// effect is the skill's headline (its Pct and Value are what the UI shows).
type SkillDef struct {
	Skill   Skill
	Name    string
	Effects []Effect
}

// Main is the headline effect.
func (d SkillDef) Main() Effect { return d.Effects[0] }

// UnitDef is one roster entry. Stats.Ranged and Stats.Kite are its traits
// (a unit that isn't ranged is melee); ranged units need a Projectile and
// Stats.ShotSpeed.
type UnitDef struct {
	ID         Class
	Name       string
	Stats      Stats
	Projectile Projectile
	Tags       Tag
	Bonus      []TagBonus // first bonus whose tag the victim has wins
	Skills     [MaxSkills]SkillDef
	// InRedPool: Red's generated armies draw from these, in roster order.
	// Changing the pool changes Red's armies, so it needs a sim.Version bump.
	InRedPool bool
}

// Roster is indexed by Class: Roster[c].ID == c.
var Roster = []UnitDef{warriorDef, archerDef, mageDef}

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

// RedPool lists the classes Red's generated armies draw from, in roster order.
func RedPool() []Class {
	var out []Class
	for _, d := range Roster {
		if d.InRedPool {
			out = append(out, d.ID)
		}
	}
	return out
}

// skillShape is what a skill id means across units: its effect kinds and
// stats, in order (numbers may differ, e.g. Toughness +10% or +20% HP).
func skillShape(d SkillDef) string {
	s := ""
	for _, e := range d.Effects {
		s += fmt.Sprintf("%v/%v;", e.Kind, e.Stat)
	}
	return s
}

// ValidateRoster checks a roster against the unit-roster rules. Ids: a unit's
// id is its position (new units append, ids are never reused); a skill id
// shared by several units must mean the same there (same effect kinds and
// stats), so a new kind of skill takes the next free id. Neither needs a proto
// change: the wire enums are open and the roster is the check.
func ValidateRoster(r []UnitDef) error {
	shapes := map[Skill]string{}
	for _, d := range r {
		for _, s := range d.Skills {
			if prev, ok := shapes[s.Skill]; ok && prev != skillShape(s) && s.Skill != 0 && len(s.Effects) > 0 {
				return fmt.Errorf("skill id %d means different things (%s vs %s, in %s): give the new skill the next free id", s.Skill, prev, skillShape(s), d.Name)
			} else if !ok && s.Skill != 0 && len(s.Effects) > 0 {
				shapes[s.Skill] = skillShape(s)
			}
		}
	}
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
		if st.Kite && !st.Ranged {
			return fail("only ranged units kite")
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
		var kinds [len(effectNames)]bool
		for si, s := range d.Skills {
			if s.Skill == 0 || s.Name == "" || len(s.Effects) == 0 {
				return fail("skill %d: needs an id, a name and at least one effect (five skills per unit)", si)
			}
			for sj := 0; sj < si; sj++ {
				if d.Skills[sj].Skill == s.Skill {
					return fail("skill id %d twice", s.Skill)
				}
			}
			for _, e := range s.Effects {
				if err := checkEffect(e); err != nil {
					return fail("skill %s: %v", s.Name, err)
				}
				if e.Kind != StatBonus {
					if kinds[e.Kind] {
						return fail("more than one %v effect", e.Kind)
					}
					kinds[e.Kind] = true
				}
				if (e.Kind == SlowOnHit || e.Kind == SplashOnHit || e.Kind == OverCover) && !st.Ranged {
					return fail("%v needs a ranged unit (it acts on projectiles)", e.Kind)
				}
			}
		}
	}
	if pool == 0 {
		return fmt.Errorf("no roster entry is in Red's pool")
	}
	return nil
}

func checkEffect(e Effect) error {
	pct := func(lo, hi int) error {
		if e.Pct < lo || e.Pct > hi {
			return fmt.Errorf("%v: %d%% outside %d..%d", e.Kind, e.Pct, lo, hi)
		}
		return nil
	}
	switch e.Kind {
	case StatBonus:
		if e.Stat == StatAttackSpeed {
			return pct(1, 100)
		}
		if e.Stat < StatHP || e.Stat > StatAttackSpeed {
			return fmt.Errorf("stat bonus on unknown stat %d", e.Stat)
		}
		return pct(-90, 400)
	case Finisher:
		return pct(1, 100)
	case SlowOnHit:
		if e.Ticks <= 0 || e.Factor <= 0 || e.Factor >= One {
			return fmt.Errorf("slow needs ticks > 0 and a factor in (0, 1)")
		}
		return pct(1, 100)
	case SplashOnHit:
		if e.Radius <= 0 {
			return fmt.Errorf("splash needs a radius")
		}
		return pct(1, 100)
	case GapCloser:
		if e.Value <= 0 || e.Speed <= 0 || e.Ticks <= 0 {
			return fmt.Errorf("gap-closer needs reach, speed and cooldown")
		}
	case Escape:
		if e.Value <= 0 || e.Radius <= 0 || e.Ticks <= 0 {
			return fmt.Errorf("escape needs distance, threat radius and cooldown")
		}
	case OverCover:
	default:
		return fmt.Errorf("unknown effect kind %d", e.Kind)
	}
	return nil
}
