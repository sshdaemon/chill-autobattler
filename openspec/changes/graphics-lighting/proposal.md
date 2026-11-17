# Proposal

## Why

Battles look flat: a single-colour sky, one hemisphere light, hard shadows spread over the whole arena, and no ambient occlusion or glow. Lighting and post-processing give the biggest visual improvement for the least effort, with no new assets. This is step 1 of `docs/graphics-roadmap.md`, which this change also adds.

## What Changes

- **Sky and image-based lighting.** A code-generated sky (three's `Sky` shader) replaces the flat background. A matching environment map for image-based lighting comes from the same sky plus soft light shapes, so there's nothing to download.
- **Sun and shadows.** One sun matches the sky. Its shadow camera follows the RTS camera's target, so shadows are sharp where you look, with softened edges on High.
- **Post-processing** with `@react-three/postprocessing` (High quality only):
  - ambient occlusion (N8AO);
  - bloom on emissive and un-tone-mapped things: mage orbs, bolts, Blink and Frost;
  - ACES tone mapping with a mild contrast and saturation grade, and a light vignette;
  - an optional subtle tilt-shift, off by default.
- **Quality setting.** Low / High in the dev panel, remembered per browser:
  - Low keeps today's look, with simpler shadows and no post-processing.
  - The game switches to Low automatically if the frame rate stays under 70% of the display's refresh rate (drei's `PerformanceMonitor`), and says so. A browser that simply caps the frame rate isn't downgraded.
  - The dev panel shows the current fps.
- **`docs/graphics-roadmap.md`**: the four-step plan, with its frame-rate, asset and readability rules.

## Out of scope

- New textures or models, water, grass, particles, character changes (steps 2–4).
- Anything in the sim or on the server.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-presentation`: adds a graphics quality setting with an automatic fallback.

## Impact

- `web/`:
  - new dependencies `@react-three/postprocessing` and `postprocessing`;
  - new files `scene/Lighting.tsx` and `scene/Effects.tsx`;
  - quality in the store and the dev panel;
  - the fog overlay keeps working under post-processing.
- `docs/graphics-roadmap.md` (new).
