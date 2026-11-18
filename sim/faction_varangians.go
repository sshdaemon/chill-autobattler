package sim

// Varangians (design/SETTING.md): huskarl, bowman and völva, finished in
// roadmap Phase 4. Playable: players can pick it and Red may draw it.
func init() {
	registerFaction(FactionDef{ID: FactionVarangians, Name: "varangians", Units: []Class{Huskarl, Bowman, Volva}, Playable: true})
}
