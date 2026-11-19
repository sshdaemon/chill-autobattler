package sim

// Focus fire: the army advances together and fights the enemy most allies
// already fight; drilled blows hit harder, but the slow advance gives the
// enemy time.
func init() {
	drilled := []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 10}}
	back := Stance{Engage: EngageLine, Target: TargetFocusReach, Effects: drilled}
	registerPlan(PlanDef{ID: PlanFocusFire, Name: "focus fire",
		Melee:  Stance{Engage: EngageLine, Target: TargetFocus, Effects: drilled},
		Ranged: back, Caster: back})
}
