package convert

import (
	"errors"
	"testing"

	chillv1 "chill/pb/chill/v1"
	"chill/sim"
)

func five() []*chillv1.UnitSpec {
	w, a, m := chillv1.UnitClass_UNIT_CLASS_WARRIOR, chillv1.UnitClass_UNIT_CLASS_ARCHER, chillv1.UnitClass_UNIT_CLASS_MAGE
	return []*chillv1.UnitSpec{{UnitClass: w, Skills: []chillv1.Skill{chillv1.Skill_SKILL_DASH}}, {UnitClass: w}, {UnitClass: a}, {UnitClass: a}, {UnitClass: m}}
}

func TestInput(t *testing.T) {
	in := &chillv1.BattleInput{SimVersion: sim.Version, Seed: 9, CombatSeed: 1 << 63, RedSkillCount: 2,
		Blue: &chillv1.Army{Formation: chillv1.Formation_FORMATION_3_2, Units: five()}}
	got, err := Input(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Blue.Formation != sim.F3x2 || got.Blue.Units[0].Class != sim.Warrior || got.Blue.Units[4].Class != sim.Mage ||
		got.Blue.Units[0].Skills[0] != sim.SkillDash || got.RedSkills != 2 || got.CombatSeed != 1<<63 || len(got.Red.Units) != 0 {
		t.Errorf("converted %+v", got)
	}
	// enum numbering must line up with sim's
	pairs := map[chillv1.Skill]sim.Skill{chillv1.Skill_SKILL_HEALTH: sim.SkillHealth, chillv1.Skill_SKILL_LOB: sim.SkillLob, chillv1.Skill_SKILL_BLAST: sim.SkillBlast}
	for p, s := range pairs {
		if sim.Skill(p) != s {
			t.Errorf("skill %v != %v", p, s)
		}
	}
	if sim.Formation(chillv1.Formation_FORMATION_5_0-chillv1.Formation_FORMATION_CUSTOM) != sim.F5x0 {
		t.Error("formation numbering")
	}

	custom := &chillv1.BattleInput{SimVersion: sim.Version, Blue: &chillv1.Army{Formation: chillv1.Formation_FORMATION_CUSTOM, Units: five()}}
	for i, u := range custom.Blue.Units {
		u.Position = &chillv1.Point{XRaw: int64(sim.I(i*2 - 4)), ZRaw: int64(sim.I(10))}
	}
	if got, err := Input(custom); err != nil || got.Blue.Units[1].Pos.X != sim.I(-2) {
		t.Errorf("custom positions: %v %+v", err, got.Blue.Units[1].Pos)
	}
	custom.Blue.Units[0].Position.ZRaw = 0 // midfield: outside Blue's zone
	if _, err := Input(custom); !errors.Is(err, sim.ErrInvalid) {
		t.Errorf("unit outside zone accepted: %v", err)
	}

	bad := []*chillv1.BattleInput{
		nil,
		{SimVersion: sim.Version},
		{SimVersion: sim.Version, Blue: &chillv1.Army{Units: five()}},                                                                 // no formation
		{SimVersion: sim.Version, Blue: &chillv1.Army{Formation: chillv1.Formation_FORMATION_SPREAD, Units: []*chillv1.UnitSpec{{}}}}, // no class
		{SimVersion: "0.0.0", Blue: &chillv1.Army{Formation: chillv1.Formation_FORMATION_SPREAD, Units: five()}},
		{SimVersion: sim.Version, RedSkillCount: 1 << 31, Blue: &chillv1.Army{Formation: chillv1.Formation_FORMATION_SPREAD, Units: five()}},
	}
	for i, b := range bad {
		if _, err := Input(b); !errors.Is(err, sim.ErrInvalid) {
			t.Errorf("bad input %d accepted: %v", i, err)
		}
	}
}

func TestOutcome(t *testing.T) {
	o := Outcome(sim.Outcome{Result: sim.RedWin, Ticks: 300, Checksum: 77, Survivors: []sim.Survivor{{Index: 6, HP: sim.I(12)}}})
	if o.Result != chillv1.BattleResult_BATTLE_RESULT_RED_WIN || o.Ticks != 300 || o.Checksum != 77 || o.Survivors[0].Index != 6 || o.Survivors[0].HpRaw != int64(sim.I(12)) {
		t.Errorf("outcome %v", o)
	}
	if chillv1.BattleResult(sim.Draw) != chillv1.BattleResult_BATTLE_RESULT_DRAW || chillv1.BattleResult(sim.BlueWin) != chillv1.BattleResult_BATTLE_RESULT_BLUE_WIN {
		t.Error("result numbering")
	}
}

// A unit added to the roster after the proto was written still travels: its
// id is just a number on the wire, and the roster decides whether it exists.
func TestNewUnitIDs(t *testing.T) {
	newUnit := chillv1.UnitClass(len(sim.Roster) + 1) // next roster position, +1 for the proto's UNSPECIFIED
	newSkill := chillv1.Skill(40)
	in := func() *chillv1.BattleInput {
		units := five()
		units[4] = &chillv1.UnitSpec{UnitClass: newUnit, Skills: []chillv1.Skill{newSkill}}
		return &chillv1.BattleInput{SimVersion: sim.Version, Seed: 3, Blue: &chillv1.Army{Formation: chillv1.Formation_FORMATION_SPREAD, Units: units}}
	}
	if _, err := Input(in()); !errors.Is(err, sim.ErrInvalid) { // Input validates against the roster
		t.Errorf("unknown unit accepted: %v", err)
	}

	old := sim.Roster
	t.Cleanup(func() { sim.Roster = old })
	extra := old[sim.Warrior]
	extra.ID, extra.Name, extra.InRedPool = sim.Class(len(old)), "test spearman", false
	extra.Skills[0].Skill = sim.Skill(newSkill)
	extra.Skills[0].Effects = []sim.Effect{{Kind: sim.StatBonus, Stat: sim.StatRange, Pct: 20}}
	sim.Roster = append(append([]sim.UnitDef(nil), old...), extra)
	if err := sim.ValidateRoster(sim.Roster); err != nil {
		t.Fatal(err)
	}
	si, err := Input(in())
	if err != nil {
		t.Fatalf("new unit and skill rejected: %v", err)
	}
	if _, err := sim.Run(si); err != nil {
		t.Errorf("battle with the new unit: %v", err)
	}
}
