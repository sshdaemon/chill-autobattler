// Package main is the balance lab: it plays batches of seeded battles through
// the real sim, in parallel, and reports win rates as Markdown.
package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"strings"
	"sync"

	"chill/sim"
)

// result of one battle, from Blue's point of view
type result struct {
	result sim.Result
	ticks  int
}

// play runs inputs in parallel; results come back in input order.
func play(ins []sim.Input) []result {
	out := make([]result, len(ins))
	next := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < runtime.NumCPU(); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range next {
				o, err := sim.Run(ins[i])
				if err != nil {
					panic(fmt.Sprintf("input %d: %v", i, err))
				}
				out[i] = result{o.Result, o.Ticks}
			}
		}()
	}
	for i := range ins {
		next <- i
	}
	close(next)
	wg.Wait()
	return out
}

// tally summarises results: Blue wins, Red wins, draws.
type tally struct{ n, blue, red, draw, ticks int }

func count(rs []result) tally {
	t := tally{n: len(rs)}
	for _, r := range rs {
		switch r.result {
		case sim.BlueWin:
			t.blue++
		case sim.RedWin:
			t.red++
		default:
			t.draw++
		}
		t.ticks += r.ticks
	}
	return t
}

func (t tally) win() float64 { return float64(t.blue) / float64(t.n) }

// margin is the ~95% half-width of a win rate over n battles.
func margin(p float64, n int) float64 { return 1.96 * math.Sqrt(p*(1-p)/float64(n)) }

func pct(p float64) string { return fmt.Sprintf("%.1f%%", 100*p) }

// labFaction fields every unit type, so the lab can mix any of them in one
// army (a new unit swapped into another faction's army). Lab-only: never
// playable, so Red keeps drawing from the game's playable factions.
const labFaction sim.Faction = 255

func init() {
	var all []sim.Class
	for d := range sim.Units() {
		all = append(all, d.ID)
	}
	sim.Factions = append(sim.Factions, sim.FactionDef{ID: labFaction, Name: "lab", Units: all})
}

func input(seed uint64) sim.Input {
	return sim.Input{Version: sim.Version, Seed: seed, CombatSeed: seed*7919 + 1, BlueFaction: labFaction}
}

// army builds recruits (level 1): unit i of a class gets that class's ability
// i, so an army of one class covers its abilities evenly.
func army(f sim.Formation, classes ...sim.Class) sim.Army {
	a := sim.Army{Formation: f}
	nth := map[sim.Class]int{}
	for _, c := range classes {
		a.Units = append(a.Units, recruit(c, nth[c]))
		nth[c]++
	}
	return a
}

// recruit is a level-1 unit with its class's (k mod 5)+1-th ability.
func recruit(c sim.Class, k int) sim.UnitSpec {
	return sim.UnitSpec{Class: c, Skills: []sim.Skill{sim.Roster[c].Skills[k%sim.MaxSkills].Skill}}
}

// mix is a faction's army of w melee, a ranged and m caster units.
func mix(f sim.Faction, w, a, m int) []sim.Class {
	t := trio(f)
	var out []sim.Class
	for r, k := range [3]int{w, a, m} {
		for i := 0; i < k; i++ {
			out = append(out, t[r])
		}
	}
	return out
}

// mixName: melee·ranged·caster counts ("2·2·1").
func mixName(w, a, m int) string { return fmt.Sprintf("%d·%d·%d", w, a, m) }

// title is a roster entry's name for tables ("druzhynnyk" -> "Druzhynnyk").
func title(c sim.Class) string {
	n := sim.Roster[c].Name
	return strings.ToUpper(n[:1]) + n[1:]
}

// core is the game's core roster: the units Red's armies are built from. The
// whole-game experiments (compositions, runs) are about these; a unit outside
// the pool is measured on its own with -unit.
func core() []sim.Class { return sim.RedPool() }

func five(c sim.Class) []sim.Class { return []sim.Class{c, c, c, c, c} }

// ---------- experiments ----------

