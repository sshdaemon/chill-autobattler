package sim

import "fmt"

// FactionDef is a faction: the unit types a side drafts from (Blue's army must
// come from its faction) and Red's generated army is drawn from. Each faction
// lives in its own file (faction_<name>.go) and registers itself under its id
// from ids.go. Its lore and looks are presentation (web/src/factions/).
type FactionDef struct {
	ID    Faction
	Name  string
	Units []Class // roster ids, in the order the UI lists them
	// Playable: players may pick it and Red may draw it. A faction whose units
	// are still placeholders isn't.
	Playable bool
}

// Factions is sorted by id.
var Factions []FactionDef

func registerFaction(f FactionDef) {
	i := len(Factions)
	for i > 0 && Factions[i-1].ID > f.ID {
		i--
	}
	Factions = append(Factions, FactionDef{})
	copy(Factions[i+1:], Factions[i:])
	Factions[i] = f
	refreshVersion()
}

// FactionOf returns a faction's entry, or nil.
func FactionOf(id Faction) *FactionDef {
	for i := range Factions {
		if Factions[i].ID == id {
			return &Factions[i]
		}
	}
	return nil
}

// FactionWith returns the first faction that fields a unit type (0 if none).
func FactionWith(c Class) Faction {
	for _, f := range Factions {
		if f.Has(c) {
			return f.ID
		}
	}
	return 0
}

// Has reports whether the faction fields a unit type.
func (f *FactionDef) Has(c Class) bool {
	for _, u := range f.Units {
		if u == c {
			return true
		}
	}
	return false
}

// Pool is the faction's units Red's generated armies draw from (InRedPool), in roster order.
func (f *FactionDef) Pool() []Class {
	var out []Class
	for d := range Units() {
		if d.InRedPool && f.Has(d.ID) {
			out = append(out, d.ID)
		}
	}
	return out
}

// RedFactionFor is Red's faction for a battle: seeded random among the
// playable factions, so it may be the same as Blue's.
func RedFactionFor(seed uint64) Faction {
	var ids []Faction
	for _, f := range Factions {
		if f.Playable {
			ids = append(ids, f.ID)
		}
	}
	return ids[Sub(seed, StreamRedFaction).IntN(len(ids))]
}

// ValidateFactions checks the factions against the roster: known units, no
// unit twice, and something for Red to draw from.
func ValidateFactions(fs []FactionDef) error {
	playable := 0
	for _, f := range fs {
		if f.Playable {
			playable++
		}
	}
	if playable == 0 {
		return fmt.Errorf("no playable faction")
	}
	for i, f := range fs {
		if f.ID == 0 || f.Name == "" || (i > 0 && fs[i-1].ID >= f.ID) {
			return fmt.Errorf("faction %d (%s): needs an id (sorted, unique) and a name", f.ID, f.Name)
		}
		seen := map[Class]bool{}
		for _, c := range f.Units {
			if RosterOf(c) == nil || seen[c] {
				return fmt.Errorf("faction %s: unit %d unknown or listed twice", f.Name, c)
			}
			seen[c] = true
		}
		if len(f.Pool()) == 0 {
			return fmt.Errorf("faction %s: no unit in Red's pool, so Red couldn't play it", f.Name)
		}
	}
	return nil
}
