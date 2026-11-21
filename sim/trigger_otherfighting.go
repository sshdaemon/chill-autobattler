package sim

// TriggerOtherFighting: a unit of the army's other group is fighting an
// enemy within its reach.
func init() {
	registerTrigger(TriggerOtherFighting, &triggerKind{name: "other group fighting",
		fired: func(b *Battle, team, g int, _ Trigger) bool { return b.groupFighting(team, 1-g, nil) },
	})
}
