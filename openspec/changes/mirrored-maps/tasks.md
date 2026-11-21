# Tasks

## 1. Maps

- [x] 1.1 `GenerateObstacles`: features in Red's half, reflected; centre-line rocks and trees once; joined ponds
- [x] 1.2 `Map.Mirrored`; Red's `findPath`, `freeSpot` and `reachable` in Blue's frame
- [x] 1.3 `RulesVersion` 0.13.0

## 2. Tests

- [x] 2.1 `TestMapRules` for the new counts and mirror images; `TestAlwaysAWayThrough` still passes
- [x] 2.2 `TestMirroredMapChoices`
- [x] 2.3 `TestGuardsAndFollowersAdvance` set up by hand (it was pinned to an old map)
- [x] 2.4 Goldens re-recorded; mutual-wipe seeds 12 and 703; `pnpm test`

## 3. Measure

- [x] 3.1 `tools/balance -exp maps` before and after; plan matrix still has no dominant plan
- [x] 3.2 Findings in `docs/balance-findings.md`

## 4. Client

- [x] 4.1 The game draws a mirrored map correctly (rebuild WASM, restart the server, check in the browser)
