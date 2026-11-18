# Proposal

## Why

Units' legs don't match how fast they move. Every unit plays the same jog clip at a fixed rate, but on our characters that clip covers 6.2 m/s. Units move at 3.0–3.6 m/s, with Swift, water, frost and sharp turns changing that from moment to moment. Measured, a planted foot slid at 44–550% of the unit's speed depending on the situation, and the mismatch changed whenever a unit sped up or slowed down.

## What Changes

- **Move clips record their natural speed.** Each move clip in `looks.json` records the ground speed it covers at 1×, measured on our characters (the speed of a planted foot relative to the body): walk 1.04 m/s, jog 6.24 m/s, sprint 9.57 m/s. Looks name a walk, a move (jog) and optionally a sprint clip.
- **Legs follow the unit's real ground speed.** A moving unit's speed is taken from the sim's positions between snapshots, smoothed. Walk and jog are blended and kept in step, so that the stride exactly covers that speed, with the blend chosen so the cadence stays natural.
  - Units smaller or larger than the base body (builds, height variety) scale their stride with their size.
  - Dashes without a skill animation use the sprint clip at the dash's speed, within a playback limit.
  - Kiting archers play the same blend backwards.
- **The look check** requires every move clip a look names to record its speed.

## Out of scope

- Sim speeds and rules (no rule change, `sim.Version` unchanged).
- Turning-in-place or strafing clips, foot IK, and root motion.
- Skill action clips (Dash charge, Blink landing), which keep their own rates.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `battle-presentation`: "Animation follows the sim": moving units' legs match their ground speed.
- `unit-looks`: adds that move clips record their speed and looks name walk, move and sprint clips.

## Impact

- `web/src/units/looks.json` (clip speeds, walk and sprint per look) and `looks.ts` types;
- `web/src/scene/Units.tsx` (the locomotion blend);
- `web/public/assets/clips.json` (+ Walk_Loop and Sprint_Loop; within budget);
- `wasm/looks_test.go`;
- no `sim/` changes.
