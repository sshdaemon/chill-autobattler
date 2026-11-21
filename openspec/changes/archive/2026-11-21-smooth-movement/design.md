# Design

## Context

Movement is `moveToward` (a straight leg when the way is clear, else an A*
path on the 0.5 m grid, found again every 0.5 s) plus per-tick decisions in
`idle` and `behave` (`sim/temperament.go`). Each decision compares a distance
to a fixed edge, every tick, with nothing remembered, so a unit near an edge
alternates. The measurement tool that found this is the new
`TestMovementSmooth`: it counts step reversals (a step pointing back against
the previous one) and stalled unit-ticks (wanting to move for 1.5 s, less than
0.5 m of progress) over 600 battles.

## Goals / Non-Goals

**Goals:** remove the flips and the mutual wait; keep every unit's intended
behaviour (bodyguards guard, wounded units fall back, hit-and-run units step
back); deterministic, with the outcome change shown by the goldens.

**Non-Goals:** local steering round other units; changing kiting.

## Decisions

- **Hysteresis with a remembered flag**, per behaviour: `Unit.guarding`
  (set past `StayReach`, cleared within `StayReach/2`), `Unit.keeping` (set
  below `Keep − 0.5 m`, held until `Keep` or the cooldown ends). The
  alternative, wider fixed edges, only moves the flip elsewhere.
- **Wounded units wait out of sight** rather than a wider fall-back edge: with
  any edge, a unit that walks back in reaches it again and loops. Waiting is
  what "falls back behind its allies" means; it still fights once it has no
  allies, so no stall.
- **Bodyguards skip wards that follow the melee** (`StayNear` melee). The
  guard and the follower each wait for the other otherwise; the follower
  already comes along.
- **Paths keyed to their goal:** `Unit.pathGoal` is where the path was found
  for; a goal more than `PathGoalSlack` (1 m) away finds a new path at once.
  Small drift (a target walking) still only moves the last leg.
- **`Map.Reachable`** clamps a goal into the arena and moves one inside a rock
  or tree to the nearest free cell; `moveToward` applies it to every goal.

## Determinism

All new state is integer/Fx on the unit, read and written in the existing
think order. Golden replays (native and WASM via `pnpm test`) are re-recorded
under `RulesVersion` 0.12.0; the mutual-wipe seeds become 1191 and 1869.
`TestPlansFight` (no stalls for any plan pair and faction pair) still passes.
`TestMovementSmooth`, `TestGuardsAndFollowersAdvance` and
`TestWoundedWaitsOutOfSight` fail on the old code and pass on the new.

## Risks / Trade-offs

- **Balance:** the Rus units change most (druzhynnyk advance instead of
  pacing; huslyar wait when wounded). Checked against main with
  `pnpm balance`.
- **More path searches** when goals jump; the regression battles run at the
  same speed.
