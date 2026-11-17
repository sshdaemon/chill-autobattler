# Tasks

## 1. Foundations

- [x] 1.1 Add `sim/data.go` with STATS, MATCHUP, SKILLS, terrain and timing constants as `Fx` and tick counts; verify a test asserts the table matches docs/GAME_CONCEPT.md §4 and §10.1
- [x] 1.2 Extend `sim.Fx` with vector helpers (add, scale, normalise, dot) and a fixed direction table; verify unit tests including squared-distance bounds at the arena corners
- [x] 1.3 Add `sim/rng.go` sub-streams via `mix(seed, tag)`; verify the same seed and tag give the same sequence and different tags differ
- [x] 1.4 Add the import lint test (no `math`, `math/rand`, `time`, `sync` in sim/); verify it fails when such an import is added

## 2. Map

- [x] 2.1 Port `randomObstacles` (ponds, rocks, trees, clearance rules) on the map sub-stream; verify the "same seed, same map" test
- [x] 2.2 Build the blocked, wet and opaque grids, plus `lineOfSight`, `clearView` and `nearestFree`; verify table tests on a hand-made map
- [x] 2.3 Verify the "always a way through" test over 1,000 seeds (flood fill from Blue's zone to Red's)

## 3. Movement

- [x] 3.1 Port A* (integer costs, wet ×10/7, heap with stable ties) and string-pulling; verify the pond-detour and wade-across tests
- [x] 3.2 Unit movement: facing turn rate, slow-turn rule, water and frost speed, separation, push-out and arena clamp; verify tests for each

## 4. Combat

- [x] 4.1 Team vision, nearest-target selection, re-targeting every 10 ticks, investigate and advance/sweep; verify the "spotted by a teammate" and "nothing in sight" tests
- [x] 4.2 Swing, wind-up and cooldown with jitter from the combat stream; melee hit timing; projectiles with homing and time of flight; verify the "target dies before the arrow lands" test
- [x] 4.3 Damage formula with matchup and flank (dot-product cone), and stagger; verify the flanked-archer and stagger tests
- [x] 4.4 Archer kiting (flee vector, bursts and rests, cornered); verify the kiting tests
- [x] 4.5 All 15 skills (stat skills at spawn; dash, blink, lob, execute, headshot, frost, blast); verify one test per skill
- [x] 4.6 Battle end: win, mutual-wipe draw, tick limit; damage stops after the end; verify the tests

## 5. Armies and input

- [x] 5.1 Formation presets and slot snapping; verify a test for every preset
- [x] 5.2 Seeded Red army, formation and skill spread; verify the "matching skills" test
- [x] 5.3 Add `BattleInput` and `BattleOutcome` to `proto/chill/v1`, `pnpm gen`, and input validation (zones, 5 units, unique skills); verify `pnpm lint:proto` and validation tests
- [x] 5.4 Replace the toy battle with `sim.New(input)`, `Step`, `Outcome` and the checksum; bump `sim.Version` to `0.2.0`

## 6. Golden replays

- [x] 6.1 Write about 20 golden inputs to `sim/testdata/golden.json` (all formations, every skill, a water-heavy map, mutual wipe, tick limit), with an `-update` flag and a version guard; verify `pnpm test:go` and `pnpm test:wasm` both pass and match. (18 cases. No valid input reaches the tick limit any more (0 of 3,000 seeds), so the limit is covered by `TestBattleEnd`; the two former stalemate seeds are kept as regression cases.)
- [x] 6.2 Add a determinism fuzz test (same input played twice gives an identical outcome); verify it runs for 30 s clean
- [x] 6.3 Add a benchmark; verify a full golden battle runs natively in under 50 ms

## 7. Server and browser

- [x] 7.1 Change `VerifyBattle` to take `BattleInput` plus the claimed outcome; verify server tests for honest, tampered-claim, tampered-input and unknown-version cases
- [ ] 7.2 WASM API: `create(bytes)`, `step(n)` returning snapshots and events, `outcome()`; verify a WASM test steps a golden input to the golden outcome
- [ ] 7.3 Worker loop at 20 Hz with 1×/2×/4× speed, and interpolated R3F debug viewer (obstacles, capsules, projectiles, health bars); verify in the browser that a battle plays to the end and "Verify on server" says fair
- [ ] 7.4 Update CLAUDE.md and README for the new API and golden workflow; verify `pnpm test` passes

## Workflow follow-up

- Archive the change once implemented and reviewed (`/openspec-archive-change`).
- Next change: the server issues seeds and records battles before they start.
