package sim

// Drilled (melee): harder blows, at the cost of guard. Only for melee: a
// back line rarely pays a damage-taken cost, so for them it'd be free.
func init() {
	registerEdge(EdgeDef{ID: EdgeDrilled, Name: "drilled", Roles: [3]bool{true, false, false},
		Effects: []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}, {Kind: DamageTaken, Pct: 8}}})
}
