package sim

// EngageCharge: with no target in sight, advance on the enemy's side of the
// field (checking where a lost target was last seen first).
func init() {
	registerEngage(EngageCharge, &engageRule{name: "charge", idle: func(b *Battle, u *Unit) { b.advance(u) }})
}
