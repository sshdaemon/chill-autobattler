package sim

import "testing"

// Golden results. This test runs natively AND as WASM (pnpm test:go), so if
// either build drifts, it fails. Update the table only on purpose, together
// with sim.Version.
func TestGoldenBattles(t *testing.T) {
	cases := []struct {
		seed          uint64
		winner, ticks int
		checksum      uint64
	}{
		{1, 0, 224, 6834344777054476648},
		{42, 0, 222, 8098190413683205046},
		{0xdeadbeefcafe, 0, 226, 8522140261139492202},
	}
	for _, c := range cases {
		r := Run(c.seed, 4000)
		if r.Winner != c.winner || r.Ticks != c.ticks || r.Checksum != c.checksum {
			t.Errorf("seed %d: got %+v, want winner %d ticks %d checksum %d", c.seed, r, c.winner, c.ticks, c.checksum)
		}
	}
}

func TestFixed(t *testing.T) {
	if got := I(3).Mul(Ratio(1, 2)); got != I(3)/2 {
		t.Errorf("3*0.5 = %v", got)
	}
	if got := Length(I(3), I(4)); got != I(5) {
		t.Errorf("|3,4| = %v", got)
	}
}
