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

// Bowman (Varangians, ranged): hunter-warrior with a long yew self bow and
// throwing spears (design/units/bowman/brief.md). Walks in on the enemy's
// archers and specials instead of giving ground; light kit, little health.
// No fall-back: a ranged unit that retreats at low health stops shooting and
// loses archer duels (design/units/bowman/balance.md).
var bowmanDef = UnitDef{
	ID: Bowman, Name: "bowman",
	Stats:      Stats{HP: I(52), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(16), Range: I(8), Speed: Ratio(35, 10), Sight: I(10), Ranged: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Temper:     Temperament{Target: TargetRangedFirst, Engage: EngageCharge, Kite: true, Keep: I(4)},
	Projectile: Arrow,
	Tags:       TagRanged,
	Skills: [MaxSkills]SkillDef{
		{BowmanThrowingSpears, "Throwing spears", []Effect{{Kind: Burst, Value: I(8), Ticks: sec(100), Factor: Ratio(16, 10), Speed: I(14)}}},
		{BowmanHunter, "Hunter", []Effect{{Kind: BonusAgainst, Pct: 20, Tag: TagRanged | TagCaster}}},
		{BowmanYewLongbow, "Yew longbow", []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 20}}},
		{BowmanArcingShot, "Arcing shot", []Effect{{Kind: OverCover}}},
		{BowmanSeaHardened, "Sea-hardened", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 20}}},
	},
	InRedPool: true,
}
