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

Specs and changes live in `openspec/`; plan with the openspec skills before larger work. Game design: `docs/GAME_CONCEPT.md`. Roadmap (factions, draft, the new roster, refactor for parallel unit work): `docs/ROADMAP.md`; setting: `design/SETTING.md`; the roster plan under review: `design/ROSTER-PLAN.md`.

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
- A skill is a list of effects from a fixed set (`EffectKind`). Each kind is a plug-in in its own file, `sim/effect_<name>.go`: it registers itself (`registerEffect`, id from `sim/ids.go`) with a check for its numbers and the hooks it needs (`beforeAct`, `approach`, `onBlow`, `onShotLands`; see `sim/effects.go`). A unit's learned effects are folded at spawn into a list in kind-id order, and the battle calls their hooks in that order; it never checks units or skills by name. A new kind adds a file and a hook point only if none fits.
- Adding a unit means adding a roster file that registers itself (`func init() { register(xDef) }`); no shared file lists units. `ValidateRoster` (run by the tests) checks it. A new kind of behaviour means a new effect kind, which is a rule change.
- New units: use the `/new-unit <name>` skill (`.claude/skills/new-unit/`): brief in `design/units/<name>/brief.md` (template in `design/units/_template/`), reference pictures in `design/units/<name>/refs/` (git-ignored, never committed or shared), design checkpoint with the user, then roster file + look + golden case, `pnpm balance:unit <name>` (bands: 25–75% head-to-head, ±10 pts swapped into the standard army, 0..+15 pts per skill), preview screenshots, PR.
- Ids: a unit's id is its roster position, allocated in `sim/ids.go` (with effect kind ids) only by planning changes, never "next free" inside a unit's own change; ids are never reused. A unit's abilities are numbered 1..5 in the order its roster entry lists them (named consts in its own file); numbers mean nothing across units. No proto change needed (`UnitSpec.skills` and `SkillPick.skill` are ability numbers; the unit enum is open; the roster is the check). A unit outside Red's pool (`InRedPool: false`) needs no version bump; changing the pool does.
- Restart `pnpm start` after adding a unit: the Go server is compiled with the roster.
- Changing `InRedPool` changes Red's generated armies, so it needs a `sim.Version` bump.
- `wasm data()` publishes the roster; the web client reads units and skills only from there.

Golden replays (`sim/testdata/golden/`, one file per case; every roster entry gets a "<unit> skills" case automatically):
- Any rule change that alters outcomes needs `sim.Version` bumped, then `go test ./sim -run TestGoldenReplays -update`.
- The update refuses to run without the version bump.

The snapshot layout is defined in `sim/snapshot.go` and decoded in `web/src/sim/snapshot.ts`. Keep them in sync; `TestSnapshotLayout` pins it.

Unit lab (dev only, `http://localhost:5173/?lab` or `?lab=<unit>`, linked from the dev HUD; tabs Unit / Animate / Walk & tune under an always-visible playback bar with frame stepping, space and arrow keys): any unit type, recipe or lineup, team, build override and variety seed, with an orbit camera, any clip played, paused, looped, slowed or scrubbed, the walk–jog blend on a moving floor with a foot-slide readout, skill animations and effects, and effect presets (`web/src/ui/UnitLab.tsx`). Unit preview (dev only): `http://localhost:5173/?preview=<unit>` shows a unit type's recipes on both teams, steps through its animations, skills, effects and portraits (`window.previewSteps`, `previewShot(step)`, `previewSave()` writes `design/units/<unit>/preview/*.png` through a dev-server endpoint in `web/preview-shots.mjs`). Unit UI text (label, icon, blurb) comes from the look's `ui`; nothing in the UI assumes three unit types.

Fog of war comes from the sim: `Battle.SeenCells` (exported by `wasm fog()`) and the snapshot's `UFSeen` flag. Never re-implement vision in TS. The camera is `scene/RtsCamera.tsx`; portraits are `scene/portraits.ts` and `ui/Portraits.tsx`.

