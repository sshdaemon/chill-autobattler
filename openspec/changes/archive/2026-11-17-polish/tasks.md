# Tasks

## 1. Lean WASM

- [x] 1.1 `pb/wire`: `DecodeInput` and `EncodeOutcome` with `protowire`; verify the equivalence test against `pb/convert` (random and edge-case inputs, errors included)
- [x] 1.2 `wasm/main.go` uses `pb/wire`, with no `proto`/`chillv1` imports; verify the WASM bridge test, the goldens under WASM, and the size before/after

## 2. Web

- [x] 2.1 Split the bundle into chunks for three, React and Connect/protobuf; verify `pnpm --filter web build` has no chunk warning, and report the sizes
- [x] 2.2 Minimap (obstacles, units with fog, camera view, click to jump); verify in the browser
- [x] 2.3 Synthesized sound effects from sim events and phase changes, with mute and volume remembered; verify in the browser (the audio context starts, mute persists)

## 3. Docs

- [x] 3.1 README/CLAUDE.md (wire decoder rule, sizes, sound and minimap); verify `pnpm test`

## Workflow follow-up

- Archive the change once merged (auto-merge after green CI).
