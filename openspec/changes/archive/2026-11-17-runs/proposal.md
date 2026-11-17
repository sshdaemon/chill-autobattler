# Proposal

## Why

Matches are still one-offs: nothing carries over, Red's strength is a dev-panel number, and there is no score. The game concept's core loop (§2, §10, §11) is a **run**: keep playing while you keep winning, survivors become veterans who learn skills, Red grows to match you, and the run ends at the first loss with a score on a scoreboard. For that to mean anything, the server has to own the run. Otherwise a client could invent veterans, set Red's skills to zero, or simply not report a loss.

## What Changes

- **Anonymous players.**
  - The browser gets a random secret player token from the server on first visit and keeps it in localStorage.
  - Every request after that sends it as `Authorization: Bearer <token>`. The server stores only its SHA-256 hash.
- **Server-side runs.**
  - `StartRun` creates a run with a random name (for example "Iron Fox 910") and ends any active run of that player as a loss.
  - `GetRun` returns the current run so a page refresh resumes it: match number, wins, score, veterans, the pending skill picks, and the current battle with what the client needs to resume it.
- **Battles belong to runs.** **BREAKING** for the battle API:
  - `StartBattle` takes the run id.
  - The server computes Red's skill count as the total skills of the run's veterans; the client no longer sends it.
  - `LockDeployment` requires the army to be the run's veterans, exactly as stored, followed by skill-less recruits, 5 units in all.
- **Results update the run.** When `VerifyBattle` gives its verdict:
  - **A fair win** adds 100 points plus 10 per surviving Blue unit, and makes every survivor a veteran. Each survivor below 5 skills then owes one skill pick.
  - **Anything else** (a loss, a draw, or a claim that isn't fair) ends the run.
- **Skill picks.** `PickSkills` takes one new skill for each survivor that is owed one, from its class's skills that it doesn't have. The next battle can't start until every pick is in.
- **Abandoned battles count as losses.** If the current battle's deadline passes without a verdict (never locked within the deploy deadline, or never verified within the verify deadline), the run ends as a loss. The check runs whenever the run or the scoreboard is read.
- **Scoreboard.** `GetScoreboard` returns the top 10 finished runs and the player's own best rank.
- **Database migrations.** `PRAGMA user_version` with numbered migration files replaces the single `CREATE TABLE IF NOT EXISTS` schema. Existing `chill.db` files upgrade in place.
- **Web.**
  - A run bar shows the name, match, wins, points and Red ★.
  - The picker locks veterans in, shown in gold with their stars.
  - The Victory screen gains skill picks.
  - A Run over screen shows the top-10 scoreboard with your row highlighted, and a New run button.
  - Reloading resumes the run.
  - Red ★ leaves the dev panel.

## Out of scope

- Real accounts, logins, display names and moving a player between devices.
- Rate limits or abuse protection beyond token checks.
- Balance changes. Red's skill spread stays seeded-random (concept §10.2), and counter-picking is still an open question.
- Real visuals; the capsules stay.

## Capabilities

### New Capabilities

- `runs`: players, runs, veterans, skill picks, Red scaling, scoring, abandonment, and the scoreboard, all decided by the server.

### Modified Capabilities

- `battle-sessions`: battles are started for a run. Red's skill count comes from the run, and locking checks the run's veterans.
- `match-flow`: the army picker keeps veterans, Victory leads to skill picks, a loss leads to Run over with the scoreboard, and a refresh resumes the run.

## Impact

- `proto/`: a new `RunService` (`CreatePlayer`, `StartRun`, `GetRun`, `PickSkills`, `GetScoreboard`); `StartBattleRequest` takes `run_id` instead of `red_skill_count`.
- `server/`:
  - `store` gains migrations and the `players` and `runs` tables, plus a `run_id`/`match_no` link on battles.
  - A new `run` package, and an auth interceptor.
  - `battle` validates veterans and updates the run on verdict.
- `web/`: the store's run state, a run bar, veterans in the picker, skill picks, Run over with the scoreboard, and resume on load.
- No sim change; `sim.Version` stays 0.3.0.
