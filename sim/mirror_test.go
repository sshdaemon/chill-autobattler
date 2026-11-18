package sim

import "testing"

// Every faction against itself (and every pair of factions) must actually
// fight: armies that only wait for each other end in a draw at the time limit.
func TestFactionMirrorsFight(t *testing.T) {
	army := func(f *FactionDef, seed uint64) []UnitSpec {
		var out []UnitSpec
		for i := 0; i < ArmySize; i++ {
			out = append(out, UnitSpec{Class: f.Units[(i+int(seed))%len(f.Units)], Skills: []Skill{Skill(1 + (i+int(seed))%MaxSkills)}})
		}
		return out
	}
	for i := range Factions {
		for j := range Factions {
			a, c := &Factions[i], &Factions[j]
			stalls := 0
			for seed := uint64(1); seed <= 6; seed++ {
				in := Input{Version: Version, BlueFaction: a.ID, Seed: seed, CombatSeed: seed * 31,
					Blue: Army{Formation: Presets[seed%7], Units: army(a, seed)},
					Red:  Army{Formation: Presets[(seed+3)%7], Units: army(c, seed+1)}}
				o, err := Run(in)
				if err != nil {
					t.Fatal(err)
				}
				if o.Result == Draw && o.Ticks >= MaxTicks {
					stalls++
				}
			}
			if stalls > 0 {
				t.Errorf("%s vs %s: %d of 6 battles ran out of time", a.Name, c.Name, stalls)
			}
		}
	}
}