// SideBias: identical armies and presets on both sides.
// One row per playable faction, n battles each; the tally is all of them.
func SideBias(n int) (string, tally) {
	var b strings.Builder
	fmt.Fprintf(&b, "| Faction | Battles | Blue wins | Red wins | Draws |\n|---|---|---|---|---|\n")
	var all tally
	row := func(name string, t tally) {
		fmt.Fprintf(&b, "| %s | %d | %s ± %s | %s | %s |\n", name, t.n, pct(t.win()), pct(margin(t.win(), t.n)), pct(float64(t.red)/float64(t.n)), pct(float64(t.draw)/float64(t.n)))
	}
	for _, fa := range playable() {
		var ins []sim.Input
		mixes := [][3]int{{2, 2, 1}, {5, 0, 0}, {0, 5, 0}}
		for i := 0; i < n; i++ {
			m := mixes[i%len(mixes)]
			f := sim.Presets[(i/len(mixes))%len(sim.Presets)]
			in := input(100_000 + uint64(i))
			in.Blue = army(f, mix(fa, m[0], m[1], m[2])...)
			in.Red = army(f, mix(fa, m[0], m[1], m[2])...)
			ins = append(ins, in)
		}
		t := count(play(ins))
		row(factionTitle(fa), t)
		all.n, all.blue, all.red, all.draw, all.ticks = all.n+t.n, all.blue+t.blue, all.red+t.red, all.draw+t.draw, all.ticks+t.ticks
	}
	row("**All**", all)
	return b.String(), all
}

// headToHead plays 5×a (Blue) against 5×c (Red), Spread on both sides.
func headToHead(a, c sim.Class, n int) tally {
	var ins []sim.Input
	for i := 0; i < n; i++ {
		in := input(200_000 + uint64(i))
		in.Blue = army(sim.Spread, five(a)...)
		in.Red = army(sim.Spread, five(c)...)
		ins = append(ins, in)
	}
	return count(play(ins))
}

// Matchups: 5×A (Blue) vs 5×B (Red), both explicit, Spread, over the core roster.
func Matchups(n int) (string, [][]tally) {
	cs := core()
	all := make([][]tally, len(cs))
	var b strings.Builder
	b.WriteString("Blue win rate, rows = Blue's class (5 units), columns = Red's class (5 units), Spread formation.\n\n| Blue \\ Red |")
	for _, c := range cs {
		fmt.Fprintf(&b, " %s |", title(c))
	}
	b.WriteString("\n|---|" + strings.Repeat("---|", len(cs)) + "\n")
	for i, a := range cs {
		fmt.Fprintf(&b, "| **%s** |", title(a))
		all[i] = make([]tally, len(cs))
		for j, c := range cs {
			t := headToHead(a, c, n)
			all[i][j] = t
			fmt.Fprintf(&b, " %s |", pct(t.win()))
		}
		b.WriteString("\n")
	}
	fmt.Fprintf(&b, "\n%d battles per cell (±%s at worst).\n", n, pct(margin(0.5, n)))
	return b.String(), all
}

type compRow struct {
	name      string
	formation sim.Formation
	t         tally
}

