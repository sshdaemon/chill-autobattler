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

// Völva (Varangians, special): the seeress with the iron staff
// (design/units/volva/brief.md). Her attack is a curse at 7 m (a staff jab
// and dark smoke on the target; a bolt in the sim, drawn as smoke), strong
// against melee. She holds until an enemy comes within 10 m, stands her ground
// (backing off or falling back let her farm line infantry; balance.md), curses
// the strongest, and slips away from melee at a run, not a teleport.
var volvaDef = UnitDef{
	ID: Volva, Name: "volva",
	Stats:      Stats{HP: I(62), DmgLo: I(12), DmgHi: I(18), Cooldown: sec(22), Range: I(7), Speed: I(3), Sight: I(9), Ranged: true, SwingTicks: sec(8), ShotSpeed: I(14)},
	Temper:     Temperament{Target: TargetStrongest, Engage: EngageHold, Trigger: I(10)},
	Projectile: Bolt,
	Tags:       TagCaster,
	Bonus:      []TagBonus{{TagMelee, Ratio(130, 100)}},
	Skills: [MaxSkills]SkillDef{
		{VolvaCurse, "Curse", []Effect{{Kind: Curse, Pct: 20, Ticks: sec(50)}}},
		{VolvaDread, "Dread", []Effect{{Kind: Fear, Radius: I(4), Ticks: sec(120), Duration: sec(15)}}},
		{VolvaForetelling, "Foretelling", []Effect{{Kind: Aura, Mod: ModEvade, Pct: 10, Radius: I(6)}}},
		{VolvaSlipAway, "Slip away", []Effect{{Kind: Escape, Value: I(6), Radius: Ratio(25, 10), Speed: I(9), Ticks: sec(80)}}},
		{VolvaStaff, "Staff of the seeress", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
