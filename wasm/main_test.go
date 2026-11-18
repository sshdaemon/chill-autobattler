//go:build js && wasm

package main

import (
	"encoding/binary"
	"encoding/json"
	"syscall/js"
	"testing"

	"google.golang.org/protobuf/proto"

	chillv1 "chill/pb/chill/v1"
	"chill/pb/convert"
	"chill/sim"
)

// Drives the JS API the way the web worker does and checks it reaches the same
// outcome as the sim run directly.
// l1: a recruit's single level-1 ability
var l1 = []uint32{1}

func TestStepToOutcome(t *testing.T) {
	register()
	api := js.Global().Get("chillSim")
	w, a, m := chillv1.UnitClass_UNIT_CLASS_WARRIOR, chillv1.UnitClass_UNIT_CLASS_ARCHER, chillv1.UnitClass_UNIT_CLASS_MAGE
	in := &chillv1.BattleInput{SimVersion: sim.Version, Seed: 921, RedSkillCount: 3, Blue: &chillv1.Army{
		Formation: chillv1.Formation_FORMATION_4_1,
		Units:     []*chillv1.UnitSpec{{UnitClass: w, Skills: l1}, {UnitClass: a, Skills: l1}, {UnitClass: m, Skills: l1}, {UnitClass: a, Skills: l1}, {UnitClass: w, Skills: l1}},
	}}
	raw, _ := proto.Marshal(in)
	if msg := api.Call("create", bytesToJS(raw)).String(); msg != "" {
		t.Fatal(msg)
	}
	if api.Call("map").Get("length").Int()%16 != 0 {
		t.Error("map data should be whole obstacles")
	}
	for i := 0; i < sim.MaxTicks; i++ {
		snap := api.Call("step", 1)
		b := make([]byte, snap.Get("length").Int())
		js.CopyBytesToGo(b, snap)
		if result := int32(uint32(b[4]) | uint32(b[5])<<8 | uint32(b[6])<<16 | uint32(b[7])<<24); result != 0 {
			break
		}
	}
	out := api.Call("outcome")
	ob := make([]byte, out.Get("length").Int())
	js.CopyBytesToGo(ob, out)
	var got chillv1.BattleOutcome
	if err := proto.Unmarshal(ob, &got); err != nil {
		t.Fatal(err)
	}
	si, _ := convert.Input(in)
	want, _ := sim.Run(si)
	if !proto.Equal(&got, convert.Outcome(want)) {
		t.Errorf("WASM API outcome %v, want %v", &got, convert.Outcome(want))
	}

	bad := in
	bad.SimVersion = "nope"
	raw, _ = proto.Marshal(bad)
	if msg := api.Call("create", bytesToJS(raw)).String(); msg == "" {
		t.Error("create should reject an invalid input")
	}
}

