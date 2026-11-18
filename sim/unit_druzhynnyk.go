package sim

// Druzhynnyk's abilities, numbered in the order of its Skills list.
const (
	DruzhynnykShieldWall Skill = iota + 1
	DruzhynnykPrincesGuard
	DruzhynnykBracedSpear
	DruzhynnykHeavyMail
	DruzhynnykAxeAndSword
)

func init() { register(druzhynnykDef) }

// Druzhynnyk (Kyivan Rus, melee): the prince's retainer; balanced and armoured, guards the archers and the singer.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var druzhynnykDef = UnitDef{
	ID: Druzhynnyk, Name: "druzhynnyk",
	Stats:  Stats{HP: I(100), DmgLo: I(10), DmgHi: I(16), Cooldown: sec(13), Range: Ratio(15, 10), Speed: Ratio(32, 10), Sight: I(7), SwingTicks: sec(9)},
	Temper: Temperament{Target: TargetProtect, Engage: EngageLine, Leash: I(8), StayNear: TagRanged | TagCaster},
	Tags:   TagMelee,
	Bonus:  []TagBonus{{TagRanged, Ratio(140, 100)}},
	Skills: [MaxSkills]SkillDef{
		{DruzhynnykShieldWall, "Shield wall", []Effect{{Kind: DamageTaken, Pct: -30, Cond: CondRanged}}},
		{DruzhynnykPrincesGuard, "Prince's guard", []Effect{{Kind: Guard, Pct: 30, Radius: Ratio(25, 10), Tag: TagRanged | TagCaster}}},
		{DruzhynnykBracedSpear, "Braced spear", []Effect{{Kind: Brace, Pct: 150}}},
		{DruzhynnykHeavyMail, "Heavy mail", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}, {Kind: StatBonus, Stat: StatSpeed, Pct: -5}}},
		{DruzhynnykAxeAndSword, "Axe and sword", []Effect{{Kind: StatBonus, Stat: StatAttackSpeed, Pct: 80}}},
	},
	InRedPool: true,
}
