package sim

// Ids that are shared across sessions working on different units, allocated
// here and only here, in planning changes (docs/ROADMAP.md), never by "take
// the next free one" in a unit's own change. A unit's id is its position in
// the roster (ids are never reused); its abilities are numbered within its own
// roster entry (Skill), so they need no shared ids.

// Class is a roster id: Roster[c].ID == c. Each unit file registers itself
// (register in roster.go) under its id from this list.
type Class uint8

// Retired unit ids: never reused, their roster slots stay empty (OpenSpec
// change retire-classic): 0 warrior, 1 archer, 2 mage, of the Classic faction.
var retiredClasses = []Class{0, 1, 2}

const (
	// The Dnipro road (design/ROSTER-PLAN.md), allocated together in roadmap Phase 3:
	Druzhynnyk  Class = 3  // Kyivan Rus
	Strilets    Class = 4  // Kyivan Rus
	Huslyar     Class = 5  // Kyivan Rus
	Huskarl     Class = 6  // Varangians
	Bowman      Class = 7  // Varangians
	Volva       Class = 8  // Varangians
	Skoutatos   Class = 9  // Romans
	Toxotes     Class = 10 // Romans
	Siphonarios Class = 11 // Romans
	// next free: 12
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
	// Phase 3 (design/ROSTER-PLAN.md), one file each:
	DamageTaken   EffectKind = 8  // takes Pct% more (negative: less) damage, under Cond
	Aura          EffectKind = 9  // allies (or Foes) within Radius get Mod +Pct%
	Rally         EffectKind = 10 // when an ally dies within Radius, allies there get Mod +Pct% for Ticks
	Guard         EffectKind = 11 // takes Pct% of blows aimed at allies with Tag within Radius
	Brace         EffectKind = 12 // an enemy that dashes at it takes a Pct% counter-blow and is stopped short
	Fury          EffectKind = 13 // Mod +Pct% while below Threshold% health
	Curse         EffectKind = 14 // its hits make the victim take Pct% more damage for Ticks
	Fear          EffectKind = 15 // every Ticks, enemies within Radius flee for Duration
	Burst         EffectKind = 16 // every Ticks, a ×Factor attack thrown up to Value (Speed m/s), Pct% splash within Radius
	BonusAgainst  EffectKind = 17 // +Pct% damage against units with Tag
	Focus         EffectKind = 18 // +Pct% damage when Threshold or more allies fight the same target
	Phalanx       EffectKind = 19 // Mod +Pct% for each ally of its type within Radius, up to Threshold
	BurningGround EffectKind = 20 // its hits leave fire of Radius for Duration: Pct% of a blow per second
	KiteShoot     EffectKind = 21 // keeps shooting while it backs away
	Cone          EffectKind = 22 // its blow hits every enemy within range in front (cos of the half-angle: Factor)
	// next free: 23
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
	TargetFocusReach  TargetRule = 7 // like TargetFocus, among enemies within its reach (range + FocusSlack); else the nearest
	// next free: 8
)

type EngageRule uint8

const (
	EngageCharge EngageRule = 1 // go for the enemy at once
	EngageLine   EngageRule = 2 // advance at the pace of its allies
	EngageHold   EngageRule = 3 // wait until an enemy is within Trigger or an ally is fighting
	EngageFlank  EngageRule = 4 // go round the side of its target before closing in
	// next free: 5
)

// Plan is a battle plan id (plan.go): an army-wide stance a side picks
// before the battle. 0 is Instinct (no change), so an army without a plan
// fights as it always did.
type Plan uint8

const (
	PlanInstinct   Plan = 0 // every unit fights by its own temperament
	PlanShieldWall Plan = 1 // melee hold and protect the back line; ranged focus behind them
	PlanCharge     Plan = 2 // everyone charges the nearest enemy, no kiting
	PlanHunt       Plan = 3 // melee charge the enemy's ranged and casters first
	PlanFocusFire  Plan = 4 // advance together and fight the enemy most allies fight
	PlanSkirmish   Plan = 5 // melee screen the back line; ranged kite at a distance
	// Tactics: plans with groups and triggers.
	PlanHammer     Plan = 6 // left holds as a shield wall; right goes round and charges once the left fights
	PlanReserve    Plan = 7 // left advances in line; right holds and charges when the first ally falls
	PlanHoldCharge Plan = 8 // everyone holds for 15 s, then charges
	// next free: 9
	PlanCustom Plan = 255 // reserved: a side fighting by a custom plan (custom.go)
)

// Edge is an edge from the custom-plan menu (edge.go): an advantage with a
// cost, one per role. 0 is none.
type Edge uint8

const (
	EdgeBraced    Edge = 1 // melee: less damage from the front, slower
	EdgeReckless  Edge = 2 // faster, takes more damage
	EdgeDrilled   Edge = 3 // melee: more damage, takes more damage
	EdgeLongReach Edge = 4 // ranged and casters: longer range, less damage
	EdgeLightFeet Edge = 5 // ranged: faster, less health
	// next free: 6
)

// TriggerKind is what switches a tactic's group to its second stance
// (trigger.go); each lives in its own trigger_<name>.go. 0 is none.
type TriggerKind uint8

const (
	TriggerAfter         TriggerKind = 1 // Seconds have passed since the battle began
	TriggerAllyFalls     TriggerKind = 2 // a unit of its army has fallen
	TriggerOtherFighting TriggerKind = 3 // a unit of the other group is fighting within reach
	// next free: 4
)

// Faction is a faction id (faction.go): which units a side may field.
type Faction uint8

const (
	// 1: retired (Classic: warrior, archer and mage); never reused
	FactionRus        Faction = 2 // Kyivan Rus
	FactionVarangians Faction = 3
	FactionRomans     Faction = 4
	// next free: 5
)
