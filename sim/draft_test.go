package sim

import "testing"

func TestDraft(t *testing.T) {
	total := 0
	for _, n := range DraftTurns {
		total += n
	}
	if total != 2*ArmySize {
		t.Fatalf("draft turns pick %d units, want %d", total, 2*ArmySize)
	}
	firsts := [2]int{}
	for seed := uint64(1); seed <= 200; seed++ {
		f, rev := Draft(seed)
		f2, rev2 := Draft(seed)
		if f != f2 || len(rev) != ArmySize || rev[0] != rev2[0] {
			t.Fatal("same seed, different draft")
		}
		seen := map[int]bool{}
		for _, i := range rev {
			seen[i] = true
		}
		if len(seen) != ArmySize {
			t.Fatalf("seed %d: reveal order %v is not a permutation", seed, rev)
		}
		firsts[f]++
	}
	if firsts[0] < 70 || firsts[1] < 70 {
		t.Errorf("coin flip is lopsided: %v", firsts)
	}
}
