# chill-autobattler

Monorepo:
- `sim/`: Go, the deterministic battle rules (map, pathfinding, combat, skills, armies).
- `pb/`: generated protobuf/Connect Go code plus `pb/convert` (wire ⇄ sim types). Shared by `wasm/` and `server/`.
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

sim/ rules (enforced by `TestNoNondeterministicImports`):
- Fixed-point `sim.Fx` only.
- No floats, map iteration, time, goroutines, `math` or `math/rand`.
- Randomness only from the battle's `Sub(seed, tag)` streams.

Golden replays (`sim/testdata/golden.json`):
- Any rule change that alters outcomes needs `sim.Version` bumped, then `go test ./sim -run TestGoldenReplays -update`.
- The update refuses to run without the version bump.

The snapshot layout is defined in `sim/snapshot.go` and decoded in `web/src/sim/snapshot.ts`. Keep them in sync; `TestSnapshotLayout` pins it.

Units are rendered by `web/src/scene/models.ts` + `Units.tsx`: baked Quaternius looks per class and team, animated from snapshot flags. Re-bake with `tools/blender/bake_units.py`, then pack with `tools/assets/pack_units.py`.
