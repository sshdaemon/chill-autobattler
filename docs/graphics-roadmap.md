# Graphics roadmap

The goal is to make battles look **rich and believable without leaving the stylised look**: a Warcraft III-like, readable, top-down strategy view, not photorealism. Each step is its own OpenSpec change and PR, with screenshots before and after.

## Ground rules for every step

- **Presentation only.** Nothing visual may affect a battle. The sim, the golden replays and the server stay untouched.
- **Frame-rate budget:**
  - **High** quality: 60 fps on a mid-range laptop (integrated GPU from about 2022), with all 10 units fighting at 1×.
  - **Low** quality: 60 fps on older or weaker machines.
  - A quality setting (Low / High, remembered per browser) switches the expensive parts. The game drops to Low by itself if the frame rate stays low.
- **Assets:**
  - Licences: only CC0, or our own code-generated assets (Quaternius, Poly Haven, ambientCG, Kenney).
  - Files: committed with the code, compressed (KTX2 or WebP textures where it helps), and credited in the README.
  - Each step lists its download cost. The total target for a first visit is ≤ 30 MB.
- **Readability first:** units, teams, health and fog of war must stay at least as clear as today. Effects never hide the fight.

## Step 1: Lighting and post-processing (no new assets)

- **Sky and lighting:** a code-generated sky with matching image-based lighting, so the characters' materials get proper ambient light and soft highlights. It replaces the flat blue background and the hemisphere light.
- **Shadows:** soft shadows from a sun whose shadow box follows the camera, so they're sharp where you look instead of spread over the whole arena.
- **Post-processing** (`@react-three/postprocessing`):
  - ambient occlusion: tried (N8AO) and dropped in step 2; it cost 8–14 ms per frame on an Apple M3;
  - bloom: mage orbs, bolts, Blink and Frost glow;
  - colour grading and a light vignette;
  - an optional subtle tilt-shift depth of field, for the "battle table" feel.
- **Quality and fps:** Low / High in the dev panel, automatic fallback to Low, and an fps readout in dev.
- **Done when:** the scene reads clearly better in before/after screenshots, the High budget is met, and units stay as readable as before.

## Step 2: Ground, water and nature

- **Terrain:** a CC0 PBR-textured ground (grass blended with dirt, and mud near the ponds), with subtle variation so it doesn't tile visibly.
- **Grass:** instanced grass that sways in the wind, kept thin on paths and under units. It's switched off on Low.
- **Water:** a real water surface with reflection, ripples, a soft shore, and a ripple ring where a unit wades.
- **Obstacle models:** rocks and trees from CC0 packs in the same style as the characters (Quaternius nature), replacing the dodecahedrons and cones. Their footprint stays exactly the sim's circles, so what you see still matches what blocks.
- **Arena dressing:** a stone wall around the arena, a few banners and torches, and props outside the arena to frame it.
- **Done when:** the battlefield looks like a place. The extra download is ≤ 10 MB compressed and the budget still holds.

## Step 3: Combat effects

- **Particles:**
  - hit sparks, plus a bright one for flank and execute hits;
  - dust when units run and dash, and a splash when they wade;
  - a cast glow on mages, arrow trails, a Frost shimmer, Blink flashes and a Blast shockwave.
- **Death:** a short fade into the ground instead of a sudden sink.
- **Camera:** a small shake on executes, and on the last kill of a battle.
- **Readability:** effects are pooled (no per-frame allocation) and rate-limited at 4× playback. Floating numbers stay on top.
- **Done when:** every sim event has a visible cue, the budget holds in the busiest fights, and the effects never hide units.

## Step 4: Characters

- **Materials:** a rim light per team colour so units pop against the ground, better texture filtering, and a subtle outline on the selected unit.
- **Animation:** smoother blending, upper-body aiming toward the target, a hit flinch on the hit side, and a backward jog while kiting.
- **Variety:** more looks per class (more Quaternius parts and colours), and per-team trims such as banners, shoulder pads or a cloak colour.
- **Optional, a separate decision:** richer characters, either a paid stylised pack or commissioned models, judged against download size, frame rate and style consistency.
- **Done when:** the class, team and state of any unit read at a glance at normal zoom.
