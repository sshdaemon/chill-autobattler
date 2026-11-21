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
	druz, tox, both := VersionFor(Druzhynnyk), VersionFor(Toxotes), VersionFor(Druzhynnyk, Toxotes)
	if VersionFor(Toxotes, Druzhynnyk, Toxotes) != both {
		t.Error("VersionFor depends on order or duplicates")
	}
	old := Roster
	t.Cleanup(func() { Roster = old })
	Roster = append([]UnitDef(nil), old...)
	Roster[Toxotes].Stats.HP++
	if VersionFor(Druzhynnyk) != druz {
		t.Error("changing the archer changed the druzhynnyk's version")
	}
	if VersionFor(Toxotes) == tox || VersionFor(Druzhynnyk, Toxotes) == both {
		t.Error("changing the archer didn't change the versions that use it")
	}
}
