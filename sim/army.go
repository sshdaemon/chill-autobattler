package sim

import (
	"errors"
	"fmt"
)

// Formation presets: "L-R" = units on the left flank - units on the right
// flank (as seen by Blue). Custom means every unit gives its own position.
type Formation uint8

const (
	Custom Formation = iota
	Spread
	F3x2
	F2x3
	F4x1
	F1x4
	F0x5
	F5x0
)

// Presets is the order Red's random formation pick uses.
var Presets = []Formation{Spread, F3x2, F2x3, F4x1, F1x4, F0x5, F5x0}

var flankSizes = map[Formation][2]int{F3x2: {3, 2}, F2x3: {2, 3}, F4x1: {4, 1}, F1x4: {1, 4}, F0x5: {0, 5}, F5x0: {5, 0}}

func (f Formation) String() string {
	return [...]string{"Custom", "Spread", "3-2", "2-3", "4-1", "1-4", "0-5", "5-0"}[f]
}

var (
	flankX = Ratio(55, 10) // flank block centres at ±5.5
	rowGap = Ratio(24, 10)
	colGap = I(2)
)

// UnitSpec is one unit in a battle input. Pos is required for a Custom
// formation and must be nil otherwise.
type UnitSpec struct {
	Class  Class
	Skills []Skill
	Pos    *Vec
}

type Army struct {
	Formation Formation
	Units     []UnitSpec
}

// Input is everything that decides a battle. Seed (the map seed) decides the
// map and, when Red is left empty, Red's army, formation and RedSkills skills.
// CombatSeed decides every combat roll; the server reveals it only after the
// player has locked Blue's army.
type Input struct {
	Version    string
	Seed       uint64
	CombatSeed uint64
	Blue, Red  Army
	RedSkills  int
	// RedHealthPct scales Red's health (0 means 100). The server sets it from
	// the run's match number to ease the first matches.
	RedHealthPct int
}

var ErrInvalid = errors.New("invalid battle input")

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// Validate checks an input against the rules; the server rejects anything
// that fails here without playing it.
func (in Input) Validate() error {
	if in.Version != Version {
		return invalid("sim version %q, this build is %q", in.Version, Version)
	}
	if len(in.Blue.Units) == 0 {
		return invalid("blue has no units")
	}
	if len(in.Red.Units) > 0 && in.RedSkills != 0 {
		return invalid("red skill count is only used for a generated red army")
	}
	if in.RedHealthPct < 0 || in.RedHealthPct > 100 {
		return invalid("red health %d%% out of range", in.RedHealthPct)
	}
	if in.RedSkills < 0 || in.RedSkills > ArmySize*MaxSkills {
		return invalid("red skill count %d out of range", in.RedSkills)
	}
	for team, a := range [2]Army{in.Blue, in.Red} {
		if len(a.Units) > ArmySize {
			return invalid("team %d has %d units, max %d", team, len(a.Units), ArmySize)
		}
		if a.Formation > F5x0 {
			return invalid("team %d: unknown formation %d", team, a.Formation)
		}
		for i, u := range a.Units {
			if RosterOf(u.Class) == nil {
				return invalid("team %d unit %d: unknown class %d", team, i, u.Class)
			}
			if len(u.Skills) > MaxSkills {
				return invalid("team %d unit %d: too many skills", team, i)
			}
			seen := map[Skill]bool{}
			for _, s := range u.Skills {
				if _, ok := SkillDefOf(u.Class, s); !ok {
					return invalid("team %d unit %d: a %v can't learn skill %d", team, i, u.Class, s)
				}
				if seen[s] {
					return invalid("team %d unit %d: skill %d twice", team, i, s)
				}
				seen[s] = true
			}
			if a.Formation == Custom {
				if u.Pos == nil {
					return invalid("team %d unit %d: custom formation needs a position", team, i)
				}
				if !Zones[team].Contains(*u.Pos) {
					return invalid("team %d unit %d: position %v,%v is outside the deployment zone", team, i, u.Pos.X, u.Pos.Z)
				}
			} else if u.Pos != nil {
				return invalid("team %d unit %d: position given with formation %v", team, i, a.Formation)
			}
		}
	}
	return nil
}

