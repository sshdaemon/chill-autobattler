# Tasks

## 1. Sim

- [x] 1.1 Ids in `sim/ids.go`: `PlanCustom` (255, reserved), edge ids
- [x] 1.2 Step tables in `sim/data.go`; edge registry and the five edges (`sim/edge_<name>.go`)
- [x] 1.3 `Chips`, `CustomGroup`, `CustomPlan`, `Def()`; `validatePlan` split out of `ValidatePlans`; checks for chips, edges per role, no edge on a second stance, custom triggers
- [x] 1.4 `Army.Custom`; `Validate`; `Battle.planDefs`; spawning and triggers use it; `Battle.Plans` reports `PlanCustom`
- [x] 1.5 Step tables and edges in `plansFingerprint`; `RulesVersion` bump

## 2. Sim tests and goldens

- [x] 2.1 Chips → `PlanDef` equals hand-written; empty custom plan plays like Instinct; each invalid case rejected
- [x] 2.2 Golden cases: army-wide custom plan, custom tactic with each trigger kind; re-record; mutual-wipe seeds if needed
- [x] 2.3 `TestPlansFight` includes custom plans (no stalls)

## 3. Proto and server

- [x] 3.1 `CustomPlan` messages and `Army.custom_plan` in `units.proto`; `pnpm gen`
- [x] 3.2 `pb/convert` and `pb/wire`; fuzz test with random custom plans
- [x] 3.3 Server test: lock and verify a custom plan; invalid one rejected at lock

## 4. Balance

- [x] 4.1 `-exp edges` and `-exp custom` in `tools/balance`
- [x] 4.2 Tune edges (and steps if needed) until no edge passes 60% alone and no custom plan beats every built-in one
- [x] 4.3 Findings in `docs/balance-findings.md`

## 5. Bridge and client

- [x] 5.1 `wasm/main.go`: `data().chips`, `data().plans[i].chips`, `checkPlan`; `units()`/`plans()` handle `PlanCustom`
- [x] 5.2 Store: `custom`, `plan = CUSTOM`, `blueArmy` sends it, kept for the run, cleared on a new run
- [x] 5.3 `PlanEditor.tsx` and the Edit button; tabs, chips, words, validation message; countdown pause
- [x] 5.4 "Custom" in the picker, status line and banners
- [x] 5.5 Browser check: build a custom plan, lock, the battle verifies fair; a custom tactic; next match keeps it; new run clears it

## 6. Wrap up

- [x] 6.1 `pnpm test`
- [x] 6.2 Roadmap: step 5 status and notes
