//go:build js && wasm

// Browser bridge: exposes the sim to JavaScript as globalThis.chillSim.
// Build: pnpm build:wasm (GOOS=js GOARCH=wasm go build -o ../web/public/sim.wasm .)
//
//	chillSim.version                       sim version string
//	chillSim.create(input: Uint8Array)     protobuf chill.v1.BattleInput; returns "" or an error message
//	chillSim.map(): Uint8Array             obstacles (Int32 LE: kind, x, z, r)
//	chillSim.step(n): Uint8Array           advance n ticks, return a snapshot (see sim/snapshot.go)
//	chillSim.outcome(): Uint8Array         protobuf chill.v1.BattleOutcome
//	chillSim.fog(): Uint8Array            Blue's visible grid cells (56×56, row = z cell), for fog of war
//	chillSim.data(): string                balance data as JSON (stats, matchups, skills), for the UI
//	chillSim.formationSlots(f, classes)    Blue's preset positions (Int32 LE x, z per unit, raw fixed-point);
//	                                       f is a sim.Formation (1..7), classes are sim.Class values (0..2)
package main

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"strings"
	"syscall/js"

	"chill/pb/wire" // lean decoder: keeps the protobuf runtime out of the browser build
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
	si, err := wire.DecodeInput(raw)
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
	return bytesToJS(wire.EncodeOutcome(battle.Outcome()))
}

// Display copies of the balance tables, so the UI never duplicates numbers
// from sim/data.go. Fixed-point becomes decimal only here, at the edge.
// fx converts for display, rounded to 3 decimals (3.2 m/s is 3.19999 in Q48.16).
func fx(v sim.Fx) float64 { return math.Round(float64(v)/float64(sim.One)*1000) / 1000 }

func dataJSON() string {
	type effect struct {
		Kind   string  `json:"kind"`
		Stat   string  `json:"stat,omitempty"`
		Pct    int     `json:"pct,omitempty"`
		Value  float64 `json:"value,omitempty"`  // metres
		Radius float64 `json:"radius,omitempty"` // metres
		Speed  float64 `json:"speed,omitempty"`  // metres per second
		Factor float64 `json:"factor,omitempty"`
		Secs   float64 `json:"secs,omitempty"` // cooldown or duration
	}
	type skill struct {
		ID      int      `json:"id"`
		Name    string   `json:"name"`
		Pct     int      `json:"pct"`   // the headline effect's %
		Value   float64  `json:"value"` // the headline effect's value
		Effects []effect `json:"effects"`
	}
	type bonus struct {
		Tag  string  `json:"tag"`
		Mult float64 `json:"mult"`
	}
	type class struct {
		Name       string    `json:"name"`
		HP         float64   `json:"hp"`
		DmgLo      float64   `json:"dmgLo"`
		DmgHi      float64   `json:"dmgHi"`
		Cooldown   float64   `json:"cooldown"` // seconds
		Range      float64   `json:"range"`
		Speed      float64   `json:"speed"`
		Sight      float64   `json:"sight"`
		Swing      float64   `json:"swing"` // attack animation seconds (the blow lands at 45%)
		Ranged     bool      `json:"ranged"`
		Kite       bool      `json:"kite"`
		Projectile string    `json:"projectile,omitempty"` // "arrow" or "bolt"
		Tags       []string  `json:"tags"`
		Bonus      []bonus   `json:"bonus"`   // multiplier against tags; the first the victim has wins
		Matchup    []float64 `json:"matchup"` // damage multiplier vs each roster entry
		Skills     []skill   `json:"skills"`
	}
	var out struct {
		Version    string     `json:"version"`
		Classes    []class    `json:"classes"`    // the unit roster, index = sim.Class
		Formations []string   `json:"formations"` // index = sim.Formation
		ArmySize   int        `json:"armySize"`
		DeployZone [4]float64 `json:"deployZone"` // Blue: xMin, xMax, zMin, zMax
	}
	tagNames := func(t sim.Tag) []string {
		names := []string{}
		for _, d := range sim.Tags {
			if t&d.Tag != 0 {
				names = append(names, d.Name)
			}
		}
		return names
	}
	out.Version, out.ArmySize = sim.Version, sim.ArmySize
	for i := range sim.Roster {
		d := &sim.Roster[i]
		st := d.Stats
		cl := class{Name: d.Name, HP: fx(st.HP), DmgLo: fx(st.DmgLo), DmgHi: fx(st.DmgHi),
			Cooldown: float64(st.Cooldown) / sim.TicksPerSecond, Range: fx(st.Range), Speed: fx(st.Speed), Sight: fx(st.Sight),
			Swing: float64(st.SwingTicks) / sim.TicksPerSecond, Ranged: st.Ranged, Kite: st.Kite,
			Projectile: [...]string{"", "arrow", "bolt"}[d.Projectile], Tags: tagNames(d.Tags), Bonus: []bonus{}}
		for _, b := range d.Bonus {
			cl.Bonus = append(cl.Bonus, bonus{strings.Join(tagNames(b.Tag), "|"), fx(b.Mult)})
		}
		for v := range sim.Roster {
			cl.Matchup = append(cl.Matchup, fx(sim.MatchupOf(d, &sim.Roster[v])))
		}
		for _, sd := range d.Skills {
			sk := skill{ID: int(sd.Skill), Name: sd.Name, Pct: sd.Main().Pct, Value: fx(sd.Main().Value)}
			for _, e := range sd.Effects {
				ef := effect{Kind: e.Kind.String(), Pct: e.Pct, Value: fx(e.Value), Radius: fx(e.Radius), Speed: fx(e.Speed),
					Factor: fx(e.Factor), Secs: float64(e.Ticks) / sim.TicksPerSecond}
				if e.Kind == sim.StatBonus {
					ef.Stat = e.Stat.String()
				}
				sk.Effects = append(sk.Effects, ef)
			}
			cl.Skills = append(cl.Skills, sk)
		}
		out.Classes = append(out.Classes, cl)
	}
	for f := sim.Custom; f <= sim.F5x0; f++ {
		out.Formations = append(out.Formations, f.String())
	}
	z := sim.Zones[0]
	out.DeployZone = [4]float64{fx(z.XMin), fx(z.XMax), fx(z.ZMin), fx(z.ZMax)}
	b, _ := json.Marshal(out)
	return string(b)
}

func formationSlots(_ js.Value, args []js.Value) any {
	f := sim.Formation(args[0].Int())
	if f == sim.Custom || f > sim.F5x0 {
		return js.Null()
	}
	n := args[1].Length()
	classes := make([]sim.Class, n)
	for i := range classes {
		classes[i] = sim.Class(args[1].Index(i).Int())
	}
	var out []int32
	for _, p := range sim.FormationSlots(f, classes, 0) {
		out = append(out, int32(p.X), int32(p.Z))
	}
	return int32sToJS(out)
}

func register() {
	api := js.Global().Get("Object").New()
	api.Set("version", sim.Version)
	api.Set("create", js.FuncOf(create))
	api.Set("step", js.FuncOf(step))
	api.Set("outcome", js.FuncOf(outcome))
	api.Set("data", js.FuncOf(func(js.Value, []js.Value) any { return dataJSON() }))
	api.Set("fog", js.FuncOf(func(js.Value, []js.Value) any {
		if battle == nil {
			return js.Null()
		}
		seen := battle.SeenCells(0)
		b := make([]byte, len(seen))
		for i, s := range seen {
			if s {
				b[i] = 1
			}
		}
		return bytesToJS(b)
	}))
	api.Set("formationSlots", js.FuncOf(formationSlots))
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
