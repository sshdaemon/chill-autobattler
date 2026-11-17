# Tasks

## 1. Wire format and storage

- [x] 1.1 Proto: `RunService` (CreatePlayer, StartRun, GetRun, PickSkills, GetScoreboard), `Run`/`CurrentBattle`/`ScoreRow`; `StartBattleRequest.run_id` replaces `red_skill_count`; `VerifyBattleResponse.run`; then `pnpm gen`; verify `pnpm lint:proto`
- [x] 1.2 Store migrations (`user_version`, `migrations/0001`, `0002`) with sqlc reading the directory; queries for players, runs, battle↔run, the scoreboard and the abandonment sweep; verify store tests for a fresh DB, upgrading a version-1 DB, and a migration rerun being a no-op

## 2. Server

- [x] 2.1 Auth middleware (Bearer token → player id in context) and `CreatePlayer`; verify tests for an unknown or missing token → Unauthenticated, and that `CreatePlayer` needs no token
- [x] 2.2 `run` service: StartRun (name, restart ends the old run), GetRun (settles abandonment), PickSkills (validation, all-or-nothing), GetScoreboard (sweep, top 10, mine, best rank); verify tests for each spec scenario
- [x] 2.3 Battle service on runs: StartBattle needs a ready run (Red skills = veteran skills); lock checks veterans; verify updates the run in one transaction (win → score, veterans, picks; else over); battles of other players → NotFound; verify tests driving whole runs (win → picks → match 2 with veterans and Red ★ → loss → scoreboard)

## 3. Web

- [x] 3.1 Session token and auth interceptor (with one automatic re-registration on Unauthenticated); resume on load by run state; remove Red ★ from the dev panel; verify typecheck
- [x] 3.2 Run bar; picker with veterans (gold ★, minimum per class, army order veterans-first); verify in the browser
- [x] 3.3 Victory skill picks (per survivor, class skills not owned) → Next match; Defeat → Run over with the top-10 scoreboard (mine highlighted) and New run; verify in the browser through a win, picks, match 2 and a loss
- [x] 3.4 Reload during picks and during a locked fight resumes correctly; verify in the browser

## 4. Docs

- [x] 4.1 Update the README (runs, auth, anti-cheat rows, migrations) and CLAUDE.md; verify `pnpm test` passes

## Workflow follow-up

- Archive the change once merged (auto-merge after green CI).
