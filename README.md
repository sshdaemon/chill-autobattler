# Chill Autobattler

Web 3D auto-battler. The design is in `docs/GAME_CONCEPT.md`. `CLAUDE.md` describes how the repo is laid out and run.

Battles run in the browser: the Go sim is compiled to WebAssembly and runs in a Web Worker. Afterwards, the Go server re-runs the same input to check the result was fair.

```bash
pnpm install
pnpm gen        # protobuf → pb/, web/src/gen
pnpm server     # Go fairness server on :8080
pnpm dev        # WASM build + web debug viewer on :5173
pnpm test       # Go (native + WASM) + TS typecheck
```

`tools/blender/bake_units.py` is the Blender unit-bake script from the prototype. It expects the Quaternius assets under `assets/quaternius/`, which aren't copied yet.
