// Package wire decodes BattleInput and encodes BattleOutcome with only the
// low-level protowire functions, straight into the sim's own types. The
// browser build uses it so it doesn't link the protobuf runtime (about 6 MB
// of WASM). It must behave exactly like pb/convert on the generated types:
// wire_test.go checks that over many inputs. Field numbers and enum offsets
// follow proto/chill/v1/*.proto.
package wire

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"

	"chill/sim"
)

var errBad = errors.New("malformed protobuf")

// fields calls fn for every field in b; fn gets the number, type and the
// remaining bytes, and returns how many it consumed (or <0 on error).
func fields(b []byte, fn func(num protowire.Number, typ protowire.Type, b []byte) int) error {
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return errBad
		}
		b = b[n:]
		m := fn(num, typ, b)
		if m == -2 { // not ours: skip like any proto3 parser
			m = protowire.ConsumeFieldValue(num, typ, b)
		}
		if m < 0 {
			return errBad
		}
		b = b[m:]
	}
	return nil
}

const skip = -2

func varint(typ protowire.Type, b []byte, out *uint64) int {
	if typ != protowire.VarintType {
		return skip
	}
	v, n := protowire.ConsumeVarint(b)
	*out = v
	return n
}

func bytesField(typ protowire.Type, b []byte, out *[]byte) int {
	if typ != protowire.BytesType {
		return skip
	}
	v, n := protowire.ConsumeBytes(b)
	*out = v
	return n
}

type point struct {
	set  bool
	x, z int64
}

type unitSpec struct {
	class  int32
	skills []int32
	pos    point
}

type army struct {
	present   bool
	formation int32
	units     []unitSpec
	plan      uint32 // battle plan (field 3)
	custom    bool   // custom_plan (field 4) is present
	customRaw []byte // its bytes, every occurrence concatenated (= merged, in protobuf)
}

func decodePoint(b []byte) (point, error) {
	p := point{set: true}
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		var v uint64
		switch num {
		case 1:
			n := varint(typ, b, &v)
			p.x = protowire.DecodeZigZag(v)
			return n
		case 2:
			n := varint(typ, b, &v)
			p.z = protowire.DecodeZigZag(v)
			return n
		}
		return skip
	})
	return p, err
}

func decodeUnit(b []byte) (unitSpec, error) {
	var u unitSpec
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		var v uint64
		switch num {
		case 1:
			n := varint(typ, b, &v)
			u.class = int32(v)
			return n
		case 2: // repeated enum: packed (proto3 default) or one per tag
			if typ == protowire.VarintType {
				n := varint(typ, b, &v)
				u.skills = append(u.skills, int32(v))
				return n
			}
			var packed []byte
			n := bytesField(typ, b, &packed)
			for len(packed) > 0 {
				s, m := protowire.ConsumeVarint(packed)
				if m < 0 {
					return -1
				}
				u.skills = append(u.skills, int32(s))
				packed = packed[m:]
			}
			return n
		case 3:
			var sub []byte
			n := bytesField(typ, b, &sub)
			if n < 0 {
				return n
			}
			p, err := decodePoint(sub)
			if err != nil {
				return -1
			}
			u.pos = p
			return n
		}
		return skip
	})
	return u, err
}

func decodeArmy(b []byte) (army, error) {
	a := army{present: true}
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		var v uint64
		switch num {
		case 1:
			n := varint(typ, b, &v)
			a.formation = int32(v)
			return n
		case 2:
			var sub []byte
			n := bytesField(typ, b, &sub)
			if n < 0 {
				return n
			}
			u, err := decodeUnit(sub)
			if err != nil {
				return -1
			}
			a.units = append(a.units, u)
			return n
		case 3:
			n := varint(typ, b, &v)
			a.plan = uint32(v) // proto uint32: the low 32 bits
			return n
		case 4:
			var sub []byte
			n := bytesField(typ, b, &sub)
			if n >= 0 {
				a.custom, a.customRaw = true, append(a.customRaw, sub...)
			}
			return n
		}
		return skip
	})
	return a, err
}

// Custom plans. A message field that occurs more than once merges, which in
// protobuf is the same as decoding its occurrences' bytes concatenated; a
// repeated message field's occurrences are separate elements.

type chips [9]uint32 // PlanChips fields 1..9

func decodeChips(b []byte) (chips, error) {
	var c chips
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		if num < 1 || num > 9 {
			return skip
		}
		var v uint64
		n := varint(typ, b, &v)
		c[num-1] = uint32(v)
		return n
	})
	return c, err
}

