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

// Toxotes (Romans, ranged): archer of the line; shoots together with the others from behind the shields.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var toxotesDef = UnitDef{
	ID: Toxotes, Name: "toxotes",
	Stats:      Stats{HP: I(55), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(17), Range: I(9), Speed: Ratio(34, 10), Sight: I(11), Ranged: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Temper:     Temperament{Target: TargetFocus, Engage: EngageLine, Kite: true, Keep: I(8), FallBack: 30, StayNear: TagMelee},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{ToxotesSolenarion, "Solenarion", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 35}, {Kind: StatBonus, Stat: StatDamage, Pct: -20}}},
		{ToxotesVolley, "Volley", []Effect{{Kind: Focus, Pct: 15, Threshold: 2}}},
		{ToxotesSharpArrows, "Sharp arrows", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 15}}},
		{ToxotesArcingShot, "Arcing shot", []Effect{{Kind: OverCover}}},
		{ToxotesQuickDraw, "Quick draw", []Effect{{Kind: StatBonus, Stat: StatAttackSpeed, Pct: 80}}},
	},
	InRedPool: true,
}
