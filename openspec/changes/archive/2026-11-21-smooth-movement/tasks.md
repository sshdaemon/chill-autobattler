# Tasks

## 1. Movement

- [x] 1.1 `Map.Reachable`; `moveToward` uses it and finds a new path when its goal moves more than `PathGoalSlack`
- [x] 1.2 Bodyguard hysteresis (`Unit.guarding`), and bodyguards skip wards that follow the front line
- [x] 1.3 Wounded units wait out of their target's sight
- [x] 1.4 Keep hysteresis (`Unit.keeping`)
- [x] 1.5 `RulesVersion` 0.12.0

## 2. Tests

- [x] 2.1 `TestMovementSmooth` (600 battles: reversals under 2%, stalls at most 1,500 unit-ticks; old code: 3.45%, 12,717)
- [x] 2.2 `TestReachable`, `TestGuardsAndFollowersAdvance`, `TestWoundedWaitsOutOfSight`, `TestKeepGoesOutToKeep`
- [x] 2.3 Golden replays re-recorded; new mutual-wipe seeds; `pnpm test` (native and WASM)

## 3. Balance

- [x] 3.1 `pnpm balance` against main; note changes in `docs/balance-findings.md`
