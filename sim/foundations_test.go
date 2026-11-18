package sim

import (
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 1.1: the data tables match docs/GAME_CONCEPT.md §4 and §10.1.
func TestDataMatchesConcept(t *testing.T) {
	type row struct {
		hp, lo, hi               int
		cdTenths                 int
		rangeTenths, speedTenths int
		sight                    int
		ranged, kite             bool
	}
	want := [...]row{
		Warrior: {100, 10, 16, 13, 15, 32, 7, false, false},
		Archer:  {55, 9, 14, 17, 90, 36, 11, true, true},
		Mage:    {55, 14, 20, 22, 70, 30, 9, true, false},
	}
	for c, w := range want {
		s := Roster[c].Stats
		if s.HP != I(w.hp) || s.DmgLo != I(w.lo) || s.DmgHi != I(w.hi) || s.Cooldown != sec(w.cdTenths) ||
			s.Range != Ratio(w.rangeTenths, 10) || s.Speed != Ratio(w.speedTenths, 10) || s.Sight != I(w.sight) ||
			s.Ranged != w.ranged || s.Kite != w.kite {
			t.Errorf("%v stats = %+v, want %+v", Class(c), s, w)
		}
	}
	pct := map[[2]Class]int{
		{Warrior, Archer}: 145, {Warrior, Mage}: 95, {Mage, Warrior}: 135,
	}
	for a := Warrior; a <= Mage; a++ {
		for v := Warrior; v <= Mage; v++ {
			p, ok := pct[[2]Class{a, v}]
			if !ok {
				p = 100
			}
			if m := MatchupOf(&Roster[a], &Roster[v]); m != Ratio(p, 100) {
				t.Errorf("matchup %v>%v = %v, want %d%%", a, v, m, p)
			}
		}
	}
	skills := map[Class]string{
		Warrior: "Toughness Swift Strength Dash Executioner",
		Archer:  "Long shot,Sharp arrows,Arcing shot,Headshot,Toughness",
		Mage:    "Frost Power Blink Blast Toughness",
	}
	for c, names := range skills {
		var got []string
		for _, d := range Roster[c].Skills {
			got = append(got, d.Name)
		}
		sep := " "
		if strings.Contains(names, ",") {
			sep = ","
		}
		if strings.Join(got, sep) != names {
			t.Errorf("%v skills = %v, want %v", c, got, names)
		}
	}
	if d, _ := SkillDefOf(Warrior, WarriorExecutioner); d.Main().Pct != 35 {
		t.Errorf("executioner chance %v", d.Main().Pct)
	}
	if d, _ := SkillDefOf(Archer, ArcherHeadshot); d.Main().Pct != 25 {
		t.Errorf("headshot chance %v", d.Main().Pct)
	}
}

// 1.2: fixed-point vectors and the precomputed angle tables.
func TestVectors(t *testing.T) {
	if got := I(3).Mul(Ratio(1, 2)); got != I(3)/2 {
		t.Errorf("3*0.5 = %v", got)
	}
	if got := V(I(3), I(4)).Len(); got != I(5) {
		t.Errorf("|3,4| = %v", got)
	}
	if n := V(I(3), I(4)).Norm(); n != V(Ratio(3, 5), Ratio(4, 5)) {
		t.Errorf("norm = %v", n)
	}
	if !V(0, 0).Norm().IsZero() {
		t.Error("norm of zero should be zero")
	}
	// squared distance across the arena corners must not overflow
	a, b := V(-ArenaHalf, -ArenaHalf), V(ArenaHalf, ArenaHalf)
	if d := Dist(a, b); d.Floor() != 39 { // 28·√2 = 39.6
		t.Errorf("diagonal = %v", d)
	}
	if !Within(a, b, I(40)) || Within(a, b, I(39)) {
		t.Error("Within wrong across the arena")
	}
	if Fx(-1).Floor() != -1 || Ratio(-3, 2).Floor() != -2 || Ratio(3, 2).Ceil() != 2 || I(2).Ceil() != 2 {
		t.Error("floor/ceil")
	}
	if s := Ratio(-3, 2).String(); s != "-1.5000" {
		t.Errorf("String = %s", s)
	}
	// tables agree with math (tests may use floats; sim code may not)
	check := func(name string, got Fx, want float64) {
		if math.Abs(float64(got)/65536-want) > 2.0/65536 {
			t.Errorf("%s = %v, want %f", name, got, want)
		}
	}
	check("turn cos", TurnPerTick.Cos, math.Cos(0.4))
	check("turn sin", TurnPerTick.Sin, math.Sin(0.4))
	check("cos slow turn", CosSlowTurn, math.Cos(1.2))
	check("cos aim", CosAim, math.Cos(0.5))
	for i := 0; i < 64; i++ {
		d := Dir64(i)
		a := 2 * math.Pi * float64(i) / 64
		check("dir64 x "+strconv.Itoa(i), d.X, math.Sin(a))
		check("dir64 z "+strconv.Itoa(i), d.Z, math.Cos(a))
	}
	for i, tt := range append(append([]Turn{}, KiteTurns...), BlinkTurns...) {
		check("turn table "+strconv.Itoa(i), tt.Cos.Mul(tt.Cos)+tt.Sin.Mul(tt.Sin), 1)
	}
	// rotation: facing +Z, turning toward +X
	if got := V(0, One).Rot(Turn{0, One}); got != V(One, 0) {
		t.Errorf("rot 90° = %v", got)
	}
	face := TurnToward(V(0, One), V(One, 0), TurnPerTick)
	if face.X <= 0 || face.Z <= 0 {
		t.Errorf("turning toward +X went the wrong way: %v", face)
	}
	if got := TurnToward(V(0, One), V(Ratio(1, 10), One).Norm(), TurnPerTick); got != V(Ratio(1, 10), One).Norm() {
		t.Errorf("small turn should snap: %v", got)
	}
}

// 1.3: derived RNG streams.
func TestStreams(t *testing.T) {
	a, b := Sub(42, StreamMap), Sub(42, StreamMap)
	for i := 0; i < 100; i++ {
		if a.Uint64() != b.Uint64() {
			t.Fatal("same seed and tag gave different sequences")
		}
	}
	c, d := Sub(42, StreamMap), Sub(42, StreamCombat)
	same := 0
	for i := 0; i < 100; i++ {
		if c.Uint64() == d.Uint64() {
			same++
		}
	}
	if same > 0 {
		t.Errorf("different tags overlapped %d times", same)
	}
	r := NewRNG(7)
	for i := 0; i < 1000; i++ {
		if f := r.Frac(); f < 0 || f >= One {
			t.Fatalf("Frac out of range: %v", f)
		}
		if v := r.Between(I(-9), I(9)); v < I(-9) || v >= I(9) {
			t.Fatalf("Between out of range: %v", v)
		}
	}
}

// 1.4: sim/ stays free of anything that could break determinism.
func TestNoNondeterministicImports(t *testing.T) {
	banned := map[string]bool{"math": true, "math/rand": true, "math/rand/v2": true, "time": true, "sync": true, "sync/atomic": true, "unsafe": true, "os": true}
	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		af, err := parser.ParseFile(token.NewFileSet(), f, src, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range af.Imports {
			if p, _ := strconv.Unquote(imp.Path.Value); banned[p] {
				t.Errorf("%s imports %q, which is not allowed in sim/", f, p)
			}
		}
		for _, bad := range []string{"float32", "float64", "go func", "go\tfunc", "range map"} {
			if strings.Contains(string(src), bad) {
				t.Errorf("%s contains %q, which is not allowed in sim/", f, bad)
			}
		}
	}
}
