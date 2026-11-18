package sim

// Huskarl's abilities, numbered in the order of its Skills list.
const (
	HuskarlDaneAxeSweep Skill = iota + 1
	HuskarlCharge
	HuskarlBattleFury
	HuskarlShieldBreaker
	HuskarlExecutioner
)

func init() { register(huskarlDef) }

// Huskarl (Varangians, melee): sworn household warrior with a Dane axe; shock
// troops with big blows and less armour that charge at once, never fall back
// and cut down the wounded (design/units/huskarl/brief.md, balance.md).
var huskarlDef = UnitDef{
	ID: Huskarl, Name: "huskarl",
	Stats:  Stats{HP: I(94), DmgLo: I(14), DmgHi: I(20), Cooldown: sec(15), Range: Ratio(17, 10), Speed: Ratio(36, 10), Sight: I(7), SwingTicks: sec(10)},
	Temper: Temperament{Target: TargetWounded, Engage: EngageCharge},
	Tags:   TagMelee,
	Bonus:  []TagBonus{{TagRanged, Ratio(145, 100)}},
	Skills: [MaxSkills]SkillDef{
		{HuskarlDaneAxeSweep, "Dane axe sweep", []Effect{{Kind: SplashOnHit, Pct: 40, Radius: Ratio(15, 10)}}},
		{HuskarlCharge, "Charge", []Effect{{Kind: GapCloser, Value: I(10), Speed: I(16), Ticks: 200}}},
		{HuskarlBattleFury, "Battle fury", []Effect{{Kind: Fury, Mod: ModDamage, Pct: 30, Threshold: 50}}},
		{HuskarlShieldBreaker, "Shield-breaker", []Effect{{Kind: Curse, Pct: 15, Ticks: sec(40)}}},
		{HuskarlExecutioner, "Executioner", []Effect{{Kind: Finisher, Pct: 25}}},
	},
	InRedPool: true,
}
