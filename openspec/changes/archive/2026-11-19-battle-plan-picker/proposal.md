# Proposal

## Why

Battle plans exist in the sim (change `battle-plans`), but the client always
sends Instinct, so players can't use them. This is step 2 of
`docs/battle-plans-roadmap.md`: let the player pick a plan during deployment,
and show Red's plan once the fight starts.

## What Changes

- **Plan picker:** the deployment bar gets a row of plan buttons (Instinct,
  Shield wall, All-out charge, Hunt the archers, Focus fire, Skirmish), each
  with a one-line description. The chosen plan goes into Blue's locked army.
- **Remembered for the run:** the plan stays chosen from match to match in a
  run. A new run starts on Instinct.
- **Red's plan revealed at the fight:** the "Fight!" banner names Red's plan,
  and the status line shows both sides' plans during the fight. Red's plan is
  never shown before the army is locked.
- **WASM bridge:** `data()` lists the plans (id and sim name), and a new
  `plans()` returns the current battle's two plans.
- Plan titles and descriptions are presentation, in `web/src/plans.ts`.
- **Fix:** the browser's lean protobuf decoder (`pb/wire`) didn't read
  `Army.plan`, so the client would play Instinct while the server re-ran the
  chosen plan and judged the battle unfair. It now decodes the plan, and its
  equivalence fuzz test against `pb/convert` includes plans.

Out of scope:

- Tuning the plans (step 3), groups and triggers (step 4), custom plans
  (step 5), free text (step 6).
- Any sim, server or protobuf schema change: those landed in `battle-plans`.

## Capabilities

### New Capabilities

### Modified Capabilities
- `match-flow`: deployment adds the plan picker; the fight start reveals
  Red's plan.

## Impact

- `wasm/main.go`: plans in `data()`, new `plans()`.
- `pb/wire`: decode `Army.plan`; fuzz test covers it.
- `web/src/sim/sim.worker.ts`: the `started` message carries the plans.
- `web/src/store.ts`: Blue's plan state, `setPlan`, plan in `blueArmy`, the
  battle's plans from the worker.
- `web/src/ui/DeployBar.tsx`, `web/src/ui/Overlays.tsx`, `web/index.html`
  (styles), new `web/src/plans.ts`.
