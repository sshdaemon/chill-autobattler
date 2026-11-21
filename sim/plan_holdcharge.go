package sim

// Hold, then charge: the whole army holds, braced, for 10 seconds, letting the
// enemy come and break on it (ranged shoot whatever comes into reach), then
// charges the nearest enemy.
func init() {
	braced := []Effect{{Kind: DamageTaken, Pct: -8, Cond: CondFront}}
	wait := Stance{Engage: EngageHold, Trigger: I(6), Target: TargetNearest, Effects: braced}
	shoot := Stance{Engage: EngageHold, Trigger: I(10), Target: TargetFocusReach}
	charge := Stance{Engage: EngageCharge, Target: TargetNearest}
	g := GroupDef{
		Opening: Roles{Melee: wait, Ranged: shoot, Caster: shoot},
		When:    Trigger{Kind: TriggerAfter, Seconds: 10},
		Then:    Roles{Melee: charge, Ranged: charge, Caster: charge},
	}
	registerPlan(PlanDef{ID: PlanHoldCharge, Name: "hold, then charge", Groups: [2]GroupDef{g, g}})
}
