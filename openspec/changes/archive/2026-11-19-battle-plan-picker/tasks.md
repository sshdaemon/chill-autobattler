# Tasks

## 1. WASM bridge

- [x] 1.0 The browser decoder (`pb/wire`) reads `Army.plan` like `pb/convert`. Verify: `TestDecodeMatchesConvert` with plans fails without the fix and passes with it.
- [x] 1.1 `data()` lists the plans (`plans: [{id, name}]`); `chillSim.plans()` returns the current battle's `[blue, red]` plans. Verify: `pnpm build:wasm` and `pnpm test:wasm`.

## 2. Client

- [x] 2.1 The worker's `started` message carries the plans; the store keeps them. Verify: `pnpm --filter web typecheck`.
- [x] 2.2 `web/src/plans.ts` with titles and one-line descriptions keyed by sim name. Verify: typecheck.
- [x] 2.3 Store: `plan` state and `setPlan` (deploying only), plan in `blueArmy`, reset to Instinct on a new run. Verify: typecheck.
- [x] 2.4 Deployment bar: plan buttons with the chosen one marked and its description. Verify in the browser: picking Shield wall marks it.
- [x] 2.5 Fight banner names Red's plan; the status line shows both plans from the countdown on, and nothing shows Red's plan while deploying. Verify in the browser.

## 3. End to end

- [x] 3.1 Play a battle with a plan in the browser: it is verified fair, and the plan is still chosen at the next match. Verify: the result says "verified fair by the server".
- [x] 3.2 Mark step 2 done in `docs/battle-plans-roadmap.md`. Verify: the doc says so.

## Workflow follow-up

- Archive this change after the PR merges, in its own PR, merged after green CI.
