# Design

## Context

- The sim already computes shared team vision each tick (`Battle.spotted`) with `Map.ClearView`.
- The client receives Int32 snapshots: unit records of 11 ints, plus shots and events.
- The prototype's fog was a 56×56 `DataTexture` (one texel per nav cell) refreshed every 0.1 s. Its portraits were rendered once with a second `WebGLRenderer`, and its camera was the `RtsCamera` in `web3d-demo/common.js`.

## Goals / Non-Goals

**Goals:** the prototype's fog, portraits, info panel and camera, with vision coming from the sim (one implementation of the rules).

**Non-Goals:** a minimap, fog for Red, sound (see proposal).

## Decisions

**Vision from the sim, never re-implemented in TS.**
- `Battle.SeenCells(team) []bool` marks every grid cell within some living teammate's sight with `ClearView`, the same rules as `canSee`.
- The snapshot's unit flags gain `UFSeen` (bit 8) for units in the team's `spotted` set (computed at the start of the tick that produced the snapshot).
- Both are read-only outputs. They are neither in the outcome nor in the checksum, so `sim.Version` and the golden replays don't change.
- The cost is about 5 units × up to ~700 cells × ~40 samples each. The worker asks for it every 2 ticks (10 Hz, like the prototype), not every tick.

**Snapshot unit record → 13 ints:** `… target, skills, kills, dealt` (dealt is raw fixed-point). `TestSnapshotLayout` and the TS decoder change together.

**Fog overlay.**
- A 56×56 RGBA `DataTexture` with linear filtering, so the edges are soft. Alpha is 0.62 where unseen, and the plane sits just above the ground (renderOrder 5).
- While deploying, the client lights Blue's zone from the exported deploy zone; the sim has no vision before the fight.
- Red units are drawn only when the fog is off or `UFSeen` is set. Their portraits use the same flag.
- **V** toggles the fog; it's a dev key, as in the prototype.

**Portraits.**
- One offscreen `WebGLRenderer` (108×108, `preserveDrawingBuffer`) renders each look × team once: a fresh character from `makeCharacter`, idle at t=0, with the camera 0.8 m in front of the `Head` bone.
- The results are cached as data URLs, and cards show a placeholder icon until theirs is ready.
- The renderer is disposed afterwards, to keep a second GL context from lingering.

**Selection.**
- The store holds `selected` (a unit index). Clicking a unit mesh outside deployment selects it, and so does clicking a portrait; double-clicking a portrait asks the camera to centre on it.
- The info panel reads the latest snapshot. For an unseen Red unit it freezes on the last values seen, so it doesn't leak information.
- **Status text** comes from the flags, highest priority first: dead → staggered → dashing ("Dashing") → swinging ("Attacking") → kiting ("Falling back from melee") → moving ("Advancing"). Wet adds "wading, 70% speed" and slowed adds "slowed by frost"; otherwise "Holding".
- **Skills:** names for Blue come from the run's veterans, matched by index. Red shows a count only.

**RTS camera.**
- A port of the prototype's `RtsCamera` as an R3F component:
  - fixed 56° pitch, distance 10–45 m;
  - pan speed scales with zoom, clamped to the arena ±16 m;
  - edge pan within 14 px; follow uses exponential smoothing.
- It replaces OrbitControls. Because it doesn't react to dragging, unit dragging during deployment no longer conflicts with it, and the camera stays usable while deploying.

**How determinism is preserved and tested.**
- The only sim changes are read-only outputs. `SeenCells` gets a unit test: a rock blocks a cell behind it, and open ground within sight is seen.
- The golden replays and `sim.Version` stay unchanged (the checksum ignores the new outputs). `TestSnapshotLayout` pins 13 ints.

## Risks / Trade-offs

- [The fog grid costs CPU in the worker] → It's computed every 2 ticks only, and visible cells are cached per tick.
- [Portrait rendering creates a second WebGL context briefly] → It's disposed right after, and capped to the looks in use.
- [Edge panning while dragging a unit near the screen edge] → Accepted; it matches the RTS genre.
