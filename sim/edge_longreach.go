package sim

// Long reach: ranged and casters reach further, with weaker blows.
func init() {
	registerEdge(EdgeDef{ID: EdgeLongReach, Name: "long reach", Roles: [3]bool{false, true, true},
		Effects: []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 20}, {Kind: StatBonus, Stat: StatDamage, Pct: -5}, {Kind: StatBonus, Stat: StatHP, Pct: -5}}})
}
