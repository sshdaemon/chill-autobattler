# Design

## Context

The client plays battles in a worker through the WASM sim. During deployment
it previews the battle with combat seed 0; after the lock it plays with the
real combat seed. A generated Red's plan comes from the combat seed, so the
preview's Red plan is meaningless and must not be shown.

## Goals / Non-Goals

**Goals:** pick a plan with one click; see Red's plan only once it is real.

**Non-Goals:** explaining plans in depth (a one-line description is enough),
any sim change.

## Decisions

- **The sim owns the plan list; the web owns the words.** `data().plans` gives
  `{id, name}` in id order. `web/src/plans.ts` maps the sim name to a title and
  description, falling back to the sim name, so a new sim plan shows up even
  before it gets words.
- **Red's plan comes from the sim, not computed in the client.** A new
  `chillSim.plans()` returns the battle's `[blue, red]` after `create`; the
  worker sends it in its `started` message. The store keeps it, but the UI
  reads it only in the countdown, fighting and done phases. Computing it in
  TypeScript would duplicate the seeded draw.
- **Remembered for the run:** the plan lives in the store next to the
  formation and is reset to Instinct in `startRun`. A reload in the middle of
  a locked battle replays `lockedBlue`, which already carries the plan; a
  reload during deployment keeps the store's default (Instinct), like the
  formation.
- **Placement:** the plan buttons are a second row in the deployment bar
  above the formations, so the bar keeps its width on narrow screens.

- **Red's plan during the countdown:** the deployment preview uses combat
  seed 0, so its Red plan is a placeholder. When the army is locked, the client
  sets the battle up again with the real combat seed before the countdown, so
  the "Fight!" banner can name Red's real plan.

## Determinism

No sim change: the plan is part of the locked input the server re-runs. The
existing golden replays and the `battle-plans` server test cover it.

## Risks / Trade-offs

- [The deployment bar gets taller and covers more of the field] → compact
  text buttons, with the description in a single line under the row.
