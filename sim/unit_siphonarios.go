package sim

// Siphonarios's abilities, numbered in the order of its Skills list.
const (
	SiphonariosBurningGround Skill = iota + 1
	SiphonariosFirePots
	SiphonariosPressure
	SiphonariosPaddedCoat
	SiphonariosFearOfFire
)

func init() { register(siphonariosDef) }

// Siphonarios (Romans, special): the Greek fire operator; a short fire jet that hits everyone in front, punishing massed melee.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var siphonariosDef = UnitDef{
	ID: Siphonarios, Name: "siphonarios",
	Stats:  Stats{HP: I(60), DmgLo: I(8), DmgHi: I(12), Cooldown: sec(24), Range: I(5), Speed: Ratio(29, 10), Sight: I(8), SwingTicks: sec(8)},
	Temper: Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(6), Keep: I(3), FallBack: 40, StayNear: TagMelee},
	Tags:   TagCaster,
	Bonus:  []TagBonus{{TagMelee, Ratio(135, 100)}},
	Traits: []Effect{{Kind: Cone, Factor: Ratio(866, 1000)}}, // the fire jet: a 60° cone
	Skills: [MaxSkills]SkillDef{
		{SiphonariosBurningGround, "Burning ground", []Effect{{Kind: BurningGround, Pct: 30, Radius: Ratio(12, 10), Duration: sec(30)}}},
		{SiphonariosFirePots, "Fire pots", []Effect{{Kind: Burst, Value: I(8), Ticks: sec(100), Factor: One, Speed: I(10), Radius: I(2), Pct: 60}}},
		{SiphonariosPressure, "Pressure", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 30}}},
		{SiphonariosPaddedCoat, "Padded coat", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
		{SiphonariosFearOfFire, "Fear of fire", []Effect{{Kind: SlowOnHit, Pct: 100, Ticks: sec(20), Factor: Ratio(7, 10)}}},
	},
	InRedPool: true,
}
