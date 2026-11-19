package sim

// Skirmish: the melee screen the back line without chasing; ranged kite and keep
// their distance, light and quick on their feet with a longer reach.
// Bleeds a slow enemy.
func init() {
	registerPlan(PlanDef{ID: PlanSkirmish, Name: "skirmish",
		Melee: Stance{Engage: EngageLine, Leash: I(10)},
		Ranged: Stance{Kite: KiteOn, Keep: I(8),
			Effects: []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 30}, {Kind: StatBonus, Stat: StatSpeed, Pct: 15}}},
		Caster: Stance{Keep: I(8), Effects: []Effect{{Kind: StatBonus, Stat: StatRange, Pct: 30}}}})
}
