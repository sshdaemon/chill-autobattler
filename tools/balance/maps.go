package main

import (
	"fmt"
	"math"
	"strings"

	"chill/sim"
)

// mapSeeds are the maps the map experiment fixes (a battle's map comes from
// its Seed alone when both armies are given).
var mapSeeds = []uint64{101, 202, 303, 404, 505, 606, 707, 808, 909, 1010}

// mapLine describes a map: how many rocks, trees and ponds, and how much
// obstacle area lies in Blue's half (z > 0) against Red's.
func mapLine(seed uint64) string {
	var rocks, trees int
	ponds := map[int]bool{}
	var blue, red float64
	for _, o := range sim.GenerateObstacles(seed) {
		switch o.Kind {
		case sim.Rock:
			rocks++
		case sim.Tree:
			trees++
		case sim.Water:
			ponds[o.Pond] = true
		}
		r := float64(o.R) / 65536
		if o.Pos.Z > 0 {
			blue += math.Pi * r * r
		} else {
			red += math.Pi * r * r
		}
	}
	return fmt.Sprintf("%d rocks, %d trees, %d ponds; obstacle area Blue half %.0f m², Red half %.0f m²", rocks, trees, len(ponds), blue, red)
}

// mapCell is faction a's win rate against faction b (draws count half) on one
// map: the same armies every battle (each faction's units 2-2-1, recruits, in
// 3-2, Instinct), only the combat rolls change. With swap, a plays Red every
// other battle so side bias cancels; without, a is always Blue.
func mapCell(seed uint64, a, b *sim.FactionDef, n int, swap bool) float64 {
	var ins []sim.Input
	for i := 0; i < n; i++ {
		blue, red := a, b
		if swap && i%2 == 1 {
			blue, red = b, a
		}
		ins = append(ins, sim.Input{Version: sim.Version, Seed: seed, CombatSeed: seed*1_000_003 + uint64(i)*7919 + 1,
			BlueFaction: blue.ID, Blue: planArmy(blue, sim.F3x2, sim.PlanInstinct), Red: planArmy(red, sim.F3x2, sim.PlanInstinct)})
	}
	w := 0.0
	for i, r := range play(ins) {
		aBlue := !swap || i%2 == 0
		switch {
		case r.result == sim.Draw:
			w += 0.5
		case (r.result == sim.BlueWin) == aBlue:
			w++
		}
	}
	return w / float64(n)
}

// MapEffects asks whether the random map decides battles: on each of the
// fixed maps, (1) the same army on both sides (does a map favour a side?) and
// (2) every pair of playable factions, sides swapped (does a map swing a
// matchup more than the combat rolls do?). n battles per cell.
func MapEffects(n int) string {
	var b strings.Builder
	var fs []*sim.FactionDef
	for i := range sim.Factions {
		if sim.Factions[i].Playable {
			fs = append(fs, &sim.Factions[i])
		}
	}
	se := 0.5 / math.Sqrt(float64(n)) // one cell's noise at 50%
	fmt.Fprintf(&b, "%d maps, %d battles per cell; armies: each faction's units 2-2-1, recruits, 3-2, Instinct; only the combat seed changes within a cell. Noise (one standard error) at 50%%: ±%.1f pts; ⚠ marks a cell more than 3 standard errors from its expectation.\n\n", len(mapSeeds), n, 100*se)

	b.WriteString("### Maps\n\n| Map | Layout |\n|---|---|\n")
	for _, s := range mapSeeds {
		fmt.Fprintf(&b, "| %d | %s |\n", s, mapLine(s))
	}

	// (1) side bias per map: mirror matches
	b.WriteString("\n### Side advantage per map\n\nBlue's win rate with the same army on both sides. A fair map gives 50%.\n\n| Map |")
	for _, f := range fs {
		fmt.Fprintf(&b, " %s |", f.Name)
	}
	b.WriteString(" All factions |\n|---|" + strings.Repeat("---|", len(fs)+1) + "\n")
	for _, s := range mapSeeds {
		fmt.Fprintf(&b, "| %d |", s)
		sum := 0.0
		for _, f := range fs {
			p := mapCell(s, f, f, n, false)
			sum += p
			fmt.Fprintf(&b, " %s%s |", pct(p), offBy3(p, 0.5, se))
		}
		all := sum / float64(len(fs))
		fmt.Fprintf(&b, " **%s**%s |\n", pct(all), offBy3(all, 0.5, se/math.Sqrt(float64(len(fs)))))
	}

	// (2) matchup swing per map
	b.WriteString("\n### Matchups per map\n\nThe first faction's win rate against the second, sides swapped every other battle. Spread: the standard deviation of the ten maps' numbers, against what the combat rolls alone would give (×1 means maps add nothing).\n\n| Matchup |")
	for _, s := range mapSeeds {
		fmt.Fprintf(&b, " %d |", s)
	}
	b.WriteString(" Mean | Min–max | Spread |\n|---|" + strings.Repeat("---|", len(mapSeeds)+3) + "\n")
	for i := 0; i < len(fs); i++ {
		for j := i + 1; j < len(fs); j++ {
			var ps []float64
			for _, s := range mapSeeds {
				ps = append(ps, mapCell(s, fs[i], fs[j], n, true))
			}
			mean, lo, hi := 0.0, 1.0, 0.0
			for _, p := range ps {
				mean += p
				lo, hi = math.Min(lo, p), math.Max(hi, p)
			}
			mean /= float64(len(ps))
			sd := 0.0
			for _, p := range ps {
				sd += (p - mean) * (p - mean)
			}
			sd = math.Sqrt(sd / float64(len(ps)-1))
			noise := math.Sqrt(mean * (1 - mean) / float64(n))
			fmt.Fprintf(&b, "| %s vs %s |", fs[i].Name, fs[j].Name)
			for _, p := range ps {
				fmt.Fprintf(&b, " %s%s |", pct(p), offBy3(p, mean, noise))
			}
			fmt.Fprintf(&b, " %s | %s–%s | ±%.1f pts (×%.1f) |\n", pct(mean), pct(lo), pct(hi), 100*sd, sd/noise)
		}
	}
	return b.String()
}

// offBy3 marks p more than 3 standard errors from want.
func offBy3(p, want, se float64) string {
	if math.Abs(p-want) > 3*se {
		return " ⚠"
	}
	return ""
}
