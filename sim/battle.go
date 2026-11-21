// Package sim is the deterministic battle logic. It is shared by the browser
// (compiled to WASM, see ../wasm) and the server (replay verification).
// Rules: no floats, no map iteration, no time, no goroutines, randomness
// only from the battle's seeded streams. foundations_test.go enforces this.
package sim

import (
	"fmt"
	"hash/fnv"
	"sort"
)

// RulesVersion must change whenever the rules code changes, so old replays
// are checked with the logic that produced them. Changes to units' data need
// no bump: Version carries a fingerprint of the roster.
const RulesVersion = "0.13.0"

// Version is RulesVersion plus a fingerprint of the roster and the factions
// ("0.7.0+1a2b3c4d"), so changing any unit's numbers changes it without
// editing a shared constant. Set as unit and faction files register.
var Version = RulesVersion

func refreshVersion() {
	h := fnv.New32a()
	fmt.Fprintf(h, "%s;%#v;%s", Fingerprint(Roster...), Factions, plansFingerprint())
	Version = fmt.Sprintf("%s+%08x", RulesVersion, h.Sum32())
}

// Fingerprint is a short hash of roster entries' data (stats, temperament,
// tags, skills and their effects), in the order given.
func Fingerprint(defs ...UnitDef) string {
	h := fnv.New32a()
	for _, d := range defs {
		fmt.Fprintf(h, "%#v;", d)
	}
	return fmt.Sprintf("%08x", h.Sum32())
}

// VersionFor is the version of the rules a battle between these units depends
// on: RulesVersion plus the fingerprint of just those units (sorted by id).
// Golden replays record it per case, so changing one unit only touches the
// cases that use it.
func VersionFor(classes ...Class) string {
	cs := append([]Class(nil), classes...)
	sort.Slice(cs, func(i, j int) bool { return cs[i] < cs[j] })
	var defs []UnitDef
	for i, c := range cs {
		if i > 0 && c == cs[i-1] {
			continue
		}
		if d := RosterOf(c); d != nil {
			defs = append(defs, *d)
		}
	}
	return RulesVersion + "+" + Fingerprint(defs...) + "." + plansFingerprint()
}

type Result int8

const (
	Ongoing Result = iota
	BlueWin
	RedWin
	Draw
)

type Battle struct {
	Input  Input
	Map    *Map
	Units  []*Unit // Blue in input order, then Red
	Shots  []*Shot
	Tick   int
	Result Result
	Plans  [2]Plan // each side's plan; a generated Red's is drawn (RedPlanFor)

	switched [2][2]bool // a tactic's group has switched to its second stance, per team and group

	rng     *RNG
	spotted [2][]int // enemies each team sees, in unit order
	zones   []zone   // burning ground
	events  []Event
}

// redPlanFor is RedPlanFor; tests may pin it.
var redPlanFor = RedPlanFor

// New validates the input and sets up the battle at tick 0.
func New(in Input) (*Battle, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	b := &Battle{Input: in, Map: NewMap(GenerateObstacles(in.Seed)), rng: Sub(in.CombatSeed, StreamCombat)}
	red := in.Red
	if len(red.Units) == 0 {
		red = RedArmy(in.Seed, in.RedSkills)
		red.Plan = redPlanFor(in.CombatSeed)
	}
	b.Plans = [2]Plan{in.Blue.Plan, red.Plan}
	for team, army := range [2]Army{in.Blue, red} {
		classes := make([]Class, len(army.Units))
		for i, s := range army.Units {
			classes[i] = s.Class
		}
		var slots []Vec
		if army.Formation != Custom {
			slots = FormationSlots(army.Formation, classes, team)
		}
		pos := make([]Vec, len(army.Units))
		for i, s := range army.Units {
			if slots != nil {
				pos[i] = slots[i]
			} else {
				pos[i] = *s.Pos
			}
			if !b.Map.FreeAt(pos[i]) { // nudge off rocks and trees
				pos[i] = b.freeSpot(team, pos[i])
			}
		}
		var groups []int
		if p := PlanOf(army.Plan); p != nil && p.IsTactic() {
			groups = GroupsOf(pos)
		}
		for i, s := range army.Units {
			g := 0
			if groups != nil {
				g = groups[i]
			}
			u := newUnitGroup(len(b.Units), team, s, pos[i], army.Plan, g)
			if team == 1 && in.RedHealthPct > 0 && in.RedHealthPct < 100 {
				u.MaxHP = u.MaxHP.MulDiv(in.RedHealthPct, 100)
				u.HP = u.MaxHP
			}
			b.Units = append(b.Units, u)
		}
	}
	for _, u := range b.Units { // stagger first decisions so units don't act in lockstep
		u.retarget = b.rng.IntN(RetargetTicks)
		u.cd = StartCooldownLo + b.rng.IntN(StartCooldownHi-StartCooldownLo)
	}
	return b, nil
}

