package sim

// Archer's abilities, numbered in the order of its Skills list.
const (
	ArcherLongShot Skill = iota + 1
	ArcherSharpArrows
	ArcherArcingShot
	ArcherHeadshot
	ArcherToughness
)

func init() { register(archerDef) }

// Archer: long-range, kites away from melee (docs/GAME_CONCEPT.md §4).
var archerDef = UnitDef{
	ID: Archer, Name: "archer",
	Stats: Stats{HP: I(55), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(17), Range: I(9), Speed: Ratio(36, 10), Sight: I(11),
		Ranged: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Temper:     Temperament{Target: TargetNearest, Engage: EngageCharge, Kite: true},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{ArcherLongShot, "Long shot", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 25}}},
		{ArcherSharpArrows, "Sharp arrows", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}}},
		{ArcherArcingShot, "Arcing shot", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}, {Kind: OverCover}}},
		{ArcherHeadshot, "Headshot", []Effect{{Kind: Finisher, Pct: 25}}},
		{ArcherToughness, "Toughness", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
