package wire

import (
	"errors"
	"math/rand/v2"
	"reflect"
	"testing"

	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	chillv1 "chill/pb/chill/v1"
	"chill/pb/convert"
	"chill/sim"
)

func randArmy(r *rand.Rand) *chillv1.Army {
	if r.IntN(4) == 0 {
		return nil
	}
	a := &chillv1.Army{Formation: chillv1.Formation(r.IntN(11))}
	switch r.IntN(4) { // battle plan: mostly a real one, sometimes unknown or out of range
	case 0, 1:
		a.Plan = uint32(r.IntN(len(sim.Plans)))
	case 2:
		a.Plan = uint32(r.IntN(300))
	}
	for i, n := 0, r.IntN(7); i < n; i++ {
		u := &chillv1.UnitSpec{UnitClass: chillv1.UnitClass(r.IntN(5))}
		for k, m := 0, 1+r.IntN(3); k < m; k++ { // mostly valid: 1..3 abilities, sometimes unknown or repeated
			u.Skills = append(u.Skills, uint32(1+r.IntN(6)))
		}
		if r.IntN(2) == 0 {
			u.Position = &chillv1.Point{XRaw: r.Int64N(int64(sim.I(30))) - int64(sim.I(15)), ZRaw: r.Int64N(int64(sim.I(30))) - int64(sim.I(15))}
		}
		a.Units = append(a.Units, u)
	}
	return a
}

func randInput(r *rand.Rand) *chillv1.BattleInput {
	in := &chillv1.BattleInput{Seed: r.Uint64(), CombatSeed: r.Uint64(), BlueFaction: uint32(max(1, r.IntN(8)-5)), Blue: randArmy(r), Red: randArmy(r)} // faction: mostly classic, sometimes unknown
	if r.IntN(5) > 0 {
		in.SimVersion = sim.Version
	}
	if r.IntN(3) == 0 {
		in.RedSkillCount = uint32(r.IntN(32))
	}
	if r.IntN(3) == 0 {
		in.RedHealthPct = uint32(r.IntN(120))
	}
	return in
}

// The browser's decoder must agree with the server's on every input.
func TestDecodeMatchesConvert(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	valid := 0
	check := func(in *chillv1.BattleInput, raw []byte) {
		want, werr := convert.Input(in)
		got, gerr := DecodeInput(raw)
		if (werr == nil) != (gerr == nil) || errors.Is(werr, sim.ErrInvalid) != errors.Is(gerr, sim.ErrInvalid) {
			t.Fatalf("errors differ for %v:\nconvert: %v\nwire:    %v", in, werr, gerr)
		}
		if werr == nil {
			valid++
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("decoded differently for %v:\nconvert: %+v\nwire:    %+v", in, want, got)
			}
		}
	}
	for i := 0; i < 20000; i++ {
		in := randInput(r)
		raw, _ := proto.Marshal(in)
		check(in, raw)
	}
	// a typical valid input, and one per formation and class
	for f := chillv1.Formation(1); f <= 8; f++ {
		in := &chillv1.BattleInput{SimVersion: sim.Version, BlueFaction: uint32(sim.FactionClassic), Seed: 7, CombatSeed: 9, RedHealthPct: 85, RedSkillCount: 3, Blue: &chillv1.Army{Formation: f}}
		for c := chillv1.UnitClass(1); c <= 3; c++ {
			u := &chillv1.UnitSpec{UnitClass: c, Skills: []uint32{1}}
			if f == chillv1.Formation_FORMATION_CUSTOM {
				u.Position = &chillv1.Point{XRaw: int64(sim.I(int(c) * 2)), ZRaw: int64(sim.I(10))}
			}
			in.Blue.Units = append(in.Blue.Units, u)
		}
		raw, _ := proto.Marshal(in)
		check(in, raw)
	}
	if valid < 50 {
		t.Fatalf("only %d valid inputs exercised; widen the generator", valid)
	}
	// nil input and garbage
	if _, err := DecodeInput(nil); err == nil {
		t.Error("empty input must be rejected (blue has no units)")
	}
	if _, err := DecodeInput([]byte{0xff, 0xff}); err == nil {
		t.Error("garbage must be rejected")
	}
}

// proto3 parsers must accept repeated enums unpacked too.
func TestUnpackedSkills(t *testing.T) {
	unit := protowire.AppendTag(nil, 1, protowire.VarintType)
	unit = protowire.AppendVarint(unit, uint64(chillv1.UnitClass_UNIT_CLASS_WARRIOR))
	for _, s := range []uint32{4, 1} {
		unit = protowire.AppendTag(unit, 2, protowire.VarintType)
		unit = protowire.AppendVarint(unit, uint64(s))
	}
	army := protowire.AppendTag(nil, 1, protowire.VarintType)
	army = protowire.AppendVarint(army, uint64(chillv1.Formation_FORMATION_SPREAD))
	army = protowire.AppendTag(army, 2, protowire.BytesType)
	army = protowire.AppendBytes(army, unit)
	in := protowire.AppendTag(nil, 1, protowire.BytesType)
	in = protowire.AppendString(in, sim.Version)
	in = protowire.AppendTag(in, 3, protowire.BytesType)
	in = protowire.AppendBytes(in, army)
	in = protowire.AppendTag(in, 8, protowire.VarintType)
	in = protowire.AppendVarint(in, uint64(sim.FactionClassic))
	got, err := DecodeInput(in)
	var pm chillv1.BattleInput
	if e := proto.Unmarshal(in, &pm); e != nil {
		t.Fatal(e)
	}
	want, werr := convert.Input(&pm)
	if err != nil || werr != nil || !reflect.DeepEqual(got, want) || len(got.Blue.Units[0].Skills) != 2 {
		t.Errorf("unpacked skills: %v %v\n%+v\n%+v", err, werr, got, want)
	}
}

func TestEncodeOutcome(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for i := 0; i < 2000; i++ {
		o := sim.Outcome{Result: sim.Result(r.IntN(4)), Ticks: r.IntN(4000), Checksum: r.Uint64()}
		for k, n := 0, r.IntN(6); k < n; k++ {
			o.Survivors = append(o.Survivors, sim.Survivor{Index: r.IntN(10), HP: sim.Fx(r.Int64N(int64(sim.I(120))))})
		}
		var got chillv1.BattleOutcome
		if err := proto.Unmarshal(EncodeOutcome(o), &got); err != nil {
			t.Fatal(err)
		}
		if want := convert.Outcome(o); !proto.Equal(&got, want) {
			t.Fatalf("outcome %+v:\n got %v\nwant %v", o, &got, want)
		}
	}
}
