package sim

import "testing"

// Mirror-image armies on a mirrored map march, path and push as exact mirror
// images until the first contact (a dash, a swing, a blow): nothing in
// movement favours a side. From contact on, units react to each other within a
// tick and which side acts first is a coin flip (Battle.Step), fair over many
// battles but not mirror-exact in one. Red's start stagger is copied from
// Blue's so both start in step.
func TestMirrorImageBattles(t *testing.T) {
	armies := map[Faction][]Class{
		FactionVarangians: {Huskarl, Bowman, Volva, Bowman, Huskarl},
		FactionRomans:     {Skoutatos, Toxotes, Siphonarios, Toxotes, Skoutatos},
		FactionRus:        {Druzhynnyk, Strilets, Huslyar, Strilets, Druzhynnyk},
	}
	for f, cs := range armies {
		for seed := uint64(1); seed <= 25; seed++ {
			var us []UnitSpec
			for i, c := range cs {
				us = append(us, UnitSpec{Class: c, Skills: []Skill{Roster[c].Skills[(i+int(seed))%MaxSkills].Skill}})
			}
			a := Army{Formation: Presets[seed%uint64(len(Presets))], Units: us, Plan: Plan(seed % uint64(len(Plans)))}
			b, err := New(Input{Version: Version, BlueFaction: f, Seed: seed * 101, CombatSeed: seed, Blue: a, Red: a})
			if err != nil {
				t.Fatal(err)
			}
			n := len(cs)
			for i := 0; i < n; i++ {
				b.Units[i+n].retarget, b.Units[i+n].cd = b.Units[i].retarget, b.Units[i].cd
			}
			for b.Result == Ongoing && b.Tick < MaxTicks {
				b.Step()
				contact := false
				for _, u := range b.Units {
					contact = contact || u.HP != u.MaxHP || u.swing != nil || u.dash != nil
				}
				if contact {
					break
				}
				for i := 0; i < n; i++ {
					u, r := b.Units[i], b.Units[i+n]
					if u.Pos != flipZ(r.Pos) || u.Face != flipZ(r.Face) {
						t.Fatalf("%v seed %d tick %d: %s %d at %v facing %v, its mirror image at %v facing %v",
							f, seed, b.Tick, u.Def.Name, i, u.Pos, u.Face, flipZ(r.Pos), flipZ(r.Face))
					}
				}
			}
		}
	}
}

func TestToCellMirrors(t *testing.T) {
	for raw := -ArenaHalf + 1; raw < ArenaHalf; raw += Ratio(1, 8) + 3 {
		if ToCell(-raw) != GridN-1-ToCell(raw) {
			t.Fatalf("ToCell(%v) = %d, ToCell(%v) = %d: not mirror images", raw, ToCell(raw), -raw, ToCell(-raw))
		}
	}
	for k := -27; k <= 27; k++ { // cell edges
		if v := Half * Fx(k); v != 0 && ToCell(-v) != GridN-1-ToCell(v) {
			t.Fatalf("cell edge %v: %d vs %d", v, ToCell(v), ToCell(-v))
		}
	}
}

func TestMulIsOdd(t *testing.T) {
	for _, a := range []Fx{1, 3, Half + 1, Ratio(17, 10), I(5) + 7} {
		for _, b := range []Fx{Ratio(1, 3), Ratio(7, 10) + 1, I(2) + 5} {
			if (-a).Mul(b) != -(a.Mul(b)) {
				t.Errorf("(-%v)*%v = %v, want %v", a, b, (-a).Mul(b), -(a.Mul(b)))
			}
		}
	}
}
