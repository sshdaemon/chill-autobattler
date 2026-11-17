//go:build js && wasm

package main

import (
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
