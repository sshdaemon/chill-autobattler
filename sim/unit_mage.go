package sim

// Mage's abilities, numbered in the order of its Skills list.
const (
	MageFrost Skill = iota + 1
	MagePower
	MageBlink
	MageBlast
	MageToughness
)

func init() { register(mageDef) }

// Mage: slow bolts, strong against melee (docs/GAME_CONCEPT.md §4).
var mageDef = UnitDef{
	ID: Mage, Name: "mage",
	Stats: Stats{HP: I(55), DmgLo: I(14), DmgHi: I(20), Cooldown: sec(22), Range: I(7), Speed: I(3), Sight: I(9),
		Ranged: true, SwingTicks: sec(8), ShotSpeed: I(14)},
	Projectile: Bolt,
	Tags:       TagCaster,
	Bonus:      []TagBonus{{TagMelee, Ratio(135, 100)}},
	Skills: [MaxSkills]SkillDef{
		{MageFrost, "Frost", []Effect{{Kind: SlowOnHit, Pct: 100, Ticks: 60, Factor: Ratio(6, 10)}}},
		{MagePower, "Power", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}}},
		{MageBlink, "Blink", []Effect{{Kind: Escape, Value: I(8), Radius: Ratio(25, 10), Ticks: sec(80)}}},
		{MageBlast, "Blast", []Effect{{Kind: SplashOnHit, Pct: 35, Radius: I(2)}}},
		{MageToughness, "Toughness", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
