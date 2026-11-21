package sim

import (
	"errors"
	"testing"
)

// a test-only effect kind: its unit targets ranged and caster enemies first
const testHunter EffectKind = 249

func init() {
	registerEffect(testHunter, &effectKind{name: "test hunter", temper: func(t *Temperament, _ Effect) { t.Target = TargetRangedFirst }})
}

func TestPlansValid(t *testing.T) {
	if err := ValidatePlans(Plans, Roster); err != nil {
		t.Fatal(err)
	}
	for i, p := range Plans {
		if int(p.ID) != i {
			t.Fatalf("plan %s has id %d at %d", p.Name, p.ID, i)
		}
	}
	if len(Plans) != 9 {
		t.Errorf("%d built-in plans, want 9", len(Plans))
	}
}

func TestBadPlanRejected(t *testing.T) {
	kitingMelee := append([]PlanDef(nil), Plans...)
	kitingMelee[PlanCharge].Melee.Kite = KiteOn
	if ValidatePlans(kitingMelee, Roster) == nil {
		t.Error("a plan that makes melee units kite was accepted")
	}
	holdNoTrigger := append([]PlanDef(nil), Plans...)
	holdNoTrigger[PlanCharge].Ranged = Stance{Engage: EngageHold}
	r := cloneRoster()
	for i := range r { // no unit brings its own trigger
		r[i].Temper.Trigger = 0
		if r[i].Temper.Engage == EngageHold {
			r[i].Temper.Engage = EngageLine
		}
	}
	if ValidatePlans(holdNoTrigger, r) == nil {
		t.Error("a plan that makes units hold without a trigger was accepted")
	}
	rangedOnMelee := append([]PlanDef(nil), Plans...)
	rangedOnMelee[PlanCharge].Melee.Effects = []Effect{{Kind: OverCover}}
	if ValidatePlans(rangedOnMelee, Roster) == nil {
		t.Error("a plan that gives melee units a ranged-only effect was accepted")
	}
	changesInstinct := append([]PlanDef(nil), Plans...)
	changesInstinct[PlanInstinct].Melee.Engage = EngageCharge
	if ValidatePlans(changesInstinct, Roster) == nil {
		t.Error("an Instinct that changes units was accepted")
	}
}

func TestUnknownPlanRejected(t *testing.T) {
	for team := 0; team < 2; team++ {
		in := Input{Version: Version, BlueFaction: FactionClassic, Seed: 1, CombatSeed: 2, Blue: Army{Formation: F3x2, Units: mix()}}
		if team == 0 {
			in.Blue.Plan = 200
		} else {
			in.Red = Army{Formation: F3x2, Units: mix(), Plan: Plan(len(Plans))}
		}
		if _, err := Run(in); !errors.Is(err, ErrInvalid) {
			t.Errorf("team %d: unknown plan gave %v, want ErrInvalid", team, err)
		}
	}
}

func TestPlanChangesTemperament(t *testing.T) {
	// the skoutatos holds by its own temperament; All-out charge makes it charge
	if Roster[Skoutatos].Temper.Engage != EngageHold {
		t.Fatal("test assumes the skoutatos holds")
	}
	u := newUnitPlan(0, 0, UnitSpec{Class: Skoutatos}, Vec{}, PlanCharge)
	if u.Temper.Engage != EngageCharge || u.Temper.Target != TargetNearest {
		t.Errorf("under All-out charge the skoutatos has %+v", u.Temper)
	}
	// a field the plan leaves alone keeps the unit's own value
	own := Roster[Huslyar].Temper
	u = newUnitPlan(0, 0, UnitSpec{Class: Huslyar}, Vec{}, PlanCharge)
	if u.Temper.StayNear != own.StayNear || u.Temper.Keep != own.Keep {
		t.Errorf("All-out charge changed the huslyar's own fields: %+v", u.Temper)
	}
	// Instinct changes nothing
	for _, d := range Roster {
		if got := newUnitPlan(0, 0, UnitSpec{Class: d.ID}, Vec{}, PlanInstinct).Temper; got != d.Temper {
			t.Errorf("%s under Instinct: %+v, want %+v", d.Name, got, d.Temper)
		}
	}
	// ranged roles
	if u := newUnitPlan(0, 0, UnitSpec{Class: Archer}, Vec{}, PlanCharge); u.Temper.Kite {
		t.Error("All-out charge should stop archers kiting")
	}
	if u := newUnitPlan(0, 0, UnitSpec{Class: Mage}, Vec{}, PlanShieldWall); u.Temper.StayNear != TagMelee {
		t.Error("Shield wall should keep casters near the melee")
	}
}

