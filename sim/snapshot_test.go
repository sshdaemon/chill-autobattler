package sim

import "testing"

// The snapshot layout is decoded by web/src/sim/snapshot.ts; this pins its size.
func TestSnapshotLayout(t *testing.T) {
	b := testBattle(nil, tu{0, Warrior, p(0, 100), []Skill{WarriorDash, WarriorToughness}}, tu{1, Archer, p(0, -100), nil})
	snap := b.Snapshot()
	if len(snap) != SnapHeader+2*SnapUnit {
		t.Fatalf("snapshot has %d ints, want %d", len(snap), SnapHeader+2*SnapUnit)
	}
	if got := snap[SnapHeader+10]; got != 2 {
		t.Errorf("first unit's skill count = %d, want 2", got)
	}
	if SnapUnit != 13 {
		t.Error("SnapUnit changed: update web/src/sim/snapshot.ts too")
	}
}

func TestSeenCells(t *testing.T) {
	// a blue archer (sight 11) at the bottom, a rock in front of it
	b := testBattle([]Obstacle{{Kind: Rock, Pos: p(0, 20), R: I(1)}}, tu{0, Archer, p(0, 60), nil}, tu{1, Warrior, p(0, -20), nil})
	seen := b.SeenCells(0)
	at := func(x, z int) bool { i, j := cellOf(p(x, z)); return seen[j*GridN+i] }
	if !at(30, 40) || !at(0, 60) {
		t.Error("open ground within sight should be seen")
	}
	if at(0, -20) {
		t.Error("the cell straight behind the rock should be hidden")
	}
	if at(0, -100) {
		t.Error("cells beyond sight range should be hidden")
	}
	b.updateVision()
	if b.SeenBy(0, b.Units[1]) {
		t.Error("the warrior behind the rock is not seen")
	}
	if !b.SeenBy(0, b.Units[0]) {
		t.Error("own units are always seen")
	}
}
