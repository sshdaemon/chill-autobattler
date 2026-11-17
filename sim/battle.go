// Package sim is the deterministic battle logic. It is shared by the browser
// (compiled to WASM, see ../wasm) and the server (replay verification).
// Rules: no floats, no map iteration, no time, no goroutines, randomness
// only from the battle's seeded streams. foundations_test.go enforces this.
package sim

import "hash/fnv"

// Version must change whenever the rules change, so old replays are checked
// with the logic that produced them.
const Version = "0.4.0"

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

	rng     *RNG
	spotted [2][]int // enemies each team sees, in unit order
	events  []Event
}

// New validates the input and sets up the battle at tick 0.
func New(in Input) (*Battle, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}
	b := &Battle{Input: in, Map: NewMap(GenerateObstacles(in.Seed)), rng: Sub(in.CombatSeed, StreamCombat)}
	red := in.Red
	if len(red.Units) == 0 {
		red = RedArmy(in.Seed, in.RedSkills)
	}
	for team, army := range [2]Army{in.Blue, red} {
		classes := make([]Class, len(army.Units))
		for i, s := range army.Units {
			classes[i] = s.Class
		}
		var slots []Vec
		if army.Formation != Custom {
			slots = FormationSlots(army.Formation, classes, team)
		}
		for i, s := range army.Units {
			p := slots
			pos := Vec{}
			if p != nil {
				pos = p[i]
			} else {
				pos = *s.Pos
			}
			if !b.Map.FreeAt(pos) { // nudge off rocks and trees
				pos = b.Map.FreeSpot(pos)
			}
			u := newUnit(len(b.Units), team, s, pos)
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

// Step advances the battle by one tick (1/20 s).
func (b *Battle) Step() {
	if b.Result != Ongoing {
		return
	}
	b.Tick++
	b.updateVision()
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

// nearestEnemy picks the closest enemy the unit's team sees; ties go to the lower index.
func (b *Battle) nearestEnemy(u *Unit) int {
	best, bd := -1, Fx(0)
	for _, i := range b.spotted[u.Team] {
		if d := Dist(u.Pos, b.Units[i].Pos); best < 0 || d < bd {
			best, bd = i, d
		}
	}
	return best
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
