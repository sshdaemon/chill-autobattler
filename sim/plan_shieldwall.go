package sim

// Shield wall: the melee hold behind their shields and meet whoever comes;
// ranged and casters stay behind them and shoot what is in reach. Braced
// shields take blows from the front better, but the army gives up the first
// move. Counters rushes and hunters; a focused push breaks it.
func init() {
	back := Stance{Engage: EngageLine, Target: TargetFocusReach, StayNear: TagMelee}
	registerPlan(PlanDef{ID: PlanShieldWall, Name: "shield wall",
		Melee: Stance{Engage: EngageHold, Trigger: I(8), Target: TargetNearest,
			Effects: []Effect{{Kind: DamageTaken, Pct: -8, Cond: CondFront}}},
		Ranged: back, Caster: back})
}
