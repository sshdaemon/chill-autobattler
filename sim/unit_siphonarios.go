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

// Siphonarios (Romans, special): the Greek fire operator (design/units/siphonarios/brief.md).
// He waits behind his own melee until enemies come close, then jets everyone
// in a 60° cone in front of him; strong against massed melee, helpless against
// anything that outranges him but for his fire pots.
var siphonariosDef = UnitDef{
	ID: Siphonarios, Name: "siphonarios",
	Stats:  Stats{HP: I(76), DmgLo: I(9), DmgHi: I(13), Cooldown: sec(24), Range: I(5), Speed: Ratio(29, 10), Sight: I(8), SwingTicks: sec(8)},
	Temper: Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(6), StayNear: TagMelee},
	Tags:   TagCaster,
	Bonus:  []TagBonus{{TagMelee, Ratio(135, 100)}},
	Traits: []Effect{{Kind: Cone, Factor: Ratio(866, 1000)}}, // the fire jet: a 60° cone
	Skills: [MaxSkills]SkillDef{
		{SiphonariosBurningGround, "Burning ground", []Effect{{Kind: BurningGround, Pct: 22, Radius: Ratio(12, 10), Duration: sec(30)}}},
		{SiphonariosFirePots, "Fire pots", []Effect{{Kind: Burst, Value: I(8), Ticks: sec(100), Factor: One, Speed: I(10), Radius: I(2), Pct: 60}}},
		{SiphonariosPressure, "Pressure", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 30}}},
		{SiphonariosPaddedCoat, "Padded coat", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 25}}},
		{SiphonariosFearOfFire, "Fear of fire", []Effect{{Kind: SlowOnHit, Pct: 100, Ticks: sec(20), Factor: Ratio(6, 10)}}},
	},
	InRedPool: true,
}
