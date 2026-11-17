# Tasks

## 1. Data from the sim

- [x] 1.1 Add `data()` and `formationSlots()` to `wasm/`; verify a WASM test checks `formationSlots` against `sim.FormationSlots` and that `data()` parses with 3 classes and 5 skills each
- [x] 1.2 The worker fetches the data on start-up and slots on request; verify `pnpm --filter web typecheck`

## 2. Match flow

- [x] 2.1 Phase machine in the store (picker → starting → deploying → locking → countdown → fighting → done), with the deploy timer, countdown and automatic verification; verify typecheck
- [x] 2.2 Army picker overlay: class cards (stats, damage taken), −/+ counts, 5 slots, default 2/2/1, last army remembered; verify in the browser
- [x] 2.3 Deploy bar: preset buttons with diagrams of your army, Custom, the 15 s timer and Ready; Red hidden while deploying; verify in the browser
- [x] 2.4 Drag Blue units within the zone, switching to Custom, with snapping done by the sim; verify in the browser that a dragged unit moves and Custom is selected
- [x] 2.5 Fight banner, result overlay (Victory/Defeat/Draw, survivors, verdict, retry, Next match) and a dev panel (speed, Red ★); verify in the browser that a full match ends with "verified fair" and Next match returns to the picker

## 3. Local launch and docs

- [x] 3.1 Add `pnpm start` (WASM build, then server and web together, both stopped on Ctrl-C); verify it starts both and Ctrl-C stops both
- [x] 3.2 Update the README (how to play, `pnpm start`) and CLAUDE.md; verify `pnpm test` passes

## Workflow follow-up

- Archive the change once merged.
