package sim

import (
	"errors"
	"strings"
	"testing"
)

func TestRosterIsValid(t *testing.T) {
	if err := ValidateRoster(Roster); err != nil {
		t.Fatal(err)
	}
	if got := RedPool(); len(got) != 3 || got[0] != Warrior || got[1] != Archer || got[2] != Mage {
		t.Errorf("red pool %v, want warrior, archer, mage", got)
	}
}

// One broken entry per rule; each must be reported.
func TestRosterRejectsBadEntries(t *testing.T) {
	cases := map[string]func(d *UnitDef){
		"id":           func(d *UnitDef) { d.ID = 5 },
		"no name":      func(d *UnitDef) { d.Name = "" },
		"stats":        func(d *UnitDef) { d.Stats.HP = 0 },
		"shot speed":   func(d *UnitDef) { d.Stats.ShotSpeed = 0 },
		"projectile":   func(d *UnitDef) { d.Projectile = NoProjectile },
		"tags":         func(d *UnitDef) { d.Tags = 0 },
		"undeclared":   func(d *UnitDef) { d.Tags |= 1 << 6 },
		"bonus":        func(d *UnitDef) { d.Bonus = []TagBonus{{1 << 7, One}} },
		"skill twice":  func(d *UnitDef) { d.Skills[1].Skill = d.Skills[0].Skill },
		"empty skill":  func(d *UnitDef) { d.Skills[4] = SkillDef{} },
		"chance":       func(d *UnitDef) { d.Skills[3].Effects = []Effect{{Kind: Finisher, Pct: 150}} },
		"slow factor":  func(d *UnitDef) { d.Skills[3].Effects = []Effect{{Kind: SlowOnHit, Pct: 50, Ticks: 10, Factor: One}} },
		"unknown kind": func(d *UnitDef) { d.Skills[3].Effects = []Effect{{Kind: 99}} },
		"kind twice":   func(d *UnitDef) { d.Skills[0].Effects = []Effect{{Kind: Finisher, Pct: 10}} },
		"numbering":    func(d *UnitDef) { d.Skills[4].Skill = 7 },
	}
	for name, breakIt := range cases {
		r := cloneRoster()
		breakIt(&r[Archer]) // archer: skill 3 is its finisher (Headshot)
		if err := ValidateRoster(r); err == nil {
			t.Errorf("%s: broken entry accepted", name)
		}
	}
	r := cloneRoster()
	r[Warrior].Skills[3].Effects = []Effect{{Kind: SlowOnHit, Pct: 50, Ticks: 10, Factor: Half}}
	if err := ValidateRoster(r); err == nil || !strings.Contains(err.Error(), "ranged") {
		t.Errorf("slow on hit on a melee unit: %v", err)
	}
	r = cloneRoster()
	for i := range r {
		r[i].InRedPool = false
	}
	if ValidateRoster(r) == nil {
		t.Error("empty red pool accepted")
	}
}

func cloneRoster() []UnitDef {
	r := append([]UnitDef(nil), Roster...)
	for i := range r {
		r[i].Bonus = append([]TagBonus(nil), r[i].Bonus...)
		for k := range r[i].Skills {
			r[i].Skills[k].Effects = append([]Effect(nil), r[i].Skills[k].Effects...)
		}
	}
	return r
}

// withRoster swaps in extra roster entries for one test, appended after the
// real ones (their ids follow, whatever the real roster's size).
func withRoster(t *testing.T, extra ...UnitDef) {
	old := Roster
	Roster = append([]UnitDef(nil), old...)
	for _, d := range extra {
		d.ID = Class(len(Roster))
		Roster = append(Roster, d)
	}
	t.Cleanup(func() { Roster = old })
	if err := ValidateRoster(Roster); err != nil {
		t.Fatal(err)
	}
}

// A unit made only of existing effects with new numbers: no battle code knows
// about it, yet it dashes in, slows what it hits and fights to the end.
var testRanger = UnitDef{
	Name: "frost ranger", // ID set by withRoster
	Stats: Stats{HP: I(70), DmgLo: I(8), DmgHi: I(12), Cooldown: sec(15), Range: I(6), Speed: I(3), Sight: I(10),
		Ranged: true, SwingTicks: sec(6), ShotSpeed: I(16)},
	Projectile: Arrow,
	Tags:       TagRanged | TagCaster,
	Bonus:      []TagBonus{{TagMelee, Ratio(120, 100)}},
	Skills: [MaxSkills]SkillDef{
		{1, "Leap", []Effect{{Kind: GapCloser, Value: I(12), Speed: I(12), Ticks: 100}}},
		{2, "Chill", []Effect{{Kind: SlowOnHit, Pct: 100, Ticks: 40, Factor: Half}}},
		{3, "Hide", []Effect{{Kind: StatBonus, Stat: StatHP, Pct: 50}}},
		{4, "Pierce", []Effect{{Kind: Finisher, Pct: 50}}},
		{5, "Shatter", []Effect{{Kind: SplashOnHit, Pct: 20, Radius: I(3)}}},
	},
}

func TestDataOnlyUnit(t *testing.T) {
	withRoster(t, testRanger)
	ranger := Class(len(Roster) - 1)
	u := newUnit(0, 0, UnitSpec{Class: ranger, Skills: []Skill{3, 2}}, Vec{})
	if u.MaxHP != I(105) || u.kit.slow == nil || u.kit.slow.Factor != Half {
		t.Fatalf("loadout: hp %v slow %+v", u.MaxHP, u.kit.slow)
	}
	if m := MatchupOf(&Roster[Warrior], &Roster[ranger]); m != Ratio(145, 100) {
		t.Errorf("warrior vs a ranged-tagged unit: %v, want 1.45", m)
	}
	spec := UnitSpec{Class: ranger, Skills: []Skill{1, 2, 3, 4, 5}}
	in := Input{Version: Version, Seed: 7, CombatSeed: 8,
		Blue: Army{Formation: Spread, Units: []UnitSpec{spec, spec, spec, spec, spec}},
		Red:  Army{Formation: F3x2, Units: mix()}}
	b, err := New(in)
	if err != nil {
		t.Fatal(err)
	}
	var dashed, slowed bool
	for b.Result == Ongoing {
		b.Step()
		for _, e := range b.TakeEvents() {
			dashed = dashed || e.Kind == EvDash && b.Units[e.Unit].Class == ranger
			if e.Kind == EvFrost {
				slowed = true
				if v := b.Units[e.Unit]; v.slowFactor != Half {
					t.Errorf("slowed with factor %v, want 0.5", v.slowFactor)
				}
			}
		}
	}
	if !slowed || b.Tick > MaxTicks {
		t.Errorf("slowed %v dashed %v ticks %d", slowed, dashed, b.Tick)
	}
	t.Logf("result %v after %d ticks, dashed %v", b.Result, b.Tick, dashed)
}

func TestUnknownUnitRejected(t *testing.T) {
	in := Input{Version: Version, Seed: 1, Blue: Army{Formation: Spread, Units: []UnitSpec{{Class: Class(len(Roster) + 3)}}}}
	if err := in.Validate(); !errors.Is(err, ErrInvalid) {
		t.Errorf("class 7 accepted: %v", err)
	}
}