func TestDataAndSlots(t *testing.T) {
	register()
	api := js.Global().Get("chillSim")
	var d struct {
		Classes []struct {
			Name       string `json:"name"`
			HP         float64
			Projectile string
			Tags       []string
			Matchup    []float64
			Skills     []struct {
				Name    string
				Pct     int
				Value   float64
				Effects []struct {
					Kind, Stat  string
					Pct         int
					Value, Secs float64
					Radius      float64
					Factor      float64
				}
			}
		}
		Formations []string
	}
	if err := json.Unmarshal([]byte(api.Call("data").String()), &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Classes) < 3 || d.Classes[0].HP != 100 || len(d.Classes[2].Skills) != 5 || d.Formations[sim.F3x2] != "3-2" {
		t.Errorf("data %+v", d)
	}
	w, ar, m := d.Classes[sim.Warrior], d.Classes[sim.Archer], d.Classes[sim.Mage]
	if w.Matchup[sim.Archer] != 1.45 || w.Matchup[sim.Mage] != 0.95 || m.Matchup[sim.Warrior] != 1.35 || ar.Matchup[sim.Mage] != 1 {
		t.Errorf("matchups %v %v %v", w.Matchup, ar.Matchup, m.Matchup)
	}
	if ar.Projectile != "arrow" || m.Projectile != "bolt" || w.Projectile != "" || len(m.Tags) != 1 || m.Tags[0] != "caster" {
		t.Errorf("traits %+v", d.Classes)
	}
	// the old headline fields keep their values
	swift, lob, blink, frost := w.Skills[1], ar.Skills[2], m.Skills[2], m.Skills[0]
	if swift.Pct != 40 || lob.Pct != 10 || blink.Value != 8 || frost.Pct != 100 || w.Skills[4].Pct != 35 {
		t.Errorf("headline pct/value changed: %+v %+v %+v %+v", swift, lob, blink, frost)
	}
	if len(swift.Effects) != 2 || swift.Effects[1].Stat != "attackSpeed" || swift.Effects[1].Pct != 80 ||
		frost.Effects[0].Kind != "slowOnHit" || frost.Effects[0].Secs != 3 || frost.Effects[0].Factor != 0.6 ||
		blink.Effects[0].Kind != "escape" || blink.Effects[0].Secs != 8 || blink.Effects[0].Radius != 2.5 {
		t.Errorf("effects %+v %+v %+v", swift.Effects, frost.Effects, blink.Effects)
	}
	classes := []sim.Class{sim.Warrior, sim.Archer, sim.Mage, sim.Archer, sim.Warrior}
	jsClasses := js.Global().Get("Array").New()
	for _, c := range classes {
		jsClasses.Call("push", int(c))
	}
	for _, f := range sim.Presets {
		v := api.Call("formationSlots", int(f), jsClasses)
		b := make([]byte, v.Get("length").Int())
		js.CopyBytesToGo(b, v)
		want := sim.FormationSlots(f, classes, 0)
		for i, p := range want {
			x := int32(binary.LittleEndian.Uint32(b[8*i:]))
			z := int32(binary.LittleEndian.Uint32(b[8*i+4:]))
			if x != int32(p.X) || z != int32(p.Z) {
				t.Errorf("%v unit %d: got %d,%d want %v", f, i, x, z, p)
			}
		}
	}
}

// units() lists every unit's class and skills, Red's generated army included.
func TestUnits(t *testing.T) {
	register()
	api := js.Global().Get("chillSim")
	w, a := chillv1.UnitClass_UNIT_CLASS_WARRIOR, chillv1.UnitClass_UNIT_CLASS_ARCHER
	in := &chillv1.BattleInput{SimVersion: sim.Version, Seed: 77, RedSkillCount: 6, Blue: &chillv1.Army{
		Formation: chillv1.Formation_FORMATION_SPREAD,
		Units:     []*chillv1.UnitSpec{{UnitClass: w, Skills: []uint32{uint32(sim.WarriorDash)}}, {UnitClass: a, Skills: l1}},
	}}
	raw, _ := proto.Marshal(in)
	if msg := api.Call("create", bytesToJS(raw)).String(); msg != "" {
		t.Fatal(msg)
	}
	var got []struct {
		Cls    int
		Skills []int
	}
	if err := json.Unmarshal([]byte(api.Call("units").String()), &got); err != nil {
		t.Fatal(err)
	}
	si, _ := convert.Input(in)
	b, _ := sim.New(si)
	if len(got) != len(b.Units) {
		t.Fatalf("%d units, battle has %d", len(got), len(b.Units))
	}
	red := 0
	for i, u := range b.Units {
		if got[i].Cls != int(u.Class) || len(got[i].Skills) != len(u.Skills) {
			t.Errorf("unit %d: got %+v, want class %d skills %v", i, got[i], u.Class, u.Skills)
			continue
		}
		for k, s := range u.Skills {
			if got[i].Skills[k] != int(s) {
				t.Errorf("unit %d skill %d: %d, want %d", i, k, got[i].Skills[k], s)
			}
		}
		if u.Team == 1 {
			red += len(u.Skills)
		}
	}
	if red != sim.ArmySize+6 || got[0].Skills[0] != int(sim.WarriorDash) {
		t.Errorf("red abilities %d (want one each plus 6), blue warrior %v", red, got[0].Skills)
	}
}
