package sim

import "testing"

func TestFactions(t *testing.T) {
	if err := ValidateFactions(Factions); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string][]FactionDef{
		"unknown unit":  {{ID: 1, Name: "x", Units: []Class{99}, Playable: true}},
		"unit twice":    {{ID: 1, Name: "x", Units: []Class{Druzhynnyk, Druzhynnyk}, Playable: true}},
		"no red pool":   {{ID: 1, Name: "x", Playable: true}},
		"none playable": {{ID: 1, Name: "x", Units: []Class{Druzhynnyk}}},
		"unsorted":      {{ID: 2, Name: "a", Units: []Class{Druzhynnyk}, Playable: true}, {ID: 1, Name: "b", Units: []Class{Druzhynnyk}}},
	} {
		if ValidateFactions(bad) == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// Blue's units must come from its faction; Red's come from a seeded faction.
func TestFactionRules(t *testing.T) {
	l1 := []Skill{1}
	in := Input{Version: Version, Seed: 3, BlueFaction: FactionRus, Blue: Army{Formation: Spread, Units: []UnitSpec{{Class: Druzhynnyk, Skills: l1}}}}
	if err := in.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, f := range []Faction{0, 1} { // none, and the retired Classic
		in.BlueFaction = f
		if in.Validate() == nil {
			t.Errorf("faction %d accepted", f)
		}
	}
	old := Factions
	t.Cleanup(func() { Factions = old })
	Factions = []FactionDef{{ID: FactionRus, Name: "rus", Units: []Class{Toxotes, Volva}, Playable: true}}
	in.BlueFaction = FactionRus
	if in.Validate() == nil {
		t.Error("a druzhynnyk accepted in a faction without one")
	}
	for seed := uint64(1); seed < 30; seed++ {
		for _, u := range RedArmy(seed, 3).Units {
			if u.Class == Druzhynnyk {
				t.Fatalf("seed %d: Red fielded a druzhynnyk its faction doesn't have", seed)
			}
		}
	}
}

// Classic (faction 1) is retired: never offered, never Red's.
func TestClassicRetired(t *testing.T) {
	if FactionOf(1) != nil {
		t.Fatal("faction 1 (Classic) is registered")
	}
	for seed := uint64(0); seed < 2000; seed++ {
		if f := RedFactionFor(seed); f == 1 || FactionOf(f) == nil || !FactionOf(f).Playable {
			t.Fatalf("seed %d: Red's faction %d", seed, f)
		}
	}
	for _, c := range retiredClasses {
		if RosterOf(c) != nil || FactionWith(c) != 0 {
			t.Errorf("retired class %d is still in the roster or a faction", c)
		}
	}
}
