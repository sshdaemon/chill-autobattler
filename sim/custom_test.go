package sim

import (
	"errors"
	"reflect"
	"testing"
)

// customPlansForTests: an army-wide custom plan, and custom tactics with each
// kind of trigger (golden cases and the no-stall test use them).
func customPlansForTests() []struct {
	name string
	plan *CustomPlan
} {
	wide := &CustomPlan{Roles: [3]Chips{
		{Target: TargetNearest, Engage: EngageHold, Trigger: 2, Edge: EdgeBraced},
		{Kite: KiteOn, Keep: 3, Edge: EdgeLongReach},
		{Target: TargetFocusReach, Edge: EdgeLongReach}}}
	tactic := func(when CustomTrigger) *CustomPlan {
		hold := Chips{Engage: EngageHold, Trigger: 1, Target: TargetNearest}
		charge := Chips{Engage: EngageCharge, Target: TargetNearest}
		return &CustomPlan{Groups: []CustomGroup{
			{Opening: [3]Chips{{Engage: EngageLine, Target: TargetFocus, Edge: EdgeReckless}, {}, {}}},
			{Opening: [3]Chips{hold, hold, hold}, When: when, Then: [3]Chips{{Engage: EngageFlank, Target: TargetNearest}, charge, charge}},
		}}
	}
	return []struct {
		name string
		plan *CustomPlan
	}{
		{"army-wide", wide},
		{"tactic after 10 s", tactic(CustomTrigger{Kind: TriggerAfter, Seconds: 10})},
		{"tactic ally falls", tactic(CustomTrigger{Kind: TriggerAllyFalls})},
		{"tactic other group fighting", tactic(CustomTrigger{Kind: TriggerOtherFighting})},
	}
}

func TestEdgesValid(t *testing.T) {
	if err := ValidateEdges(Edges, Roster); err != nil {
		t.Fatal(err)
	}
	if len(Edges) != 5 {
		t.Errorf("%d edges, want 5", len(Edges))
	}
	for _, e := range Edges {
		gain, cost := false, false
		for _, x := range e.Effects {
			up := x.Pct > 0
			if x.Kind == DamageTaken {
				up = !up
			}
			gain, cost = gain || up, cost || !up
		}
		if !gain || !cost {
			t.Errorf("edge %s needs an advantage and a cost: %+v", e.Name, e.Effects)
		}
	}
}

// Chips give exactly the plan a hand-written PlanDef would.
func TestCustomPlanDef(t *testing.T) {
	cp := customPlansForTests()[0].plan
	got, err := cp.Def()
	if err != nil {
		t.Fatal(err)
	}
	want := PlanDef{ID: PlanCustom, Name: "custom",
		Melee:  Stance{Target: TargetNearest, Engage: EngageHold, Trigger: I(8), Effects: EdgeOf(EdgeBraced).Effects},
		Ranged: Stance{Kite: KiteOn, Keep: I(8), Effects: EdgeOf(EdgeLongReach).Effects},
		Caster: Stance{Target: TargetFocusReach, Effects: EdgeOf(EdgeLongReach).Effects}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("chips gave\n%+v\nwant\n%+v", got, want)
	}
	tac, err := customPlansForTests()[1].plan.Def()
	if err != nil {
		t.Fatal(err)
	}
	if !tac.IsTactic() || tac.Groups[1].When != (Trigger{Kind: TriggerAfter, Seconds: 10}) || tac.Groups[1].Then.Melee.Engage != EngageFlank {
		t.Errorf("custom tactic: %+v", tac.Groups)
	}
}

