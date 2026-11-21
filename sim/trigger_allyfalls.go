package sim

// TriggerAllyFalls: a unit of the group's army has fallen.
func init() {
	registerTrigger(TriggerAllyFalls, &triggerKind{name: "ally falls",
		fired: func(b *Battle, team, _ int, _ Trigger) bool {
			for _, u := range b.Units {
				if u.Team == team && u.Dead {
					return true
				}
			}
			return false
		},
	})
}
