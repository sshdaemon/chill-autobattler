package sim

// All balance numbers and tuning constants, ported from the prototype
// (docs/GAME_CONCEPT.md §4, §5, §7, §10). Durations are in ticks.

const (
	TicksPerSecond = 20
	MaxTicks       = 3 * 60 * TicksPerSecond // a battle that runs 3 minutes is a draw
	ArmySize       = 5
	MaxSkills      = 5
)

// sec converts seconds (as tenths, to stay in integers) to ticks: sec(13) = 1.3 s.
func sec(tenths int) int { return tenths * TicksPerSecond / 10 }

// perTick converts a per-second Fx amount to per tick.
func perTick(v Fx) Fx { return v / TicksPerSecond }

type Class uint8

const (
	Warrior Class = iota
	Archer
	Mage
	numClasses
)

func (c Class) String() string { return [...]string{"warrior", "archer", "mage"}[c] }

type Stats struct {
	HP           Fx
	DmgLo, DmgHi Fx
	Cooldown     int // ticks
	Range        Fx
	Speed        Fx // metres per second
	Sight        Fx
	Ranged, Kite bool
	SwingTicks   int // attack animation length; the blow lands at HitAtPct of it
	ShotSpeed    Fx  // projectile metres per second (ranged only)
}

var ClassStats = [numClasses]Stats{
	Warrior: {HP: I(100), DmgLo: I(10), DmgHi: I(16), Cooldown: sec(13), Range: Ratio(15, 10), Speed: Ratio(32, 10), Sight: I(7),
		SwingTicks: sec(9)}, // min(cooldown*0.9, ~0.9 s clip)
	Archer: {HP: I(45), DmgLo: I(9), DmgHi: I(14), Cooldown: sec(17), Range: I(9), Speed: Ratio(34, 10), Sight: I(11),
		Ranged: true, Kite: true, SwingTicks: sec(6), ShotSpeed: I(18)},
	Mage: {HP: I(55), DmgLo: I(14), DmgHi: I(20), Cooldown: sec(22), Range: I(7), Speed: I(3), Sight: I(9),
		Ranged: true, SwingTicks: sec(8), ShotSpeed: I(14)},
}

// Matchup[attacker][victim] damage multiplier (rock-paper-scissors).
var Matchup = [numClasses][numClasses]Fx{
	Warrior: {One, Ratio(3, 2), One},
	Archer:  {Ratio(3, 4), One, Ratio(5, 4)},
	Mage:    {Ratio(5, 4), Ratio(3, 4), One},
}

// Skill identifies an effect; each class has 5 of them (SkillsOf).
type Skill uint8

const (
	SkillHealth  Skill = iota + 1 // Toughness
	SkillSpeed                    // Swift
	SkillDamage                   // Strength / Sharp arrows / Power
	SkillDash                     // warrior
	SkillExecute                  // Executioner / Headshot
	SkillRange                    // Long shot
	SkillLob                      // Arcing shot
	SkillFrost                    // mage
	SkillBlink                    // mage
	SkillBlast                    // mage
)

type SkillDef struct {
	Skill Skill
	Name  string
	Pct   int // percent bonus or chance, for percentage skills
	Value Fx  // otherwise the number the skill uses (metres, seconds)
}

// SkillsOf lists each class's skills in the order Red's random picks use.
var SkillsOf = [numClasses][MaxSkills]SkillDef{
	Warrior: {
		{SkillHealth, "Toughness", 15, 0},
		{SkillSpeed, "Swift", 25, 0},
		{SkillDamage, "Strength", 10, 0},
		{SkillDash, "Dash", 0, I(10)}, // reach in metres
		{SkillExecute, "Executioner", 25, 0},
	},
	Archer: {
		{SkillRange, "Long shot", 10, 0},
		{SkillDamage, "Sharp arrows", 15, 0},
		{SkillLob, "Arcing shot", 0, 0},
		{SkillExecute, "Headshot", 15, 0},
		{SkillHealth, "Toughness", 20, 0},
	},
	Mage: {
		{SkillFrost, "Frost", 50, 0}, // chance
		{SkillDamage, "Power", 15, 0},
		{SkillBlink, "Blink", 0, I(15)}, // cooldown seconds
		{SkillBlast, "Blast", 50, 0},    // splash damage %
		{SkillHealth, "Toughness", 20, 0},
	},
}

// SkillDefOf returns the class's definition of a skill, ok=false if the class
// can't learn it.
func SkillDefOf(c Class, s Skill) (SkillDef, bool) {
	for _, d := range SkillsOf[c] {
		if d.Skill == s {
			return d, true
		}
	}
	return SkillDef{}, false
}

