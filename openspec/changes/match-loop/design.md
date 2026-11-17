# Design

## Context

- `web/` has a zustand store with phases `idle → starting → deploying → locking → fighting → done`, a dropdown formation selector, and a Web Worker that runs the WASM sim.
- The battle flow is `StartBattle` / `LockDeployment` / `VerifyBattle` (see the battle-sessions spec).
- The sim already supports Custom formations with explicit positions; it snaps them to the nearest free cell and validates them against the zone.

## Goals / Non-Goals

**Goals:** a playable match loop with no balance numbers duplicated in TypeScript, and nothing in the client able to change an outcome.

**Non-Goals:** runs and progression, real visuals, and sim changes (see proposal).

## Decisions

**Phases.**
- The sequence is `picker → starting → deploying → locking → countdown → fighting → done`, with Next match going back to `picker`.
- The store owns the phase, and components render from it.
- Timers (the 15 s deploy and the 1 s countdown) live in the store, not in components, so a re-render can't double-fire Fight.

**Balance data comes from the sim.**
- `wasm/` gains `data()`, which returns a JSON document built from `sim.ClassStats`, `sim.Matchup` and `sim.SkillsOf`. Fixed-point values are converted to decimals only at this display boundary.
- It also gains `formationSlots(formation, classes)` → Int32 positions, using `sim.FormationSlots`.
- The worker asks for the data once at start-up, and for slots whenever the army changes. The picker cards and the deploy-bar diagrams then always match the rules.
- *Alternative:* a TS copy of the tables. Rejected, because it would drift from `sim/data.go`.

**Drag.**
- During `deploying`, Blue's unit meshes take pointer events. A drag raycasts onto the ground plane and clamps the point to Blue's zone, with a margin of one unit radius.
- On release, the store switches to Custom, with every Blue unit at its current preview position and the dragged unit at its new point. Positions are converted to raw fixed-point with `Math.round(v × 65536)`.
- The worker re-creates the preview, and the sim snaps the unit onto a walkable cell. The client never decides walkability.
- The camera stays fixed for the whole deployment. OrbitControls listens on the canvas and sees `pointerdown` before R3F's handler can stop it, so turning it off only while dragging still let a drag orbit the view. It frees up once the fight starts.

**Fog while deploying.** Red's units aren't rendered in `deploying`. This is presentation only: the browser has Red's army because the sim needs it (a known limit in the battle-sessions design).

**Countdown.** When `LockDeployment` resolves, a "Fight!" banner shows for 1 s, then the worker creates the real input with the revealed combat seed and starts stepping. The 1 s wait is purely visual; the sim starts at tick 0 either way.

**Automatic verification.** On `done`, the store calls `VerifyBattle` once. A failure shows "couldn't verify" with a retry button; the server's first-verdict-final rule makes retrying safe.

**Debug controls.** Speed (1×/2×/4×) and Red's skill count move into a small "dev" panel. Red's skill count is still client-chosen until runs exist.

**`pnpm start`.** A root script builds the WASM, then runs `pnpm server` and the web dev server as background jobs under `trap 'kill 0' EXIT`, so Ctrl-C stops both. No new dependency (such as `concurrently`) is needed.

**How determinism is preserved and tested.** Nothing in `sim/` changes; the golden replays and `sim.Version` stay as they are. The new WASM exports are read-only, and a WASM test checks that `formationSlots` matches `sim.FormationSlots` and that `data()` parses with the expected classes. Every outcome still comes only from the sim, and is verified by the server after each match.

## Risks / Trade-offs

- [Dragging lets a unit end up on a different cell than where it was dropped] → The preview redraws from the sim within a frame, so what you see is what is locked.
- [15 s may feel short on first use] → It's a constant in the store, easy to tune.
- [The UI grows in one store file] → Split the store by phase if it passes ~300 lines.
