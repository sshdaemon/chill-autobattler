package sim

// All-out charge: everyone goes for the nearest enemy at once, faster than
// usual and without kiting, but reckless: they take more damage.
func init() {
	reckless := []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 20}, {Kind: DamageTaken, Pct: 3}}
	rush := Stance{Engage: EngageCharge, Target: TargetNearest, Effects: reckless}
	shoot := rush
	shoot.Kite = KiteOff
	registerPlan(PlanDef{ID: PlanCharge, Name: "all-out charge", Melee: rush, Ranged: shoot, Caster: rush})
}
