package sim

// Reckless: faster, but taking more damage.
func init() {
	registerEdge(EdgeDef{ID: EdgeReckless, Name: "reckless", Roles: [3]bool{true, true, true},
		Effects: []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 25}, {Kind: StatBonus, Stat: StatDamage, Pct: 3}, {Kind: DamageTaken, Pct: 6}}})
}
