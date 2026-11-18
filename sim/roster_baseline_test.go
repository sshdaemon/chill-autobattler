package sim

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"
)

// Spawn stats and Red armies recorded before units and skills became roster
// data (testdata/roster_baseline.json). The roster must reproduce them
// exactly: a different rounding or draw order would change replays.
var baseline = flag.Bool("baseline", false, "rewrite testdata/roster_baseline.json")

const baselineFile = "testdata/roster_baseline.json"

type spawnStats struct {
	Class  Class
	Skills []Skill
	Stats  Stats
	MaxHP  Fx
}

type baselineData struct {
	Spawns []spawnStats
	Red    []Army
}

func baselineNow() baselineData {
	var d baselineData
	add := func(c Class, skills ...Skill) {
		u := newUnit(0, 0, UnitSpec{Class: c, Skills: skills}, Vec{})
		d.Spawns = append(d.Spawns, spawnStats{c, skills, u.Stats, u.MaxHP})
	}
	for c := Warrior; c <= Mage; c++ {
		add(c)
		for _, s := range baselineOrder[c] {
			add(c, s)
		}
		add(c, skillIDs(c)...)
	}
	add(Archer, ArcherArcingShot, ArcherSharpArrows)
	add(Archer, ArcherSharpArrows, ArcherArcingShot)
	add(Warrior, WarriorDash, WarriorSwift)
	add(Mage, MageBlast, MageFrost, MagePower)
	for seed := uint64(1); seed <= 200; seed++ {
		d.Red = append(d.Red, RedArmy(seed, int(seed%26)))
	}
	return d
}

// The order the baseline was recorded in: by the global skill ids skills had
// before they were numbered per unit (Toughness, Swift, Strength, Dash,
// Executioner, Long shot, Arcing shot, Frost, Blink, Blast).
var baselineOrder = map[Class][]Skill{
	Warrior: {WarriorToughness, WarriorSwift, WarriorStrength, WarriorDash, WarriorExecutioner},
	Archer:  {ArcherToughness, ArcherSharpArrows, ArcherHeadshot, ArcherLongShot, ArcherArcingShot},
	Mage:    {MageToughness, MagePower, MageFrost, MageBlink, MageBlast},
}

func skillIDs(c Class) []Skill {
	var out []Skill
	for _, d := range Roster[c].Skills {
		out = append(out, d.Skill)
	}
	return out
}

func TestRosterBaseline(t *testing.T) {
	got := baselineNow()
	if *baseline {
		raw, _ := json.MarshalIndent(got, "", " ")
		if err := os.WriteFile(baselineFile, append(raw, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	raw, err := os.ReadFile(baselineFile)
	if err != nil {
		t.Fatal(err)
	}
	var want baselineData
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatal(err)
	}
	for i := range want.Spawns {
		if !reflect.DeepEqual(got.Spawns[i], want.Spawns[i]) {
			t.Errorf("spawn %d: got %+v\nwant %+v", i, got.Spawns[i], want.Spawns[i])
		}
	}
	for i := range want.Red {
		if !reflect.DeepEqual(got.Red[i], want.Red[i]) {
			t.Errorf("red army for seed %d: got %+v\nwant %+v", i+1, got.Red[i], want.Red[i])
		}
	}
	if len(got.Spawns) != len(want.Spawns) || len(got.Red) != len(want.Red) {
		t.Errorf("baseline has %d spawns and %d armies, test made %d and %d", len(want.Spawns), len(want.Red), len(got.Spawns), len(got.Red))
	}
}
