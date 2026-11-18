package sim

// Archer: long-range, kites away from melee (docs/GAME_CONCEPT.md §4).
var archerDef = UnitDef{
	ID: Archer, Name: "archer",
	Stats: Stats{HP: I(55), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(17), Range: I(9), Speed: Ratio(36, 10), Sight: I(11),
		Ranged: true, Kite: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{SkillRange, "Long shot", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 25}}},
		{SkillDamage, "Sharp arrows", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}}},
		{SkillLob, "Arcing shot", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}, {Kind: OverCover}}},
		{SkillExecute, "Headshot", []Effect{{Kind: Finisher, Pct: 25}}},
		{SkillHealth, "Toughness", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
