# Tasks

## 1. Assets

- [x] 1.1 `tools/assets/pack_nature.py` (Blender export of the chosen models plus downscaled textures) writes `web/public/assets/nature/`; verify the total size ≤ 10 MB, and credit Quaternius and ambientCG in the README

## 2. Ground and water

- [x] 2.1 Ground shader: grass, dirt and mud blend with anti-tiling, extending past the arena; verify no visible tiling at normal zoom, and mud around ponds
- [x] 2.2 Water: one surface for all ponds (union of circles), a soft shore, ripples, reflection, and rings around wading units; verify overlapping circles read as one pond and a wading unit makes rings

## 3. Nature

- [x] 3.1 Rocks and trees from the pack, matched to the sim footprints and stable per map; verify a unit walking around a rock touches its model edge, not empty space
- [x] 3.2 Grass (High only): instanced, with wind, no grass on obstacles or ponds, darkened by fog; verify the fog still reads and Low has no grass
- [x] 3.3 Dressing: stone wall, team banners, torches, and props outside the arena

## 4. Wrap-up

- [x] 4.1 Verify before/after screenshots, the fps budget (High equal to Low in the pane), `pnpm test`, the web build, and a full match verified fair
