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
	b.Shots = []*Shot{{Shooter: 1, Victim: 0, Dur: 3}}
	snap = b.Snapshot()
	if len(snap) != SnapHeader+2*SnapUnit+SnapShot || SnapShot != 9 {
		t.Fatal("SnapShot changed: update web/src/sim/snapshot.ts too")
	}
	if shot := snap[SnapHeader+2*SnapUnit:]; shot[7] != 0 || shot[8] != 1 {
		t.Errorf("shot victim, shooter = %d, %d, want 0, 1", shot[7], shot[8])
	}
	// a thrown spear is a lobbed burst shot (web/src/sim/snapshot.ts reads bit 4)
	b.Shots = nil
	b.throw(b.Units[0], b.Units[1], &Effect{Kind: Burst, Value: I(8), Ticks: sec(80), Factor: I(2), Speed: I(14)})
	snap = b.Snapshot()
	if f := snap[SnapHeader+2*SnapUnit]; f != SFLob|SFBurst || SFBurst != 4 {
		t.Errorf("thrown shot flags = %d, want lob|burst (%d)", f, SFLob|SFBurst)
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
