package main

import (
	"fmt"
	"strings"

	"chill/sim"
)

// planArmy is a faction's mixed army: its three units as 2-2-1, recruits.
func planArmy(f *sim.FactionDef, form sim.Formation, plan sim.Plan) sim.Army {
	var cs []sim.Class
	for i := 0; i < sim.ArmySize; i++ {
		cs = append(cs, f.Units[i*len(f.Units)/sim.ArmySize])
	}
	a := army(form, cs...)
	a.Plan = plan
	return a
}

// planCell is plan a's win rate against plan b (draws count half): the same
// army on both sides, sides swapped every other battle so side bias cancels.
func planCell(fi int, f *sim.FactionDef, a, b sim.Plan, n int) float64 {
	var ins []sim.Input
	for i := 0; i < n; i++ {
		blue, red := a, b
		if i%2 == 1 {
			blue, red = b, a
		}
		form := sim.Presets[i%len(sim.Presets)]
		in := input(300_000 + uint64(fi)*10_000 + uint64(i))
		in.BlueFaction = f.ID
		in.Blue = planArmy(f, form, blue)
		in.Red = planArmy(f, form, red)
		ins = append(ins, in)
	}
	w := 0.0
	for i, r := range play(ins) {
		switch {
		case r.result == sim.Draw:
			w += 0.5
		case (r.result == sim.BlueWin) == (i%2 == 0):
			w++
		}
	}
	return w / float64(n)
}

// PlanMatrix: for each playable faction, every plan against every other with
// the same army on both sides, n battles per pair. A plan "beats" another at
// 55% or more; one that beats every other plan dominates, which the balance
// goal (docs/battle-plans-roadmap.md, step 3) rules out.
func PlanMatrix(n int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Row plan's win rate against the column plan; the same army (the faction's units 2-2-1, recruits) on both sides, sides swapped every other battle, %d battles per pair. Beats: 55%% or more.\n", n)
	np := len(sim.Plans)
	for fi := range sim.Factions {
		f := &sim.Factions[fi]
		if !f.Playable {
			continue
		}
		m := make([][]float64, np)
		for x := range m {
			m[x] = make([]float64, np)
			m[x][x] = 0.5
		}
		for x := 0; x < np; x++ {
			for y := x + 1; y < np; y++ {
				m[x][y] = planCell(fi, f, sim.Plans[x].ID, sim.Plans[y].ID, n)
				m[y][x] = 1 - m[x][y]
			}
		}
		fmt.Fprintf(&b, "\n### %s\n\n| Plan |", f.Name)
		for _, p := range sim.Plans {
			fmt.Fprintf(&b, " %s |", p.Name)
		}
		b.WriteString(" Avg | Beats |\n|---|" + strings.Repeat("---|", np+2) + "\n")
		for x, p := range sim.Plans {
			fmt.Fprintf(&b, "| **%s** |", p.Name)
			sum, beats := 0.0, 0
			for y := range sim.Plans {
				if x == y {
					b.WriteString(" — |")
					continue
				}
				sum += m[x][y]
				if m[x][y] >= 0.55 {
					beats++
				}
				fmt.Fprintf(&b, " %s |", pct(m[x][y]))
			}
			note := fmt.Sprint(beats)
			if beats == np-1 {
				note += " ❌ dominates"
			}
			fmt.Fprintf(&b, " %s | %s |\n", pct(sum/float64(np-1)), note)
		}
	}
	return b.String()
}
