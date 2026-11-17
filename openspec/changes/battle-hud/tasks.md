# Tasks

## 1. Sim and WASM outputs

- [x] 1.1 `Battle.SeenCells(team)`; snapshot `UFSeen` flag plus kills and damage dealt (13 ints); `wasm fog()`; verify a `SeenCells` test (rock shadow, open ground), `TestSnapshotLayout`, and that the golden replays are unchanged (native + WASM)

## 2. Fog of war

- [x] 2.1 The worker sends the fog grid every 2 ticks; the fog overlay texture; Red units and their bars drawn only when seen; Blue's zone lit while deploying; V toggles; verify in the browser

## 3. Portraits and info panel

- [x] 3.1 Portrait rendering (once per look and team) and portrait bars (health, number, ★, dead ✝, unseen "?"); verify in the browser
- [x] 3.2 Selection (unit or portrait click) and the info panel (status, kills, damage, takes-from, skills); verify in the browser that a selected unit's status updates

## 4. Camera

- [x] 4.1 RtsCamera (WASD, arrows, edge pan, wheel zoom, F follow, Space jump, double-click to centre) replacing OrbitControls; verify in the browser

## 5. Docs

- [x] 5.1 README controls and CLAUDE.md; verify `pnpm test` passes

## Workflow follow-up

- Archive the change once merged (auto-merge after green CI).
