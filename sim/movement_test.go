package sim

import "testing"

// movementInputs are varied battles: every faction, formation and plan.
func movementInputs(n int) []Input {
	var ins []Input
	for s := uint64(1); s <= uint64(n); s++ {
		f := &Factions[int(s)%len(Factions)]
		var units []UnitSpec
		for i := 0; i < ArmySize; i++ {
			units = append(units, UnitSpec{Class: f.Units[(i+int(s/4))%len(f.Units)], Skills: []Skill{Skill(1 + (i+int(s))%MaxSkills)}})
		}
		ins = append(ins, Input{Version: Version, BlueFaction: f.ID, Seed: s * 13, CombatSeed: s*7 + 3,
			Blue: Army{Formation: Presets[s%7], Units: units, Plan: Plan(s % uint64(len(Plans)))}, RedSkills: int(s % 6)})
	}
	return ins
}

// movementStats plays the battles and counts:
//   - reversals: ticks where a walking unit's step points back against its
//     previous step (visible shaking), out of all walking unit-ticks;
//   - stalled: unit-ticks a unit spent wanting to move for 1.5 s on end while
//     getting less than half a metre anywhere.
func movementStats(ins []Input) (reversals, walking, stalled int) {
	const window = 30
	for _, in := range ins {
		b, err := New(in)
		if err != nil {
			panic(err)
		}
		n := len(b.Units)
		last := make([]Vec, n)
		hist := make([][]Vec, n)
		want := make([]int, n) // ticks in a row it has wanted to move
		for b.Result == Ongoing {
			prev := make([]Vec, n)
			for i, u := range b.Units {
				prev[i] = u.Pos
			}
			b.Step()
			for i, u := range b.Units {
				free := !u.Dead && u.dash == nil && u.fear == 0
				step := u.Pos.Sub(prev[i])
				if !free || step.Len() < Ratio(1, 100) {
					last[i] = Vec{}
				} else {
					walking++
					if last[i].Len() > 0 && last[i].Norm().Dot(step.Norm()) < -Half {
						reversals++
					}
					last[i] = step
				}
				hist[i] = append(hist[i], u.Pos)
				if free && u.moveTarget != nil && !u.Kiting && u.swing == nil {
					want[i]++
				} else {
					want[i] = 0
				}
				if t := len(hist[i]) - 1; want[i] > window && Dist(hist[i][t-window], hist[i][t]) < Half {
					stalled++
				}
			}
		}
	}
	return
}

// Units don't shake in place or stall while trying to move (they did: a
// bodyguard flipping between its ward and its advance, a wounded unit flipping
// at the edge of its sight, a goal past the arena edge). The numbers before the
// fix, for these 600 battles: 3.45% reversals and 12,717 stalled unit-ticks.
func TestMovementSmooth(t *testing.T) {
	if testing.Short() {
		t.Skip("plays 600 battles")
	}
	rev, walking, stalled := movementStats(movementInputs(600))
	t.Logf("%d reversals in %d walking unit-ticks (%.2f%%), %d stalled unit-ticks", rev, walking, float64(rev)*100/float64(walking), stalled)
	if rev*100 > walking*2 {
		t.Errorf("units reverse their step %.2f%% of walking ticks, want under 2%%", float64(rev)*100/float64(walking))
	}
	if stalled > 1500 {
		t.Errorf("%d stalled unit-ticks in 600 battles, want at most 1500", stalled)
	}
}

func TestReachable(t *testing.T) {
	m := NewMap([]Obstacle{{Kind: Rock, Pos: V(I(3), I(3)), R: I(1)}})
	if p := m.Reachable(V(I(20), I(-20))); p != V(ClampLim, -ClampLim) {
		t.Errorf("a goal past the arena edge: %v, want the corner inside", p)
	}
	if p := m.Reachable(V(I(3), I(3))); m.insideSolid(p, UnitRadius) {
		t.Errorf("a goal inside a rock stayed in it: %v", p)
	}
	if p := V(I(1), I(1)); m.Reachable(p) != p {
		t.Error("an open goal moved")
	}
}

// A bodyguard (a druzhynnyk) turned back for an archer that itself stays
// behind the melee: each waited for the other and both paced on the spot. In
// this battle Red's druzhynnyk paced from 9 s on, for about 7 s.
func TestGuardsAndFollowersAdvance(t *testing.T) {
	in := movementInputs(469)[468]
	b, err := New(in)
	if err != nil {
		t.Fatal(err)
	}
	u := b.Units[5]
	if u.Def.ID != Druzhynnyk {
		t.Fatalf("test assumes Red's first unit is a druzhynnyk, got %s", u.Def.Name)
	}
	for b.Tick < 9*TicksPerSecond {
		b.Step()
	}
	from := u.Pos
	for b.Tick < 14*TicksPerSecond && b.Result == Ongoing {
		b.Step()
	}
	if moved := Dist(from, u.Pos); moved < I(2) {
		t.Errorf("the druzhynnyk got %v m in 5 s", moved)
	}
}

func TestWoundedWaitsOutOfSight(t *testing.T) {
	in := Input{Version: Version, BlueFaction: FactionRus, Seed: 5, CombatSeed: 6,
		Blue: Army{Formation: F3x2, Units: []UnitSpec{{Class: Huslyar, Skills: []Skill{1}}, {Class: Druzhynnyk, Skills: []Skill{1}}}},
		Red:  Army{Formation: F3x2, Units: mix()}}
	b, _ := New(in)
	u, foe := b.Units[0], b.Units[2]
	if u.Temper.FallBack == 0 {
		t.Fatal("test assumes the huslyar falls back")
	}
	u.HP = u.MaxHP / 4
	foe.Pos = u.Pos.Add(V(0, -(u.Stats.Sight + I(1))))
	if !b.behave(u, foe) || u.moveTarget != nil {
		t.Error("a wounded unit out of its target's sight should wait, not walk back in")
	}
	foe.Pos = u.Pos.Add(V(0, -(u.Stats.Sight - I(1))))
	if !b.behave(u, foe) || u.moveTarget == nil {
		t.Error("a wounded unit in its target's sight should fall back")
	}
}

func TestKeepGoesOutToKeep(t *testing.T) {
	in := Input{Version: Version, BlueFaction: FactionRus, Seed: 5, CombatSeed: 6,
		Blue: Army{Formation: F3x2, Units: []UnitSpec{{Class: Huslyar, Skills: []Skill{1}}}},
		Red:  Army{Formation: F3x2, Units: mix()}}
	b, _ := New(in)
	u, foe := b.Units[0], b.Units[1]
	u.cd = 10
	between := u.Temper.Keep - Half/2 // inside Keep, but not by half a metre
	foe.Pos = u.Pos.Add(V(0, -between))
	if b.behave(u, foe) && u.keeping {
		t.Fatal("started stepping back without coming within Keep less half a metre")
	}
	u.keeping = true // already stepping back: it goes on out to Keep
	if !b.behave(u, foe) || !u.keeping {
		t.Error("a unit stepping back stopped short of Keep")
	}
}
