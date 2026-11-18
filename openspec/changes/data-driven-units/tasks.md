# Tasks

## 1. Lock today's behaviour

- [ ] 1.1 Add golden replay cases recorded with the current code that use every skill: each skill on its own, Frost + Blast on one mage, Sharp arrows + Lob on one archer, Swift + Dash, Executioner and Headshot near the 40% threshold, and Red with 0, 3 and 25 skills. Verify that `pnpm test:go` and `pnpm test:wasm` pass, with `sim.Version` still "0.4.0" and no existing outcome changed.
- [ ] 1.2 Add a test that records today's spawn stats for every class with every single skill, plus the combinations from 1.1 (exact `Fx` values for HP, damage range, range, speed, cooldown and swing), and today's `RedArmy` output for seeds 1–200. Verify it passes on the current code.

## 2. Roster types and data

- [ ] 2.1 Add `sim/roster.go` with `UnitDef`, `Tag`/`TagBonus`, `Projectile`, `Effect`/`EffectKind`/`StatKind` and the new `SkillDef` with `Effects`, plus `RosterOf(Class)` and `SkillDefOf(Class, Skill)`. Verify that `go build ./...` succeeds.
- [ ] 2.2 Write `unit_warrior.go`, `unit_archer.go` and `unit_mage.go` with today's stats, traits, tags, bonuses and the five skills each, moving the effect constants (dash, blink, frost, blast) into effect parameters. Remove `ClassStats`, `Matchup` and `SkillsOf`. Verify the build succeeds.
- [ ] 2.3 Add `ValidateRoster` (duplicate ids, five skills, ranged needs projectile and speed, declared tags, positive stats, parameter ranges, at most one of each non-stat effect kind), with a test of the real roster and one failing case per rule. Verify with `go test ./sim -run Roster`.

## 3. Sim reads the roster

- [ ] 3.1 Compile a loadout in `newUnit`: stat bonuses grouped by stat in a fixed order, then in skill-list order, each a separate `MulDiv`. Store finisher, slow, splash, gap-closer, escape and over-cover. Verify that the spawn-stats test from 1.2 passes unchanged.
- [ ] 3.2 Switch `think`, `tryBlink` and `kite` to read the loadout and the `Melee` and `Kite` traits. Switch `damage` to the finisher and the tag matchup, and `shoot`/`updateShots` to `Projectile`, over-cover, slow and splash, keeping the hook order in design.md §3. Verify that the golden replays pass natively and under WASM with no `-update`.
- [ ] 3.3 Make `RedArmy` and `Input.Validate` use the roster (`InRedPool`, unknown ids rejected), with a comment that changing the pool needs a version bump. Verify that the `RedArmy` test from 1.2 passes and that an input with class id 7 is rejected.
- [ ] 3.4 Add a test-only "spearman" roster entry (gap-closer, slow on hit, stat bonus) swapped in by a helper. Verify a battle with it runs to the end, slows victims, and needed no sim code change. Also verify that the tag matchup matrix equals the old 3×3 table.
- [ ] 3.5 Run `TestNoNondeterministicImports` over the new files. Update the comments in `data.go` and `unit.go` that described the old tables. Verify with `pnpm test:go`.

## 4. Callers outside the sim

- [ ] 4.1 Update `wasm/main.go data()` to build from the roster. Keep the existing fields, add `traits`, `tags`, `bonus` and the `effects` of each skill, and extend `wasm/main_test.go` to check the new fields and the unchanged old ones. Verify with `pnpm test:wasm`.
- [ ] 4.2 Update `server/run` skill-pick validation, `pb/convert` and `pb/wire` (if they reference the class count), and `tools/balance` to use the roster. Verify with `pnpm test:go` and a short `pnpm balance` run whose report numbers equal the current `docs/balance-report.md`.
- [ ] 4.3 Run `pnpm test` (Go, WASM and the web typecheck) and `pnpm start`, and play one battle in the browser. Verify the picker shows the same classes and skill texts, and the battle runs.

## 5. Docs

- [ ] 5.1 Update CLAUDE.md (sim section: the roster, effect kinds, "add a unit = a roster file", the Red pool version rule) and `docs/GAME_CONCEPT.md` §4/§10 pointers to the roster files. Verify by reading the updated sections against the code.

## Workflow follow-up

- Open a PR for review. Merge it when the user says so, then archive the change in a separate PR (auto-merged after green CI).
