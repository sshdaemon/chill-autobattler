package sim

// Hunt the archers: the melee run past the enemy's front line for its ranged
// and casters, faster than usual but exposed while they do; the back line
// focuses on what is in reach. Counters ranged-heavy armies.
func init() {
	back := Stance{Target: TargetFocusReach}
	registerPlan(PlanDef{ID: PlanHunt, Name: "hunt the archers",
		Melee: Stance{Engage: EngageCharge, Target: TargetRangedFirst,
			Effects: []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 25}, {Kind: DamageTaken, Pct: 5}}},
		Ranged: back, Caster: back})
}
