package sim

// Warrior: melee bruiser, strong against ranged units (docs/GAME_CONCEPT.md §4).
var warriorDef = UnitDef{
	ID: Warrior, Name: "warrior",
	Stats: Stats{HP: I(100), DmgLo: I(10), DmgHi: I(16), Cooldown: sec(13), Range: Ratio(15, 10), Speed: Ratio(32, 10), Sight: I(7),
		SwingTicks: sec(9)}, // min(cooldown*0.9, ~0.9 s clip)
	Tags:  TagMelee,
	Bonus: []TagBonus{{TagRanged, Ratio(145, 100)}, {TagCaster, Ratio(95, 100)}},
	Skills: [MaxSkills]SkillDef{
		{SkillHealth, "Toughness", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 10}}},
		{SkillSpeed, "Swift", []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 40}, {Kind: StatBonus, Stat: StatAttackSpeed, Pct: 80}}},
		{SkillDamage, "Strength", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 15}}},
		{SkillDash, "Dash", []Effect{{Kind: GapCloser, Value: I(10), Speed: I(16), Ticks: 200}}},
		{SkillExecute, "Executioner", []Effect{{Kind: Finisher, Pct: 35}}},
	},
	InRedPool: true,
}
