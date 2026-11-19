# Tasks

## 1. Before changing anything

- [x] 1.1 Record every golden case's result, tick count and final-state hash from the current code into a temporary file in the scratchpad, for the plan-0 check in 4.1. Verify: the file lists every case in `sim/testdata/golden/`.

## 2. Plans in the sim

- [x] 2.1 Add `type Plan uint8` and plan ids (`PlanInstinct = 0` to `PlanSkirmish = 5`, with a "next free" note) to `sim/ids.go`. Verify: `go build ./sim`.
- [x] 2.2 Add `sim/plan.go`: `Stance`, `Toggle`, `PlanDef`, `Plans`, `registerPlan` (sorted by id, refreshes the version), `PlanOf`, and the unit-role helper. Verify: `go build ./sim`.
- [x] 2.3 Add one file per plan (`plan_instinct.go`, `plan_shieldwall.go`, `plan_charge.go`, `plan_hunt.go`, `plan_focusfire.go`, `plan_skirmish.go`) with the starting values from design.md. Verify: `go vet ./sim`.
- [x] 2.4 Add a start-up check that applies each plan to each roster entry, with and without each of its skills, and runs `checkTemperament`. Wire it into `ValidateRoster`. Verify: a test that registers a bad plan (kiting melee) gets an error.
- [x] 2.5 Add `Plan` to `Army`. `Input.Validate` rejects an unknown plan in either army. Verify: a test with plan 200 gets `ErrInvalid`.
- [x] 2.6 Change `temperOf` to apply the side's stance between the roster and the ability hooks. Keep the resolved plans on the battle (`Battle.Plans`). Verify: tests for "Charge overrides a holding unit", "fields the plan leaves alone" and "a skill on top of a plan".
- [x] 2.7 Add `StreamRedPlan` at the end of the stream list. In `New`, draw Red's plan from the combat seed when Red's army is generated; otherwise use `in.Red.Plan`. Verify: tests for same combat seed → same plan, and all six plans drawn over 200 seeds.
- [x] 2.8 Hash the plan table in `refreshVersion` and `VersionFor`, and set `RulesVersion` to 0.9.0. Verify: `TestVersion*` passes and the version changes when a plan value changes.

## 3. Protobuf and conversion

- [x] 3.1 Add `uint32 plan = 3` to `Army` in `proto/chill/v1/units.proto` with a comment (0 = Instinct) and regenerate Go and TS with buf. Verify: `buf lint` and `go build ./...`; `pnpm -C web typecheck`.
- [x] 3.2 Carry the plan both ways in `pb/convert`. Verify: a round-trip test of an army with Skirmish.
- [x] 3.3 Check the server path: a battle with Blue on a plan verifies, and one with an unknown plan is rejected. Verify: a test in `server/battle`.

## 4. Goldens and stall tests

- [x] 4.1 Add a test-only hook that forces Red's plan to Instinct. Check every golden case against the record from 1.1. Verify: all match; then remove the temporary file.
- [x] 4.2 Add one golden case per plan with explicit armies on both sides. Re-record goldens (`go test ./sim -run TestGoldenReplays -update`) and the Red baseline (`go test ./sim -run TestRosterBaseline -baseline`). Find new mutual-wipe seeds if the check fails. Verify: `go test ./sim` passes natively and the WASM golden run passes.
- [x] 4.3 Extend the mirror test to every plan pair over every faction pair with a few seeds. Verify: no battle runs out of time.

## 5. Balance tool

- [x] 5.1 Add a `plans` experiment to `tools/balance`: for each faction, the same army on both sides, every plan against every plan, win % for the row's plan. Verify: `go run ./tools/balance -exp plans -n 200` prints a 6×6 matrix per faction.
- [ ] 5.2 Run the matrix with the default count and add the result to `docs/balance-findings.md` as the starting point for step 3. Verify: the doc has the matrix and any plan that wins against every other plan is called out.

## 6. Docs

- [x] 6.1 Mark step 1 done in `docs/battle-plans-roadmap.md` with the PR number. Verify: the doc says so.

## Workflow follow-up

- Archive this change after the PR merges, in its own PR, merged after green CI.
