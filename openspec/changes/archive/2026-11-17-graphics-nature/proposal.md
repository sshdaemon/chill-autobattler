# Proposal

## Why

After step 1 the light is right, but the battlefield is still a flat green square with blue discs for ponds, grey dodecahedrons for rocks and cones for trees. Step 2 of `docs/graphics-roadmap.md` makes it look like a place, in the same Quaternius style as the characters.

## What Changes

- **Ground.** CC0 ambientCG textures (Grass004, Ground037 dirt and Ground054 mud), blended in a shader:
  - dirt patches from world-space noise;
  - mud around the ponds;
  - two texture scales mixed, so the tiling doesn't show.

  The ground extends past the arena, so the field sits in a landscape.
- **Grass (High only).** Instanced grass blades that sway in the wind. There's none on ponds, rocks and trees, and it's thinner in the deployment zones. Grass under fog of war is darkened like the ground.
- **Water.** One water surface draws all ponds from the sim's circles, so overlapping circles read as one pond. It has:
  - a darker centre and a soft shore;
  - moving ripples and a sky reflection with a sun glint;
  - a ripple ring around each unit that wades (the sim's wet flag).
- **Obstacles.** Quaternius Stylized Nature MegaKit rocks (CC0) and EZ-Tree trees (MIT, baked offline) replace the primitives:
  - the variant and rotation are picked from the obstacle's position, so a map always looks the same;
  - each model is scaled so its footprint matches the sim's circle. What you see still matches what blocks.
- **Arena dressing:**
  - a low stone wall made of the pack's rocks;
  - a banner per team at its deployment edge;
  - torches with glowing flames;
  - trees, bushes and rocks outside the arena.
- **Assets:**
  - `tools/assets/pack_nature.py` picks the models and textures, resizes and compresses them, and writes `web/public/assets/nature/`;
  - the source packs stay outside the repo;
  - the README credits Quaternius and ambientCG.

## Out of scope

- Particles and combat effects (step 3), and character changes (step 4).
- Anything in the sim or on the server. Obstacle footprints, the map and the battle stay exactly the same.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-presentation`: the battlefield's look, and the rule that obstacle models match the sim's footprints.

## Impact

- `web/`: new `scene/Ground.tsx`, `scene/Grass.tsx`, `scene/Water.tsx`, `scene/Nature.tsx` (obstacles and dressing) and `scene/nature.ts` (loading); `Arena.tsx` is rewritten.
- `tools/assets/pack_nature.py` (new), and about 3–6 MB of new assets in `web/public/assets/nature/`.
