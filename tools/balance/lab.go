// Package main is the balance lab: it plays batches of seeded battles through
// the real sim, in parallel, and reports win rates as Markdown.
package main

import (
	"fmt"
	"math"
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
	for w := 0; w < workers; w++ {
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
// army (a new unit swapped into the classic army). Lab-only: never playable,
// so Red keeps drawing from the game's playable factions.
const labFaction sim.Faction = 255

func init() {
	var all []sim.Class
	for _, d := range sim.Roster {
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

func mix(w, a, m int) []sim.Class {
	var out []sim.Class
	for i := 0; i < w; i++ {
		out = append(out, sim.Warrior)
	}
	for i := 0; i < a; i++ {
		out = append(out, sim.Archer)
	}
	for i := 0; i < m; i++ {
		out = append(out, sim.Mage)
	}
	return out
}

func mixName(w, a, m int) string { return fmt.Sprintf("%dW %dA %dM", w, a, m) }

// title is a roster entry's name for tables ("warrior" -> "Warrior").
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
func SideBias(n int) (string, tally) {
	var ins []sim.Input
	mixes := [][3]int{{2, 2, 1}, {5, 0, 0}, {0, 5, 0}}
	for i := 0; i < n; i++ {
		m := mixes[i%len(mixes)]
		f := sim.Presets[(i/len(mixes))%len(sim.Presets)]
		in := input(100_000 + uint64(i))
		in.Blue = army(f, mix(m[0], m[1], m[2])...)
		in.Red = army(f, mix(m[0], m[1], m[2])...)
		ins = append(ins, in)
	}
	t := count(play(ins))
	var b strings.Builder
	fmt.Fprintf(&b, "| Battles | Blue wins | Red wins | Draws |\n|---|---|---|---|\n")
	fmt.Fprintf(&b, "| %d | %s ± %s | %s | %s |\n", t.n, pct(t.win()), pct(margin(t.win(), t.n)), pct(float64(t.red)/float64(t.n)), pct(float64(t.draw)/float64(t.n)))
	return b.String(), t
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
	for w := 0; w <= 5; w++ {
		for a := 0; a <= 5-w; a++ {
			m := 5 - w - a
			for _, f := range sim.Presets {
				var ins []sim.Input
				for i := 0; i < n; i++ {
					in := input(300_000 + uint64(i))
					in.Blue = army(f, mix(w, a, m)...)
					ins = append(ins, in)
				}
				rows = append(rows, compRow{mixName(w, a, m), f, count(play(ins))})
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
	b.WriteString("Blue win rate against the game's seeded random Red army (no skills, like match 1), averaged over the 7 formations, and with the best formation.\n\n| Mix | Avg win | Best formation | Best win |\n|---|---|---|---|\n")
	for _, k := range names {
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", k, pct(avg[k]), best[k].formation, pct(best[k].t.win()))
	}
	fmt.Fprintf(&b, "\n%d battles per mix and formation (±%s at worst).\n", n, pct(margin(0.5, n)))
	return b.String(), rows
}

type skillRow struct {
	class      sim.Class
	name       string
	base, with tally
}

// Skills: baseline 2W 2A 1M (3-2) of recruits vs seeded Red; variant: the
// first unit of the skill's class takes it as its level-1 ability instead.
// Same seeds, so the difference is the choice of ability.
func Skills(n int) (string, []skillRow) {
	base := func(i int) sim.Input {
		in := input(400_000 + uint64(i))
		in.Blue = army(sim.F3x2, mix(2, 2, 1)...)
		return in
	}
	var bins []sim.Input
	for i := 0; i < n; i++ {
		bins = append(bins, base(i))
	}
	baseT := count(play(bins))
	first := map[sim.Class]int{sim.Warrior: 0, sim.Archer: 2, sim.Mage: 4}
	var rows []skillRow
	for _, c := range core() {
		for _, d := range sim.Roster[c].Skills {
			var ins []sim.Input
			for i := 0; i < n; i++ {
				in := base(i)
				u := &in.Blue.Units[first[c]]
				u.Skills = []sim.Skill{d.Skill}
				ins = append(ins, in)
			}
			rows = append(rows, skillRow{c, d.Name, baseT, count(play(ins))})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].with.win() > rows[j].with.win() })
	var b strings.Builder
	fmt.Fprintf(&b, "Baseline: 2W 2A 1M in 3-2 against the seeded Red army, **%s** win rate (%d battles). Every unit is a recruit with one ability; in each row one unit of the skill's class takes that skill as its ability instead of its default. Same seeds and armies, so Δ is the choice of ability.\n\n", pct(baseT.win()), baseT.n)
	b.WriteString("| Skill | Class | Win rate | Δ vs baseline |\n|---|---|---|---|\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %+.1f pts |\n", r.name, title(r.class), pct(r.with.win()), 100*(r.with.win()-r.base.win()))
	}
	fmt.Fprintf(&b, "\nPaired seeds make Δ much tighter than ±%s; treat |Δ| < 2 pts as noise.\n", pct(margin(0.5, n)))
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
