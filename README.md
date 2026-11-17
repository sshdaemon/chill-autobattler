# Chill Autobattler

Web 3D auto-battler. See `docs/GAME_CONCEPT.md` for the design and `CLAUDE.md` for how the repo is laid out and run.

```bash
pnpm install
pnpm gen        # protobuf → server/gen, web/src/gen
pnpm server     # Go fairness server on :8080
pnpm dev        # WASM build + web app on :5173
pnpm test
```

`tools/blender/bake_units.py` is the Blender unit-bake script from the prototype; it expects the Quaternius assets under `assets/quaternius/` (not copied yet).
