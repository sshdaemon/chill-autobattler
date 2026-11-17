# Design

## Context

- Battles are server-issued (`battle-sessions`) and stored in SQLite. The schema is applied with `CREATE TABLE IF NOT EXISTS`.
- The client picks Red's skill count and can start a battle at any time.
- The web match loop (`match-flow`) runs picker → deploy → fight → result, and every match stands alone.
- Motivation is in proposal.md.

## Goals / Non-Goals

**Goals:**
- The server is the only authority on veterans, skills, Red's strength, scores and when a run ends.
- A reload resumes the run.
- Abandoning a losing fight doesn't help.

**Non-Goals:** accounts, cross-device play, rate limits, balance changes (see proposal).

## Decisions

**Player tokens.**
- `CreatePlayer` returns 32 random bytes as hex. The server stores `sha256(token)` as the lookup key, so a database leak doesn't hand out usable tokens.
- An HTTP middleware wraps the Connect mux. It reads `Authorization: Bearer …`, resolves the player and puts the player id in the request context. `CreatePlayer` and `/healthz` are exempt; every other RPC without a known token gets `Unauthenticated`.
- Handlers check that the run or battle they touch belongs to that player; otherwise they answer `NotFound`, so ids can't be probed.
- *Alternative:* a Connect interceptor. Rejected because the middleware is independent of the Connect API version and simple to test.

**The run state machine** lives in a `runs` row:
- `state`:
  - `ready`: may start a battle.
  - `picking`: skill picks are owed.
  - `over`: finished.
- `match_no`, `wins`, `score`.
- `veterans`: protobuf `Veterans`, a list of `UnitSpec` with class and skills, in army order.
- `current_battle_id`: the battle started for `match_no` and not yet resolved.
- `end_reason`: "defeat", "draw", "not fair", "abandoned" or "restarted".

**Transitions** happen in one SQLite transaction each, through sqlc's `WithTx`:
- `StartBattle`: `ready` with no current battle → creates the battle and sets `current_battle_id`.
- `VerifyBattle`: records the verdict and applies the result to the run.
  - A fair Blue win: `wins+1`, `score += 100 + 10·survivors`, the veterans become the surviving Blue units, then `picking` if any survivor has fewer than 5 skills, otherwise `ready`. `match_no+1`, and the current battle is cleared.
  - Anything else: `over`, with an end reason.
- `PickSkills`: `picking` → `ready`. It takes one pick for every owed veteran in a single request; partial picks are rejected.
- `StartRun`: ends the player's active run (`restarted`) and inserts the new one.

**Veteran check at lock.**
- The first `len(veterans)` units of Blue's army must equal the veterans in order: the same class and the same skill set.
- The rest must have no skills, and the army must have exactly 5 units.
- The client builds the army in that order: veterans first, then recruits by class. This runs alongside the sim's own validation.

**Red's skill count** at `StartBattle` is the sum of the veterans' skills. The sim's seeded `RedArmy` already spreads that count (concept §10.2), so the sim doesn't change.

**Abandonment is settled lazily, plus a sweep.**
- `settle(run)` runs before any read or write of a run. It ends the run as `abandoned` when its current battle is past its deploy deadline without a lock, or past its verify deadline without a verdict.
- `GetScoreboard` first runs one `UPDATE … WHERE state != 'over' AND current battle expired`, so abandoned runs reach the board even if their player never returns.
- No background job is needed.

**Migrations.**
- `store/migrations/0001_battles.sql` (the existing table) and `0002_runs.sql` (`players`, `runs`, and `battles.run_id` + `match_no` via `ALTER TABLE ADD COLUMN`).
- `Open` applies the files above `PRAGMA user_version` in order, each in a transaction that also bumps `user_version`.
- sqlc reads the same directory as its schema.
- Existing dev databases from before this change start at version 0. Their `CREATE TABLE IF NOT EXISTS` is a no-op, and the `ALTER` adds the columns.

**Run names** are an adjective, a noun and a number from 100 to 999, using the prototype's word lists and drawn with `crypto/rand`. Names don't have to be unique.

**Scoreboard.**
- The top 10 come from `ORDER BY score DESC, ended_at ASC`, each with `mine = player_id = ?`.
- The player's best rank is `1 + count(over runs with a higher score)`.

**Web.**
- A tiny `session.ts` holds the token (localStorage) and a Connect interceptor that adds the header. If the server answers `Unauthenticated` (say, a wiped DB), the client drops the token and creates a new player once.
- On load: `GetRun`, or `StartRun` if there is none, then route by run state:
  - `ready` → picker.
  - `picking` → skill picks.
  - A current battle, unlocked and before its deadline → deployment with that battle.
  - A current battle, locked → replay it from tick 0 and verify (the sim is deterministic, so this is the same battle).
  - `over` → Run over.
- The picker builds the army as veterans first. Veterans show in gold with their ★ count, and a class count can't go below its veterans.
- Red ★ comes from the server and shows in the run bar; the dev control is removed.

**How determinism is preserved and tested.**
- The sim is untouched; the golden replays and `sim.Version` 0.3.0 stay.
- The server tests drive whole runs through the real handlers:
  - The battles are played with `sim.Run` exactly like the client does, so honest claims verify.
  - Wins carry veterans, Red's skill count matches the veterans' skills, and losses, draws, unfair claims and abandonment all end the run.
- The store tests cover migrating a version-1 database.

## Risks / Trade-offs

- [A cleared localStorage loses the player's runs] → Accepted for anonymous play; accounts can adopt tokens later.
- [Run state and battle state in two tables could disagree] → Every transition that touches both runs in one transaction.
- [Lazy settling means an abandoned run stays "active" until something reads it] → The scoreboard sweep covers what anyone can see.
