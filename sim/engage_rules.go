package sim

// The other engage rules (EngageCharge is in engage_charge.go).
func init() {
	// EngageLine: advance like a charge, but wait while it is more than LineSlack
	// ahead of its allies' centre; fights whatever it meets.
	registerEngage(EngageLine, &engageRule{name: "line", idle: func(b *Battle, u *Unit) {
		if c, ok := b.allyCentre(u); ok && u.Pos.Sub(c).Dot(back(u)) < -LineSlack {
			u.moveTarget, u.path = nil, nil
			return
		}
		b.advance(u)
	}})
	// EngageHold: stand until an enemy comes within Trigger or an ally is
	// fighting, then fight like a line.
	hold := func(b *Battle, u *Unit) bool { return b.teamFighting(u) }
	registerEngage(EngageHold, &engageRule{name: "hold",
		idle: func(b *Battle, u *Unit) {
			if hold(b, u) {
				engageRules[EngageLine].idle(b, u)
				return
			}
			u.moveTarget, u.path = nil, nil
		},
		engage: func(b *Battle, u, t *Unit) bool {
			return Within(u.Pos, t.Pos, u.Temper.Trigger) || hold(b, u)
		},
	})
}