Unit looks (`web/src/units/looks/<unit>.json`, one file per unit, plus `_shared.json` for builds and clips several units use; presentation only, matched to the roster by file name): each unit's models, animation set (idle, move, attack variants, hit, death, cheer) and, per skill, an action clip, an attack clip, a lean-back pitch and an effect preset. Adding a unit = a roster file + a look entry; `wasm/looks_test.go` (pnpm test:wasm) fails when they drift, when a named clip or preset is missing, or when the clip bundles (`web/public/assets/clips/<look file>.json`) exceed their budget (1.5 MB raw, 400 KB gzipped, together). Clips come from Quaternius's Universal Animation Library 1 and 2 (CC0, same rig as our characters): unzip them into `~/fun/assets-src/ual1` and `ual2`, list the clips in the unit's look file, or `_shared.json` if several units use them (`from` joins a combo strike with its recovery, `hit` is when the blow lands), then run `python3 -I tools/assets/ual_clips.py`.

Units are rendered by `web/src/scene/models.ts` + `Units.tsx`: each unit is assembled at runtime from its look's recipe (outfit parts, head, hair, colour variant, items, build; the unit's look file) out of the parts library in `web/public/assets/parts/` (`parts.json` lists parts, items with their mount bone and pose, and outfit materials with colour variants and cloth masks). Parts sharing a material are merged into one skinned mesh; the first part's skeleton is the unit's. Builds are data (`looks/_shared.json` `builds`): girth pushes the surface out per bone in the vertex shader (before skinning, so animation can't shear it), plus shoulder width, head size and height. Team colour is applied on outfit cloth through the mask. Animated from snapshot flags through the look (attacks are timed so the blow lands at `data().hitAt` of the sim's swing; moving units blend the look's walk and move clips, kept in step, so the stride covers their ground speed measured from snapshots: clip `speed` in the looks is m/s at 1× measured on our characters, see `locomotion()` in `Units.tsx`). The client learns every unit's skills (Red's too) from `wasm units()` at battle start (`store.units`). Dev: `window.animLog` lists every clip start. Each unit gets its own cloned materials (team rim light, team cloth, hair and tint variety seeded by team, class and nth, death fade); aim and flinch are applied to `spine_02`/`spine_01` after the mixer update. Re-bake the parts with `blender --background --factory-startup --python tools/blender/bake_parts.py` from the CC0 packs unzipped in `~/fun/assets-src/{outfits,basechars,items}` (Quaternius outfits, base characters and props, KayKit Adventurers); items (bone, mount pose, pack source, modelled or cloth) are data in `tools/blender/items/*.json` (one file per unit, `_shared.json` for common kit); the bake edits glTF directly (rest pose untouched), cuts the base bodies to the head, quantises vertex data, shrinks textures and derives the cloth masks. Budgets (`wasm/looks_test.go`): parts ≤ 10 MB, a default army's parts ≤ 3 MB.

The battlefield is `scene/Arena.tsx`: `Ground`, `Water` (all ponds as one shader surface from the sim's circles), `Nature` (rocks and trees placed from the sim's obstacles, variant picked by position; wall and scenery) and `Grass` (High only). Models and ground textures come from `tools/assets/pack_nature.py`; the detailed trees from `web/scripts/bake-trees.mjs` (EZ-Tree, dev-only). Obstacle models must keep matching the sim's footprints.

Combat effects: `scene/effects.ts` turns sim events into named presets (`scene/fx/*.json`, `_shared.json` plus one file per unit: sparks, bursts, dust, rings, shake; a skill's look can name its own preset for the events it causes) and unit flags into particles (`scene/Particles.tsx`, pooled GPU billboards) and camera shake (`scene/shake.ts`); the store calls it with each snapshot. Effects skip units hidden by fog. Dev helpers: `window.fxDemo()` fires every cue, `window.fxSlow(k)` slows particles, `window.benchFrame(n)` times a frame.
