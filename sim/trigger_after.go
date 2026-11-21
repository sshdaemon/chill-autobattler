package sim

import "fmt"

// TriggerAfter: Seconds have passed since the battle began.
func init() {
	registerTrigger(TriggerAfter, &triggerKind{name: "after",
		check: func(t Trigger) error {
			if t.Seconds < 1 || t.Seconds > MaxTicks/TicksPerSecond {
				return fmt.Errorf("after %d s: want 1..%d", t.Seconds, MaxTicks/TicksPerSecond)
			}
			return nil
		},
		fired: func(b *Battle, _, _ int, t Trigger) bool { return b.Tick >= t.Seconds*TicksPerSecond },
	})
}
