package sim

// Strilets's abilities, numbered in the order of its Skills list.
const (
	StriletsLongShot Skill = iota + 1
	StriletsSteppeShooting
	StriletsBodkinPoints
	StriletsArcingShot
	StriletsQuickDraw
)

func init() { register(striletsDef) }

// Strilets (Kyivan Rus, ranged): archer of the druzhyna with a composite
// recurved bow of steppe type; light and quick, gives ground while shooting
// (design/units/strilets/brief.md). No fall-back: a ranged unit that retreats
// at low health stops shooting and loses archer duels (balance.md).
var striletsDef = UnitDef{
	ID: Strilets, Name: "strilets",
	Stats:      Stats{HP: I(57), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(17), Range: Ratio(85, 10), Speed: Ratio(36, 10), Sight: I(11), Ranged: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Temper:     Temperament{Target: TargetNearest, Engage: EngageLine, Kite: true, Keep: I(7)},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{StriletsLongShot, "Long shot", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 20}}},
		{StriletsSteppeShooting, "Steppe shooting", []Effect{{Kind: KiteShoot}, {Kind: StatBonus, Stat: StatSpeed, Pct: 10}}},
		{StriletsBodkinPoints, "Bodkin points", []Effect{{Kind: BonusAgainst, Pct: 30, Tag: TagMelee}}},
		{StriletsArcingShot, "Arcing shot", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}, {Kind: OverCover}}},
		{StriletsQuickDraw, "Quick draw", []Effect{{Kind: StatBonus, Stat: StatAttackSpeed, Pct: 84}}},
	},
	InRedPool: true,
}
