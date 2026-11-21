package run

import (
	"slices"
	"testing"

	chillv1 "chill/pb/chill/v1"
	"chill/sim"
)

// Veterans below the top level owe a pick; Red is matched to abilities above level 1.
func TestLevels(t *testing.T) {
	w := chillv1.UnitClass(sim.Druzhynnyk + 1) // on the wire: class id + 1
	vets := []*chillv1.UnitSpec{
		{UnitClass: w, Skills: []uint32{1}},       // level 1
		{UnitClass: w, Skills: []uint32{1, 2}},    // level 2
		{UnitClass: w, Skills: []uint32{1, 2, 3}}, // level 3: the top
	}
	if got := Owed(vets); !slices.Equal(got, []uint32{0, 1}) {
		t.Errorf("owed %v, want [0 1] (the level-3 veteran owes nothing)", got)
	}
	if got := SkillTotal(vets); got != 3 {
		t.Errorf("Red's count %d, want 3 (abilities above level 1)", got)
	}
	if sim.MaxLevel != 3 {
		t.Errorf("MaxLevel %d", sim.MaxLevel)
	}
}
