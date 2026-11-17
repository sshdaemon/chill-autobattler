package main

import "testing"

// The parallel runner must not share state: the same experiment twice gives
// the same numbers.
func TestReproducible(t *testing.T) {
	a, ta := SideBias(60)
	b, tb := SideBias(60)
	if a != b || ta != tb {
		t.Fatalf("side bias differs between runs:\n%s\n%s", a, b)
	}
	m1, _ := Matchups(20)
	m2, _ := Matchups(20)
	if m1 != m2 {
		t.Fatal("matchups differ between runs")
	}
}
