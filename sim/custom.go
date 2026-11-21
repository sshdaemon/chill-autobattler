package sim

import "fmt"

// A custom plan is a plan the player builds from chips (docs/battle-plans-
// roadmap.md, step 5): per role a stance from a small closed vocabulary and at
// most one edge from the menu, optionally split into a left and a right group
// with an opening stance, one trigger and a second stance each. Def turns it
// into a PlanDef, checked by the same rules as the built-in plans.

// Step is a coarse distance chip: 0 own, then close/short, mid/long, far.
type Step uint8

// Chips is one role's stance as chips; zero values keep the unit's own.
type Chips struct {
	Target   TargetRule
	Engage   EngageRule
	Trigger  Step  // hold trigger (TriggerSteps)
	Keep     Step  // keep distance (KeepSteps)
	Leash    Step  // leash (LeashSteps)
	FallBack uint8 // FallBackSteps index
	StayNear uint8 // StayNearSteps index
	Kite     Toggle
	Edge     Edge
}

// CustomTrigger is a custom group's trigger: a kind and, for "after", one of
// TriggerSeconds.
type CustomTrigger struct {
	Kind    TriggerKind
	Seconds int
}

type CustomGroup struct {
	Opening, Then [3]Chips // melee, ranged, caster
	When          CustomTrigger
}

type CustomPlan struct {
	Roles  [3]Chips      // melee, ranged, caster
	Groups []CustomGroup // none, or left and right
}

var roleNames = [3]string{"melee", "ranged", "caster"}

// stance checks a role's chips and builds its stance; a second stance may not
// take an edge.
func (c Chips) stance(role int, edges bool) (Stance, error) {
	var s Stance
	if c.Target != 0 && targetRules[c.Target] == nil {
		return s, fmt.Errorf("unknown target rule %d", c.Target)
	}
	if c.Engage != 0 && engageRules[c.Engage] == nil {
		return s, fmt.Errorf("unknown engage rule %d", c.Engage)
	}
	if int(c.Trigger) >= len(TriggerSteps) || int(c.Keep) >= len(KeepSteps) || int(c.Leash) >= len(LeashSteps) ||
		int(c.FallBack) >= len(FallBackSteps) || int(c.StayNear) >= len(StayNearSteps) || c.Kite > KiteOff {
		return s, fmt.Errorf("a distance or choice outside the chips: %+v", c)
	}
	if c.Kite != KeepKite && role != 1 {
		return s, fmt.Errorf("only ranged units kite")
	}
	s = Stance{Target: c.Target, Engage: c.Engage, Trigger: TriggerSteps[c.Trigger], Keep: KeepSteps[c.Keep], Leash: LeashSteps[c.Leash],
		FallBack: FallBackSteps[c.FallBack], StayNear: StayNearSteps[c.StayNear], Kite: c.Kite}
	if c.Edge != 0 {
		e := EdgeOf(c.Edge)
		switch {
		case e == nil:
			return s, fmt.Errorf("unknown edge %d", c.Edge)
		case !edges:
			return s, fmt.Errorf("a second stance can't take an edge")
		case !e.Roles[role]:
			return s, fmt.Errorf("edge %s isn't for %s units", e.Name, roleNames[role])
		}
		s.Effects = e.Effects
	}
	return s, nil
}

func rolesOf(cs [3]Chips, edges bool) (Roles, error) {
	var r Roles
	out := [3]*Stance{&r.Melee, &r.Ranged, &r.Caster}
	for i, c := range cs {
		s, err := c.stance(i, edges)
		if err != nil {
			return r, fmt.Errorf("%s: %v", roleNames[i], err)
		}
		*out[i] = s
	}
	return r, nil
}

func (t CustomTrigger) check() error {
	switch t.Kind {
	case 0, TriggerAllyFalls, TriggerOtherFighting:
		if t.Seconds != 0 {
			return fmt.Errorf("trigger %v takes no seconds", t.Kind)
		}
		return nil
	case TriggerAfter:
		for _, s := range TriggerSeconds {
			if t.Seconds == s {
				return nil
			}
		}
		return fmt.Errorf("after %d s: want one of %v", t.Seconds, TriggerSeconds)
	}
	return fmt.Errorf("unknown trigger %d", t.Kind)
}

