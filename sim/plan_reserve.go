package sim

// Reserve: the left group advances in line and takes the first blows; the
// right group holds back, not drawn in by the left's fight, and comes in
// fresh, hitting harder, once the first ally falls.
func init() {
	front := GroupDef{Opening: Roles{
		Melee:  Stance{Engage: EngageLine, Target: TargetNearest},
		Ranged: Stance{Engage: EngageLine, Target: TargetFocusReach},
		Caster: Stance{Engage: EngageLine, Target: TargetFocusReach}}}
	fresh := []Effect{{Kind: StatBonus, Stat: StatDamage, Pct: 5}}
	wait := Stance{Engage: EngageHold, Trigger: I(8), Target: TargetNearest, Effects: fresh}
	shoot := Stance{Engage: EngageHold, Trigger: I(10), Target: TargetFocusReach, Effects: fresh}
	reserve := GroupDef{
		Opening: Roles{Melee: wait, Ranged: shoot, Caster: shoot},
		When:    Trigger{Kind: TriggerAllyFalls},
		Then: Roles{
			Melee:  Stance{Engage: EngageCharge, Target: TargetNearest},
			Ranged: Stance{Engage: EngageCharge, Target: TargetFocusReach},
			Caster: Stance{Engage: EngageCharge, Target: TargetFocusReach}},
	}
	registerPlan(PlanDef{ID: PlanReserve, Name: "reserve", Groups: [2]GroupDef{front, reserve}})
}
