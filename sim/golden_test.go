package sim

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

// Golden replays: fixed inputs with their recorded outcomes in
// testdata/golden/, one file per case. `pnpm test:go` runs them natively and `pnpm test:wasm`
// under WASM, so both builds must reproduce every outcome bit for bit.
//
// Each case records the version of what it depends on: RulesVersion plus the
// fingerprint of the units it uses (VersionFor), so changing one unit's data
// only touches its own cases. After a deliberate change (a unit's numbers, or
// rules code with RulesVersion bumped), run
//
//	go test ./sim -run TestGolden -update
//
// Changing a recorded outcome while its version stays the same fails, because
// old replays would then be verified with different rules. Adding cases is fine.
var update = flag.Bool("update", false, "rewrite the golden files (testdata/golden/)")

// One file per case (testdata/golden/<case>.json), so changes to different
// units touch different files.
const goldenDir = "testdata/golden"

type goldenFileData struct {
	Version string        `json:"version"`
	Case    string        `json:"case"`
	Outcome goldenOutcome `json:"outcome"`
}

// caseVersion is the version a case depends on: the rules and the units that
// can appear in it (Blue's, Red's, or Red's pool when Red is generated).
func caseVersion(in Input) string {
	var cs []Class
	for _, a := range []Army{in.Blue, in.Red} {
		for _, u := range a.Units {
			cs = append(cs, u.Class)
		}
	}
	if len(in.Red.Units) == 0 {
		cs = append(cs, FactionOf(RedFactionFor(in.Seed)).Pool()...)
	}
	return VersionFor(cs...)
}

func goldenPath(name string) string {
	return goldenDir + "/" + strings.ReplaceAll(name, " ", "-") + ".json"
}

// readGolden returns the recorded outcomes by case name and the versions they were recorded with.
func readGolden() (map[string]goldenOutcome, map[string]string, error) {
	outs, vers := map[string]goldenOutcome{}, map[string]string{}
	entries, err := os.ReadDir(goldenDir)
	if err != nil {
		return nil, nil, err
	}
	for _, e := range entries {
		raw, err := os.ReadFile(goldenDir + "/" + e.Name())
		if err != nil {
			return nil, nil, err
		}
		var f goldenFileData
		if err := json.Unmarshal(raw, &f); err != nil {
			return nil, nil, fmt.Errorf("%s: %v", e.Name(), err)
		}
		outs[f.Case], vers[f.Case] = f.Outcome, f.Version
	}
	return outs, vers, nil
}

type goldenOutcome struct {
	Result    Result     `json:"result"`
	Ticks     int        `json:"ticks"`
	Checksum  string     `json:"checksum"` // uint64 as text: JSON numbers lose precision
	Survivors []Survivor `json:"survivors"`
}

func mix() []UnitSpec {
	l1 := []Skill{1} // recruits: level 1, their first ability
	return []UnitSpec{{Class: Warrior, Skills: l1}, {Class: Archer, Skills: l1}, {Class: Mage, Skills: l1}, {Class: Archer, Skills: l1}, {Class: Warrior, Skills: l1}}
}

func allSkills(c Class, perUnit int) []UnitSpec {
	var out []UnitSpec
	for i := 0; i < ArmySize; i++ {
		u := UnitSpec{Class: c}
		for k := 0; k < perUnit; k++ {
			u.Skills = append(u.Skills, Roster[c].Skills[(i+k)%MaxSkills].Skill)
		}
		out = append(out, u)
	}
	return out
}

