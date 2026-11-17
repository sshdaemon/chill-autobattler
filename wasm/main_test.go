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
func TestStepToOutcome(t *testing.T) {
	register()
	api := js.Global().Get("chillSim")
	w, a, m := chillv1.UnitClass_UNIT_CLASS_WARRIOR, chillv1.UnitClass_UNIT_CLASS_ARCHER, chillv1.UnitClass_UNIT_CLASS_MAGE
	in := &chillv1.BattleInput{SimVersion: sim.Version, Seed: 921, RedSkillCount: 3, Blue: &chillv1.Army{
		Formation: chillv1.Formation_FORMATION_4_1,
		Units:     []*chillv1.UnitSpec{{UnitClass: w}, {UnitClass: a}, {UnitClass: m}, {UnitClass: a}, {UnitClass: w}},
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
			Name   string `json:"name"`
			HP     float64
			Skills []struct{ Name string }
		}
		Formations []string
	}
	if err := json.Unmarshal([]byte(api.Call("data").String()), &d); err != nil {
		t.Fatal(err)
	}
	if len(d.Classes) != 3 || d.Classes[0].HP != 100 || len(d.Classes[2].Skills) != 5 || d.Formations[sim.F3x2] != "3-2" {
		t.Errorf("data %+v", d)
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
