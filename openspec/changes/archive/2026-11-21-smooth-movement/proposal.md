# Proposal

## Why

Units sometimes get stuck or shake on the spot instead of walking a slightly
longer way. Measured over 600 varied battles (every faction, formation and
plan): 150 battles (25%) had a unit that wanted to move for 1.5 s or more but
got less than half a metre, 12,717 such unit-ticks in all, and 3.45% of all
walking ticks reversed the previous tick's step (visible shaking).

Tracing the worst cases showed that rocks were rarely the cause. Units flip
between two decisions at a boundary, one tick each way:

- **Bodyguards and front-line followers wait for each other.** A druzhynnyk
  turns back when it gets 4 m from its archer; under Shield wall the archer
  only follows the melee once that is 4.5 m ahead. Neither moves far enough
  for the other (61% of the stuck episodes).
- **A bodyguard flips at the edge of its reach** (4 m from its ward): back to
  the ward one tick, onward the next.
- **A wounded unit flips at the edge of its sight:** in sight of its target it
  falls back, out of sight it walks back in (most of the huslyar's episodes).
- **A unit keeping its distance flips** at Keep less half a metre: stepping
  back one tick, closing in the next.
- **Paths aren't tied to their goal:** when a unit's goal changes, it keeps
  the old path's waypoints and only swaps the end point.
- **Goals a unit can't reach:** a fall-back spot past the arena edge leaves
  the unit sliding along the wall.

## What Changes

- A bodyguard doesn't wait for a ward that itself follows the front line.
- A bodyguard that turns back heads for its ward until it is within half its
  reach, not just back inside it.
- A wounded unit that has fallen back out of its target's sight waits there,
  facing it, instead of walking back in (while it has allies).
- A unit that starts stepping back to keep its distance goes on until it is
  out at that distance (while its attack is cooling down).
- A path is found again when its goal moves more than 1 m from where it was
  found.
- A movement goal outside the arena or inside a rock or tree is moved to the
  nearest spot a unit can stand on.
- A regression test plays 600 battles and checks shaking and stalls stay low.
- **BREAKING (rules):** `RulesVersion` 0.12.0; goldens re-recorded.

Out of scope:

- Units steering round each other (paths still ignore other units; units
  push apart as before).
- Archers kiting: backing off from melee and turning to shoot is intended.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-rules`: steady movement (no flipping at boundaries, paths that follow
  their goal, reachable goals), bodyguards and followers, wounded units
  waiting out of sight.

## Impact

- `sim/unit.go`, `sim/temperament.go`, `sim/grid.go`, `sim/data.go`.
- `sim/movement_test.go` (new), golden replays, `RulesVersion`.
- Balance: Rus units (druzhynnyk, huslyar) and units that keep their distance
  behave differently; checked with `pnpm balance`.
