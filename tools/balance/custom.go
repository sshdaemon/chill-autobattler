package main

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"

	"chill/sim"
)

// A side's plan for the lab: a built-in plan id, or a custom plan.
type planSide struct {
	plan   sim.Plan
	custom *sim.CustomPlan
}

func (p planSide) army(f *sim.FactionDef, form sim.Formation) sim.Army {
	a := planArmy(f, form, p.plan)
	a.Custom = p.custom
	return a
}

// sideCell is a's win rate against b (draws half), the faction's 2-2-1 army on
// both sides, sides swapped every other battle; seeds from base.
func sideCell(f *sim.FactionDef, a, b planSide, n int, base uint64) float64 {
	var ins []sim.Input
	for i := 0; i < n; i++ {
		blue, red := a, b
		if i%2 == 1 {
			blue, red = b, a
		}
		form := sim.Presets[i%len(sim.Presets)]
		in := input(base + uint64(i))
		in.BlueFaction = f.ID
		in.Blue, in.Red = blue.army(f, form), red.army(f, form)
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

func playableFactions() []*sim.FactionDef {
	var out []*sim.FactionDef
	for i := range sim.Factions {
		if sim.Factions[i].Playable {
			out = append(out, &sim.Factions[i])
		}
	}
	return out
}

// EdgeCheck: each edge, on every role it suits, with the units' own behaviour,
// against the same army fighting by Instinct. Goal: none over 60%.
func EdgeCheck(n int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Win rate of the army with the edge (on every role it suits, own behaviour) against the same army by Instinct; each faction's units 2-2-1, recruits, sides swapped, %d battles per cell. Goal: none over 60%%.\n\n| Edge |", n)
	fs := playableFactions()
	for _, f := range fs {
		fmt.Fprintf(&b, " %s |", f.Name)
	}
	b.WriteString("\n|---|" + strings.Repeat("---|", len(fs)) + "\n")
	for _, e := range sim.Edges {
		cp := &sim.CustomPlan{}
		for r := 0; r < 3; r++ {
			if e.Roles[r] {
				cp.Roles[r].Edge = e.ID
			}
		}
		fmt.Fprintf(&b, "| %s |", e.Name)
		for fi, f := range fs {
			p := sideCell(f, planSide{custom: cp}, planSide{}, n, 500_000+uint64(fi)*10_000+uint64(e.ID)*1_000)
			mark := ""
			if p > 0.60 {
				mark = " ❌"
			}
			fmt.Fprintf(&b, " %s%s |", pct(p), mark)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// randomChips draws one role's chips; second stances take no edge. Def
// decides what is valid; callers retry.
func randomChips(r *rand.Rand, role int, edge bool) sim.Chips {
	var c sim.Chips
	maybe := func(n int) int {
		if r.IntN(2) == 0 {
			return 0
		}
		return r.IntN(n)
	}
	c.Target = sim.TargetRule(maybe(8))
	c.Engage = sim.EngageRule(maybe(5))
	if c.Engage == sim.EngageHold {
		c.Trigger = sim.Step(1 + r.IntN(3))
	}
	c.Keep = sim.Step(maybe(4))
	c.Leash = sim.Step(maybe(3))
	c.FallBack = uint8(maybe(3))
	c.StayNear = uint8(maybe(3))
	if role == 1 {
		c.Kite = sim.Toggle(maybe(3))
	}
	if edge && r.IntN(3) > 0 {
		var ok []sim.Edge
		for _, e := range sim.Edges {
			if e.Roles[role] {
				ok = append(ok, e.ID)
			}
		}
		c.Edge = ok[r.IntN(len(ok))]
	}
	return c
}

func randomTrigger(r *rand.Rand) sim.CustomTrigger {
	switch r.IntN(4) {
	case 0:
		return sim.CustomTrigger{Kind: sim.TriggerAfter, Seconds: sim.TriggerSeconds[r.IntN(len(sim.TriggerSeconds))]}
	case 1:
		return sim.CustomTrigger{Kind: sim.TriggerAllyFalls}
	case 2:
		return sim.CustomTrigger{Kind: sim.TriggerOtherFighting}
	}
	return sim.CustomTrigger{}
}

func randomCustom(r *rand.Rand) *sim.CustomPlan {
	for {
		cp := &sim.CustomPlan{}
		for role := 0; role < 3; role++ {
			cp.Roles[role] = randomChips(r, role, true)
		}
		if r.IntN(3) == 0 {
			cp.Groups = make([]sim.CustomGroup, 2)
			for g := range cp.Groups {
				for role := 0; role < 3; role++ {
					if r.IntN(2) == 0 {
						cp.Groups[g].Opening[role] = randomChips(r, role, false)
					}
					if r.IntN(2) == 0 {
						cp.Groups[g].Then[role] = randomChips(r, role, false)
					}
				}
				cp.Groups[g].When = randomTrigger(r)
			}
		}
		if _, err := cp.Def(); err == nil {
			return cp
		}
	}
}

// mutate changes one role's chips (army-wide, or one group's stance or
// trigger) and returns a valid copy.
func mutate(r *rand.Rand, cp *sim.CustomPlan) *sim.CustomPlan {
	for {
		c := &sim.CustomPlan{Roles: cp.Roles, Groups: append([]sim.CustomGroup(nil), cp.Groups...)}
		role := r.IntN(3)
		switch {
		case len(c.Groups) == 0 || r.IntN(2) == 0:
			c.Roles[role] = randomChips(r, role, true)
		case r.IntN(3) == 0:
			c.Groups[r.IntN(2)].When = randomTrigger(r)
		case r.IntN(2) == 0:
			c.Groups[r.IntN(2)].Opening[role] = randomChips(r, role, false)
		default:
			c.Groups[r.IntN(2)].Then[role] = randomChips(r, role, false)
		}
		if _, err := c.Def(); err == nil {
			return c
		}
	}
}

// vsPresets is a custom plan's win rate against each built-in plan.
func vsPresets(f *sim.FactionDef, cp *sim.CustomPlan, n int, base uint64) []float64 {
	var out []float64
	for _, p := range sim.Plans {
		out = append(out, sideCell(f, planSide{custom: cp}, planSide{plan: p.ID}, n, base+uint64(p.ID)*1_000))
	}
	return out
}

func worst(ps []float64) (lo, avg float64) {
	lo = 1
	for _, p := range ps {
		lo = min(lo, p)
		avg += p
	}
	return lo, avg / float64(len(ps))
}

func describe(cp *sim.CustomPlan) string {
	chip := func(c sim.Chips) string {
		var parts []string
		if c.Target != 0 {
			parts = append(parts, c.Target.String())
		}
		if c.Engage != 0 {
			parts = append(parts, c.Engage.String())
		}
		if c.Trigger != 0 {
			parts = append(parts, fmt.Sprintf("trigger %v", sim.TriggerSteps[c.Trigger]))
		}
		if c.Keep != 0 {
			parts = append(parts, fmt.Sprintf("keep %v", sim.KeepSteps[c.Keep]))
		}
		if c.Leash != 0 {
			parts = append(parts, fmt.Sprintf("leash %v", sim.LeashSteps[c.Leash]))
		}
		if c.FallBack != 0 {
			parts = append(parts, fmt.Sprintf("fall back %d%%", sim.FallBackSteps[c.FallBack]))
		}
		if c.StayNear != 0 {
			parts = append(parts, []string{"", "stay behind melee", "guard ranged"}[c.StayNear])
		}
		if c.Kite != 0 {
			parts = append(parts, []string{"", "kite", "no kite"}[c.Kite])
		}
		if c.Edge != 0 {
			parts = append(parts, "edge "+c.Edge.String())
		}
		if len(parts) == 0 {
			return "own"
		}
		return strings.Join(parts, ", ")
	}
	roles := func(cs [3]sim.Chips) string {
		return fmt.Sprintf("melee: %s; ranged: %s; caster: %s", chip(cs[0]), chip(cs[1]), chip(cs[2]))
	}
	s := roles(cp.Roles)
	for g, cg := range cp.Groups {
		s += fmt.Sprintf(" | group %d: %s; then (%v %d) %s", g, roles(cg.Opening), cg.When.Kind, cg.When.Seconds, roles(cg.Then))
	}
	return s
}

// CustomSearch looks for a custom plan that beats every built-in plan, per
// playable faction: random plans, the best hill-climbed one chip at a time,
// the winner checked with n battles per pair. Goal: none beats all nine.
func CustomSearch(n int) string {
	var b strings.Builder
	randoms, quick, climb, steps := max(20, n/8), max(20, n/25), max(40, n/10), max(10, n/50)
	fmt.Fprintf(&b, "Per faction: %d random custom plans at %d battles per built-in plan, the best 3 hill-climbed %d steps at %d, the best checked at %d. A custom plan is ranked by its worst result against the nine built-in plans. Goal: the best still loses to or ties one (under 55%%).\n", randoms, quick, steps, climb, n)
	for fi, f := range playableFactions() {
		r := rand.New(rand.NewPCG(uint64(fi)+1, 77))
		type cand struct {
			cp       *sim.CustomPlan
			lo, mean float64
		}
		var cs []cand
		for i := 0; i < randoms; i++ {
			cp := randomCustom(r)
			lo, mean := worst(vsPresets(f, cp, quick, 600_000+uint64(fi)*100_000))
			cs = append(cs, cand{cp, lo, mean})
		}
		sort.Slice(cs, func(i, j int) bool { return cs[i].lo > cs[j].lo || cs[i].lo == cs[j].lo && cs[i].mean > cs[j].mean })
		best := cs[0]
		for k := 0; k < min(3, len(cs)); k++ {
			c := cs[k]
			c.lo, c.mean = worst(vsPresets(f, c.cp, climb, 700_000+uint64(fi)*100_000))
			for s := 0; s < steps; s++ {
				m := mutate(r, c.cp)
				lo, mean := worst(vsPresets(f, m, climb, 700_000+uint64(fi)*100_000))
				if lo > c.lo || lo == c.lo && mean > c.mean {
					c = cand{m, lo, mean}
				}
			}
			if c.lo > best.lo || c.lo == best.lo && c.mean > best.mean {
				best = c
			}
		}
		final := vsPresets(f, best.cp, n, 800_000+uint64(fi)*100_000)
		lo, mean := worst(final)
		verdict := "ok"
		if lo >= 0.55 {
			verdict = "❌ beats every built-in plan"
		}
		fmt.Fprintf(&b, "\n### %s\n\nBest found: %s\n\nWorst %s, average %s: %s\n\n|", f.Name, describe(best.cp), pct(lo), pct(mean), verdict)
		for _, p := range sim.Plans {
			fmt.Fprintf(&b, " %s |", p.Name)
		}
		b.WriteString("\n|" + strings.Repeat("---|", len(sim.Plans)) + "\n|")
		for _, p := range final {
			fmt.Fprintf(&b, " %s |", pct(p))
		}
		b.WriteString("\n")
	}
	return b.String()
}
