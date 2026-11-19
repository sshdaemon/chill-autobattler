package sim

// Focus fire: the army advances together and fights the enemy most allies
// already fight. A slow, deadly block.
func init() {
	back := Stance{Engage: EngageLine, Target: TargetFocusReach}
	registerPlan(PlanDef{ID: PlanFocusFire, Name: "focus fire",
		Melee:  Stance{Engage: EngageLine, Target: TargetFocus},
		Ranged: back, Caster: back})
}
