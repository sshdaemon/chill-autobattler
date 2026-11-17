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
	out := sim.Input{Version: in.SimVersion, Seed: in.Seed, RedSkills: int(in.RedSkillCount)}
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
	out := sim.Army{Formation: sim.Formation(a.Formation - chillv1.Formation_FORMATION_CUSTOM)}
	for i, u := range a.Units {
		if u.UnitClass == chillv1.UnitClass_UNIT_CLASS_UNSPECIFIED {
			return sim.Army{}, fmt.Errorf("%w: %s unit %d has no class", sim.ErrInvalid, name, i)
		}
		spec := sim.UnitSpec{Class: sim.Class(u.UnitClass - 1)}
		for _, s := range u.Skills {
			spec.Skills = append(spec.Skills, sim.Skill(s)) // same numbering as sim.Skill
		}
		if u.Position != nil {
			p := sim.V(sim.Fx(u.Position.XRaw), sim.Fx(u.Position.ZRaw))
			spec.Pos = &p
		}
		out.Units = append(out.Units, spec)
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
