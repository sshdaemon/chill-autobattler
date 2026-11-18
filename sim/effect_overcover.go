package sim

// OverCover: the unit's shots arc over rocks and trees, needing no clear line
// of fire (Arcing shot). The team still has to see the target.
func init() {
	registerEffect(OverCover, &effectKind{name: "overCover", check: needsRanged, overCover: true})
}
