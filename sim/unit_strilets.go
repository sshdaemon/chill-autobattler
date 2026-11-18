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

// Strilets (Kyivan Rus, ranged): archer of the druzhyna; mobile, shoots while backing away.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var striletsDef = UnitDef{
	ID: Strilets, Name: "strilets",
	Stats:      Stats{HP: I(55), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(17), Range: I(9), Speed: Ratio(36, 10), Sight: I(11), Ranged: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Temper:     Temperament{Target: TargetNearest, Engage: EngageLine, Kite: true, Keep: I(7), FallBack: 30},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{StriletsLongShot, "Long shot", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 25}}},
		{StriletsSteppeShooting, "Steppe shooting", []Effect{{Kind: KiteShoot}}},
		{StriletsBodkinPoints, "Bodkin points", []Effect{{Kind: BonusAgainst, Pct: 30, Tag: TagMelee}}},
		{StriletsArcingShot, "Arcing shot", []Effect{{Kind: OverCover}}},
		{StriletsQuickDraw, "Quick draw", []Effect{{Kind: StatBonus, Stat: StatAttackSpeed, Pct: 80}}},
	},
	InRedPool: true,
}
