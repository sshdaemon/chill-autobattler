package sim

// Skirmish: the melee hold and fall back when hurt; ranged kite and keep
// their distance. Bleeds the enemy.
func init() {
	registerPlan(PlanDef{ID: PlanSkirmish, Name: "skirmish",
		Melee:  Stance{Engage: EngageHold, Trigger: I(6), FallBack: 40},
		Ranged: Stance{Kite: KiteOn, Keep: I(8)},
		Caster: Stance{Keep: I(8)}})
}
