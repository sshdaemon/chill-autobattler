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
// wireClass is a unit's value on the wire: its class id + 1.
func wireClass(c sim.Class) chillv1.UnitClass { return chillv1.UnitClass(c + 1) }

// l1: a recruit's single level-1 ability
var l1 = []uint32{1}

func TestStepToOutcome(t *testing.T) {
	register()
	api := js.Global().Get("chillSim")
	w, a, m := wireClass(sim.Druzhynnyk), wireClass(sim.Strilets), wireClass(sim.Huslyar)
	in := &chillv1.BattleInput{SimVersion: sim.Version, BlueFaction: uint32(sim.FactionRus), Seed: 921, RedSkillCount: 3, Blue: &chillv1.Army{
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
		Classes []*struct {
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
	druz, str, hus := d.Classes[sim.Druzhynnyk], d.Classes[sim.Strilets], d.Classes[sim.Huslyar]
	volva, siph := d.Classes[sim.Volva], d.Classes[sim.Siphonarios]
	if len(d.Classes) != len(sim.Roster) || druz.HP != 100 || len(hus.Skills) != 5 || d.Formations[sim.F3x2] != "3-2" {
		t.Errorf("data %+v", d)
	}
	for c := sim.Class(0); c <= 2; c++ { // retired: Classic's warrior, archer and mage
		if d.Classes[c] != nil || druz.Matchup[c] != 0 {
			t.Errorf("retired class %d: %+v, matchup %v", c, d.Classes[c], druz.Matchup[c])
		}
	}
	if druz.Matchup[sim.Strilets] != 1.4 || volva.Matchup[sim.Druzhynnyk] != 1.3 || str.Matchup[sim.Huslyar] != 1 {
		t.Errorf("matchups %v %v %v", druz.Matchup, volva.Matchup, str.Matchup)
	}
	if str.Projectile != "arrow" || volva.Projectile != "bolt" || druz.Projectile != "" || len(hus.Tags) != 1 || hus.Tags[0] != "caster" {
		t.Errorf("traits %+v %+v %+v", str, volva, druz)
	}
	// skills' headline fields and effects
	axe, slip, fear := druz.Skills[4], volva.Skills[3], siph.Skills[4]
	if axe.Pct != 80 || slip.Value != 6 || fear.Pct != 100 {
		t.Errorf("headline pct/value: %+v %+v %+v", axe, slip, fear)
	}
	if len(axe.Effects) != 1 || axe.Effects[0].Stat != "attackSpeed" || axe.Effects[0].Pct != 80 ||
		fear.Effects[0].Kind != "slowOnHit" || fear.Effects[0].Secs != 2 || fear.Effects[0].Factor != 0.6 ||
		slip.Effects[0].Kind != "escape" || slip.Effects[0].Secs != 8 || slip.Effects[0].Radius != 2.5 {
		t.Errorf("effects %+v %+v %+v", axe.Effects, fear.Effects, slip.Effects)
	}
	classes := []sim.Class{sim.Druzhynnyk, sim.Strilets, sim.Huslyar, sim.Strilets, sim.Druzhynnyk}
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
	w, a := wireClass(sim.Druzhynnyk), wireClass(sim.Strilets)
	in := &chillv1.BattleInput{SimVersion: sim.Version, BlueFaction: uint32(sim.FactionRus), Seed: 77, RedSkillCount: 6, Blue: &chillv1.Army{
		Formation: chillv1.Formation_FORMATION_SPREAD,
		Units:     []*chillv1.UnitSpec{{UnitClass: w, Skills: []uint32{uint32(sim.DruzhynnykBracedSpear)}}, {UnitClass: a, Skills: l1}},
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
	if red != sim.ArmySize+6 || got[0].Skills[0] != int(sim.DruzhynnykBracedSpear) {
		t.Errorf("red abilities %d (want one each plus 6), blue druzhynnyk %v", red, got[0].Skills)
	}
}
