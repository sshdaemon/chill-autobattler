package sim

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"
)

// Golden replays: fixed inputs with their recorded outcomes in
// testdata/golden.json. `pnpm test:go` runs them natively and `pnpm test:wasm`
// under WASM, so both builds must reproduce every outcome bit for bit.
//
// After a deliberate rule change: bump sim.Version, then
//
//	go test ./sim -run TestGolden -update
//
// Updating without a version bump fails, because old replays would then be
// verified with different rules.
var update = flag.Bool("update", false, "rewrite testdata/golden.json")

const goldenFile = "testdata/golden.json"

type goldenFileData struct {
	Version  string                   `json:"version"`
	Outcomes map[string]goldenOutcome `json:"outcomes"`
}

type goldenOutcome struct {
	Result    Result     `json:"result"`
	Ticks     int        `json:"ticks"`
	Checksum  string     `json:"checksum"` // uint64 as text: JSON numbers lose precision
	Survivors []Survivor `json:"survivors"`
}

func mix() []UnitSpec {
	return []UnitSpec{{Class: Warrior}, {Class: Archer}, {Class: Mage}, {Class: Archer}, {Class: Warrior}}
}

func allSkills(c Class, perUnit int) []UnitSpec {
	var out []UnitSpec
	for i := 0; i < ArmySize; i++ {
		u := UnitSpec{Class: c}
		for k := 0; k < perUnit; k++ {
			u.Skills = append(u.Skills, SkillsOf[c][(i+k)%MaxSkills].Skill)
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
		return Input{Version: Version, Seed: seed, Blue: blue, RedSkills: redSkills}
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
	add("warrior skills", in(11, Army{Formation: F3x2, Units: allSkills(Warrior, 1)}, 2))
	add("archer skills", in(12, Army{Formation: Spread, Units: allSkills(Archer, 1)}, 2))
	add("mage skills", in(13, Army{Formation: F2x3, Units: allSkills(Mage, 1)}, 2))
	add("veterans 5 stars", in(14, Army{Formation: F3x2, Units: []UnitSpec{
		{Class: Warrior, Skills: []Skill{SkillHealth, SkillSpeed, SkillDamage, SkillDash, SkillExecute}},
		{Class: Archer, Skills: []Skill{SkillRange, SkillDamage, SkillLob, SkillExecute, SkillHealth}},
		{Class: Mage, Skills: []Skill{SkillFrost, SkillDamage, SkillBlink, SkillBlast, SkillHealth}},
	}}, 15))
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
	for _, s := range []uint64{334, 921} { // these end with both armies dying on the same tick
		add("mutual wipe", in(s, Army{Formation: Presets[s%7], Units: mix()}, int(s%6)))
		cases[len(cases)-1].name += " " + itoa(int(s))
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
	got := map[string]goldenOutcome{}
	for _, c := range goldenCases() {
		o, err := Run(c.in)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got[c.name] = toGolden(o)
	}
	var file goldenFileData
	raw, err := os.ReadFile(goldenFile)
	if err == nil {
		err = json.Unmarshal(raw, &file)
	}
	if *update {
		if err == nil && file.Version == Version && !reflect.DeepEqual(file.Outcomes, got) {
			t.Fatalf("outcomes changed but sim.Version is still %q: bump it before updating the golden file", Version)
		}
		out, _ := json.MarshalIndent(goldenFileData{Version: Version, Outcomes: got}, "", "  ")
		if err := os.WriteFile(goldenFile, append(out, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if err != nil {
		t.Fatalf("reading %s: %v (run with -update to create it)", goldenFile, err)
	}
	if file.Version != Version {
		t.Fatalf("golden file is for sim %q, this is %q: run with -update", file.Version, Version)
	}
	for name, want := range file.Outcomes {
		if g, ok := got[name]; !ok {
			t.Errorf("%s: case missing", name)
		} else if !reflect.DeepEqual(g, want) {
			t.Errorf("%s: got %+v\nwant %+v", name, g, want)
		}
	}
	if len(got) != len(file.Outcomes) {
		t.Errorf("%d cases, golden file has %d: run with -update", len(got), len(file.Outcomes))
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
		in := Input{Version: Version, Seed: seed, RedSkills: int(redSkills) % (ArmySize*MaxSkills + 1),
			Blue: Army{Formation: Presets[int(formation)%len(Presets)]}}
		for i := 0; i < ArmySize; i++ {
			c := Class((int(classes) >> i) % int(numClasses))
			in.Blue.Units = append(in.Blue.Units, UnitSpec{Class: c, Skills: []Skill{SkillsOf[c][(int(seed)+i)%MaxSkills].Skill}})
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
