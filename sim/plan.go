package sim

import (
	"fmt"
	"hash/fnv"
)

// PlanDef is a battle plan: an army-wide stance a side picks before the
// battle (docs/battle-plans-roadmap.md). It changes its side's units'
// temperaments by role, after the roster entry and before learned abilities.
// Each plan lives in its own file (plan_<name>.go) and registers itself under
// its id from ids.go. Its name and description for players are presentation
// (web/src/).
type PlanDef struct {
	ID                    Plan
	Name                  string
	Melee, Ranged, Caster Stance
}

// Stance is a plan's changes for one role. A zero field keeps the unit's own
// value, so a plan only says what it changes.
type Stance struct {
	Target   TargetRule
	Engage   EngageRule
	Trigger  Fx // set it whenever Engage is EngageHold
	Leash    Fx
	Keep     Fx
	FallBack int
	StayNear Tag
	Kite     Toggle
}

// Toggle is a yes/no field a stance may leave alone.
type Toggle uint8

const (
	KeepKite Toggle = iota // the unit's own value
	KiteOn
	KiteOff
)

// Plans is sorted by id; Plans[i].ID == i.
var Plans []PlanDef

func registerPlan(p PlanDef) {
	i := len(Plans)
	for i > 0 && Plans[i-1].ID > p.ID {
		i--
	}
	Plans = append(Plans, PlanDef{})
	copy(Plans[i+1:], Plans[i:])
	Plans[i] = p
	refreshVersion()
}

// PlanOf returns a plan's entry, or nil.
func PlanOf(id Plan) *PlanDef {
	if int(id) < len(Plans) && Plans[id].ID == id {
		return &Plans[id]
	}
	return nil
}

func (p Plan) String() string {
	if d := PlanOf(p); d != nil {
		return d.Name
	}
	return fmt.Sprintf("plan(%d)", p)
}

// stanceFor is the plan's stance for a unit's role: caster if it has the
// caster tag, else ranged if it shoots, else melee.
func (p *PlanDef) stanceFor(def *UnitDef) Stance {
	switch {
	case def.Tags&TagCaster != 0:
		return p.Caster
	case def.Stats.Ranged:
		return p.Ranged
	}
	return p.Melee
}

// apply changes a temperament by the stance's non-zero fields.
func (s Stance) apply(t *Temperament) {
	if s.Target != 0 {
		t.Target = s.Target
	}
	if s.Engage != 0 {
		t.Engage = s.Engage
	}
	if s.Trigger != 0 {
		t.Trigger = s.Trigger
	}
	if s.Leash != 0 {
		t.Leash = s.Leash
	}
	if s.Keep != 0 {
		t.Keep = s.Keep
	}
	if s.FallBack != 0 {
		t.FallBack = s.FallBack
	}
	if s.StayNear != 0 {
		t.StayNear = s.StayNear
	}
	switch s.Kite {
	case KiteOn:
		t.Kite = true
	case KiteOff:
		t.Kite = false
	}
}

// ValidatePlans checks every plan against every roster entry: the
// temperament it gives, alone and with each of the entry's skills, must be
// valid (no kiting melee, no holding without a trigger). Ids must match
// positions, starting with Instinct at 0.
func ValidatePlans(plans []PlanDef, roster []UnitDef) error {
	if len(plans) == 0 || plans[0] != (PlanDef{ID: PlanInstinct, Name: plans[0].Name}) {
		return fmt.Errorf("plan 0 must be Instinct and change nothing")
	}
	for i := range plans {
		p := &plans[i]
		if int(p.ID) != i || p.Name == "" {
			return fmt.Errorf("plan %d (%s): ids must match positions and every plan needs a name", p.ID, p.Name)
		}
		for j := range roster {
			def := roster[j]
			for k := -1; k < len(def.Skills); k++ {
				var skills []Skill
				if k >= 0 {
					skills = []Skill{def.Skills[k].Skill}
				}
				var kit loadout
				kit.learn(&def, skills)
				d := def
				d.Temper = temperOf(&def, p, &kit)
				if err := checkTemperament(&d); err != nil {
					return fmt.Errorf("plan %s on %s (skills %v): %v", p.Name, def.Name, skills, err)
				}
			}
		}
	}
	return nil
}

// RedPlanFor is a generated Red's plan: seeded random among all plans, from
// the combat seed, which the server reveals only after Blue's army is locked,
// so the player drafts against Red's units, not its strategy.
func RedPlanFor(combatSeed uint64) Plan {
	return Plans[Sub(combatSeed, StreamRedPlan).IntN(len(Plans))].ID
}

// plansFingerprint is a short hash of the plan table: plans are rules, so
// changing one changes the sim version.
func plansFingerprint() string {
	h := fnv.New32a()
	fmt.Fprintf(h, "%#v", Plans)
	return fmt.Sprintf("%08x", h.Sum32())
}
