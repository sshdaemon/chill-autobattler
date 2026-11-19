package sim

// All-out charge: everyone goes for the nearest enemy at once, no kiting.
func init() {
	rush := Stance{Engage: EngageCharge, Target: TargetNearest}
	shoot := rush
	shoot.Kite = KiteOff
	registerPlan(PlanDef{ID: PlanCharge, Name: "all-out charge", Melee: rush, Ranged: shoot, Caster: rush})
}