// On a mirrored map Red finds paths and free spots in Blue's frame: the
// question is reflected, answered, and the answer reflected back, so every
// choice the grid makes between equal options (which side of a rock, which
// free cell) is the mirror image of Blue's and neither side gains from it.
func (b *Battle) inBlueFrame(team int) bool { return team == 1 && b.Map.Mirrored }

func (b *Battle) freeSpot(team int, p Vec) Vec {
	if b.inBlueFrame(team) {
		return flipZ(b.Map.FreeSpot(flipZ(p)))
	}
	return b.Map.FreeSpot(p)
}

func (b *Battle) reachable(team int, p Vec) Vec {
	if b.inBlueFrame(team) {
		return flipZ(b.Map.Reachable(flipZ(p)))
	}
	return b.Map.Reachable(p)
}

func (b *Battle) findPath(team int, from, to Vec) []Vec {
	if !b.inBlueFrame(team) {
		return b.Map.FindPath(from, to)
	}
	path := b.Map.FindPath(flipZ(from), flipZ(to))
	for i := range path {
		path[i] = flipZ(path[i])
	}
	return path
}

// GroupsOf splits an army into a tactic's two groups by starting position:
// units at or left of the army's mean x (−x, the formation's first flank)
// form group 0, the others group 1.
func GroupsOf(pos []Vec) []int {
	if len(pos) == 0 {
		return nil
	}
	var sum Fx
	for _, p := range pos {
		sum += p.X
	}
	n := Fx(len(pos))
	out := make([]int, len(pos))
	for i, p := range pos {
		if p.X*n > sum { // x > mean, without rounding the mean
			out[i] = 1
		}
	}
	return out
}

// Step advances the battle by one tick (1/20 s).
func (b *Battle) Step() {
	if b.Result != Ongoing {
		return
	}
	b.Tick++
	b.checkTriggers()
	b.updateVision()
	b.computeMods()
	for _, u := range b.Units {
		if !u.Dead {
			b.think(u)
		}
	}
	live := b.alive(-1)
	for _, u := range live {
		b.move(u)
	}
	for i := 0; i < len(live); i++ { // units push apart, in index order
		for j := i + 1; j < len(live); j++ {
			a, c := live[i], live[j]
			d := c.Pos.Sub(a.Pos)
			dist := d.Len()
			if dist > 0 && dist < Separation {
				push := d.Norm().Scale((Separation - dist) / 2)
				a.Pos, c.Pos = a.Pos.Sub(push), c.Pos.Add(push)
			}
		}
	}
	for _, u := range live {
		u.Pos = b.Map.PushOut(u.Pos)
	}
	b.updateShots()
	b.updateZones()

	blue, red := len(b.alive(0)), len(b.alive(1))
	switch {
	case blue == 0 && red == 0:
		b.Result = Draw
	case blue == 0:
		b.Result = RedWin
	case red == 0:
		b.Result = BlueWin
	case b.Tick >= MaxTicks:
		b.Result = Draw
	}
}

