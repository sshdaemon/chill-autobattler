# Proposal

## Why

The game is feature-complete for a first playtest, but rough at the edges:
- **Downloads:** `sim.wasm` is 8.0 MB (2.0 MB gzipped), mostly the protobuf runtime, used to decode one message and encode another. The web app is a single 1.36 MB JS chunk, and Vite warns about it.
- **Overview:** there's no overview of the field beyond the camera view.
- **Sound:** battles are silent.

## What Changes

- **Lean WASM.** A small `pb/wire` package decodes `BattleInput` and encodes `BattleOutcome` with only `protowire`, the low-level wire functions, into the sim's own types. The browser build no longer links the protobuf runtime or the generated types. The server keeps using the full runtime, and a test proves both paths decode identically.
- **Bundle splitting.** three.js, React and the Connect/protobuf client become separate cacheable chunks.
- **Minimap.** A small top-down map in a corner shows:
  - the arena, obstacles and water;
  - Blue's units and the Red units Blue can see (fog of war applies);
  - the camera's view.

  Clicking it moves the camera there.
- **Sound effects.** They are synthesized with Web Audio at runtime, so there are no audio files and no licensing:
  - sword hits, arrow release, magic bolt, death, Blink and Dash;
  - a "Fight!" sting, and victory and defeat jingles.

  A mute toggle and the volume are remembered per browser. Sounds are driven by the sim's events, as the floating numbers are.

## Out of scope

- Music, voice, recorded samples.
- TinyGo or other compilers.
- Any rule, server or wire-format change: the bytes on the wire stay the same.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-presentation`: adds the minimap and sound effects.

## Impact

- `pb/wire` (new) and `wasm/main.go` (use it; drop the `proto` and `chillv1` imports).
- `web/vite.config.ts` (manual chunks); `web/src/ui/Minimap.tsx`, `web/src/sound.ts` (new); the dev panel gets the mute toggle.
