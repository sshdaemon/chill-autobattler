# chill-autobattler

Monorepo:
- `sim/`: Go, the deterministic battle rules (map, pathfinding, combat, skills, armies).
- `pb/`: generated protobuf/Connect Go code, `pb/convert` (generated types ⇄ sim, used by the server) and `pb/wire` (the same mapping straight from bytes with `protowire`, used by `wasm/`, so the browser doesn't link the protobuf runtime).
  - `pb/wire` must match `pb/convert`; `TestDecodeMatchesConvert` checks it.
  - After any `proto/` change, update `pb/wire` too.
- `wasm/`: the browser build of the sim, exposed as `globalThis.chillSim` (create / map / step / outcome).
- `server/`: Go, Connect RPC.
  - `server/auth`: anonymous player tokens. Every RPC except `CreatePlayer` needs `Authorization: Bearer`.
  - `server/run`: runs, with veterans, skill picks, Red scaling, scores and the scoreboard. The server is the authority on all of it.
  - `server/battle`: battles of a run. `StartBattle` (map seed), `LockDeployment` (veterans must lead the army; reveals the combat seed), `VerifyBattle` (re-runs the stored input, updates the run in the same transaction; the first verdict is final).
  - `server/store`: SQLite (modernc, pure Go) with sqlc. Migrations are numbered files in `migrations/`, tracked with `user_version`; never edit an applied one. Regenerate with `go generate ./server/store`.
- `web/`: Vite, React and react-three-fiber.
  - The sim runs in a Web Worker at 20 Hz, and the scene interpolates snapshots.
  - The match flow (picker → deploy → countdown → fight → result) is a phase machine in `web/src/store.ts`; its UI is in `web/src/ui/`.
  - Balance numbers and formation slots come from the WASM build (`data()`, `formationSlots()`); never copy them into TS.
- `proto/`: buf schema. Run `pnpm gen` to write `pb/` and `web/src/gen/`.

Specs and changes live in `openspec/`; plan with the openspec skills before larger work. Game design: `docs/GAME_CONCEPT.md`.

Commands:
- `pnpm start`: the whole game locally. It builds the WASM, then runs the server on :8080 and Vite on :5173; Ctrl-C stops both.
- `pnpm dev`: builds the WASM, then starts Vite on :5173.
- `pnpm server`: the Go server on :8080. Its DB file is `CHILL_DB`, default `server/chill.db`.
- `pnpm test`: Go tests natively, sim and bridge tests under WASM, and a TS typecheck.
- `pnpm balance`: the balance lab (`tools/balance`). It plays ~160k seeded battles and writes `docs/balance-report.md`. Run it before and after any balance change.

sim/ rules (enforced by `TestNoNondeterministicImports`):
- Fixed-point `sim.Fx` only.
- No floats, map iteration, time, goroutines, `math` or `math/rand`.
- Randomness only from the battle's `Sub(seed, tag)` streams.

Unit roster (`sim/roster.go`, one `sim/unit_*.go` per unit type):
- A unit type is data: stats, traits (ranged, kite, projectile), matchup tags and bonuses, and five skills.
- A skill is a list of effects from a fixed set (`EffectKind`). The battle code reads a loadout folded at spawn and never checks units or skills by name.
- Adding a unit means adding a roster file; `ValidateRoster` (run by the tests) checks it. A new kind of behaviour means a new effect kind, which is a rule change.
- Changing `InRedPool` changes Red's generated armies, so it needs a `sim.Version` bump.
- `wasm data()` publishes the roster; the web client reads units and skills only from there.

Golden replays (`sim/testdata/golden.json`):
- Any rule change that alters outcomes needs `sim.Version` bumped, then `go test ./sim -run TestGoldenReplays -update`.
- The update refuses to run without the version bump.

The snapshot layout is defined in `sim/snapshot.go` and decoded in `web/src/sim/snapshot.ts`. Keep them in sync; `TestSnapshotLayout` pins it.

Fog of war comes from the sim: `Battle.SeenCells` (exported by `wasm fog()`) and the snapshot's `UFSeen` flag. Never re-implement vision in TS. The camera is `scene/RtsCamera.tsx`; portraits are `scene/portraits.ts` and `ui/Portraits.tsx`.

Unit looks (`web/src/units/looks.json`, presentation only, matched to the roster by name): each unit's models, animation set (idle, move, attack variants, hit, death, cheer) and, per skill, an action clip, an attack clip, a lean-back pitch and an effect preset. Adding a unit = a roster file + a look entry; `wasm/looks_test.go` (pnpm test:wasm) fails when they drift, when a named clip or preset is missing, or when `clips.json` exceeds its budget (1.5 MB raw, 400 KB gzipped). Clips come from Quaternius's Universal Animation Library 1 and 2 (CC0, same rig as our characters): unzip them into `~/fun/assets-src/ual1` and `ual2`, list the clips in `looks.json` (`from` joins a combo strike with its recovery, `hit` is when the blow lands), then run `python3 -I tools/assets/ual_clips.py`.

Units are rendered by `web/src/scene/models.ts` + `Units.tsx`: each unit is assembled at runtime from its look's recipe (outfit parts, head, hair, colour variant, items, build; `looks.json`) out of the parts library in `web/public/assets/parts/` (`parts.json` lists parts, items with their mount bone and pose, and outfit materials with colour variants and cloth masks). Parts sharing a material are merged into one skinned mesh; the first part's skeleton is the unit's. Builds are data (`looks.json` `builds`): girth pushes the surface out per bone in the vertex shader (before skinning, so animation can't shear it), plus shoulder width, head size and height. Team colour is applied on outfit cloth through the mask. Animated from snapshot flags through the look (attacks are timed so the blow lands at `data().hitAt` of the sim's swing). The client learns every unit's skills (Red's too) from `wasm units()` at battle start (`store.units`). Dev: `window.animLog` lists every clip start. Each unit gets its own cloned materials (team rim light, team cloth, hair and tint variety seeded by team, class and nth, death fade); aim and flinch are applied to `spine_02`/`spine_01` after the mixer update. Re-bake the parts with `blender --background --factory-startup --python tools/blender/bake_parts.py` from the CC0 packs unzipped in `~/fun/assets-src/{outfits,basechars,items}` (Quaternius outfits, base characters and props, KayKit Adventurers); the bake edits glTF directly (rest pose untouched), cuts the base bodies to the head, quantises vertex data, shrinks textures and derives the cloth masks. Budgets (`wasm/looks_test.go`): parts ≤ 10 MB, a default army's parts ≤ 3 MB.

The battlefield is `scene/Arena.tsx`: `Ground`, `Water` (all ponds as one shader surface from the sim's circles), `Nature` (rocks and trees placed from the sim's obstacles, variant picked by position; wall and scenery) and `Grass` (High only). Models and ground textures come from `tools/assets/pack_nature.py`; the detailed trees from `web/scripts/bake-trees.mjs` (EZ-Tree, dev-only). Obstacle models must keep matching the sim's footprints.

Combat effects: `scene/effects.ts` turns sim events into named presets (`scene/fx.json`: sparks, bursts, dust, rings, shake; a skill's look can name its own preset for the events it causes) and unit flags into particles (`scene/Particles.tsx`, pooled GPU billboards) and camera shake (`scene/shake.ts`); the store calls it with each snapshot. Effects skip units hidden by fog. Dev helpers: `window.fxDemo()` fires every cue, `window.fxSlow(k)` slows particles, `window.benchFrame(n)` times a frame.
