package sim

// KiteShoot: the unit keeps shooting while it backs away from melee
// (Steppe shooting). Needs a kiting temperament to matter.
func init() {
	registerEffect(KiteShoot, &effectKind{name: "kiteShoot", check: needsRanged, kiteShoot: true})
}