// RedArmy generates Red's army from the seed: 5 random units from Red's
// pool (listed in roster order), a random preset, and `skills` skills spread
// at random, never twice on one unit. Changing the pool changes every Red
// army, so it needs a sim.Version bump.
func RedArmy(seed uint64, skills int) Army {
	r := Sub(seed, StreamRedArmy)
	pool := RedPool()
	counts := make([]int, len(pool))
	for i := 0; i < ArmySize; i++ {
		counts[r.IntN(len(pool))]++
	}
	var a Army
	for k, c := range pool {
		for n := 0; n < counts[k]; n++ {
			a.Units = append(a.Units, UnitSpec{Class: c})
		}
	}
	rs := Sub(seed, StreamRedSkills)
	for n := skills; n > 0; n-- {
		var open []int
		for i, u := range a.Units {
			if len(u.Skills) < MaxSkills {
				open = append(open, i)
			}
		}
		if len(open) == 0 {
			break
		}
		u := &a.Units[open[rs.IntN(len(open))]]
		var left []Skill
		for _, d := range RosterOf(u.Class).Skills {
			if !hasSkill(u.Skills, d.Skill) {
				left = append(left, d.Skill)
			}
		}
		u.Skills = append(u.Skills, left[rs.IntN(len(left))])
	}
	a.Formation = Presets[Sub(seed, StreamRedFormation).IntN(len(Presets))]
	return a
}

func hasSkill(list []Skill, s Skill) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// FormationSlots gives unsnapped positions for a team's units (in order).
// Melee stand in front, ranged behind; flank presets share classes out so
// both flanks get a mix.
func FormationSlots(f Formation, classes []Class, team int) []Vec {
	z := Zones[team]
	dir, frontZ := Fx(One), z.ZMin+One // Blue's front is toward -z
	if team == 1 {
		dir, frontZ = -One, z.ZMax-One
	}
	order := make([]int, 0, len(classes)) // melee first, then by index
	for _, ranged := range []bool{false, true} {
		for i, c := range classes {
			if d := RosterOf(c); d != nil && d.Stats.Ranged == ranged {
				order = append(order, i)
			}
		}
	}
	slots := make([]Vec, len(classes))
	n := len(order)
	if f == Spread {
		w := z.XMax - z.XMin - I(2)
		pos := make([]Vec, n)
		for i := range pos {
			x := z.XMin + One + w/2
			if n > 1 {
				x = z.XMin + One + w.MulDiv(i, n-1)
			}
			zz := frontZ
			if i%2 == 1 {
				zz = frontZ + dir.Mul(rowGap).MulDiv(3, 2)
			}
			pos[i] = V(x, zz)
		}
		var ranked []Vec // front spots first
		for i := 0; i < n; i += 2 {
			ranked = append(ranked, pos[i])
		}
		for i := 1; i < n; i += 2 {
			ranked = append(ranked, pos[i])
		}
		for k, ui := range order {
			slots[ui] = ranked[k]
		}
		return slots
	}
	sizes := flankSizes[f]
	var groups [2][]int
	for _, ui := range order { // fill whichever flank is emptiest relative to its size
		g := -1
		for _, cand := range []int{0, 1} {
			if len(groups[cand]) >= sizes[cand] {
				continue
			}
			if g < 0 || len(groups[g])*sizes[cand] > len(groups[cand])*sizes[g] {
				g = cand
			}
		}
		if g < 0 { // more units than the preset has room for: overflow right
			g = 1
		}
		groups[g] = append(groups[g], ui)
	}
	for g, members := range groups {
		cx := flankX
		if g == 0 {
			cx = -flankX
		}
		perRow := len(members)
		if perRow > 3 {
			perRow = 3
		}
		for k, ui := range members {
			r, c := k/perRow, k%perRow
			inRow := len(members) - r*perRow
			if inRow > perRow {
				inRow = perRow
			}
			slots[ui] = V(cx+colGap.MulDiv(2*c-(inRow-1), 2), frontZ+dir.Mul(rowGap).MulDiv(r, 1))
		}
	}
	return slots
}
