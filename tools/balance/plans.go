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

// PlanMatrix: for each playable faction, the same army on both sides, Blue
// on the row's plan against Red on the column's plan, n battles per cell over
// all formation presets. "Avg" is the row plan's mean over the columns; the
// diagonal shows the side bias.
func PlanMatrix(n int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Blue win rate, rows = Blue's plan, columns = Red's plan; the same army (the faction's units 2-2-1, recruits) on both sides, %d battles per cell. A plan that wins every row cell dominates.\n", n)
	for fi := range sim.Factions {
		f := &sim.Factions[fi]
		if !f.Playable {
			continue
		}
		fmt.Fprintf(&b, "\n### %s\n\n| Blue \\ Red |", f.Name)
		for _, p := range sim.Plans {
			fmt.Fprintf(&b, " %s |", p.Name)
		}
		b.WriteString(" Avg |\n|---|" + strings.Repeat("---|", len(sim.Plans)+1) + "\n")
		for _, pa := range sim.Plans {
			fmt.Fprintf(&b, "| **%s** |", pa.Name)
			sum := 0.0
			for _, pr := range sim.Plans {
				var ins []sim.Input
				for i := 0; i < n; i++ {
					form := sim.Presets[i%len(sim.Presets)]
					in := input(300_000 + uint64(fi)*10_000 + uint64(i))
					in.BlueFaction = f.ID
					in.Blue = planArmy(f, form, pa.ID)
					in.Red = planArmy(f, form, pr.ID)
					ins = append(ins, in)
				}
				w := count(play(ins)).win()
				sum += w
				fmt.Fprintf(&b, " %s |", pct(w))
			}
			fmt.Fprintf(&b, " %s |\n", pct(sum/float64(len(sim.Plans))))
		}
	}
	return b.String()
}
