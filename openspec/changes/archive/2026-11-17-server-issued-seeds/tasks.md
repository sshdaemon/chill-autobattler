# Tasks

## 1. Sim: separate combat seed

- [x] 1.1 Add `Input.CombatSeed`, derive `StreamCombat` from it, and bump `sim.Version` to 0.3.0; verify a new test shows two combat seeds give an identical map and start positions but different outcomes
- [x] 1.2 Give each golden case a fixed combat seed and regenerate `sim/testdata/golden.json`; verify `pnpm test:go` and `pnpm test:wasm` both pass with the new file

## 2. Wire format

- [x] 2.1 Add `combat_seed` to `BattleInput`; add `StartBattle`, `LockDeployment` and a `VerifyBattle` that takes `battle_id` + `claimed`; run `pnpm gen`; map `combat_seed` in `pb/convert`; verify `pnpm lint:proto` and the convert tests pass

## 3. Server storage

- [x] 3.1 Add `server/store`: embedded `schema.sql`, `queries.sql`, sqlc config and generated code (sqlc as a `go tool`), and an `Open(path)` that applies the schema; verify a store test round-trips a battle, including seeds ≥ 2⁶³, and reopens the file

## 4. Battle service

- [x] 4.1 Implement `StartBattle` (crypto/rand seeds, version check, deploy deadline); verify tests for a new battle, an unknown version, and seeds that differ between battles
- [x] 4.2 Implement `LockDeployment` with a conditional update; verify tests for lock and reveal, second lock, too late (injected clock) and invalid army
- [x] 4.3 Implement `VerifyBattle` from the stored input with verify-once semantics; verify tests for an honest claim, a re-rolled claim, unknown/unlocked battles, a second claim, expiry, and a restart between lock and verify (reopen the DB)
- [x] 4.4 Wire the service into `cmd/server` with `CHILL_DB` (default `chill.db`) and remove `server/verify`; add `*.db` handling to `.gitignore`; verify the server starts and `/healthz` answers

## 5. Web viewer flow

- [x] 5.1 Add the store actions `newMap`, `setFormation` (preview with combat seed 0), `fight` (lock, then play with the real combat seed) and `verify(battle_id)`, and the HUD buttons for them; verify `pnpm --filter web typecheck`
- [x] 5.2 Check in the browser that New map shows a deployment preview, changing the formation moves Blue, Fight plays the battle and Verify says fair

## 6. Docs

- [x] 6.1 Update README (flow, anti-cheat, storage, sqlc) and CLAUDE.md; verify `pnpm test` passes

## Workflow follow-up

- Archive the change once merged.
