package sim

// Shield wall: the melee hold and guard the back line; ranged and casters
// stay behind them and shoot what is in reach. Counters rushes.
func init() {
	back := Stance{Engage: EngageLine, Target: TargetFocusReach, StayNear: TagMelee}
	registerPlan(PlanDef{ID: PlanShieldWall, Name: "shield wall",
		Melee:  Stance{Engage: EngageHold, Trigger: I(8), Target: TargetProtect},
		Ranged: back, Caster: back})
}