// Compositions: every mix × preset vs the game's seeded Red (no skills).
func Compositions(n int) (string, []compRow) {
	var rows []compRow
	for _, fa := range playable() {
		for w := 0; w <= 5; w++ {
			for a := 0; a <= 5-w; a++ {
				m := 5 - w - a
				for _, f := range sim.Presets {
					var ins []sim.Input
					for i := 0; i < n; i++ {
						in := input(300_000 + uint64(i))
						in.Blue = army(f, mix(fa, w, a, m)...)
						ins = append(ins, in)
					}
					rows = append(rows, compRow{factionTitle(fa) + " " + mixName(w, a, m), f, count(play(ins))})
				}
			}
		}
	}
	// best formation per mix, sorted by win rate
	best := map[string]compRow{}
	avg := map[string]float64{}
	for _, r := range rows {
		if b, ok := best[r.name]; !ok || r.t.win() > b.t.win() {
			best[r.name] = r
		}
		avg[r.name] += r.t.win() / float64(len(sim.Presets))
	}
	var names []string
	for k := range best {
		names = append(names, k)
	}
	sort.Slice(names, func(i, j int) bool {
		if avg[names[i]] != avg[names[j]] {
			return avg[names[i]] > avg[names[j]]
		}
		return names[i] < names[j]
	})
	var b strings.Builder
	b.WriteString("Blue win rate against the game's seeded random Red army (no skills, like match 1), averaged over the 7 formations, and with the best formation. Mixes are melee·ranged·caster units of the faction.\n\n| Mix | Avg win | Best formation | Best win |\n|---|---|---|---|\n")
	for _, k := range names {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", k, pct(avg[k]), best[k].formation, pct(best[k].t.win()))
	}
	fmt.Fprintf(&b, "\n%d battles per mix and formation (±%s at worst).\n", n, pct(margin(0.5, n)))
	return b.String(), rows
}

type skillRow struct {
	faction    sim.Faction
	class      sim.Class
	name       string
	base, with tally
}

// Skills, per playable faction: baseline 2·2·1 (3-2) of recruits vs seeded
// Red; variant: the first unit of the skill's class takes it as its level-1
// ability instead. Same seeds, so the difference is the choice of ability.
func Skills(n int) (string, []skillRow) {
	var rows []skillRow
	var b strings.Builder
	for _, fa := range playable() {
		base := func(i int) sim.Input {
			in := input(400_000 + uint64(i))
			in.Blue = army(sim.F3x2, mix(fa, 2, 2, 1)...)
			return in
		}
		var bins []sim.Input
		for i := 0; i < n; i++ {
			bins = append(bins, base(i))
		}
		baseT := count(play(bins))
		var fr []skillRow
		for _, c := range trio(fa) {
			for _, d := range sim.RosterOf(c).Skills {
				var ins []sim.Input
				for i := 0; i < n; i++ {
					in := base(i)
					u := &in.Blue.Units[slotOf(roleOf(c))]
					u.Skills = []sim.Skill{d.Skill}
					ins = append(ins, in)
				}
				fr = append(fr, skillRow{fa, c, d.Name, baseT, count(play(ins))})
			}
		}
		sort.SliceStable(fr, func(i, j int) bool { return fr[i].with.win() > fr[j].with.win() })
		fmt.Fprintf(&b, "### %s\n\nBaseline: 2·2·1 in 3-2 against the seeded Red army, **%s** win rate (%d battles).\n\n| Skill | Class | Win rate | Δ vs baseline |\n|---|---|---|---|\n", factionTitle(fa), pct(baseT.win()), baseT.n)
		for _, r := range fr {
			fmt.Fprintf(&b, "| %s | %s | %s | %+.1f pts |\n", r.name, title(r.class), pct(r.with.win()), 100*(r.with.win()-r.base.win()))
		}
		b.WriteString("\n")
		rows = append(rows, fr...)
	}
	fmt.Fprintf(&b, "Every unit is a recruit with one ability; in each row one unit of the skill's class takes that skill as its ability instead of its default. Same seeds and armies, so Δ is the choice of ability. Paired seeds make Δ much tighter than ±%s; treat |Δ| < 2 pts as noise.\n", pct(margin(0.5, n)))
	return b.String(), rows
}

// Lengths: battle durations and draws across the composition experiment.
func Lengths(rows []compRow) string {
	var n, draws, ticks int
	for _, r := range rows {
		n += r.t.n
		draws += r.t.draw
		ticks += r.t.ticks
	}
	return fmt.Sprintf("Across %d composition battles: average length **%.1f s**, draws %s (mutual wipes; none hit the %d s limit unless shown here).\n",
		n, float64(ticks)/float64(n)/sim.TicksPerSecond, pct(float64(draws)/float64(n)), sim.MaxTicks/sim.TicksPerSecond)
}