func decodeRoles(b []byte) ([3]chips, error) {
	var out [3]chips
	raw := make([][]byte, 3)
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		if num < 1 || num > 3 {
			return skip
		}
		var sub []byte
		n := bytesField(typ, b, &sub)
		if n >= 0 {
			raw[num-1] = append(raw[num-1], sub...)
		}
		return n
	})
	if err != nil {
		return out, err
	}
	for i, r := range raw {
		c, err := decodeChips(r)
		if err != nil {
			return out, err
		}
		out[i] = c
	}
	return out, nil
}

type group struct {
	opening, then     [3]chips
	trigger, triggerS uint32
}

func decodeGroup(b []byte) (group, error) {
	var g group
	var open, then []byte
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		var sub []byte
		var v uint64
		switch num {
		case 1, 4:
			n := bytesField(typ, b, &sub)
			if n >= 0 {
				if num == 1 {
					open = append(open, sub...)
				} else {
					then = append(then, sub...)
				}
			}
			return n
		case 2:
			n := varint(typ, b, &v)
			g.trigger = uint32(v)
			return n
		case 3:
			n := varint(typ, b, &v)
			g.triggerS = uint32(v)
			return n
		}
		return skip
	})
	if err != nil {
		return g, err
	}
	if g.opening, err = decodeRoles(open); err != nil {
		return g, err
	}
	g.then, err = decodeRoles(then)
	return g, err
}

func decodeCustom(b []byte) ([3]chips, []group, error) {
	var roles []byte
	var groups []group
	var gerr error
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		var sub []byte
		switch num {
		case 1:
			n := bytesField(typ, b, &sub)
			if n >= 0 {
				roles = append(roles, sub...)
			}
			return n
		case 2:
			n := bytesField(typ, b, &sub)
			if n >= 0 {
				g, err := decodeGroup(sub)
				gerr = errors.Join(gerr, err)
				groups = append(groups, g)
			}
			return n
		}
		return skip
	})
	if err = errors.Join(err, gerr); err != nil {
		return [3]chips{}, nil, err
	}
	r, err := decodeRoles(roles)
	return r, groups, err
}

func simChips(cs [3]chips) ([3]sim.Chips, error) {
	var out [3]sim.Chips
	for i, c := range cs {
		for _, v := range c {
			if v > 255 {
				return out, fmt.Errorf("chip value %d out of range", v)
			}
		}
		out[i] = sim.Chips{Target: sim.TargetRule(c[0]), Engage: sim.EngageRule(c[1]), Trigger: sim.Step(c[2]), Keep: sim.Step(c[3]),
			Leash: sim.Step(c[4]), FallBack: uint8(c[5]), StayNear: uint8(c[6]), Kite: sim.Toggle(c[7]), Edge: sim.Edge(c[8])}
	}
	return out, nil
}

// toCustom decodes and converts a custom plan, like convert's customPlan.
func toCustom(raw []byte) (*sim.CustomPlan, error) {
	roles, groups, err := decodeCustom(raw)
	if err != nil {
		return nil, fmt.Errorf("bad input bytes: %w", err)
	}
	out := &sim.CustomPlan{}
	if out.Roles, err = simChips(roles); err != nil {
		return nil, fmt.Errorf("%w: %v", sim.ErrInvalid, err)
	}
	for i, g := range groups {
		if g.trigger > 255 || g.triggerS > 1000 {
			return nil, fmt.Errorf("%w: group %d: trigger %d after %d s", sim.ErrInvalid, i, g.trigger, g.triggerS)
		}
		cg := sim.CustomGroup{When: sim.CustomTrigger{Kind: sim.TriggerKind(g.trigger), Seconds: int(g.triggerS)}}
		if cg.Opening, err = simChips(g.opening); err != nil {
			return nil, fmt.Errorf("%w: group %d: %v", sim.ErrInvalid, i, err)
		}
		if cg.Then, err = simChips(g.then); err != nil {
			return nil, fmt.Errorf("%w: group %d: %v", sim.ErrInvalid, i, err)
		}
		out.Groups = append(out.Groups, cg)
	}
	return out, nil
}