func TestPlanEffects(t *testing.T) {
	for _, d := range Roster {
		own := newUnitPlan(0, 0, UnitSpec{Class: d.ID}, Vec{}, PlanInstinct)
		rush := newUnitPlan(0, 0, UnitSpec{Class: d.ID}, Vec{}, PlanCharge)
		if rush.Stats.Speed <= own.Stats.Speed {
			t.Errorf("%s under All-out charge: speed %v, under Instinct %v", d.Name, rush.Stats.Speed, own.Stats.Speed)
		}
	}
	// effects come after the unit's own: a skill's stat bonus and the plan's both apply
	c := Archer
	base := newUnitPlan(0, 0, UnitSpec{Class: c}, Vec{}, PlanInstinct).Stats.Range
	if got := newUnitPlan(0, 0, UnitSpec{Class: c}, Vec{}, PlanSkirmish).Stats.Range; got <= base {
		t.Errorf("Skirmish should lengthen an archer's range: %v vs %v", got, base)
	}
}

func TestAbilityOnTopOfPlan(t *testing.T) {
	hunter := Roster[Warrior]
	hunter.Name = "hunting warrior"
	hunter.Skills[0].Effects = []Effect{{Kind: testHunter}}
	withRoster(t, hunter)
	c := Class(len(Roster) - 1)
	if u := newUnitPlan(0, 0, UnitSpec{Class: c}, Vec{}, PlanFocusFire); u.Temper.Target != TargetFocus {
		t.Errorf("without the ability Focus fire sets focus, got %v", u.Temper.Target)
	}
	if u := newUnitPlan(0, 0, UnitSpec{Class: c, Skills: []Skill{1}}, Vec{}, PlanFocusFire); u.Temper.Target != TargetRangedFirst {
		t.Errorf("the ability should win over the plan, got %v", u.Temper.Target)
	}
}

func TestRedPlan(t *testing.T) {
	seen := map[Plan]bool{}
	for s := uint64(1); s <= 200; s++ {
		p := RedPlanFor(s)
		if RedPlanFor(s) != p {
			t.Fatal("same combat seed, different plan")
		}
		seen[p] = true
	}
	if len(seen) != len(Plans) {
		t.Errorf("200 seeds drew %d of %d plans", len(seen), len(Plans))
	}
	// a generated Red gets the drawn plan; a given Red keeps its own
	in := Input{Version: Version, BlueFaction: FactionClassic, Seed: 7, CombatSeed: 9, Blue: Army{Formation: F3x2, Units: mix(), Plan: PlanSkirmish}}
	b, err := New(in)
	if err != nil {
		t.Fatal(err)
	}
	if b.Plans != [2]Plan{PlanSkirmish, RedPlanFor(9)} {
		t.Errorf("plans %v", b.Plans)
	}
	// the map seed alone doesn't decide it
	other := in
	other.CombatSeed = 10
	differs := false
	for cs := uint64(10); cs < 40 && !differs; cs++ {
		other.CombatSeed = cs
		b2, _ := New(other)
		differs = b2.Plans[1] != b.Plans[1]
	}
	if !differs {
		t.Error("Red's plan doesn't change with the combat seed")
	}
	in.Red = Army{Formation: F2x3, Units: mix(), Plan: PlanHunt}
	if b, _ := New(in); b.Plans[1] != PlanHunt {
		t.Errorf("a given Red should keep its plan, got %v", b.Plans[1])
	}
}

// Any two plans, on any two factions, meet and fight before the time limit.
func TestPlansFight(t *testing.T) {
	army := func(f *FactionDef, seed uint64) []UnitSpec {
		var out []UnitSpec
		for i := 0; i < ArmySize; i++ {
			out = append(out, UnitSpec{Class: f.Units[(i+int(seed))%len(f.Units)], Skills: []Skill{Skill(1 + (i+int(seed))%MaxSkills)}})
		}
		return out
	}
	for _, pa := range Plans {
		for _, pb := range Plans {
			stalls := 0
			for i := range Factions {
				for j := range Factions {
					a, c := &Factions[i], &Factions[j]
					for seed := uint64(1); seed <= 2; seed++ {
						in := Input{Version: Version, BlueFaction: a.ID, Seed: seed + uint64(i*4+j), CombatSeed: seed * 31,
							Blue: Army{Formation: Presets[seed%7], Units: army(a, seed), Plan: pa.ID},
							Red:  Army{Formation: Presets[(seed+3)%7], Units: army(c, seed+1), Plan: pb.ID}}
						o, err := Run(in)
						if err != nil {
							t.Fatal(err)
						}
						if o.Result == Draw && o.Ticks >= MaxTicks {
							stalls++
						}
					}
				}
			}
			if stalls > 0 {
				t.Errorf("%s vs %s: %d battles ran out of time", pa.Name, pb.Name, stalls)
			}
		}
	}
}