// An empty custom plan plays exactly like Instinct.
func TestEmptyCustomIsInstinct(t *testing.T) {
	base := Input{Version: Version, BlueFaction: FactionRus, Seed: 41, CombatSeed: 42,
		Blue: Army{Formation: F3x2, Units: []UnitSpec{{Class: Druzhynnyk, Skills: []Skill{1}}, {Class: Strilets, Skills: []Skill{2}}, {Class: Huslyar, Skills: []Skill{3}}}}}
	a, err := Run(base)
	if err != nil {
		t.Fatal(err)
	}
	custom := base
	custom.Blue.Custom = &CustomPlan{}
	b, err := Run(custom)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Errorf("empty custom plan: %+v, Instinct: %+v", b, a)
	}
	bt, _ := New(custom)
	if bt.Plans[0] != PlanCustom {
		t.Errorf("Plans[0] = %v, want custom", bt.Plans[0])
	}
}

func TestBadCustomPlansRejected(t *testing.T) {
	bad := map[string]*CustomPlan{
		"unknown target":          {Roles: [3]Chips{{Target: 99}}},
		"keep step off the chips": {Roles: [3]Chips{{}, {Keep: 4}}},
		"melee kite":              {Roles: [3]Chips{{Kite: KiteOn}}},
		"ranged edge on melee":    {Roles: [3]Chips{{Edge: EdgeLongReach}}},
		"unknown edge":            {Roles: [3]Chips{{Edge: 99}}},
		"edge on a second stance": {Groups: []CustomGroup{{}, {When: CustomTrigger{Kind: TriggerAllyFalls}, Then: [3]Chips{{Edge: EdgeReckless}}}}},
		"after 7 s":               {Groups: []CustomGroup{{}, {When: CustomTrigger{Kind: TriggerAfter, Seconds: 7}}}},
		"seconds on ally falls":   {Groups: []CustomGroup{{}, {When: CustomTrigger{Kind: TriggerAllyFalls, Seconds: 5}}}},
		"one group":               {Groups: []CustomGroup{{}}},
	}
	for name, cp := range bad {
		in := Input{Version: Version, BlueFaction: FactionClassic, Seed: 1, CombatSeed: 2, Blue: Army{Formation: F3x2, Units: mix(), Custom: cp}}
		if _, err := Run(in); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: got %v, want ErrInvalid", name, err)
		}
	}
	withID := Input{Version: Version, BlueFaction: FactionClassic, Seed: 1, CombatSeed: 2, Blue: Army{Formation: F3x2, Units: mix(), Plan: PlanCharge, Custom: &CustomPlan{}}}
	if _, err := Run(withID); !errors.Is(err, ErrInvalid) {
		t.Errorf("a custom plan with a plan id: %v", err)
	}
	// holding without a trigger distance: archers have none of their own
	if _, err := (&CustomPlan{Roles: [3]Chips{{}, {Engage: EngageHold}}}).Def(); err == nil {
		t.Error("ranged told to hold with no trigger distance was accepted")
	}
	if _, err := (&CustomPlan{Roles: [3]Chips{{}, {Engage: EngageHold, Trigger: 3}}}).Def(); err != nil {
		t.Errorf("ranged holding with a far trigger was rejected: %v", err)
	}
}

// Every built-in plan has a valid nearest-chips version; Skirmish's ranged
// come out as kite on, keep far, long reach.
func TestChipsOfPresets(t *testing.T) {
	for i := range Plans {
		c := ChipsOf(&Plans[i])
		if _, err := c.Def(); err != nil {
			t.Errorf("%s as chips: %v", Plans[i].Name, err)
		}
	}
	sk := ChipsOf(PlanOf(PlanSkirmish)).Roles[1]
	if sk.Kite != KiteOn || sk.Keep != 3 || sk.Edge != EdgeLongReach {
		t.Errorf("Skirmish ranged as chips: %+v", sk)
	}
	if c := ChipsOf(PlanOf(PlanInstinct)); !reflect.DeepEqual(c, CustomPlan{}) {
		t.Errorf("Instinct as chips: %+v", c)
	}
	if w := ChipsOf(PlanOf(PlanShieldWall)).Roles[0]; w.Edge != EdgeBraced || w.Trigger != 2 || w.Engage != EngageHold {
		t.Errorf("Shield wall melee as chips: %+v", w)
	}
}
