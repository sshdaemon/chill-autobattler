package sim

import "testing"

// a test-only effect kind: its unit stops kiting (abilities may change temperament)
const testStandFirm EffectKind = 250

func init() {
	registerEffect(testStandFirm, &effectKind{name: "test stand firm", temper: func(t *Temperament, _ Effect) { t.Kite = false }})
}

func TestAbilityChangesTemperament(t *testing.T) {
	steady := Roster[Toxotes]
	steady.Name = "steady archer"
	steady.Skills[0].Effects = []Effect{{Kind: testStandFirm}}
	withRoster(t, steady)
	c := Class(len(Roster) - 1)
	if u := newUnit(0, 0, UnitSpec{Class: c}, Vec{}); !u.Temper.Kite {
		t.Error("without the ability it kites, like an archer")
	}
	if u := newUnit(0, 0, UnitSpec{Class: c, Skills: []Skill{1}}, Vec{}); u.Temper.Kite {
		t.Error("the ability should stop it kiting")
	}
}

func TestTemperamentChecks(t *testing.T) {
	r := cloneRoster()
	r[Druzhynnyk].Temper.Kite = true
	if err := ValidateRoster(r); err == nil {
		t.Error("a melee unit that kites was accepted")
	}
}
