package sim

import "testing"

func TestFactions(t *testing.T) {
	if err := ValidateFactions(Factions); err != nil {
		t.Fatal(err)
	}
	for name, bad := range map[string][]FactionDef{
		"unknown unit": {{ID: 1, Name: "x", Units: []Class{99}}},
		"unit twice":   {{ID: 1, Name: "x", Units: []Class{Warrior, Warrior}}},
		"no red pool":  {{ID: 1, Name: "x"}},
		"unsorted":     {{ID: 2, Name: "a", Units: []Class{Warrior}}, {ID: 1, Name: "b", Units: []Class{Warrior}}},
	} {
		if ValidateFactions(bad) == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

// Blue's units must come from its faction; Red's come from a seeded faction.
func TestFactionRules(t *testing.T) {
	l1 := []Skill{1}
	in := Input{Version: Version, Seed: 3, BlueFaction: FactionClassic, Blue: Army{Formation: Spread, Units: []UnitSpec{{Class: Warrior, Skills: l1}}}}
	if err := in.Validate(); err != nil {
		t.Fatal(err)
	}
	in.BlueFaction = 0
	if in.Validate() == nil {
		t.Error("no faction accepted")
	}
	old := Factions
	t.Cleanup(func() { Factions = old })
	Factions = []FactionDef{{ID: FactionClassic, Name: "classic", Units: []Class{Archer, Mage}}}
	in.BlueFaction = FactionClassic
	if in.Validate() == nil {
		t.Error("a warrior accepted in a faction without warriors")
	}
	for seed := uint64(1); seed < 30; seed++ {
		for _, u := range RedArmy(seed, 3).Units {
			if u.Class == Warrior {
				t.Fatalf("seed %d: Red fielded a warrior its faction doesn't have", seed)
			}
		}
	}
}