func goldenCases() []struct {
	name string
	in   Input
} {
	in := func(seed uint64, blue Army, redSkills int) Input {
		return Input{Version: Version, BlueFaction: FactionClassic, Seed: seed, CombatSeed: seed*7919 + 1, Blue: blue, RedSkills: redSkills}
	}
	cases := []struct {
		name string
		in   Input
	}{}
	add := func(name string, i Input) {
		cases = append(cases, struct {
			name string
			in   Input
		}{name, i})
	}
	for k, f := range Presets {
		add("formation "+f.String(), in(uint64(k+1), Army{Formation: f, Units: mix()}, 0))
	}
	// one case per roster entry, each in its own file, so new units add files rather than edit shared ones
	for _, d := range Roster {
		f := Presets[int(d.ID)%len(Presets)]
		if int(d.ID) < 3 {
			f = []Formation{F3x2, Spread, F2x3}[d.ID] // the original three's cases, as recorded
		}
		c := in(11+uint64(d.ID), Army{Formation: f, Units: allSkills(d.ID, 1)}, 2)
		c.BlueFaction = FactionWith(d.ID)
		add(d.Name+" skills", c)
	}
	add("veterans 3 stars", in(14, Army{Formation: F3x2, Units: []UnitSpec{
		{Class: Warrior, Skills: []Skill{WarriorSwift, WarriorDash, WarriorExecutioner}},
		{Class: Archer, Skills: []Skill{ArcherLongShot, ArcherArcingShot, ArcherHeadshot}},
		{Class: Mage, Skills: []Skill{MageFrost, MageBlink, MageBlast}},
	}}, 10))
	// every skill and the combinations whose order matters (data-driven-units)
	with := func(c Class, skills ...Skill) []UnitSpec {
		out := make([]UnitSpec, ArmySize)
		for i := range out {
			out[i] = UnitSpec{Class: c, Skills: skills}
		}
		return out
	}
	add("mages frost and blast", in(41, Army{Formation: F2x3, Units: with(Mage, MageBlast, MageFrost)}, 5))
	add("mages blink and power", in(42, Army{Formation: Spread, Units: with(Mage, MageBlink, MagePower, MageToughness)}, 5))
	add("archers sharp arrows and lob", in(43, Army{Formation: F3x2, Units: with(Archer, ArcherArcingShot, ArcherSharpArrows)}, 5))
	add("archers long shot and headshot", in(44, Army{Formation: F1x4, Units: with(Archer, ArcherLongShot, ArcherHeadshot, ArcherToughness)}, 5))
	add("warriors swift and dash", in(45, Army{Formation: F4x1, Units: with(Warrior, WarriorDash, WarriorSwift)}, 5))
	add("warriors executioner", in(46, Army{Formation: F5x0, Units: with(Warrior, WarriorExecutioner, WarriorStrength, WarriorToughness)}, 5))
	for _, n := range []int{0, 3, 10} {
		add("red with "+itoa(n)+" skills", in(uint64(50+n), Army{Formation: Spread, Units: mix()}, n))
	}
	custom := mix()
	for i := range custom {
		p := V(I(i*4-8), I(8+i%2*3))
		custom[i].Pos = &p
	}
	add("custom positions", in(15, Army{Formation: Custom, Units: custom}, 0))
	pvp := in(16, Army{Formation: F4x1, Units: mix()}, 0)
	pvp.Red = Army{Formation: F1x4, Units: allSkills(Mage, 2)}
	add("explicit red army", pvp)
	add("wet map", in(121, Army{Formation: Spread, Units: mix()}, 0))
	eased := in(31, Army{Formation: F3x2, Units: mix()}, 1)
	eased.RedHealthPct = 85 // a run's first match
	add("red at 85% health", eased)
	for _, s := range []uint64{121, 1664} { // these end with both armies dying on the same tick
		add("mutual wipe", in(s, Army{Formation: Presets[s%7], Units: mix()}, int(s%6)))
		cases[len(cases)-1].name += " " + itoa(int(s))
	}
	for _, p := range Plans { // one per plan, against an explicit Red on a different plan
		c := in(200+uint64(p.ID), Army{Formation: F3x2, Units: mix(), Plan: p.ID}, 0)
		c.Red = Army{Formation: F2x3, Units: mix(), Plan: Plans[(int(p.ID)+1)%len(Plans)].ID}
		add("plan "+p.Name, c)
	}
	for _, s := range []uint64{90, 657} { // stalled until the tick limit before line of fire used the sight test
		add("ranged standoff", in(s, Army{Formation: Presets[s%7], Units: mix()}, int(s%6)))
		cases[len(cases)-1].name += " " + itoa(int(s))
	}
	return cases
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for ; n > 0; n /= 10 {
		s = string(rune('0'+n%10)) + s
	}
	return s
}

// The combat seed changes only the fight: same map, same start, different battle.
func TestCombatSeedChangesOnlyTheFight(t *testing.T) {
	base := Input{Version: Version, BlueFaction: FactionClassic, Seed: 5, CombatSeed: 1, RedSkills: 2, Blue: Army{Formation: F3x2, Units: mix()}}
	other := base
	other.CombatSeed = 2
	a, _ := New(base)
	b, _ := New(other)
	if !reflect.DeepEqual(a.Map.Obstacles, b.Map.Obstacles) {
		t.Fatal("combat seed changed the map")
	}
	for i := range a.Units {
		if a.Units[i].Pos != b.Units[i].Pos || a.Units[i].Class != b.Units[i].Class || !reflect.DeepEqual(a.Units[i].Skills, b.Units[i].Skills) {
			t.Fatalf("combat seed changed unit %d's start", i)
		}
	}
	oa, _ := Run(base)
	ob, _ := Run(other)
	if oa.Checksum == ob.Checksum {
		t.Error("different combat seeds should play out differently")
	}
}

