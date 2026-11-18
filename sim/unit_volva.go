package sim

// Volva's abilities, numbered in the order of its Skills list.
const (
	VolvaCurse Skill = iota + 1
	VolvaDread
	VolvaForetelling
	VolvaSlipAway
	VolvaStaff
)

func init() { register(volvaDef) }

// Völva (Varangians, special): the seeress; her curses and dread weaken and scatter the enemy's front line.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var volvaDef = UnitDef{
	ID: Volva, Name: "volva",
	Stats:      Stats{HP: I(55), DmgLo: I(8), DmgHi: I(12), Cooldown: sec(22), Range: I(7), Speed: I(3), Sight: I(9), Ranged: true, SwingTicks: sec(8), ShotSpeed: I(14)},
	Temper:     Temperament{Target: TargetStrongest, Engage: EngageHold, Trigger: I(10), Keep: I(7), FallBack: 40},
	Projectile: Bolt,
	Tags:       TagCaster,
	Bonus:      []TagBonus{{TagMelee, Ratio(130, 100)}},
	Skills: [MaxSkills]SkillDef{
		{VolvaCurse, "Curse", []Effect{{Kind: Curse, Pct: 20, Ticks: sec(50)}}},
		{VolvaDread, "Dread", []Effect{{Kind: Fear, Radius: I(4), Ticks: sec(120), Duration: sec(15)}}},
		{VolvaForetelling, "Foretelling", []Effect{{Kind: Aura, Mod: ModEvade, Pct: 10, Radius: I(6)}}},
		{VolvaSlipAway, "Slip away", []Effect{{Kind: Escape, Value: I(6), Radius: Ratio(25, 10), Ticks: sec(80)}}},
		{VolvaStaff, "Staff of the seeress", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
