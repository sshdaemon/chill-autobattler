package sim

// Instinct: every unit fights by its own temperament. The default.
func init() { registerPlan(PlanDef{ID: PlanInstinct, Name: "instinct"}) }
