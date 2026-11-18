package sim

// Warrior's abilities, numbered in the order of its Skills list.
const (
	WarriorToughness Skill = iota + 1
	WarriorSwift
	WarriorStrength
	WarriorDash
	WarriorExecutioner
)

func init() { register(warriorDef) }

// Warrior: melee bruiser, strong against ranged units (docs/GAME_CONCEPT.md §4).
var warriorDef = UnitDef{
	ID: Warrior, Name: "warrior",
	Stats: Stats{HP: I(100), DmgLo: I(10), DmgHi: I(16), Cooldown: sec(13), Range: Ratio(15, 10), Speed: Ratio(32, 10), Sight: I(7),
		SwingTicks: sec(9)}, // min(cooldown*0.9, ~0.9 s clip)
	Temper: Temperament{Target: TargetNearest, Engage: EngageCharge},
	Tags:   TagMelee,
	Bonus:  []TagBonus{{TagRanged, Ratio(145, 100)}, {TagCaster, Ratio(95, 100)}},
	Skills: [MaxSkills]SkillDef{
		{WarriorToughness, "Toughness", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 10}}},
		{WarriorSwift, "Swift", []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 40}, {Kind: StatBonus, Stat: StatAttackSpeed, Pct: 80}}},
		{WarriorStrength, "Strength", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 15}}},
		{WarriorDash, "Dash", []Effect{{Kind: GapCloser, Value: I(10), Speed: I(16), Ticks: 200}}},
		{WarriorExecutioner, "Executioner", []Effect{{Kind: Finisher, Pct: 35}}},
	},
	InRedPool: true,
}
