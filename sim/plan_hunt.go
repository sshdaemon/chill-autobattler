package sim

// Hunt the archers: the melee charge the enemy's ranged and casters first;
// the back line focuses on what is in reach. Counters ranged-heavy armies.
func init() {
	back := Stance{Target: TargetFocusReach}
	registerPlan(PlanDef{ID: PlanHunt, Name: "hunt the archers",
		Melee:  Stance{Engage: EngageCharge, Target: TargetRangedFirst},
		Ranged: back, Caster: back})
}
