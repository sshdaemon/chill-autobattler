# Design

## Asset pipeline

The source packs live outside the repo, in `../assets-src/quaternius-nature` and `../assets-src/ambientcg`; `NATURE_DIR` and `AMBIENTCG_DIR` override the locations. `tools/assets/pack_nature.py`:

1. **Models:** runs Blender headless to import the chosen glTF models and export one `nature.glb`, with each model as a named root node.
   - The textures are downscaled first: bark to 512 px, leaves and rocks to 512 or 1024 px.
   - Leaf textures stay PNG (alpha); the rest are JPEG.
2. **Ground textures:** colour and normal maps at 1024 px, saved as JPEG (quality 82).

Textures for leaves, bark (with its normal map) and rocks are kept at 1024 px, so trees and rocks hold up close; the total is about 7 MB, within the roadmap's ≤ 10 MB for this step.

## Ground

- **The plane:** one 120 m `MeshStandardMaterial` plane with `onBeforeCompile`.
- **Texture mixing:**
  - Grass and dirt are sampled in world space at two scales. A value-noise function mixes the scales, which hides tiling.
  - A second noise mixes in dirt.
  - Mud comes in by the signed distance to the nearest pond circle, from up to 16 circles passed as a uniform.
- **Normal maps:** blended with the same weights.
- **Outside the arena:** past the arena edge the ground stays grass with a little more dirt, and the scene fog fades it.

## Water

- **One surface for all ponds:** a single plane over the arena. Its fragment shader computes `d`, the signed distance to the union of the pond circles, and discards where `d > 0`.
  - Alpha and colour come from `-d`: a light shallow rim, a darker centre, and a foam line at the shore.
  - Normals come from two scrolling noise layers.
  - Reflection is a Fresnel mix with the sky colour, plus a sun specular highlight.
- **Wading rings:** the positions of wet units (at most 10) go in as a uniform each frame. Each wet unit adds an expanding ring.
- **Fog:** the fog overlay above still darkens it.

## Obstacles and dressing

- **Rock:**
  - One of `Rock_Medium_1..3`.
  - Scale: the model's XZ bounding radius maps to `r`, slightly under (0.95), so the visible rock never pokes outside the blocked area by much.
  - Height is kept proportional.
- **Tree:**
  - One of `CommonTree_1..5` or `Pine_1..5`.
  - The trunk sits at the centre; the sim's 0.8 m circle is the trunk area. Trees are scaled to a fixed height range of 4.5–6 m.
- **Picking variants:** the variant, rotation and height come from a hash of the integer position. They're presentation-only and stable for a map.
- **Wall:** small rocks placed every 0.9 m along the 28 m border, slightly jittered, as instanced meshes per variant.
- **Outside:** about 40 trees, 20 bushes and 20 rocks scattered in a ring 16–40 m from the centre, from a fixed seed.
- **Banners:** a pole with a waving cloth, using a vertex sway. Blue's is at +z and Red's at −z.
- **Torches:** at the four corners. The flame is an emissive cone in HDR colour so it blooms, with a small flicker. There are no real point lights, because of the budget.

## Grass

- **Blades:** a tapered 5-vertex blade (3 triangles), instanced about 18,000 times on High and off on Low.
- **Placement:** from a fixed seed over the arena plus a 6 m margin. Rejected:
  - inside ponds, rocks and trees (with a 0.2 m margin);
  - 60% of the blades in the deployment zones.
- **Material:** `MeshStandardMaterial` with `onBeforeCompile`:
  - wind sway in the vertex shader, scaled by blade height;
  - a colour gradient from root to tip;
  - the fog texture sampled in world XZ, darkening the same way as the overlay.
- **Shadows:** grass receives shadows but doesn't cast them, which saves budget.

## Fog of war

`Fog.tsx` exports its `DataTexture` so grass can darken itself. The overlay plane stays as it is for the ground and water.

## Budget

On High: about 18,000 blades (54k triangles); obstacles ~5k triangles each, at most about 12 in the arena plus 60 outside; the wall is instanced. Draw calls stay under about 150.

## Readability and tone

- Health bars and unit rings ignore depth and draw last, so trees, grass and water never hide them.
- Only Blue's deployment zone is shown, and only while deploying.
- The overall tone is darker: the sun, the sky light and the environment light are dimmed, and the ground, grass, rocks and pine needles are toned down to match.
- Leaves sway in the wind (a vertex shader on the leaf materials).
