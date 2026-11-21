package sim

import (
	"reflect"
	"testing"
)

func tacticInput(blue, red Plan) Input {
	return Input{Version: Version, BlueFaction: FactionClassic, Seed: 11, CombatSeed: 12,
		Blue: Army{Formation: F3x2, Units: mix(), Plan: blue},
		Red:  Army{Formation: F3x2, Units: mix(), Plan: red}}
}

func TestGroupsOf(t *testing.T) {
	if got := GroupsOf([]Vec{V(I(-3), 0), V(I(1), 0), V(I(2), 0)}); !reflect.DeepEqual(got, []int{0, 1, 1}) {
		t.Errorf("groups %v", got)
	}
	// a unit on the mean joins the left group
	if got := GroupsOf([]Vec{V(I(-2), 0), V(0, 0), V(I(2), 0)}); !reflect.DeepEqual(got, []int{0, 0, 1}) {
		t.Errorf("tie: groups %v", got)
	}
	// the 3-2 flanks are the groups, on both teams
	b, err := New(tacticInput(PlanHammer, PlanReserve))
	if err != nil {
		t.Fatal(err)
	}
	for team := 0; team < 2; team++ {
		count := [2]int{}
		for _, u := range b.Units {
			if u.Team == team {
				count[u.Group]++
				if (u.Pos.X < 0) != (u.Group == 0) {
					t.Errorf("team %d unit %d at x %v in group %d", team, u.Index, u.Pos.X, u.Group)
				}
			}
		}
		if count != [2]int{3, 2} {
			t.Errorf("team %d groups %v, want 3 and 2", team, count)
		}
	}
	// without a tactic everyone is group 0
	b, _ = New(tacticInput(PlanCharge, PlanInstinct))
	for _, u := range b.Units {
		if u.Group != 0 {
			t.Fatalf("unit %d in group %d without a tactic", u.Index, u.Group)
		}
	}
}

func TestDraggedUnitChangesGroup(t *testing.T) {
	in := tacticInput(PlanHammer, PlanInstinct)
	var classes []Class
	for _, s := range in.Blue.Units {
		classes = append(classes, s.Class)
	}
	slots := FormationSlots(F3x2, classes, 0)
	in.Blue.Formation = Custom
	for i := range in.Blue.Units {
		p := slots[i]
		in.Blue.Units[i].Pos = &p
	}
	b, err := New(in)
	if err != nil {
		t.Fatal(err)
	}
	moved := -1
	for _, u := range b.Units[:5] {
		if u.Group == 0 {
			moved = u.Index
		}
	}
	far := V(I(9), slots[moved].Z)
	in.Blue.Units[moved].Pos = &far
	b, _ = New(in)
	if b.Units[moved].Group != 1 {
		t.Errorf("a unit dragged to the far right stayed in group %d", b.Units[moved].Group)
	}
}

// runUntil steps until cond or the end, and reports the tick it held.
func runUntil(b *Battle, cond func() bool) int {
	for b.Result == Ongoing {
		if cond() {
			return b.Tick
		}
		b.Step()
	}
	return -1
}

func TestTriggerAfter(t *testing.T) {
	b, _ := New(tacticInput(PlanHoldCharge, PlanShieldWall))
	tick := runUntil(b, func() bool { return b.switched[0][0] })
	if tick != 10*TicksPerSecond {
		t.Fatalf("switched at tick %d, want %d", tick, 10*TicksPerSecond)
	}
	for _, u := range b.Units[:5] {
		if !u.Dead && u.Temper.Engage != EngageCharge {
			t.Errorf("unit %d after the switch: %v", u.Index, u.Temper.Engage)
		}
	}
}

func TestReserve(t *testing.T) {
	b, _ := New(tacticInput(PlanReserve, PlanCharge))
	// while the left fights and nobody has fallen, the right group keeps holding
	leftFought := false
	runUntil(b, func() bool {
		fallen := false
		for _, u := range b.Units[:5] {
			fallen = fallen || u.Dead
		}
		if fallen {
			return true
		}
		if b.groupFighting(0, 0, nil) {
			leftFought = true
		}
		if b.switched[0][1] {
			t.Fatal("the reserve switched before an ally fell")
		}
		return false
	})
	if !leftFought {
		t.Error("the left group never fought before the first ally fell")
	}
	b.Step()
	if !b.switched[0][1] {
		t.Fatal("the reserve didn't switch after an ally fell")
	}
	if b.switched[0][0] {
		t.Error("the left group, which has no trigger, switched")
	}
	// effects stay: the reserve's damage bonus outlives the switch
	for _, u := range b.Units[:5] {
		if u.Group == 1 && !u.Dead {
			if fresh := newUnitPlan(0, 0, UnitSpec{Class: u.Class, Skills: u.Skills}, Vec{}, PlanInstinct); u.Stats.DmgHi <= fresh.Stats.DmgHi {
				t.Errorf("%s lost its reserve damage bonus", u.Def.Name)
			}
		}
	}
}

func TestHammerFlanks(t *testing.T) {
	// the hammer's melee swing out wider than the same units under Instinct
	widest := func(p Plan) Fx {
		b, _ := New(tacticInput(p, PlanShieldWall))
		var w Fx
		for b.Result == Ongoing && b.Tick < 30*TicksPerSecond {
			b.Step()
			for _, u := range b.Units[:5] {
				if !u.Dead && u.Group == 1 && !u.Stats.Ranged && u.Pos.X.Abs() > w {
					w = u.Pos.X.Abs()
				}
			}
		}
		return w
	}
	if h, i := widest(PlanHammer), widest(PlanInstinct); h <= i {
		t.Errorf("hammer's widest %v, instinct's %v", h, i)
	}
}

func TestBadTacticRejected(t *testing.T) {
	thenEffects := append([]PlanDef(nil), Plans...)
	thenEffects[PlanReserve].Groups[1].Then.Melee.Effects = []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 10}}
	if ValidatePlans(thenEffects, Roster) == nil {
		t.Error("a second stance with effects was accepted")
	}
	badTrigger := append([]PlanDef(nil), Plans...)
	badTrigger[PlanHoldCharge].Groups[0].When = Trigger{Kind: 200}
	if ValidatePlans(badTrigger, Roster) == nil {
		t.Error("an unknown trigger was accepted")
	}
	tooLate := append([]PlanDef(nil), Plans...)
	tooLate[PlanHoldCharge].Groups[0].When = Trigger{Kind: TriggerAfter, Seconds: 0}
	if ValidatePlans(tooLate, Roster) == nil {
		t.Error("an after-0-seconds trigger was accepted")
	}
	holdNoTrigger := append([]PlanDef(nil), Plans...)
	holdNoTrigger[PlanHammer].Groups[1].Then.Ranged = Stance{Engage: EngageHold}
	r := cloneRoster()
	for i := range r {
		r[i].Temper.Trigger = 0
		if r[i].Temper.Engage == EngageHold {
			r[i].Temper.Engage = EngageLine
		}
	}
	if ValidatePlans(holdNoTrigger, r) == nil {
		t.Error("a second stance that holds without a trigger was accepted")
	}
	for _, p := range Plans {
		if want := p.ID >= PlanHammer; p.IsTactic() != want {
			t.Errorf("%s: tactic %v", p.Name, p.IsTactic())
		}
	}
}
