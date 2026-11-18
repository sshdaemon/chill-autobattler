package sim

import (
	"strings"
	"testing"
)

// Changing one unit's data changes the versions that depend on it and no others.
func TestVersionFingerprint(t *testing.T) {
	if !strings.HasPrefix(Version, RulesVersion+"+") || len(Version) != len(RulesVersion)+9 {
		t.Fatalf("Version %q is not the rules version plus a fingerprint", Version)
	}
	warrior, archer, both := VersionFor(Warrior), VersionFor(Archer), VersionFor(Warrior, Archer)
	if VersionFor(Archer, Warrior, Archer) != both {
		t.Error("VersionFor depends on order or duplicates")
	}
	old := Roster
	t.Cleanup(func() { Roster = old })
	Roster = append([]UnitDef(nil), old...)
	Roster[Archer].Stats.HP++
	if VersionFor(Warrior) != warrior {
		t.Error("changing the archer changed the warrior's version")
	}
	if VersionFor(Archer) == archer || VersionFor(Warrior, Archer) == both {
		t.Error("changing the archer didn't change the versions that use it")
	}
}