func toGolden(o Outcome) goldenOutcome {
	g := goldenOutcome{Result: o.Result, Ticks: o.Ticks, Survivors: o.Survivors}
	g.Checksum = utoa(o.Checksum)
	return g
}

func utoa(v uint64) string {
	if v == 0 {
		return "0"
	}
	s := ""
	for ; v > 0; v /= 10 {
		s = string(rune('0'+v%10)) + s
	}
	return s
}

func TestGoldenReplays(t *testing.T) {
	got, want := map[string]goldenOutcome{}, map[string]string{} // outcomes and versions now
	for _, c := range goldenCases() {
		o, err := Run(c.in)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got[c.name], want[c.name] = toGolden(o), caseVersion(c.in)
	}
	recorded, versions, err := readGolden()
	if *update {
		for name := range recorded { // new cases may be added; recorded ones must not change
			if g, ok := got[name]; ok && versions[name] == want[name] && !reflect.DeepEqual(g, recorded[name]) {
				t.Fatalf("%s: outcome changed but its version is still %q: bump sim.RulesVersion before updating the golden files", name, want[name])
			}
		}
		os.MkdirAll(goldenDir, 0o755)
		for name, o := range got {
			if r, ok := recorded[name]; ok && versions[name] == want[name] && reflect.DeepEqual(r, o) {
				continue // unchanged: leave the file alone
			}
			out, _ := json.MarshalIndent(goldenFileData{Version: want[name], Case: name, Outcome: o}, "", "  ")
			if err := os.WriteFile(goldenPath(name), append(out, '\n'), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		for name := range recorded {
			if _, ok := got[name]; !ok {
				os.Remove(goldenPath(name))
			}
		}
		return
	}
	if err != nil {
		t.Fatalf("reading %s: %v (run with -update to create it)", goldenDir, err)
	}
	for name, rec := range recorded {
		if g, ok := got[name]; !ok {
			t.Errorf("%s: case missing (delete its file, or run with -update)", name)
		} else if versions[name] != want[name] {
			t.Errorf("%s: recorded for %q, its units and rules are now %q: run with -update", name, versions[name], want[name])
		} else if !reflect.DeepEqual(g, rec) {
			t.Errorf("%s: got %+v\nwant %+v", name, g, rec)
		}
	}
	for name := range got {
		if _, ok := recorded[name]; !ok {
			t.Errorf("%s: not recorded: run with -update", name)
		}
	}
	// the set must keep covering the spec's interesting endings
	draws := 0
	for _, o := range got {
		if o.Result == Draw {
			draws++
		}
	}
	if draws < 2 {
		t.Errorf("golden set should include mutual wipes, has %d draws", draws)
	}
}

// Same input, played twice in one process, gives the identical outcome:
// catches hidden state, map iteration or anything not reset between battles.
func FuzzDeterminism(f *testing.F) {
	f.Add(uint64(1), uint8(0), uint8(0), uint8(0))
	f.Add(uint64(90), uint8(3), uint8(4), uint8(5))
	f.Fuzz(func(t *testing.T, seed uint64, formation, classes, redSkills uint8) {
		in := Input{Version: Version, BlueFaction: FactionClassic, Seed: seed, CombatSeed: ^seed, RedSkills: int(redSkills) % (ArmySize*(MaxLevel-1) + 1),
			Blue: Army{Formation: Presets[int(formation)%len(Presets)]}}
		for i := 0; i < ArmySize; i++ {
			c := Class((int(classes) >> i) % 3)
			in.Blue.Units = append(in.Blue.Units, UnitSpec{Class: c, Skills: []Skill{Roster[c].Skills[(int(seed)+i)%MaxSkills].Skill}})
		}
		a, err := Run(in)
		if err != nil {
			t.Fatal(err)
		}
		b, _ := Run(in)
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("same input, different outcomes:\n%+v\n%+v", a, b)
		}
		if a.Result == Ongoing || a.Ticks > MaxTicks {
			t.Fatalf("battle did not end properly: %+v", a)
		}
	})
}

func BenchmarkBattle(b *testing.B) {
	cases := goldenCases()
	for i := 0; i < b.N; i++ {
		if _, err := Run(cases[i%len(cases)].in); err != nil {
			b.Fatal(err)
		}
	}
}

// A new unit outside Red's pool is content, not a rule change: every recorded
// battle still plays out the same, under the same version.
func TestGoldenWithExtraUnit(t *testing.T) {
	withRoster(t, testRanger)
	recorded, _, err := readGolden()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range goldenCases() {
		want, ok := recorded[c.name]
		if !ok {
			continue // the extra unit's own case
		}
		o, err := Run(c.in)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if g := toGolden(o); !reflect.DeepEqual(g, want) {
			t.Errorf("%s changed with an extra roster entry: %+v", c.name, g)
		}
	}
}
