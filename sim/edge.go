package sim

import "fmt"

// EdgeDef is an edge a custom plan's role may take (one per role): effects
// for the battle, an advantage with its cost, for the roles in Roles.
// Each edge lives in its own edge_<name>.go and registers under its id from
// ids.go. Edges are rules: they go into the plans fingerprint.
type EdgeDef struct {
	ID      Edge
	Name    string
	Roles   [3]bool // melee, ranged, caster: who may take it
	Effects []Effect
}

// Edges is sorted by id; Edges[i].ID == i+1.
var Edges []EdgeDef

func registerEdge(e EdgeDef) {
	i := len(Edges)
	for i > 0 && Edges[i-1].ID > e.ID {
		i--
	}
	Edges = append(Edges, EdgeDef{})
	copy(Edges[i+1:], Edges[i:])
	Edges[i] = e
	refreshVersion()
}

// EdgeOf returns an edge's entry, or nil (0 is none).
func EdgeOf(id Edge) *EdgeDef {
	if id > 0 && int(id) <= len(Edges) && Edges[id-1].ID == id {
		return &Edges[id-1]
	}
	return nil
}

func (e Edge) String() string {
	if d := EdgeOf(e); d != nil {
		return d.Name
	}
	return fmt.Sprintf("edge(%d)", e)
}

// ValidateEdges checks ids and that each edge's effects suit every unit of
// the roles it is offered to.
func ValidateEdges(edges []EdgeDef, roster []UnitDef) error {
	for i, e := range edges {
		if int(e.ID) != i+1 || e.Name == "" || len(e.Effects) == 0 {
			return fmt.Errorf("edge %d (%s): ids must be 1.. in order, with a name and effects", e.ID, e.Name)
		}
		for j := range roster {
			d := &roster[j]
			if d.Retired() || !e.Roles[roleOf(d)] {
				continue
			}
			for _, x := range e.Effects {
				if err := checkEffect(x, d); err != nil {
					return fmt.Errorf("edge %s on %s: %v", e.Name, d.Name, err)
				}
			}
		}
	}
	return nil
}

// roleOf is a unit's role for plans: 0 melee, 1 ranged, 2 caster (caster
// first, as in stanceFor).
func roleOf(d *UnitDef) int {
	switch {
	case d.Tags&TagCaster != 0:
		return 2
	case d.Stats.Ranged:
		return 1
	}
	return 0
}
