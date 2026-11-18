package sim

// Kyivan Rus (design/SETTING.md): druzhynnyk, strilets and huslyar, finished
// in roadmap Phase 4. Playable: players can pick it and Red may draw it.
func init() {
	registerFaction(FactionDef{ID: FactionRus, Name: "rus", Units: []Class{Druzhynnyk, Strilets, Huslyar}, Playable: true})
}
