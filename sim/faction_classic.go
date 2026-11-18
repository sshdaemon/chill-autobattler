package sim

// Classic: the original three units, until the Dnipro-road factions are
// playable (docs/ROADMAP.md).
func init() {
	registerFaction(FactionDef{ID: FactionClassic, Name: "classic", Units: []Class{Warrior, Archer, Mage}})
}
