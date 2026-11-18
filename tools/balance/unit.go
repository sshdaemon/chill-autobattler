package main

import (
	"fmt"
	"strings"

	"chill/sim"
)

// Balance bands for a new unit (openspec unit-authoring, "New units are
// balanced against the roster").
const (
	h2hLo, h2hHi = 0.25, 0.75 // 5 of it against 5 of each other unit
	swapBand     = 0.10       // swapped into the standard army: within ±10 points
	skillBand    = 0.10       // each ability as the level-1 pick: within ±10 points of the unit's average
)

type check struct {
	what, result, band string
	ok                 bool
}

// standard is today's default army (2W 2A 1M in 3-2) against the game's seeded
// Red; swap replaces one unit of the same role (melee: a warrior, ranged: an archer).
func standard(i int) sim.Input {
	in := input(500_000 + uint64(i))
	in.Blue = army(sim.F3x2, mix(2, 2, 1)...)
	return in
}

func swapSlot(u sim.Class) int {
	if sim.Roster[u].Stats.Ranged {
		return 2 // the first archer
	}
	return 0 // the first warrior
}

// UnitReport measures a roster entry against the bands. ok is false if any
// band fails.
func UnitReport(name string, n int) (string, bool, error) {
	u := sim.Class(255)
	for _, d := range sim.Roster {
		if d.Name == name {
			u = d.ID
		}
	}
	if u == 255 {
		return "", false, fmt.Errorf("no roster entry %q", name)
	}
	var checks []check
	for _, d := range sim.Roster {
		if d.ID == u {
			continue
		}
		blue, red := headToHead(u, d.ID, n), headToHead(d.ID, u, n) // both sides, so side bias cancels
		w := (blue.win() + 1 - red.win() - float64(red.draw)/float64(red.n)) / 2
		checks = append(checks, check{"5× " + name + " vs 5× " + d.Name, pct(w), "25–75%", w >= h2hLo && w <= h2hHi})
	}

	slot := swapSlot(u)
	var base, swapped []sim.Input
	for i := 0; i < n; i++ {
		base = append(base, standard(i))
		in := standard(i)
		in.Blue.Units[slot] = recruit(u, i) // its abilities in turn, so the swap measures the unit, not one ability
		swapped = append(swapped, in)
	}
	bt, st := count(play(base)), count(play(swapped))
	d := st.win() - bt.win()
	checks = append(checks, check{fmt.Sprintf("standard army with one %s swapped for %s", title(mix(2, 2, 1)[slot]), name),
		fmt.Sprintf("%s (base %s, %+.1f pts)", pct(st.win()), pct(bt.win()), 100*d), "±10 pts", d >= -swapBand && d <= swapBand})

	// each ability as its level-1 pick, against the average of all five: none
	// should be a trap or a must-pick
	var rates [sim.MaxSkills]float64
	avg := 0.0
	for k, sd := range sim.Roster[u].Skills {
		var with []sim.Input
		for i := 0; i < n; i++ {
			in := standard(i)
			in.Blue.Units[slot] = sim.UnitSpec{Class: u, Skills: []sim.Skill{sd.Skill}}
			with = append(with, in)
		}
		rates[k] = count(play(with)).win()
		avg += rates[k] / sim.MaxSkills
	}
	for k, sd := range sim.Roster[u].Skills {
		d := rates[k] - avg
		checks = append(checks, check{"level-1 pick " + sd.Name, fmt.Sprintf("%s (%+.1f pts vs its average)", pct(rates[k]), 100*d), "±10 pts", d >= -skillBand && d <= skillBand})
	}

	var b strings.Builder
	ok := true
	fmt.Fprintf(&b, "# Balance: %s\n\n%d battles per row, fixed seeds (sim %s). Bands: openspec unit-authoring.\n\n| Check | Result | Band | |\n|---|---|---|---|\n", name, n, sim.Version)
	for _, c := range checks {
		mark := "pass"
		if !c.ok {
			mark, ok = "**FAIL**", false
		}
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", c.what, c.result, c.band, mark)
	}
	return b.String(), ok, nil
}
