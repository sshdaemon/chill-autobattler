package sim

import "testing"

// withTemper gives every unit of a test battle the given temperament.
func withTemper(b *Battle, team int, t Temperament) {
	for _, u := range b.Units {
		if u.Team == team {
			u.Temper = t
		}
	}
}

func TestTargetRules(t *testing.T) {
	p := func(x, z int) Vec { return V(I(x), I(z)) }
	// Blue warrior at the origin; Red: a warrior close, an archer further, a wounded mage close
	b := testBattle(nil,
		tu{0, Warrior, p(0, 0), nil},
		tu{0, Archer, p(-6, 4), nil}, // a ranged ally, for protect
		tu{1, Warrior, p(0, -2), nil},
		tu{1, Archer, p(4, -4), nil},
		tu{1, Mage, p(3, 0), nil},
		tu{1, Warrior, p(-6, 6), nil}, // next to Blue's archer
	)
	b.Units[4].HP = b.Units[4].MaxHP / 4
	b.Units[1].Target = 3 // Blue's archer already fights Red's archer
	b.updateVision()
	u := b.Units[0]
	for rule, want := range map[TargetRule]int{
		TargetNearest: 2, TargetWounded: 4, TargetRangedFirst: 4, TargetStrongest: 4, TargetFocus: 3, TargetProtect: 5,
	} {
		if got := targetRules[rule].pick(b, u); got != want {
			t.Errorf("%v picked unit %d, want %d", rule, got, want)
		}
	}
}

func TestHoldWaitsForTrigger(t *testing.T) {
	b := testBattle(nil, tu{0, Warrior, V(0, I(8)), nil}, tu{1, Warrior, V(0, I(-1)), nil})
	withTemper(b, 0, Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(5)})
	withTemper(b, 1, Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(5)})
	start := b.Units[0].Pos
	b.steps(40)
	if b.Units[0].Pos != start {
		t.Fatalf("a holding unit moved with its enemy 9 m away: %v", b.Units[0].Pos)
	}
	withTemper(b, 1, Temperament{Target: TargetNearest, Engage: EngageCharge})
	b.steps(60)
	if Dist(b.Units[0].Pos, b.Units[1].Pos) > I(3) {
		t.Errorf("the holding unit should fight once the enemy comes within its trigger")
	}
}

func TestFallBackAndKeep(t *testing.T) {
	// a wounded unit with allies retreats toward them instead of fighting
	b := testBattle(nil, tu{0, Warrior, V(0, 0), nil}, tu{0, Archer, V(0, I(10)), nil}, tu{1, Warrior, V(0, I(-2)), nil})
	withTemper(b, 0, Temperament{Target: TargetNearest, Engage: EngageCharge, FallBack: 50})
	b.Units[0].HP = b.Units[0].MaxHP / 3
	z := b.Units[0].Pos.Z
	b.steps(20)
	if b.Units[0].Pos.Z <= z {
		t.Errorf("a unit below its fall-back health should retreat toward its allies (z %v → %v)", z, b.Units[0].Pos.Z)
	}
	// a unit that keeps its distance steps back between shots
	b = testBattle(nil, tu{0, Mage, V(0, 0), nil}, tu{1, Warrior, V(0, I(-3)), nil})
	withTemper(b, 0, Temperament{Target: TargetNearest, Engage: EngageCharge, Keep: I(7)})
	withTemper(b, 1, Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(1)}) // stays put
	b.steps(80)
	if d := Dist(b.Units[0].Pos, b.Units[1].Pos); d < I(5) {
		t.Errorf("keep 7 m: still %v m from its target", d)
	}
}

func TestLineAndStayNear(t *testing.T) {
	// a line unit with nothing in sight doesn't run ahead of its allies
	b := testBattle(nil, tu{0, Warrior, V(0, 0), nil}, tu{0, Warrior, V(I(2), I(10)), nil}, tu{1, Warrior, V(0, I(-25)), nil})
	withTemper(b, 0, Temperament{Target: TargetNearest, Engage: EngageLine})
	b.Units[1].Temper = Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(1)} // the ally stays back
	b.steps(30)
	if b.Units[0].Target >= 0 || b.Units[0].Pos.Z < -Half {
		t.Errorf("a line unit ran ahead of its ally: z %v (target %d)", b.Units[0].Pos.Z, b.Units[0].Target)
	}
	b.Units[1].Pos = V(I(2), I(-1)) // the ally moves up: the line follows
	b.steps(20)
	if b.Units[0].Pos.Z > -Half {
		t.Errorf("with its ally level, a line unit should advance: z %v", b.Units[0].Pos.Z)
	}
	// a support unit stays behind its front line
	b = testBattle(nil, tu{0, Mage, V(I(8), I(10)), nil}, tu{0, Warrior, V(0, I(5)), nil}, tu{1, Warrior, V(0, I(-25)), nil})
	withTemper(b, 0, Temperament{Target: TargetNearest, Engage: EngageCharge, StayNear: TagMelee})
	b.Units[1].Temper = Temperament{Target: TargetNearest, Engage: EngageHold, Trigger: I(1)}
	b.steps(60)
	if d := Dist(b.Units[0].Pos, b.Units[1].Pos.Add(V(0, I(3)))); d > I(3) {
		t.Errorf("stay near melee: %v m from its spot behind the front line", d)
	}
}
