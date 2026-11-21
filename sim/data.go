package sim

// Tuning constants shared by every unit, ported from the prototype
// (docs/GAME_CONCEPT.md §5, §7). Units, their stats, matchups and skills are
// roster data (roster.go, unit_*.go). Durations are in ticks.

const (
	TicksPerSecond = 20
	MaxTicks       = 3 * 60 * TicksPerSecond // a battle that runs 3 minutes is a draw
	ArmySize       = 5
	MaxSkills      = 5 // abilities each unit type offers
	MaxLevel       = 3 // abilities a unit knows at most: one per level (a recruit is level 1)
)

// sec converts seconds (as tenths, to stay in integers) to ticks: sec(13) = 1.3 s.
func sec(tenths int) int { return tenths * TicksPerSecond / 10 }

// perTick converts a per-second Fx amount to per tick.
func perTick(v Fx) Fx { return v / TicksPerSecond }

type Stats struct {
	HP           Fx
	DmgLo, DmgHi Fx
	Cooldown     int // ticks
	Range        Fx
	Speed        Fx // metres per second
	Sight        Fx
	Ranged       bool
	SwingTicks   int // attack animation length; the blow lands at HitAtPct of it
	AuraPct      int // auras reach +AuraPct% (StatAuraRadius)
	ShotSpeed    Fx  // projectile metres per second (ranged only)
}

// Skill is an ability's number within its unit's roster entry: 1..MaxSkills,
// in the order the entry lists them (unit_<name>.go names them). Numbers mean
// nothing across units, so units never share or compete for skill ids.
type Skill uint8

// Combat and movement constants.
var (
	FlankMultiplier = I(2)
	UnitRadius      = Ratio(35, 100)
	Separation      = Ratio(9, 10)
	MeleeSlack      = Ratio(6, 10) // a melee blow still lands if the victim stepped this far out of range
	WaterSpeed      = Ratio(7, 10)
	SlowTurnSpeed   = Ratio(3, 10) // speed factor while turning sharply
	BackpedalSpeed  = One          // kiting archers step back without turning round

	KitePanic = Ratio(45, 10)
	KiteStep  = Ratio(25, 10)

	DashMinGap = Ratio(15, 10) // dash only if the target is further than range + this

	WaypointReach = Ratio(4, 10)
	AdvanceReach  = Ratio(12, 10)
	ProjectileFwd = Ratio(4, 10) // projectiles start this far in front of the shooter
)

const (
	HitAtPct        = 45 // % into the swing where the blow lands / the shot leaves
	ExecuteBelowPct = 40 // Executioner / Headshot only work on victims below this % of max health
	RetargetTicks   = 10 // 0.5 s; also how often paths are recomputed
	StaggerTicks    = 6  // 0.3 s
	KiteMaxTicks    = 40 // 1.2 s
	KiteRestTicks   = 24 // 2 s
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
	ArenaHalf    = I(14)     // arena is 28 × 28 m around the origin
	Cell         = Half      // nav grid cell size
	EdgeLimit    = I(14 - 1) // dash/blink targets must stay inside this
	ClampLim     = I(14) - UnitRadius - Ratio(2, 10)
	EnemyEdge    = I(14 - 2) // units without a target march to this z on the enemy side
	WoundedReach = I(4)      // TargetWounded looks for the most wounded enemy this close
	ProtectReach = I(8)      // TargetProtect: enemies this close to a ranged or caster ally
	FocusSlack   = I(2)      // TargetFocusReach: how far beyond its range it still joins a shared target
	LineSlack    = I(2)      // EngageLine: how far ahead of its allies' centre a unit may get
	StayReach    = I(4)      // StayNear (bodyguard): how far it may get from its ward
	HoldPatience = sec(40)   // EngageHold: how long it waits for a seen enemy before going in
	FlankWide    = I(6)      // EngageFlank: how far beside its target it swings
	FlankNear    = I(4)      // EngageFlank: this close to its target it goes straight in
	Zones        = [2]Zone{  // deployment: Blue at the bottom (+z), Red at the top (-z)
		{XMin: I(-10), XMax: I(10), ZMin: I(7), ZMax: I(13)},
		{XMin: I(-10), XMax: I(10), ZMin: I(-13), ZMax: I(-7)},
	}
)

const GridN = 56 // 28 m / 0.5 m

type Zone struct{ XMin, XMax, ZMin, ZMax Fx }

func (z Zone) Contains(p Vec) bool {
	return p.X >= z.XMin && p.X <= z.XMax && p.Z >= z.ZMin && p.Z <= z.ZMax
}

// RunRedHealth is Red's health % for match m of a run (0 = full). The first
// matches are eased so a run doesn't hinge on a coin flip; tuned with
// tools/balance (docs/balance-findings.md). The server sets it per battle and
// stores it, so changing the curve never changes already issued battles.
func RunRedHealth(match int) int {
	if match >= 1 && match <= len(runRedHealth) {
		return runRedHealth[match-1]
	}
	return 0
}

var runRedHealth = []int{85, 90, 95}
