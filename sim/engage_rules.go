package sim

// The other engage rules (EngageCharge is in engage_charge.go).
func init() {
	// EngageLine: advance like a charge, but keep its place in the formation:
	// wait while it is more than LineSlack further ahead of its allies' centre
	// than it stood at the start (a front-row unit isn't held back by the rows
	// behind it); fights whatever it meets.
	registerEngage(EngageLine, &engageRule{name: "line", idle: func(b *Battle, u *Unit) {
		if c, ok := b.centreOf(u, true); ok { // the centre of the allies that lead, not of those that follow
			ahead := c.Sub(u.Pos).Dot(back(u)) // how far it is ahead of the centre now
			if !u.lineSet {
				u.lineAhead, u.lineSet = ahead, true
			}
			if ahead > u.lineAhead+LineSlack {
				u.moveTarget, u.path = nil, nil
				return
			}
		}
		b.advance(u)
	}})
	// EngageHold: with nothing in sight it marches like a line; once it sees an
	// enemy it waits for it to come within Trigger, for an ally to start
	// fighting, or HoldPatience to run out (two holding armies still meet).
	registerEngage(EngageHold, &engageRule{name: "hold",
		idle: func(b *Battle, u *Unit) { engageRules[EngageLine].idle(b, u) },
		engage: func(b *Battle, u, t *Unit) bool {
			if Within(u.Pos, t.Pos, u.Temper.Trigger) || b.teamFighting(u) || u.held >= HoldPatience {
				return true
			}
			u.held++
			return false
		},
	})
}