// DecodeInput parses a chill.v1.BattleInput and validates it, exactly like
// convert.Input(proto.Unmarshal(b)).
func DecodeInput(b []byte) (sim.Input, error) {
	var (
		version                   []byte
		seed, combat              uint64
		redSkills, redHP, faction uint64
		blue, red                 army
		blueErr, redErr           error
	)
	err := fields(b, func(num protowire.Number, typ protowire.Type, b []byte) int {
		var sub []byte
		switch num {
		case 1:
			return bytesField(typ, b, &version)
		case 2:
			return varint(typ, b, &seed)
		case 3:
			n := bytesField(typ, b, &sub)
			if n >= 0 {
				blue, blueErr = mergeArmy(blue, sub)
			}
			return n
		case 4:
			n := bytesField(typ, b, &sub)
			if n >= 0 {
				red, redErr = mergeArmy(red, sub)
			}
			return n
		case 5:
			return varint(typ, b, &redSkills)
		case 6:
			return varint(typ, b, &combat)
		case 7:
			return varint(typ, b, &redHP)
		case 8:
			return varint(typ, b, &faction)
		}
		return skip
	})
	if err == nil {
		err = errors.Join(blueErr, redErr)
	}
	if err != nil {
		return sim.Input{}, fmt.Errorf("bad input bytes: %w", err)
	}
	if uint32(faction) > 255 {
		return sim.Input{}, fmt.Errorf("%w: faction %d out of range", sim.ErrInvalid, uint32(faction))
	}
	if uint32(redHP) > 100 {
		return sim.Input{}, fmt.Errorf("%w: red health %d%% out of range", sim.ErrInvalid, uint32(redHP))
	}
	if uint32(redSkills) > sim.ArmySize*sim.MaxSkills {
		return sim.Input{}, fmt.Errorf("%w: red skill count %d out of range", sim.ErrInvalid, uint32(redSkills))
	}
	in := sim.Input{Version: string(version), Seed: seed, CombatSeed: combat, RedSkills: int(uint32(redSkills)), RedHealthPct: int(uint32(redHP)), BlueFaction: sim.Faction(uint8(faction))}
	if in.Blue, err = toSim(blue, "blue", true); err != nil {
		return sim.Input{}, err
	}
	if in.Red, err = toSim(red, "red", false); err != nil {
		return sim.Input{}, err
	}
	return in, in.Validate()
}

// mergeArmy follows proto's rule that a repeated occurrence of a message
// field merges into the earlier one.
func mergeArmy(prev army, b []byte) (army, error) {
	a, err := decodeArmy(b)
	if err != nil {
		return prev, err
	}
	if prev.present {
		if a.formation == 0 {
			a.formation = prev.formation
		}
		if a.plan == 0 {
			a.plan = prev.plan
		}
		a.custom, a.customRaw = a.custom || prev.custom, append(append([]byte(nil), prev.customRaw...), a.customRaw...)
		a.units = append(prev.units, a.units...)
	}
	return a, nil
}

func toSim(a army, name string, required bool) (sim.Army, error) {
	if len(a.units) == 0 {
		if required {
			return sim.Army{}, fmt.Errorf("%w: %s has no units", sim.ErrInvalid, name)
		}
		return sim.Army{}, nil
	}
	if a.formation == 0 {
		return sim.Army{}, fmt.Errorf("%w: %s formation not set", sim.ErrInvalid, name)
	}
	if a.plan > 255 {
		return sim.Army{}, fmt.Errorf("%w: %s plan %d", sim.ErrInvalid, name, a.plan)
	}
	out := sim.Army{Formation: sim.Formation(a.formation - 1), Plan: sim.Plan(a.plan)} // pb FORMATION_CUSTOM = 1
	if a.custom {
		c, err := toCustom(a.customRaw)
		if err != nil {
			return sim.Army{}, fmt.Errorf("%s custom plan: %w", name, err)
		}
		out.Custom = c
	}
	for i, u := range a.units {
		if u.class == 0 {
			return sim.Army{}, fmt.Errorf("%w: %s unit %d has no class", sim.ErrInvalid, name, i)
		}
		spec := sim.UnitSpec{Class: sim.Class(u.class - 1)}
		for _, s := range u.skills {
			spec.Skills = append(spec.Skills, sim.Skill(s))
		}
		if u.pos.set {
			p := sim.V(sim.Fx(u.pos.x), sim.Fx(u.pos.z))
			spec.Pos = &p
		}
		out.Units = append(out.Units, spec)
	}
	return out, nil
}

// EncodeOutcome writes a chill.v1.BattleOutcome (zero fields omitted, as
// proto3 does).
func EncodeOutcome(o sim.Outcome) []byte {
	var b []byte
	uvar := func(num protowire.Number, v uint64) {
		if v != 0 {
			b = protowire.AppendTag(b, num, protowire.VarintType)
			b = protowire.AppendVarint(b, v)
		}
	}
	uvar(1, uint64(int64(o.Result))) // Ongoing=0 is UNSPECIFIED
	uvar(2, uint64(uint32(o.Ticks)))
	for _, s := range o.Survivors {
		var m []byte
		if s.Index != 0 {
			m = protowire.AppendTag(m, 1, protowire.VarintType)
			m = protowire.AppendVarint(m, uint64(uint32(s.Index)))
		}
		if s.HP != 0 {
			m = protowire.AppendTag(m, 2, protowire.VarintType)
			m = protowire.AppendVarint(m, protowire.EncodeZigZag(int64(s.HP)))
		}
		b = protowire.AppendTag(b, 3, protowire.BytesType)
		b = protowire.AppendBytes(b, m)
	}
	uvar(4, o.Checksum)
	return b
}
