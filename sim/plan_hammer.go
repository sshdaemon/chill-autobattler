package sim

// Hammer and anvil: the left group holds as a shield wall (the anvil); the
// right group's melee wait until the anvil is fighting, then swing round the
// enemy's side and hit it there (the hammer). Its ranged shoot what is in
// reach from behind.
func init() {
	back := Stance{Engage: EngageLine, Target: TargetFocusReach, StayNear: TagMelee}
	anvil := Roles{
		Melee: Stance{Engage: EngageHold, Trigger: I(8), Target: TargetNearest,
			Effects: []Effect{{Kind: DamageTaken, Pct: -8, Cond: CondFront}}},
		Ranged: back, Caster: back}
	shoot := Stance{Engage: EngageLine, Target: TargetFocusReach}
	hammer := GroupDef{
		Opening: Roles{
			Melee:  Stance{Engage: EngageHold, Trigger: I(6), Target: TargetNearest, Effects: []Effect{{Kind: StatBonus, Stat: StatSpeed, Pct: 15}}},
			Ranged: shoot, Caster: shoot},
		When: Trigger{Kind: TriggerOtherFighting},
		Then: Roles{Melee: Stance{Engage: EngageFlank, Target: TargetNearest}, Ranged: shoot, Caster: shoot},
	}
	registerPlan(PlanDef{ID: PlanHammer, Name: "hammer and anvil", Groups: [2]GroupDef{{Opening: anvil}, hammer}})
}
