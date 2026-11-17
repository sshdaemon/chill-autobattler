# chill-autobattler

Monorepo:
- `sim/`: Go, the deterministic battle rules (map, pathfinding, combat, skills, armies).
- `pb/`: generated protobuf/Connect Go code plus `pb/convert` (wire ⇄ sim types). Shared by `wasm/` and `server/`.
- `wasm/`: the browser build of the sim, exposed as `globalThis.chillSim` (create / map / step / outcome).
- `server/`: Go, Connect RPC.
  - `server/battle` issues battles: `StartBattle` (map seed), `LockDeployment` (stores the army, reveals the combat seed), `VerifyBattle` (re-runs the stored input; the first verdict is final).
  - `server/store` is SQLite (modernc, pure Go) with sqlc. Regenerate with `go generate ./server/store`.
- `web/`: Vite, React and react-three-fiber. The sim runs in a Web Worker at 20 Hz; the scene interpolates snapshots.
- `proto/`: buf schema. Run `pnpm gen` to write `pb/` and `web/src/gen/`.

Specs and changes live in `openspec/`; plan with the openspec skills before larger work. Game design: `docs/GAME_CONCEPT.md`.

Commands:
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

The snapshot layout is defined in `sim/snapshot.go` and decoded in `web/src/sim/snapshot.ts`. Keep them in sync.
