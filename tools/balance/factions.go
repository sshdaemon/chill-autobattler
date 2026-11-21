package main

import (
	"fmt"
	"strings"

	"chill/sim"
)

// The lab measures each playable faction on its own three units, one per
// role. Roles as the sim's plans see them: 0 melee, 1 ranged, 2 caster.
const (
	melee = iota
	ranged
	caster
)

var roleNames = [3]string{"melee", "ranged", "caster"}

func roleOf(c sim.Class) int {
	d := sim.RosterOf(c)
	switch {
	case d.Tags&sim.TagCaster != 0:
		return caster
	case d.Stats.Ranged:
		return ranged
	}
	return melee
}

// playable lists the playable factions, by id.
func playable() []sim.Faction {
	var out []sim.Faction
	for _, f := range sim.Factions {
		if f.Playable {
			out = append(out, f.ID)
		}
	}
	return out
}

// trio is a faction's units by role. The lab's mixes need exactly one unit of
// each role; a faction that breaks this stops the tool.
func trio(f sim.Faction) [3]sim.Class {
	def := sim.FactionOf(f)
	var out [3]sim.Class
	var seen [3]bool
	for _, c := range def.Units {
		r := roleOf(c)
		if seen[r] {
			panic(fmt.Sprintf("faction %s has two %s units: the balance lab needs one per role", def.Name, roleNames[r]))
		}
		out[r], seen[r] = c, true
	}
	for r, ok := range seen {
		if !ok {
			panic(fmt.Sprintf("faction %s has no %s unit: the balance lab needs one per role", def.Name, roleNames[r]))
		}
	}
	return out
}

// factionTitle is a faction's name for tables ("rus" -> "Rus").
func factionTitle(f sim.Faction) string {
	n := sim.FactionOf(f).Name
	return strings.ToUpper(n[:1]) + n[1:]
}

// slotOf is where the standard 2·2·1 army keeps the first unit of a role.
func slotOf(role int) int { return [3]int{0, 2, 4}[role] }