func (b *Battle) alive(team int) []*Unit {
	var out []*Unit
	for _, u := range b.Units {
		if !u.Dead && (team < 0 || u.Team == team) {
			out = append(out, u)
		}
	}
	return out
}

// shared team vision: an enemy is spotted if any living teammate sees it
func (b *Battle) updateVision() {
	for team := 0; team < 2; team++ {
		b.spotted[team] = b.spotted[team][:0]
		eyes := b.alive(team)
		for _, e := range b.Units {
			if e.Dead || e.Team == team {
				continue
			}
			for _, f := range eyes {
				if b.canSee(f, e) {
					b.spotted[team] = append(b.spotted[team], e.Index)
					break
				}
			}
		}
	}
}

func (b *Battle) canSee(f, e *Unit) bool {
	return Dist(f.Pos, e.Pos) <= f.Stats.Sight && b.Map.ClearView(f.Pos, e.Pos)
}

// Outcome is what a finished (or current) battle produced.
type Outcome struct {
	Result    Result
	Ticks     int
	Survivors []Survivor
	Checksum  uint64
}

type Survivor struct {
	Index int
	HP    Fx
}

func (b *Battle) Outcome() Outcome {
	o := Outcome{Result: b.Result, Ticks: b.Tick, Checksum: b.Checksum()}
	for _, u := range b.Units {
		if !u.Dead {
			o.Survivors = append(o.Survivors, Survivor{u.Index, u.HP})
		}
	}
	return o
}

// Checksum hashes the whole state; equal checksums == identical battles.
func (b *Battle) Checksum() uint64 {
	h := fnv.New64a()
	var buf [8]byte
	w := func(v int64) {
		for i := range buf {
			buf[i] = byte(v >> (8 * i))
		}
		h.Write(buf[:])
	}
	w(int64(b.Tick))
	w(int64(b.Result))
	for _, u := range b.Units {
		w(int64(u.Pos.X))
		w(int64(u.Pos.Z))
		w(int64(u.Face.X))
		w(int64(u.Face.Z))
		w(int64(u.HP))
		w(int64(u.cd))
		w(int64(u.Target))
		if u.Dead {
			w(1)
		} else {
			w(0)
		}
	}
	w(int64(len(b.Shots)))
	for _, s := range b.Shots {
		w(int64(s.Victim))
		w(int64(s.T))
	}
	return h.Sum64()
}

// Run plays a battle to the end.
func Run(in Input) (Outcome, error) {
	b, err := New(in)
	if err != nil {
		return Outcome{}, err
	}
	for b.Result == Ongoing {
		b.Step()
	}
	return b.Outcome(), nil
}

// Events are what happened since the last TakeEvents call, for the renderer
// (floating numbers, skill labels). They never affect the battle.
type EventKind uint8

const (
	EvHit EventKind = iota + 1
	EvDeath
	EvDash
	EvBlink
	EvFrost
	EvBlast
	EvExecute
	EvShot
	EvMiss  // a blow was dodged (Unit: who dodged)
	EvFear  // Unit is struck with dread by Other
	EvBurst // Unit throws a burst attack (spear, fire pot) at Other
	EvJet   // Unit's fire jet (a cone attack) hits around Other
	EvBurn  // burning ground starts under Unit (Other: who set it)
	EvRally // Unit is rallied by an ally's death
	EvGuard // Unit takes Amount of a blow aimed at Other
	EvBrace // Unit braced against Other's charge
	EvCurse // Unit is cursed (takes more damage) by Other
)

const (
	FlagFlank uint8 = 1 << iota
	FlagStrong
	FlagWeak
)

type Event struct {
	Kind   EventKind
	Unit   int // victim / actor
	Other  int // attacker, or -1
	Amount Fx
	Flags  uint8
}

func (b *Battle) emit(e Event) { b.events = append(b.events, e) }

func (b *Battle) TakeEvents() []Event {
	ev := b.events
	b.events = nil
	return ev
}
