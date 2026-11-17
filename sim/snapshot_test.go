package sim

import "testing"

// The snapshot layout is decoded by web/src/sim/snapshot.ts; this pins its size.
func TestSnapshotLayout(t *testing.T) {
	b := testBattle(nil, tu{0, Warrior, p(0, 100), []Skill{SkillDash, SkillHealth}}, tu{1, Archer, p(0, -100), nil})
	snap := b.Snapshot()
	if len(snap) != SnapHeader+2*SnapUnit {
		t.Fatalf("snapshot has %d ints, want %d", len(snap), SnapHeader+2*SnapUnit)
	}
	if got := snap[SnapHeader+SnapUnit-1]; got != 2 {
		t.Errorf("first unit's skill count = %d, want 2", got)
	}
	if SnapUnit != 11 {
		t.Error("SnapUnit changed: update web/src/sim/snapshot.ts too")
	}
}
