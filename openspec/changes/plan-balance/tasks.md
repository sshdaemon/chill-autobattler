# Tasks

## 1. Sim

- [x] 1.1 `Stance.Effects`, passed as extras to `foldStats` and `loadout.learn`; `ValidatePlans` checks each effect against every roster entry. Verify: a test that a ranged-only effect on melee fails, and that All-out charge units are faster.
- [x] 1.2 Plan values from design.md; `RulesVersion` 0.10.0. Verify: `go test ./sim` with goldens re-recorded.

## 2. Balance

- [x] 2.1 Plan matrix with sides swapped and a dominance mark. Verify: `go run ./tools/balance -exp plans -n 1000` shows no "dominates".
- [x] 2.2 Findings and roadmap updated with the new matrix. Verify: `docs/balance-findings.md` has it.

## 3. Client

- [x] 3.1 Plan descriptions mention each plan's edge and cost. Verify: `pnpm --filter web typecheck`.

## Workflow follow-up

- Archive this change after the PR merges, in its own PR, merged after green CI.
