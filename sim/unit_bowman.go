package sim

// Bowman's abilities, numbered in the order of its Skills list.
const (
	BowmanThrowingSpears Skill = iota + 1
	BowmanHunter
	BowmanYewLongbow
	BowmanArcingShot
	BowmanSeaHardened
)

func init() { register(bowmanDef) }

// Bowman (Varangians, ranged): hunter-warrior with a yew bow and throwing spears; goes after the enemy's ranged and specials.
// Numbers from design/ROSTER-PLAN.md; a placeholder until its unit session
// (roadmap Phase 4) designs, balances and dresses it.
var bowmanDef = UnitDef{
	ID: Bowman, Name: "bowman",
	Stats:      Stats{HP: I(60), DmgLo: I(9), DmgHi: I(13), Cooldown: sec(16), Range: I(8), Speed: Ratio(35, 10), Sight: I(10), Ranged: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Temper:     Temperament{Target: TargetRangedFirst, Engage: EngageCharge, Kite: true, Keep: I(4), FallBack: 25},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{BowmanThrowingSpears, "Throwing spears", []Effect{{Kind: Burst, Value: I(8), Ticks: sec(80), Factor: I(2), Speed: I(14)}}},
		{BowmanHunter, "Hunter", []Effect{{Kind: BonusAgainst, Pct: 20, Tag: TagRanged | TagCaster}}},
		{BowmanYewLongbow, "Yew longbow", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 20}}},
		{BowmanArcingShot, "Arcing shot", []Effect{{Kind: OverCover}}},
		{BowmanSeaHardened, "Sea-hardened", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
