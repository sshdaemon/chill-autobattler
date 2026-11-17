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
		}
		return skip
	})
	return a, err
}

// DecodeInput parses a chill.v1.BattleInput and validates it, exactly like
// convert.Input(proto.Unmarshal(b)).
func DecodeInput(b []byte) (sim.Input, error) {
	var (
		version          []byte
		seed, combat     uint64
		redSkills, redHP uint64
		blue, red        army
		blueErr, redErr  error
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
		}
		return skip
	})
	if err == nil {
		err = errors.Join(blueErr, redErr)
	}
	if err != nil {
		return sim.Input{}, fmt.Errorf("bad input bytes: %w", err)
	}
	if uint32(redHP) > 100 {
		return sim.Input{}, fmt.Errorf("%w: red health %d%% out of range", sim.ErrInvalid, uint32(redHP))
	}
	if uint32(redSkills) > sim.ArmySize*sim.MaxSkills {
		return sim.Input{}, fmt.Errorf("%w: red skill count %d out of range", sim.ErrInvalid, uint32(redSkills))
	}
	in := sim.Input{Version: string(version), Seed: seed, CombatSeed: combat, RedSkills: int(uint32(redSkills)), RedHealthPct: int(uint32(redHP))}
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
	out := sim.Army{Formation: sim.Formation(a.formation - 1)} // pb FORMATION_CUSTOM = 1
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
