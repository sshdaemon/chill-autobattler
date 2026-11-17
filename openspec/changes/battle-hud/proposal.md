# Proposal

## Why

Battles now look right, but the player still can't follow them like the concept describes (§6, §12):
- You see Red's whole army at all times, although vision is a core rule: sight ranges, shared team vision, and rocks blocking sight.
- There's no way to see who is who, how hurt your units are, or what a unit is doing.
- The camera is a free orbit, not the fixed-tilt RTS camera the prototype had.

## What Changes

- **Fog of war during the fight.**
  - Areas Blue can't see are darkened. Red units are drawn only while Blue's team vision sees them; their health bars hide too.
  - While deploying, only Blue's zone is lit.
  - The sim exports Blue's visibility grid (same sight and line-of-view rules as the battle) and a "seen by Blue" flag per unit, so the client never re-implements vision.
  - **V** toggles the fog for debugging.
- **Portraits.**
  - A face close-up per unit, rendered once per look and team.
  - Blue's portraits sit along the bottom-left and Red's along the bottom-right, each with a health bar, a number and the veteran ★.
  - Dead units are greyed with ✝. Red units Blue can't currently see show "?" and hidden health.
- **Selection and info panel.**
  - Clicking a portrait or a unit selects it, and double-clicking centres the camera on it.
  - The panel shows the name (its look), ★, class, health, a status line ("Attacking", "Falling back from melee", "Wading, 70% speed", "Slowed by frost", "Staggered", "Dead"), kills, damage dealt, the "takes from" matchup row, and its skills: names for Blue, a count for Red.
  - The snapshot gains kills and damage dealt per unit.
- **Warcraft III-style camera** replacing OrbitControls:
  - fixed tilt, pan with WASD, the arrows or the screen edge, zoom with the wheel;
  - **F** follows the selected unit, **Space** jumps to it;
  - the camera works while deploying too.

## Out of scope

- Fog for Red's side and any change to vision rules.
- Minimap, replays UI, sound.
- Unit names beyond the look names.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-presentation`: adds fog of war, portraits, unit selection and the info panel, and the RTS camera.

## Impact

- `sim/`: `Battle.SeenCells(team)` (presentation only) and the snapshot unit record gaining seen-by-Blue, kills and damage dealt (13 ints). No rule change; `sim.Version` stays.
- `wasm/`: `fog()` returns Blue's 56×56 visibility grid.
- `web/`: fog overlay, `Portraits` and `InfoPanel` UI, selection in the store, a `RtsCamera` component, keyboard shortcuts; OrbitControls removed.
