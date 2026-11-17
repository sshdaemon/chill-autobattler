# chill-autobattler

Monorepo: `sim/` (Go, deterministic battle logic) → `wasm/` (browser build) and `server/` (Go, Connect RPC, re-runs battles to check fairness); `web/` (Vite + React + react-three-fiber); `proto/` (buf). Specs and changes live in `openspec/` — plan with the openspec skills before larger work. Game design: `docs/GAME_CONCEPT.md`.

- Go is installed at `~/sdk/go/bin` (not on PATH by default): `export PATH=$HOME/sdk/go/bin:$PATH`.
- `pnpm dev` (builds WASM, starts Vite on :5173), `pnpm server` (:8080), `pnpm gen` (buf), `pnpm test` (Go native + Go under WASM + TS typecheck).
- sim/ rules: fixed-point `sim.Fx` only, no floats/maps-order/time/goroutines, randomness only from the battle RNG. Any rule change bumps `sim.Version` and the golden table in `sim/sim_test.go`.
