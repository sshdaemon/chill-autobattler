# Proposal

## Why

The battle rules exist only in the throwaway JS prototype (`../web3d-demo/battle.html`), where they are mixed with rendering, rely on floats, `Math.random` and Yuka, and run on the frame clock. The server can't check a result it can't reproduce. So before any other feature, the rules need to live in `sim/`, deterministic and identical in the browser (WASM) and on the server.

## What Changes

- Port the battle rules from the prototype into the Go `sim/` module, using fixed-point maths and a fixed 20 Hz tick:
  - Seeded map generation.
  - The three classes (warrior, archer, mage) with matchups and flanking.
  - Pathfinding on a 0.5 m grid, with water cost.
  - Spacing, shared team vision and line of fire.
  - Projectiles, archer kiting and all 15 skills.
  - Red's seeded army, formation and skill spread.
- Replace the toy dot battle (`sim.Version` `0.1.0-toy`) with the real rules. **BREAKING** for the scaffold's checksums and golden table.
- Define a **battle input** (seed, sim version, both armies with classes and skills, deployment positions) and a **battle outcome** (winner, ticks, survivors, checksum) as protobuf. The browser and the server exchange exactly these.
- Change `VerifyBattle` to take a full battle input plus the claimed outcome. **BREAKING** for the scaffold RPC.
- Expose a frame-by-frame state API from the WASM build, so the web app can render a battle while it plays.
- Add a minimal debug viewer in `web/`: capsules and the map's obstacles, driven by the WASM sim. It proves that the browser plays the same battle the server verifies.

## Out of scope

- The server issuing seeds and recording battles before they start; anti-cheat beyond re-running. Planned as the next change.
- Runs, veterans, skill picking after a win, scoring and the scoreboard.
- Army picker and deployment UI (formations are applied by the sim from an input preset).
- Real character rendering: baked Quaternius units, animations, fog-of-war visuals, portraits, floating numbers and the camera.
- Player actions during a battle (the concept allows none yet).
- Balance changes: prototype numbers are ported as they are, with differences only where fixed-point rounding forces them.

## Capabilities

### New Capabilities

- `battle-map`: seeded generation of the arena, rocks, trees and water ponds, deployment zones, and walkability, sight and shot rules per terrain.
- `battle-rules`: how a battle plays out tick by tick. Covers classes, stats and matchups, targeting and vision, movement and spacing, attacks and projectiles, flanking, kiting, skills, Red's army generation and formations, and the win condition.
- `battle-replay`: the battle input and outcome formats, determinism guarantees (the same input gives the same outcome everywhere), sim versioning, and server-side verification of a claimed outcome.

### Modified Capabilities

None. There are no existing specs.

## Impact

- `sim/`: most of the new code (map, pathfinding, rules, data tables). The toy battle is removed.
- `proto/chill/v1/`: new `BattleInput` and `BattleOutcome` messages, and a changed `VerifyBattleRequest`. Run `pnpm gen` again.
- `server/`: the verify handler takes the new input.
- `wasm/`: a new API (create battle, step, read state), replacing `run`.
- `web/`: the debug viewer replaces the placeholder scene.
- `sim/sim_test.go`: a new golden replay table, run both natively and under WASM.
