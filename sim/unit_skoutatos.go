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

// Skoutatos (Romans, melee): heavy infantryman behind a long shield; slow and
// tough, waits in line, then advances with its neighbours and never falls back
// (design/units/skoutatos/brief.md). It holds until an enemy is within 12 m,
// not the plan's 9, and has no leash: with 9 m and a 5 m leash archers shot it
// from out of its trigger and it lost every ranged matchup
// (design/units/skoutatos/balance.md).
var skoutatosDef = UnitDef{
	ID: Skoutatos, Name: "skoutatos",
	Stats:  Stats{HP: I(115), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(14), Range: Ratio(19, 10), Speed: Ratio(29, 10), Sight: I(7), SwingTicks: sec(9)},
	Temper: Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(12)},
	Tags:   TagMelee,
	Bonus:  []TagBonus{{TagRanged, Ratio(135, 100)}},
	Skills: [MaxSkills]SkillDef{
		{SkoutatosShieldWall, "Shield wall", []Effect{{Kind: DamageTaken, Pct: -35, Cond: CondFront}}},
		{SkoutatosPhalanx, "Phalanx", []Effect{{Kind: Phalanx, Mod: ModTaken, Pct: -10, Radius: I(3), Threshold: 3}}},
		{SkoutatosKontarion, "Kontarion", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 26}}},
		{SkoutatosDrill, "Drill", []Effect{{Kind: StatBonus, Stat: StatAttackSpeed, Pct: 83}}},
		{SkoutatosKlibanion, "Klibanion", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
