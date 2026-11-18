# Tasks

## 1. Roster-generic UI

- [x] 1.1 Add `ui: { label, icon, blurb }` to each look in `looks.json` (warrior, archer, mage, with today's texts), and remove `CLASS_ICON`, `CLASS_LABEL` and `CLASS_BLURB` from `gamedata.ts`. Make the picker, info panel, overlays and portraits read `lookOf(cls).ui`. Verify with the typecheck and the picker in the browser looking as before.
- [x] 1.2 Make the picker grid fit 3–6 cards and wrap on narrow screens. Verify with screenshots at 1280 px and 375 px, using a temporary fourth roster entry in dev.
- [x] 1.3 Replace class-number branches: the shot sound from `projectile`, upper-body aim from `ranged`, `fxDemo` from the roster. Verify that a search finds no `cls === <n>` or `cls !== <n>` in `web/src`, and that a battle plays and sounds as before.
- [x] 1.4 Extend `wasm/looks_test.go`: every look has a non-empty label, icon and blurb. Verify with `pnpm test:wasm`, including one failing case.

## 2. Roster ids and versioning

- [ ] 2.1 Extend `ValidateRoster` with the id rules (unit ids equal their position; skill ids unique per meaning, never reused), with tests. Add a comment to `proto/chill/v1/units.proto` that further ids come from the roster. Verify with `go test ./sim`.
- [ ] 2.2 Test that a fourth unit outside Red's pool leaves every golden replay unchanged, and that the server accepts an army with its id and rejects an unknown id. Verify with `pnpm test:go`, using the test-only roster helper.

## 3. Balance check

- [ ] 3.1 Generalise `tools/balance` tables from `[3]` and class constants to the roster, with no change to the existing report. Verify that `pnpm balance` rewrites `docs/balance-report.md` byte for byte.
- [ ] 3.2 Add `pnpm balance:unit <name>` (head-to-head, swap and skill checks, bands, non-zero exit on fail). Verify by running it on the test-only "frost ranger" entry swapped in, and checking the table and exit code.

## 4. Preview

- [ ] 4.1 Add the dev-only `?preview=<name>` route with `UnitPreview` (recipes on both teams, clip strip with skill effects, portrait) and `window.previewShot(step)`. Verify that it isn't in the production bundle (`pnpm --filter web build` plus a search for the component), and with screenshots of the warrior preview.

## 5. The skill

- [ ] 5.1 Write `.claude/skills/new-unit/SKILL.md` (steps, the two checkpoints, the stop rule, tuning limits, privacy of references, the PR contents), `checklist.md` and the effect cookbook. Add `design/units/_template/brief.md` and the `.gitignore` entry for `design/units/*/refs/`. Verify by reading it against the unit-authoring spec, and with `git check-ignore` on a sample refs file.
- [ ] 5.2 Dry run: run `/new-unit` on a sample brief ("Spearman: a sturdy melee unit with reach, good against charging melee") on a scratch branch, through to a ready PR body, without opening the PR. Verify every spec scenario was met, then delete the scratch branch and note what the run changed in the skill.

## 6. Docs and checks

- [ ] 6.1 Update CLAUDE.md and the README: how to add a unit, the brief, `/new-unit`, `balance:unit`, the preview, and the id and versioning rules. Verify by reading them against the code.
- [ ] 6.2 Run `pnpm test` and play a battle. Verify golden replays are untouched and the server reports the battle verified.

## Workflow follow-up

- Open a PR for review. Merge it when the user says so, then archive the change in a separate PR (auto-merged after green CI).
- The first real unit is a separate run of `/new-unit` from the user's own brief.
