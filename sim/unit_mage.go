package sim

// Mage: slow bolts, strong against melee (docs/GAME_CONCEPT.md §4).
var mageDef = UnitDef{
	ID: Mage, Name: "mage",
	Stats: Stats{HP: I(55), DmgLo: I(14), DmgHi: I(20), Cooldown: sec(22), Range: I(7), Speed: I(3), Sight: I(9),
		Ranged: true, SwingTicks: sec(8), ShotSpeed: I(14)},
	Projectile: Bolt,
	Tags:       TagCaster,
	Bonus:      []TagBonus{{TagMelee, Ratio(135, 100)}},
	Skills: [MaxSkills]SkillDef{
		{SkillFrost, "Frost", []Effect{{Kind: SlowOnHit, Pct: 100, Ticks: 60, Factor: Ratio(6, 10)}}},
		{SkillDamage, "Power", []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}}},
		{SkillBlink, "Blink", []Effect{{Kind: Escape, Value: I(8), Radius: Ratio(25, 10), Ticks: sec(80)}}},
		{SkillBlast, "Blast", []Effect{{Kind: SplashOnHit, Pct: 35, Radius: I(2)}}},
		{SkillHealth, "Toughness", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
