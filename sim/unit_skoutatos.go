package sim

// Skoutatos's abilities, numbered in the order of its Skills list.
const (
	SkoutatosShieldWall Skill = iota + 1
	SkoutatosPhalanx
	SkoutatosKontarion
	SkoutatosDrill
	SkoutatosKlibanion
)

func init() { register(skoutatosDef) }

// Skoutatos (Romans, melee): heavy infantryman; slow and tough, holds the line.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var skoutatosDef = UnitDef{
	ID: Skoutatos, Name: "skoutatos",
	Stats:  Stats{HP: I(115), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(14), Range: Ratio(19, 10), Speed: Ratio(29, 10), Sight: I(7), SwingTicks: sec(9)},
	Temper: Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(9), Leash: I(5)},
	Tags:   TagMelee,
	Bonus:  []TagBonus{{TagRanged, Ratio(135, 100)}},
	Skills: [MaxSkills]SkillDef{
		{SkoutatosShieldWall, "Shield wall", []Effect{{Kind: DamageTaken, Pct: -40, Cond: CondFront}}},
		{SkoutatosPhalanx, "Phalanx", []Effect{{Kind: Phalanx, Mod: ModTaken, Pct: -8, Radius: I(2), Threshold: 3}}},
		{SkoutatosKontarion, "Kontarion", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 26}}},
		{SkoutatosDrill, "Drill", []Effect{{Kind: StatBonus, Stat: StatAttackSpeed, Pct: 83}}},
		{SkoutatosKlibanion, "Klibanion", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
