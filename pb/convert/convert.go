// Package convert maps the wire format (chill.v1 protobuf) to the sim's own
// types and back. The browser (WASM) and the server both use it, so they
// interpret a battle input identically.
package convert

import (
	"fmt"

	chillv1 "chill/pb/chill/v1"
	"chill/sim"
)

// Input converts a BattleInput and validates it.
func Input(in *chillv1.BattleInput) (sim.Input, error) {
	if in == nil {
		return sim.Input{}, fmt.Errorf("%w: missing input", sim.ErrInvalid)
	}
	if in.RedSkillCount > sim.ArmySize*sim.MaxSkills {
		return sim.Input{}, fmt.Errorf("%w: red skill count %d out of range", sim.ErrInvalid, in.RedSkillCount)
	}
	if in.RedHealthPct > 100 {
		return sim.Input{}, fmt.Errorf("%w: red health %d%% out of range", sim.ErrInvalid, in.RedHealthPct)
	}
	if in.BlueFaction > 255 {
		return sim.Input{}, fmt.Errorf("%w: faction %d out of range", sim.ErrInvalid, in.BlueFaction)
	}
	out := sim.Input{Version: in.SimVersion, Seed: in.Seed, CombatSeed: in.CombatSeed, RedSkills: int(in.RedSkillCount), RedHealthPct: int(in.RedHealthPct), BlueFaction: sim.Faction(uint8(in.BlueFaction))}
	var err error
	if out.Blue, err = army(in.Blue, "blue", true); err != nil {
		return sim.Input{}, err
	}
	if out.Red, err = army(in.Red, "red", false); err != nil {
		return sim.Input{}, err
	}
	return out, out.Validate()
}

func army(a *chillv1.Army, name string, required bool) (sim.Army, error) {
	if len(a.GetUnits()) == 0 {
		if required {
			return sim.Army{}, fmt.Errorf("%w: %s has no units", sim.ErrInvalid, name)
		}
		return sim.Army{}, nil
	}
	if a.Formation == chillv1.Formation_FORMATION_UNSPECIFIED {
		return sim.Army{}, fmt.Errorf("%w: %s formation not set", sim.ErrInvalid, name)
	}
	if a.Plan > 255 {
		return sim.Army{}, fmt.Errorf("%w: %s plan %d", sim.ErrInvalid, name, a.Plan)
	}
	out := sim.Army{Formation: sim.Formation(a.Formation - chillv1.Formation_FORMATION_CUSTOM), Plan: sim.Plan(a.Plan)} // sim checks the plan exists
	if a.CustomPlan != nil {
		c, err := customPlan(a.CustomPlan)
		if err != nil {
			return sim.Army{}, fmt.Errorf("%w: %s custom plan: %v", sim.ErrInvalid, name, err)
		}
		out.Custom = c
	}
	for i, u := range a.Units {
		if u.UnitClass == chillv1.UnitClass_UNIT_CLASS_UNSPECIFIED {
			return sim.Army{}, fmt.Errorf("%w: %s unit %d has no class", sim.ErrInvalid, name, i)
		}
		spec := sim.UnitSpec{Class: sim.Class(u.UnitClass - 1)}
		for _, s := range u.Skills {
			spec.Skills = append(spec.Skills, sim.Skill(s)) // the unit's ability number, as in sim
		}
		if u.Position != nil {
			p := sim.V(sim.Fx(u.Position.XRaw), sim.Fx(u.Position.ZRaw))
			spec.Pos = &p
		}
		out.Units = append(out.Units, spec)
	}
	return out, nil
}

// customPlan converts a custom plan's chips; the sim checks what they mean.
func customPlan(c *chillv1.CustomPlan) (*sim.CustomPlan, error) {
	out := &sim.CustomPlan{}
	var err error
	if out.Roles, err = planRoles(c.Roles); err != nil {
		return nil, err
	}
	for i, g := range c.Groups {
		if g.Trigger > 255 || g.TriggerSeconds > 1000 {
			return nil, fmt.Errorf("group %d: trigger %d after %d s", i, g.Trigger, g.TriggerSeconds)
		}
		cg := sim.CustomGroup{When: sim.CustomTrigger{Kind: sim.TriggerKind(g.Trigger), Seconds: int(g.TriggerSeconds)}}
		if cg.Opening, err = planRoles(g.Opening); err != nil {
			return nil, fmt.Errorf("group %d: %v", i, err)
		}
		if cg.Then, err = planRoles(g.Then); err != nil {
			return nil, fmt.Errorf("group %d: %v", i, err)
		}
		out.Groups = append(out.Groups, cg)
	}
	return out, nil
}

func planRoles(r *chillv1.PlanRoles) ([3]sim.Chips, error) {
	var out [3]sim.Chips
	for i, c := range []*chillv1.PlanChips{r.GetMelee(), r.GetRanged(), r.GetCaster()} {
		if c == nil {
			continue
		}
		vals := []uint32{c.Target, c.Engage, c.Trigger, c.Keep, c.Leash, c.FallBack, c.StayNear, c.Kite, c.Edge}
		for _, v := range vals {
			if v > 255 {
				return out, fmt.Errorf("chip value %d out of range", v)
			}
		}
		out[i] = sim.Chips{Target: sim.TargetRule(c.Target), Engage: sim.EngageRule(c.Engage), Trigger: sim.Step(c.Trigger), Keep: sim.Step(c.Keep),
			Leash: sim.Step(c.Leash), FallBack: uint8(c.FallBack), StayNear: uint8(c.StayNear), Kite: sim.Toggle(c.Kite), Edge: sim.Edge(c.Edge)}
	}
	return out, nil
}

// Outcome converts a sim outcome to the wire format.
func Outcome(o sim.Outcome) *chillv1.BattleOutcome {
	out := &chillv1.BattleOutcome{
		Result:   chillv1.BattleResult(o.Result), // Ongoing=0 maps to UNSPECIFIED
		Ticks:    uint32(o.Ticks),
		Checksum: o.Checksum,
	}
	for _, s := range o.Survivors {
		out.Survivors = append(out.Survivors, &chillv1.Survivor{Index: uint32(s.Index), HpRaw: int64(s.HP)})
	}
	return out
}
