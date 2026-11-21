package sim

// Braced shields: the melee take blows from the front better, but carry
// their shields up and move slower.
func init() {
	registerEdge(EdgeDef{ID: EdgeBraced, Name: "braced shields", Roles: [3]bool{true, false, false},
		Effects: []Effect{{Kind: DamageTaken, Pct: -7, Cond: CondFront}, {Kind: StatBonus, Stat: StatDamage, Pct: -9}}})
}
