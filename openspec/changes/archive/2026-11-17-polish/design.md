# Design

## Context

`wasm/main.go` decodes `BattleInput` with `proto.Unmarshal` and `convert.Input`, and encodes `BattleOutcome` with `proto.Marshal`. That links the protobuf runtime and reflection, about 6 MB of the 8 MB. The web build is one chunk.

## Goals / Non-Goals

**Goals:**
- a much smaller WASM with identical wire behaviour;
- a split bundle;
- a minimap and sound with no new assets.

**Non-Goals:** rule, server or wire changes; music.

## Decisions

**`pb/wire`: hand-decoding with `protowire`.**
- It walks the tags of `BattleInput`, `Army`, `UnitSpec` and `Point`, and fills `sim.Input` directly with the same enum offsets as `pb/convert`.
- `EncodeOutcome` writes `BattleOutcome` field by field.
- Unknown fields are skipped, as proto3 parsers do.
- A test in `pb/wire` checks **equivalence**, over generated inputs: every field set and unset, custom positions, both armies, all enums and Red health. For each:
  - `wire.DecodeInput(proto.Marshal(x))` equals `convert.Input(x)`, errors included;
  - `proto.Unmarshal(wire.EncodeOutcome(o))` equals `convert.Outcome(o)`.

  The browser and the server can't diverge without this failing.
- *Alternatives:*
  - JSON on the browser path: still reflection-heavy, and a second format.
  - TinyGo: different compiler semantics; it would need its own determinism proof.

**Bundle.** `build.rolldownOptions.output.advancedChunks` (Vite 8 / Rolldown) or `manualChunks`, with groups for `three` (and `@react-three/*`), `react`, and `@connectrpc` + `@bufbuild`.

**Minimap.**
- A 160×160 `<canvas>`, redrawn at 10 Hz from the store: obstacles (static, cached on an offscreen canvas), units (Blue always; Red only with `UF.seen` while the fog is on) and the camera's look-at point with its approximate view rectangle.
- The camera exposes its target through the store (`camTarget`), so the minimap and `RtsCamera` share it.
- A click sends `centerOn` at a point (the store gains `centerAt`).

**Sound** (`sound.ts`).
- One `AudioContext`, created on the first user gesture, as browsers require.
- Each effect is a tiny synth recipe: an oscillator and/or filtered noise burst with an envelope.
- Events come from the snapshot stream already used for floating numbers (hit, death, dash, blink, frost, shot). Effects are rate-limited: at most one of each kind per 40 ms, so 4× playback doesn't machine-gun.
- Phase changes trigger the "Fight!" and result stings.
- Mute and volume are kept in localStorage, wrapped in try/catch.

**How determinism is preserved and tested.**
- Nothing in `sim/` changes; the golden replays run unchanged, natively and under WASM.
- The equivalence test pins the browser decoder to the server's. The existing WASM bridge test (create/step/outcome compared to `sim.Run`) still runs against the new decoder.

## Risks / Trade-offs

- [The hand decoder must follow the proto schema] → The equivalence test fails on any drift, and `proto/` changes are rare.
- [Autoplay rules] → The audio context starts on the first click; until then, sounds are skipped silently.