// Combat and movement constants.
var (
	FlankMultiplier = I(2)
	UnitRadius      = Ratio(35, 100)
	Separation      = Ratio(9, 10)
	MeleeSlack      = Ratio(6, 10) // a melee blow still lands if the victim stepped this far out of range
	WaterSpeed      = Ratio(7, 10)
	FrostSpeed      = Ratio(6, 10)
	SlowTurnSpeed   = Ratio(3, 10) // speed factor while turning sharply

	KitePanic = Ratio(35, 10)
	KiteStep  = Ratio(25, 10)

	DashSpeed   = I(16)
	DashMinGap  = Ratio(15, 10) // dash only if the target is further than range + this
	BlinkPanic  = Ratio(25, 10)
	BlinkDist   = I(8)
	BlastRadius = I(2)

	WaypointReach = Ratio(4, 10)
	AdvanceReach  = Ratio(12, 10)
	ProjectileFwd = Ratio(4, 10) // projectiles start this far in front of the shooter
)

const (
	HitAtPct        = 45 // % into the swing where the blow lands / the shot leaves
	RetargetTicks   = 10 // 0.5 s; also how often paths are recomputed
	StaggerTicks    = 6  // 0.3 s
	KiteMaxTicks    = 24 // 1.2 s
	KiteRestTicks   = 40 // 2 s
	DashCooldown    = 200
	FrostTicks      = 60 // 3 s
	MinShotTicks    = 4  // 0.2 s minimum projectile flight
	CooldownJitter  = 15 // ±15% on every attack cooldown
	StartCooldownLo = 10 // first attack after 0.5..1.5 s
	StartCooldownHi = 30
)

// Angles, precomputed as cos/sin (×65536) so sim/ needs no trig. The turn
// rate is 8 rad/s = 0.4 rad per tick.
var (
	TurnPerTick = Turn{60363, 25521} // 0.4 rad
	CosFront    = Half               // cos 60°: hits outside ±60° of facing are flank hits
	CosSlowTurn = Fx(23747)          // cos 1.2 rad: turning sharper than this slows the walk
	CosAim      = Fx(57513)          // cos 0.5 rad: an attack starts once aimed this well
)

// Kite and blink try these deviations from the ideal escape direction, in order.
var KiteTurns = []Turn{
	{One, 0},
	{57513, 31420}, {57513, -31420}, // ±0.5
	{35409, 55147}, {35409, -55147}, // ±1
	{-1914, 65508}, {-1914, -65508}, // ±1.6
	{-38568, 52986}, {-38568, -52986}, // ±2.2
}

var BlinkTurns = []Turn{
	{One, 0},
	{60363, 25521}, {60363, -25521}, // ±0.4
	{45659, 47013}, {45659, -47013}, // ±0.8
	{17531, 63148}, {17531, -63148}, // ±1.3
	{-27273, 59592}, {-27273, -59592}, // ±2
	{-One, 0}, // π
}

// sin(2πi/64) for i = 0..16 (first quadrant); Dir64 mirrors it.
var sin64 = [17]Fx{0, 6424, 12785, 19024, 25080, 30893, 36410, 41576, 46341, 50660, 54491, 57798, 60547, 62714, 64277, 65220, 65536}

// Dir64 returns the unit direction at angle 2πi/64 (0 = +Z, 16 = +X).
func Dir64(i int) Vec {
	i &= 63
	s := func(k int) Fx { // sin(2πk/64) for k in 0..63
		switch {
		case k <= 16:
			return sin64[k]
		case k <= 32:
			return sin64[32-k]
		case k <= 48:
			return -sin64[k-32]
		default:
			return -sin64[64-k]
		}
	}
	return Vec{s(i), s((i + 16) & 63)}
}

// Arena geometry.
var (
	ArenaHalf = I(14)     // arena is 28 × 28 m around the origin
	Cell      = Half      // nav grid cell size
	EdgeLimit = I(14 - 1) // dash/blink targets must stay inside this
	ClampLim  = I(14) - UnitRadius - Ratio(2, 10)
	EnemyEdge = I(14 - 2) // units without a target march to this z on the enemy side
	Zones     = [2]Zone{  // deployment: Blue at the bottom (+z), Red at the top (-z)
		{XMin: I(-10), XMax: I(10), ZMin: I(7), ZMax: I(13)},
		{XMin: I(-10), XMax: I(10), ZMin: I(-13), ZMax: I(-7)},
	}
)

const GridN = 56 // 28 m / 0.5 m

type Zone struct{ XMin, XMax, ZMin, ZMax Fx }

func (z Zone) Contains(p Vec) bool {
	return p.X >= z.XMin && p.X <= z.XMax && p.Z >= z.ZMin && p.Z <= z.ZMax
}
