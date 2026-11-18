# Tasks

## 1. Data
- [x] 1.1 Add Walk_Loop and Sprint_Loop to `looks.json` clips and record `speed` (walk 1.04, jog 6.24, sprint 9.57, measured on our characters) on all three; give each look `walk` and `sprint`. Verify `clips.json` stays within budget.
- [x] 1.2 Extend `wasm/looks_test.go`: walk, move and sprint clips must have a speed, with a failing case. Verify with `pnpm test:wasm`.

## 2. Locomotion
- [x] 2.1 In `Units.tsx`, measure each unit's smoothed ground speed and drive the walk–jog blend (cycle sync, weights, size scaling, kiting reversed, sprint for plain dashes, crossfades in and out). Verify with the typecheck.
- [x] 2.2 Measure foot sliding: an offline check driving characters at the sim's speeds (dry, water, frozen, Swift, slow turns) through the same blend, showing a planted foot's speed over the ground within 10% of the unit's speed. Verify, and look at moving units in a battle at 1×.

## 3. Wrap-up
- [x] 3.1 Run `pnpm test`, play a battle verified by the server, and update the docs (CLAUDE.md animation notes). Verify all pass.
