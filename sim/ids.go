package sim

// Ids that are shared across sessions working on different units, allocated
// here and only here, in planning changes (docs/ROADMAP.md), never by "take
// the next free one" in a unit's own change. A unit's id is its position in
// the roster (ids are never reused); its abilities are numbered within its own
// roster entry (Skill), so they need no shared ids.

// Class is a roster id: Roster[c].ID == c. Each unit file registers itself
// (register in roster.go) under its id from this list.
type Class uint8

const (
	Warrior Class = 0
	Archer  Class = 1
	Mage    Class = 2
	// next free: 3
)

// EffectKind is a kind of ability effect (roster.go); ids are explicit so that
// kinds added in different changes never collide.
type EffectKind uint8

const (
	StatBonus   EffectKind = 1 // Stat +Pct% (AttackSpeed: cooldown and swing ×Pct%)
	Finisher    EffectKind = 2 // Pct% chance to kill a victim below ExecuteBelowPct of max health
	SlowOnHit   EffectKind = 3 // Pct% chance per landed projectile: victim moves ×Factor for Ticks
	SplashOnHit EffectKind = 4 // victim's allies within Radius take Pct% of the hit
	GapCloser   EffectKind = 5 // rush at Speed to a target up to Value away; Ticks cooldown
	Escape      EffectKind = 6 // teleport Value away from melee within Radius; Ticks cooldown
	OverCover   EffectKind = 7 // shots need no clear line of fire
	// next free: 8
)

// TargetRule is how a unit picks whom to fight (temperament.go); EngageRule is
// what it does while it has no target. Ids are explicit for the same reason.
type TargetRule uint8

const (
	TargetNearest     TargetRule = 1 // the closest enemy the team sees
	TargetWounded     TargetRule = 2 // the most wounded enemy close by, else the nearest
	TargetRangedFirst TargetRule = 3 // the nearest ranged or caster enemy, else the nearest
	TargetStrongest   TargetRule = 4 // the enemy that hits hardest, else the nearest
	TargetFocus       TargetRule = 5 // the enemy most allies are already fighting, else the nearest
	TargetProtect     TargetRule = 6 // the enemy closest to its own ranged and caster allies, else the nearest
	// next free: 7
)

type EngageRule uint8

const (
	EngageCharge EngageRule = 1 // go for the enemy at once
	EngageLine   EngageRule = 2 // advance at the pace of its allies
	EngageHold   EngageRule = 3 // wait until an enemy is within Trigger or an ally is fighting
	// next free: 4
)

// Faction is a faction id (faction.go): which units a side may field.
type Faction uint8

const (
	FactionClassic Faction = 1 // the original warrior, archer and mage
	// next free: 2 (planned: Rus, Varangians, Romans; design/ROSTER-PLAN.md)
)