// Def builds and checks the plan the chips describe.
func (c *CustomPlan) Def() (PlanDef, error) {
	p := PlanDef{ID: PlanCustom, Name: "custom"}
	r, err := rolesOf(c.Roles, true)
	if err != nil {
		return p, err
	}
	p.Melee, p.Ranged, p.Caster = r.Melee, r.Ranged, r.Caster
	switch len(c.Groups) {
	case 0:
	case 2:
		for g, cg := range c.Groups {
			if err := cg.When.check(); err != nil {
				return p, fmt.Errorf("group %d: %v", g, err)
			}
			open, err := rolesOf(cg.Opening, true)
			if err != nil {
				return p, fmt.Errorf("group %d: %v", g, err)
			}
			then, err := rolesOf(cg.Then, false)
			if err != nil {
				return p, fmt.Errorf("group %d, second stance: %v", g, err)
			}
			p.Groups[g] = GroupDef{Opening: open, When: Trigger{Kind: cg.When.Kind, Seconds: cg.When.Seconds}, Then: then}
		}
	default:
		return p, fmt.Errorf("a custom plan has no groups or two, not %d", len(c.Groups))
	}
	if err := validatePlan(&p, Roster); err != nil {
		return p, err
	}
	return p, nil
}

// ChipsOf is a plan as the nearest chips, for the editor to start from: each
// distance snapped to the nearest step, and per role the menu edge whose
// headline effect (kind, stat, condition) matches the plan's first effect.
func ChipsOf(p *PlanDef) CustomPlan {
	var c CustomPlan
	c.Roles = chipsOfRoles(Roles{p.Melee, p.Ranged, p.Caster}, true)
	if p.IsTactic() {
		c.Groups = make([]CustomGroup, 2)
		for g, gd := range p.Groups {
			c.Groups[g].Opening = chipsOfRoles(gd.Opening, true)
			c.Groups[g].Then = chipsOfRoles(gd.Then, false)
			c.Groups[g].When = CustomTrigger{Kind: gd.When.Kind}
			if gd.When.Kind == TriggerAfter {
				c.Groups[g].When.Seconds = TriggerSeconds[nearestInt(TriggerSeconds, gd.When.Seconds)]
			}
		}
	}
	return c
}

func chipsOfRoles(r Roles, edges bool) [3]Chips {
	var out [3]Chips
	for i, s := range []Stance{r.Melee, r.Ranged, r.Caster} {
		c := Chips{Target: s.Target, Engage: s.Engage, Kite: s.Kite,
			Trigger: nearestStep(TriggerSteps[:], s.Trigger), Keep: nearestStep(KeepSteps[:], s.Keep), Leash: nearestStep(LeashSteps[:], s.Leash)}
		if s.FallBack != 0 {
			c.FallBack = uint8(1 + nearestInt(FallBackSteps[1:], s.FallBack))
		}
		switch {
		case s.StayNear == 0:
		case s.StayNear&TagMelee != 0:
			c.StayNear = 1
		default:
			c.StayNear = 2
		}
		if edges && len(s.Effects) > 0 {
			h := s.Effects[0]
			for _, e := range Edges {
				m := e.Effects[0]
				if e.Roles[i] && m.Kind == h.Kind && m.Stat == h.Stat && m.Cond == h.Cond && (m.Pct > 0) == (h.Pct > 0) {
					c.Edge = e.ID
					break
				}
			}
		}
		out[i] = c
	}
	return out
}

// nearestStep is the index of the step nearest v, 0 ("own") for v == 0.
func nearestStep(steps []Fx, v Fx) Step {
	if v == 0 {
		return 0
	}
	best := 1
	for i := 2; i < len(steps); i++ {
		if (steps[i] - v).Abs() < (steps[best] - v).Abs() {
			best = i
		}
	}
	return Step(best)
}

func nearestInt(xs []int, v int) int {
	best := 0
	for i := range xs {
		d, bd := xs[i]-v, xs[best]-v
		if d*d < bd*bd {
			best = i
		}
	}
	return best
}
