## Why

Roadmap Phase 1, step 1.1 (`docs/ROADMAP.md`): units level up to three
abilities instead of learning all five. A recruit starts at level 1 with one
ability, and a unit learns one more after each won battle, so builds differ
(three of five) and the draft (step 1.3) can include each unit's first ability.

## What Changes

- **BREAKING (rules):** every unit in a battle has 1 to 3 abilities (one per
  level). A recruit is level 1 with exactly one ability.
- Red's generated units start at level 1 too (one random ability each); Red's
  skill count now means abilities above level 1, matched to the veterans'
  abilities above level 1 (at most 10).
- After a win, only veterans below level 3 owe a pick.
- Until the draft exists, the client gives each recruit a random level-1
  ability (stable for the run, match and slot).
- `sim.RulesVersion` 0.6.0; golden replays re-recorded (new mutual-wipe seeds);
  the Red-army baseline re-recorded. The balance lab follows the new rules
  (recruit armies, runs capped at level 3, each ability checked as a level-1
  pick against the unit's average).

## Capabilities

### New Capabilities

### Modified Capabilities
- `runs`: recruits start with one ability; Red matched to abilities above level
  1; picks only below level 3.
- `battle-rules`: 1 to 3 abilities per unit; Red's units start at level 1.
- `battle-sessions`: Red's skill count is the veterans' abilities above level 1.

## Impact

`sim/army.go`, `sim/data.go` (`MaxLevel`), `server/run`, `web/src/store.ts`,
`tools/balance`, golden replays and the roster baseline, `docs/GAME_CONCEPT.md`.
