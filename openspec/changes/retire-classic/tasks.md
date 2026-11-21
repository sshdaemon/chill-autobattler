## 1. Sim: retire the ids

- [x] 1.1 `sim/ids.go`: replace `Warrior`, `Archer`, `Mage` and `FactionClassic` with a retired-ids comment block and `retiredClasses`; note faction 1 as retired
- [x] 1.2 Delete `sim/faction_classic.go`, `unit_warrior.go`, `unit_archer.go`, `unit_mage.go` (and their skill constants)
- [x] 1.3 `sim/roster.go`: `register` pads holes; `RosterOf` returns nil for a hole; add `Units()`; `validateEntries` allows holes only at retired ids; switch `RedPool`, `Pool` and every other roster loop to `Units()`
- [x] 1.4 Bump `RulesVersion` to 0.16.0; update comments that name classic units (`ids.go`, `faction.go`, `roster.go`, wasm's `formationSlots` doc)

## 2. Sim tests

- [x] 2.1 Move every test off warrior, archer and mage per the design's table (`behaviour`, `effects`, `foundations`, `rules`, `roster`, `roster_baseline`, `plan`, `custom`, `tactic`, `mirror`, `snapshot`, `temperament`, `version`, `faction`, `draft` tests); delete tests that only pinned classic numbers and are covered elsewhere
- [x] 2.2 Effect-order test (slow then splash on one hit) with a test-only roster entry, restoring the roster afterwards
- [x] 2.3 New tests: faction 1 and classes 0–2 are rejected as unknown; a hole at a non-retired id and a unit at a retired id fail `ValidateRoster`; Red's faction is never 1 over many seeds
- [x] 2.4 Check each rewritten behaviour test still fails against a deliberately broken rule
- [x] 2.5 Golden replays: rebuild the cases on faction armies, find new mutual-wipe seeds with a Rus mix, re-record with `-update`, and pass natively and under WASM
- [x] 2.6 Re-run `TestMovementSmooth` and the mirror and side checks on faction armies; record the numbers

## 3. Proto, wire and wasm

- [x] 3.1 `units.proto`: reserve values 1–3 and their names; update the comment (wire value = class id + 1); regenerate Go and TS
- [x] 3.2 `pb/convert` and `pb/wire` tests use faction units through `UnitClass(c + 1)`
- [x] 3.3 `wasm/main.go`: data() emits `null` at retired class slots and leaves faction 1 out; update `wasm/main_test.go`

## 4. Server

- [x] 4.1 Migration `0007_retire_classic.sql` ends active Classic runs with reason "faction retired" (comment on the kept `DEFAULT 1`)
- [x] 4.2 Tests: the migration ends a Classic run and leaves others alone; `StartRun` rejects faction 1; `server/battle` tests use a faction army

## 5. Client

- [x] 5.1 Delete `factions/classic.json` and `units/looks/warrior.json`, `archer.json`, `mage.json`
- [x] 5.2 `gamedata.ts`: `classes` is `(ClassInfo | null)[]`; add `classOf`; fix every reader the compiler flags
- [x] 5.3 `units/looks.ts`: placeholder look by role (first unit of that role with a look)
- [x] 5.4 `DeployBar` mini-diagram marks by role, not class id
- [x] 5.5 Run-over screen wording for the "faction retired" end reason
- [x] 5.6 Grep `web/` (UnitLab, UnitPreview, preview shots, store defaults) for leftover classic names or class ids 0–2 and fix them
- [x] 5.7 Delete part, item, texture and clip files no remaining look, `_shared.json` or arena file uses
- [x] 5.8 `check-bundle.mjs`: fail on unreferenced files under `assets/parts` and `assets/clips`; prove it fails with a planted orphan

## 6. Balance tool

- [x] 6.1 `mix`, `title` and `core` work on a faction's three units by role; check one unit per role
- [x] 6.2 Side, matchups, comps and skills: one table per playable faction, also in the default report
- [x] 6.3 Runs and curves per faction; the "best" policy ranks skills from that faction's skills table; `counters` as a role triangle
- [ ] 6.4 `lab_test.go` passes; run `-exp plans,maps,edges,custom` once and note any shifts in `docs/balance-findings.md`

## 7. Docs and design

- [x] 7.1 `docs/GAME_CONCEPT.md` §4 and §10: describe roles and point to the faction units instead of the classic stats and skills
- [x] 7.2 `docs/ROADMAP.md`, `docs/battle-plans-roadmap.md`, `CLAUDE.md`: Classic retired
- [ ] 7.3 `docs/balance-findings.md`: keep dated numbers, add a note that Classic is retired and any new numbers from 2.6 and 6.4
- [x] 7.4 Delete `design/units/warrior/` (tracked files only; refs stay local and ignored); fix units' `balance.md` notes that compare against classic units

## 8. Verify

- [x] 8.1 `go test ./...` in every module, `pnpm build` (tsc and the bundle check) and web tests pass
- [x] 8.2 Final grep for `warrior|archer|mage|classic` outside archives, dated history and generic words; every hit is justified
- [x] 8.3 Play a run in the browser: the faction screen has three factions, the draft, deploy, battle and veterans work, and Red is never Classic
