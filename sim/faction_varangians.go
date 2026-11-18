package sim

// varangians (design/SETTING.md). Not playable until its units are finished
// (roadmap Phase 4): players can't pick it and Red never draws it.
func init() {
	registerFaction(FactionDef{ID: FactionVarangians, Name: "varangians", Units: []Class{Huskarl, Bowman, Volva}})
}
