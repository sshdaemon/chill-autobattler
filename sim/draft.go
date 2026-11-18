package sim

// The draft before a battle (presentation of who picks when; the battle only
// sees the armies). Both sides pick five units in turns of DraftTurns picks,
// alternating, the first side decided by a coin flip from the map seed. Red's
// picks are its generated army (RedArmy), revealed in a seeded order, so the
// client and the server agree on the whole draft without a round trip.

// DraftTurns is how many picks each turn has: first side 1, second 2, first 2,
// second 2, first 2, second 1 (five each).
var DraftTurns = []int{1, 2, 2, 2, 2, 1}

// Draft returns which side picks first (0 Blue, 1 Red) and the order Red's
// units (indexes into RedArmy's units) are revealed in.
func Draft(seed uint64) (first int, reveal []int) {
	r := Sub(seed, StreamDraft)
	first = r.IntN(2)
	reveal = make([]int, ArmySize)
	for i := range reveal {
		reveal[i] = i
	}
	for i := len(reveal) - 1; i > 0; i-- {
		j := r.IntN(i + 1)
		reveal[i], reveal[j] = reveal[j], reveal[i]
	}
	return first, reveal
}
