# Proposal

## Why

Fights read through floating numbers and animations alone: a hit, a flanking blow, a Blast or a Blink look almost the same on screen, and dead units sink abruptly. This is step 3 of `docs/graphics-roadmap.md`: every sim event gets a visible cue, without hiding the fight or costing the frame-rate budget.

## What Changes

- **One pooled particle system:**
  - GPU-animated billboards in a fixed ring buffer, so nothing is allocated per frame;
  - one additive pool (sparks, glows) and one soft pool (dust, splashes).
- **Event cues** (from the sim's events, presentation only):
  - **hit:** sparks from the attacker's side; gold and larger on flank hits, orange on strong hits, violet from mages;
  - **execute:** a red burst and a small camera shake;
  - **death:** a dust puff, then the body fades into the ground instead of sinking abruptly;
  - **dash:** a dust burst and a dust trail while dashing;
  - **blink:** a violet flash;
  - **frost:** ice shards, then a faint shimmer while slowed;
  - **blast:** an orange burst and an expanding shockwave ring;
  - **shot:** a cast glow at the mage, and trails behind arrows and bolts.
- **Movement cues:** dust puffs at running units' feet, and splashes when they wade.
- **Camera:** a small shake on executes and a stronger one on the last kill of a battle.
- **Readability:**
  - effects only for units the player can see (fog of war);
  - spawning is rate-limited, and scaled down at 2× and 4× playback;
  - floating numbers stay on top;
  - all of it works on Low too, with fewer particles.

## Out of scope

- New sim events or rules; character changes (step 4).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-presentation`: visible cues for sim events.

## Impact

- `web/`: new `scene/Particles.tsx`, `scene/effects.ts` (event → effect) and `scene/shake.ts`; changes in `store.ts` (hands events over), `Units.tsx` (death fade, movement dust), `Shots.tsx` (trails) and `RtsCamera.tsx` (shake).
- No new assets: the particles are drawn in shaders.
