package sim

// Light feet: ranged move faster, lightly armoured.
func init() {
	registerEdge(EdgeDef{ID: EdgeLightFeet, Name: "light feet", Roles: [3]bool{false, true, false},
		Effects: []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 20}, {Kind: StatBonus, Stat: StatHP, Pct: -4}}})
}
