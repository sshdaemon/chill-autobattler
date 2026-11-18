package sim

// Huslyar's abilities, numbered in the order of its Skills list.
const (
	HuslyarSongOfGlory Skill = iota + 1
	HuslyarTaleOfTheBohatyrs
	HuslyarLament
	HuslyarMockingSong
	HuslyarLoudVoice
)

func init() { register(huslyarDef) }

// Huslyar (Kyivan Rus, support): the court singer with the husli
// (design/units/huslyar/brief.md). Weak alone by design: a belt knife and no
// matchup bonus. His song makes the allies around him hit harder, and his
// learnable songs add speed, toughness, a rally or a mocking song. Measured
// inside an army, not head to head (design/units/huslyar/balance.md).
var huslyarDef = UnitDef{
	ID: Huslyar, Name: "huslyar",
	Stats:  Stats{HP: I(110), DmgLo: I(4), DmgHi: I(6), Cooldown: sec(15), Range: Ratio(15, 10), Speed: I(3), Sight: I(9), SwingTicks: sec(8)},
	Temper: Temperament{Target: TargetNearest, Engage: EngageLine, Keep: I(5), FallBack: 50, StayNear: TagMelee},
	Tags:   TagCaster,
	Traits: []Effect{{Kind: Aura, Mod: ModDamage, Pct: 25, Radius: I(6)}}, // his song
	Skills: [MaxSkills]SkillDef{
		{HuslyarSongOfGlory, "Song of glory", []Effect{{Kind: Aura, Mod: ModAttackSpeed, Pct: 20, Radius: I(6)}}},
		{HuslyarTaleOfTheBohatyrs, "Tale of the bohatyrs", []Effect{{Kind: Aura, Mod: ModTaken, Pct: -20, Radius: I(6)}}},
		{HuslyarLament, "Lament for the fallen", []Effect{{Kind: Rally, Mod: ModDamage, Pct: 38, Radius: I(6), Ticks: sec(50)}}},
		{HuslyarMockingSong, "Mocking song", []Effect{{Kind: Aura, Mod: ModAttackSpeed, Pct: -20, Radius: I(6), Foes: true}}},
		{HuslyarLoudVoice, "Loud voice", []Effect{{Kind: StatBonus, Stat: StatAuraRadius, Pct: 65}}},
	},
	InRedPool: true,
}
