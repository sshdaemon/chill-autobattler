# Proposal

## Why

Today the client chooses the seed and can re-run battles locally as often as it likes before claiming one. A cheater can simulate thousands of seeds or deployments and submit only a winner. The server re-running the battle can't catch that, because each submitted battle is genuinely valid. To make a claimed win mean something, the server has to fix the randomness, and the client must commit its army before it learns the combat rolls.

## What Changes

- **Two seeds per battle.**
  - The *map seed* decides the map, Red's army, Red's formation and Red's skills. It is public as soon as the battle starts, because the player needs to see the map to deploy.
  - The *combat seed* decides every combat roll: damage, cooldown jitter, Executioner/Headshot and Frost, plus the starting timers and sweep goals. It is revealed only after the player locks a deployment.
  - **BREAKING** for the sim: `BattleInput` gains `combat_seed`, and `sim.Version` goes to 0.3.0 with new golden replays.
- **The server issues battles**, in three RPCs:
  - `StartBattle`: the server creates a battle with fresh random seeds and returns its id, the map seed and the deadlines. The client sends the sim version and Red's skill count for now; until runs exist, that count is client-chosen.
  - `LockDeployment`: the client sends Blue's army once, before the deploy deadline. The server validates and stores it, then reveals the combat seed.
  - `VerifyBattle`: the client sends the battle id and its claimed outcome. The server rebuilds the input from what it stored, re-runs it and records the verdict. **BREAKING**: it no longer accepts an input from the client.
- **Battles are stored in SQLite** (pure Go, `modernc.org/sqlite`, queries generated with sqlc). A battle can be locked once and verified once (a repeat returns the stored verdict). Expired battles can no longer be locked or verified.
- **The web debug viewer** follows the new flow: **New map** starts a battle and shows the deployment preview, the formation can be changed while deploying, **Fight** locks the deployment and plays it, and **Verify** checks it by battle id.

## Out of scope

- Player identity, runs, the scoreboard and server-computed Red skill counts. They come with the runs change; the server just stores what the client asked for now.
- Hiding Red's army from the client. The browser sim needs it, so a determined client can always read it. Fog of war stays a presentation feature.
- Rate limiting, abuse protection and cleaning up old battles.
- Postgres or any hosted database.

## Capabilities

### New Capabilities

- `battle-sessions`: server-issued battles. Covers seed issuing, deployment locking with its deadline, verification by battle id, single use, expiry, and persistence across restarts.

### Modified Capabilities

- `battle-replay`:
  - The battle input now carries separate map and combat seeds.
  - Verification takes a battle id and uses the server-stored input instead of one sent by the client.

## Impact

- `sim/`: `Input.CombatSeed`; the combat RNG stream is derived from it; version 0.3.0; regenerated golden file.
- `proto/chill/v1`: new `StartBattle`/`LockDeployment` RPCs, a changed `VerifyBattleRequest`, and `BattleInput.combat_seed`. Run `pnpm gen` again.
- `server/`: a new SQLite store (`server/store`, with sqlc), a battle service replacing `server/verify`, and a configurable DB path.
- `web/`: the store and HUD flow (start, then deploy, then fight, then verify), and a preview of the deployment before locking.
- New Go dependencies: `modernc.org/sqlite`, plus sqlc as a `go tool`.
