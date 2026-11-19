package sim

// Toxotes's abilities, numbered in the order of its Skills list.
const (
	ToxotesSolenarion Skill = iota + 1
	ToxotesVolley
	ToxotesSharpArrows
	ToxotesArcingShot
	ToxotesQuickDraw
)

func init() { register(toxotesDef) }

// Toxotes (Romans, ranged): archer of the line with a composite bow; shoots
// at what its comrades are shooting, as long as it is within reach
// (design/units/toxotes/brief.md). No fall-back: a ranged unit that retreats
// at low health stops shooting and loses archer duels (balance.md).
var toxotesDef = UnitDef{
	ID: Toxotes, Name: "toxotes",
	Stats:      Stats{HP: I(57), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(17), Range: I(9), Speed: Ratio(34, 10), Sight: I(11), Ranged: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Temper:     Temperament{Target: TargetFocusReach, Engage: EngageLine, Kite: true, Keep: I(7)},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{ToxotesSolenarion, "Solenarion", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 30}, {Kind: StatBonus, Stat: StatDamage, Pct: -5}}},
		{ToxotesVolley, "Volley", []Effect{{Kind: Focus, Pct: 15, Threshold: 2}}},
		{ToxotesSharpArrows, "Sharp arrows", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 15}}},
		{ToxotesArcingShot, "Arcing shot", []Effect{{Kind: OverCover}}},
		{ToxotesQuickDraw, "Quick draw", []Effect{{Kind: StatBonus, Stat: StatAttackSpeed, Pct: 80}}},
	},
	InRedPool: true,
}
