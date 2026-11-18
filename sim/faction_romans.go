package sim

// romans (design/SETTING.md). Not playable until its units are finished
// (roadmap Phase 4): players can't pick it and Red never draws it.
func init() {
	registerFaction(FactionDef{ID: FactionRomans, Name: "romans", Units: []Class{Skoutatos, Toxotes, Siphonarios}})
}
