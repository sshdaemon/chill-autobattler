//go:build js && wasm

// Browser bridge: exposes the sim to JavaScript as globalThis.chillSim.
// Build: GOOS=js GOARCH=wasm go build -o ../web/public/sim.wasm .
package main

import (
	"strconv"
	"syscall/js"

	"chill/sim"
)

func main() {
	api := js.Global().Get("Object").New()
	api.Set("version", sim.Version)
	// run(seed: string, maxTicks: number) -> {winner, ticks, checksum: string}
	// Seeds and checksums are uint64, so they cross the boundary as decimal strings.
	api.Set("run", js.FuncOf(func(_ js.Value, args []js.Value) any {
		seed, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return js.Global().Get("Error").New("bad seed: " + err.Error())
		}
		r := sim.Run(seed, args[1].Int())
		return map[string]any{"winner": r.Winner, "ticks": r.Ticks, "checksum": strconv.FormatUint(r.Checksum, 10)}
	}))
	js.Global().Set("chillSim", api)
	select {} // keep the Go runtime alive for later calls
}
