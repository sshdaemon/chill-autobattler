//go:build js && wasm

// Browser bridge: exposes the sim to JavaScript as globalThis.chillSim.
// Build: pnpm build:wasm (GOOS=js GOARCH=wasm go build -o ../web/public/sim.wasm .)
//
//	chillSim.version                       sim version string
//	chillSim.create(input: Uint8Array)     protobuf chill.v1.BattleInput; returns "" or an error message
//	chillSim.map(): Uint8Array             obstacles (Int32 LE: kind, x, z, r)
//	chillSim.step(n): Uint8Array           advance n ticks, return a snapshot (see sim/snapshot.go)
//	chillSim.outcome(): Uint8Array         protobuf chill.v1.BattleOutcome
package main

import (
	"encoding/binary"
	"syscall/js"

	"google.golang.org/protobuf/proto"

	chillv1 "chill/pb/chill/v1"
	"chill/pb/convert"
	"chill/sim"
)

var battle *sim.Battle

func bytesToJS(b []byte) js.Value {
	arr := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(arr, b)
	return arr
}

func int32sToJS(v []int32) js.Value {
	b := make([]byte, 4*len(v))
	for i, x := range v {
		binary.LittleEndian.PutUint32(b[4*i:], uint32(x))
	}
	return bytesToJS(b)
}

func create(_ js.Value, args []js.Value) any {
	raw := make([]byte, args[0].Get("length").Int())
	js.CopyBytesToGo(raw, args[0])
	var in chillv1.BattleInput
	if err := proto.Unmarshal(raw, &in); err != nil {
		return "bad input bytes: " + err.Error()
	}
	si, err := convert.Input(&in)
	if err != nil {
		return err.Error()
	}
	b, err := sim.New(si)
	if err != nil {
		return err.Error()
	}
	battle = b
	return ""
}

func step(_ js.Value, args []js.Value) any {
	if battle == nil {
		return js.Null()
	}
	for i := 0; i < args[0].Int() && battle.Result == sim.Ongoing; i++ {
		battle.Step()
	}
	return int32sToJS(battle.Snapshot())
}

func outcome(js.Value, []js.Value) any {
	if battle == nil {
		return js.Null()
	}
	raw, _ := proto.Marshal(convert.Outcome(battle.Outcome()))
	return bytesToJS(raw)
}

func register() {
	api := js.Global().Get("Object").New()
	api.Set("version", sim.Version)
	api.Set("create", js.FuncOf(create))
	api.Set("step", js.FuncOf(step))
	api.Set("outcome", js.FuncOf(outcome))
	api.Set("map", js.FuncOf(func(js.Value, []js.Value) any {
		if battle == nil {
			return js.Null()
		}
		return int32sToJS(battle.MapData())
	}))
	js.Global().Set("chillSim", api)
}

func main() {
	register()
	select {} // keep the Go runtime alive for later calls
}
