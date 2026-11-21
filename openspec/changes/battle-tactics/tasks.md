# Tasks

## 1. Sim data

- [x] 1.1 Allocate ids in `sim/ids.go`: plans 6–8 (Hammer and anvil, Reserve, Hold then charge), engage rule Flank 4, trigger kinds 1–3 (after, ally falls, other group fighting)
- [x] 1.2 `sim/plan.go`: `Roles`, `GroupDef`, `Trigger`; `PlanDef` embeds `Roles` and gains `Groups [2]GroupDef`; `IsTactic()`
- [x] 1.3 Trigger plug-ins (`sim/trigger.go` registry, `trigger_<name>.go` each) with a check for their number
- [x] 1.4 `ValidatePlans`: every role × group × stage stance valid on every roster entry with each skill; no effects in `Then`; triggers registered and in range; Instinct has no groups
- [x] 1.5 `FlankWide` in `sim/data.go`; `sim/engage_flank.go`

## 2. Sim behaviour

- [x] 2.1 Group split at `New`: team's mean x, x ≤ mean (the first flank's side) is group 0; `Unit.Group`
- [x] 2.2 Layered temperament: plan role → group opening (or `Then` once switched) → abilities; effects from plan role + opening, fixed at spawn
- [x] 2.3 Trigger check at the start of `Step` (teams, groups, in order); a switch rebuilds each living unit's temperament and resets `held` / `lineSet`
- [x] 2.4 `teamFighting` counts only the unit's own group
- [x] 2.5 The three tactic files `sim/plan_hammer.go`, `plan_reserve.go`, `plan_holdcharge.go`
- [x] 2.6 `RulesVersion` 0.11.0

## 3. Sim tests and goldens

- [x] 3.1 Before the version bump: with Red forced to the old six-plan draw and no tactics, every existing golden matches its outcome
- [x] 3.2 Unit tests: group split (3-2, dragged unit, centre-line tie, the same split on both teams); each trigger fires on its condition and only once; effects unchanged after a switch; `Then` with effects rejected; Reserve's right group holds while only the left fights; the hammer's melee swing wider than under Instinct
- [x] 3.3 `TestPlansValid` expects nine plans; `TestRedPlan` sees every plan; `TestPlansFight` covers all nine (no stalls)
- [x] 3.4 Golden cases per tactic; re-record with `-update`; find new mutual-wipe seeds if needed; native and WASM (`pnpm test`)

## 4. Balance

- [x] 4.1 `go run ./tools/balance -exp plans` over nine plans; tune the tactics until nothing dominates in any playable faction
- [x] 4.2 Write up in `docs/balance-findings.md`

## 5. Bridge and client

- [x] 5.1 `wasm/main.go`: `data().plans[].tactic`; `units()` reports `group`
- [x] 5.2 `web/src/gamedata.ts`, `web/src/plans.ts`: tactic flag and words for the three tactics
- [x] 5.3 `DeployBar.tsx`: a Tactics row; the description line covers both rows
- [x] 5.4 Deployment view: L/R group tags under Blue's units while a tactic is chosen, updated on drag and formation change
- [x] 5.5 Check in the browser: pick each tactic, tags follow drags, the banner names a Red tactic, the battle verifies fair

## 6. Wrap up

- [x] 6.1 `pnpm test` passes
- [x] 6.2 Roadmap: step 4 status and notes in `docs/battle-plans-roadmap.md`
