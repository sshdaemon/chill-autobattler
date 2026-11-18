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

// Huslyar (Kyivan Rus, support): the singer with the husli; his song makes the army around him stronger, he is almost harmless alone.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var huslyarDef = UnitDef{
	ID: Huslyar, Name: "huslyar",
	Stats:  Stats{HP: I(60), DmgLo: I(4), DmgHi: I(6), Cooldown: sec(15), Range: Ratio(15, 10), Speed: I(3), Sight: I(9), SwingTicks: sec(8)},
	Temper: Temperament{Target: TargetNearest, Engage: EngageLine, Keep: I(5), FallBack: 50, StayNear: TagMelee},
	Tags:   TagCaster,
	Traits: []Effect{{Kind: Aura, Mod: ModDamage, Pct: 10, Radius: I(6)}}, // his song
	Skills: [MaxSkills]SkillDef{
		{HuslyarSongOfGlory, "Song of glory", []Effect{{Kind: Aura, Mod: ModAttackSpeed, Pct: 15, Radius: I(6)}}},
		{HuslyarTaleOfTheBohatyrs, "Tale of the bohatyrs", []Effect{{Kind: Aura, Mod: ModTaken, Pct: -15, Radius: I(6)}}},
		{HuslyarLament, "Lament for the fallen", []Effect{{Kind: Rally, Mod: ModDamage, Pct: 30, Radius: I(6), Ticks: sec(50)}}},
		{HuslyarMockingSong, "Mocking song", []Effect{{Kind: Aura, Mod: ModAttackSpeed, Pct: -15, Radius: I(5), Foes: true}}},
		{HuslyarLoudVoice, "Loud voice", []Effect{{Kind: StatBonus, Stat: StatAuraRadius, Pct: 50}}},
	},
	InRedPool: true,
}
