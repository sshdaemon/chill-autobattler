package sim

// Romans (design/SETTING.md): skoutatos, toxotes and siphonarios, finished in
// roadmap Phase 4. Playable: players can pick it and Red may draw it.
func init() {
	registerFaction(FactionDef{ID: FactionRomans, Name: "romans", Units: []Class{Skoutatos, Toxotes, Siphonarios}, Playable: true})
}
