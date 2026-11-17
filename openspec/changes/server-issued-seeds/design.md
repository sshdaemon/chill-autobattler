# Design

## Context

After `port-battle-sim`:
- `sim.Input` has one `Seed`, and every stream is derived from it with `Sub(seed, tag)`.
- The server has a stateless `VerifyBattle(input, claimed)`.
- The web viewer makes up a seed locally and plays at once.

See proposal.md for why this is not enough.

## Goals / Non-Goals

**Goals:**
- The client can't choose or re-roll randomness, and can't learn the combat rolls before committing its army.
- The server stays a verifier: it never steps a live battle; it only stores inputs and re-runs finished ones.
- Small and local: one SQLite file, no new services.

**Non-Goals:** identity, runs, the scoreboard, cleaning up old battles, rate limits, and hiding Red's army from the client (see proposal).

## Decisions

**Two seeds, split by stream.** `Input.CombatSeed` feeds only `StreamCombat`; the map and Red's army, formation and skills keep using `Input.Seed` (now "map seed"). Everything that can't be known before the fight draws from the combat stream: damage rolls, cooldown jitter, execute/headshot, frost, the units' starting retarget and cooldown phases, and advance-sweep goals. Deployment previews can then be built with any combat seed (the viewer uses 0) and show exactly the positions the real battle starts from.
- *Alternative:* one secret seed with the map seed derived from it. Rejected: the client could then test deployments against the real rolls as soon as it has the derived values, unless the derivation stays secret, which is just two seeds with extra steps.

**The server picks both seeds with `crypto/rand`.** The server is not bound by the sim's determinism rules; only `sim/` is.

**The server rebuilds the input; it never trusts one.** `VerifyBattle` takes only `battle_id` and `claimed`. The input is rebuilt from stored columns: sim version, both seeds, Red skill count, and Blue's `Army` as protobuf bytes, exactly as locked. Validation runs at lock time with the same `convert.Input`, so an army that is stored is always playable.

**State machine, with transitions as conditional UPDATEs.**
- `started` → `locked`: `UPDATE … SET blue=?, locked_at=? WHERE id=? AND locked_at IS NULL AND deploy_deadline > now`.
- `locked` → `verified`: `UPDATE … SET verdict… WHERE id=? AND verified_at IS NULL`.
- If zero rows change, the request loses, and the handler re-reads the row to explain why: already locked, expired, already verified (then the stored verdict is returned) or unknown.
- No transactions are needed beyond single statements, and racing requests are safe.

**Storage: SQLite via `modernc.org/sqlite`, with sqlc.**
- `modernc.org/sqlite` is pure Go, so there's no cgo, it builds anywhere and works in CI.
- The schema lives in `server/store/schema.sql`, embedded and applied at startup with `CREATE TABLE IF NOT EXISTS`. Real migrations come later, when the schema first changes.
- Queries live in `server/store/queries.sql`, and sqlc generates typed Go. sqlc runs as a Go 1.24+ `tool` dependency (`go tool sqlc generate`), so nobody needs to install it.
- The DB path comes from `CHILL_DB`, default `chill.db`; tests use a temp file.
- *Alternative:* hand-written `database/sql`. It's viable at four queries, but the user chose sqlc for growth.

**Seeds as `INTEGER`.** SQLite integers are signed 64-bit, so uint64 seeds are stored by reinterpreting their bits (`int64(seed)`) and read back the same way. A round-trip test covers seeds ≥ 2⁶³.

**Clock injection.** The battle service takes `now func() time.Time`, so deadline tests don't sleep.

**Deadlines:** deploy within 2 minutes of start (the concept's 15 s deployment plus network slack); verify within 10 minutes of locking (a battle lasts at most 3 min at 1× speed). Both are service fields with these defaults.

**Web flow.**
- **New map:** `StartBattle`, then the worker builds a *preview* input with combat seed 0 and the chosen formation, and posts `step(0)` to show the map and both armies.
- **While deploying:** changing the formation rebuilds the preview.
- **Fight:** `LockDeployment`, then the worker creates the real input with the revealed combat seed and plays.
- **Verify:** `VerifyBattle(battle_id, outcome)`.
- If the server is down, the viewer shows the error; there is no offline mode in this change.

**How determinism is preserved and tested.**
- The seed split is a rule change, so `sim.Version` goes to 0.3.0 and the golden file is regenerated. The version guard enforces the bump.
- Each golden case gets a fixed combat seed, and a new sim test shows that two combat seeds give the same map and start positions but different outcomes.
- The golden set still runs natively and under WASM.
- Server tests re-run stored battles and compare with a direct `sim.Run` of the same rebuilt input.

## Risks / Trade-offs

- [The client still sees Red's army and the map before deploying, so it can search deployments against estimated rolls] → That's accepted: it searches against *estimated* rolls only, and the real combat seed is unknown until the army is locked. Server-side matchmaking can tighten this later.
- [Playing now needs the server] → That's expected for a server-verified game. Golden tests and the sim stay fully offline.
- [The SQLite file grows without cleanup] → Out of scope; battles are small rows. A cleanup job comes with runs.
- [sqlc as a tool dependency adds to `server/go.mod`] → It's only used at generation time and isn't linked into the server binary.

## Migration Plan

There's no deployed data. The old stateless `VerifyBattle` is replaced in place, and older clients stop working, which is fine before any release.
