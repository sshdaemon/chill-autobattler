# Design

## Context

See proposal.md for the motivation. Today:

- **`tools/blender/bake_units.py`** builds 8 fixed looks from an older copy of two Quaternius outfits (Peasant, Ranger) plus Superhero heads and hair:
  - it cuts the head mesh at the neck and re-skins everything to the outfit's armature;
  - `restore_rest()` puts the original joint transforms back so the shared clips fit;
  - it exports one glTF per look and team, with Red using alternative textures.
- **`pack_units.py`** shares Blue and Red geometry. Total: about 17 MB.
- **`models.ts`:**
  - loads a look;
  - clones the skinned scene per unit;
  - gives it its own materials (rim light, hair and tint variety, death fade);
  - adds code-made weapons and team trims on bones.
- **Animation:** clips (step 2) animate bone rotations, the pelvis position and the root rotation only. There are no scale tracks and no other position tracks.

**Constraints:**
- presentation only, no sim change;
- the rig must stay the UAL rig, so every clip keeps working;
- the download budget;
- about 10 units on screen at around 60 fps on the High quality setting.

## Goals / Non-Goals

**Goals:**
- Any recipe combination works without a re-bake.
- Builds change silhouettes without breaking animations.
- One set of textures for both teams.
- A smaller download than today.

**Non-Goals:**
- Per-player customisation UI.
- New rigs.
- KTX2 or Basis texture compression. That's a possible later step; JPEG at reduced size is enough for the budget.

## Decisions

### 1. Assemble at runtime from a baked parts library
`bake_parts.py` exports every outfit part, head, hairstyle and item once, as its own glTF, to `web/public/assets/parts/`. It also writes `parts.json`, which records for each part:
- its slot;
- its file;
- its material (outfit texture set);
- the colour variants it has;
- whether it is skinned or rigid.

The client builds a unit by cloning one canonical skeleton, then binding each part's skinned mesh to it by bone name.

- **Rejected: baking every recipe combination.** That means a Blender run and a new asset per recipe, the very thing step 4 needs to avoid.
- **Canonical joint list:** every skinned part is exported with the same joint list in the same order, re-skinned to the canonical armature in Blender. `restore_rest()` from today's bake is kept for the rest-pose fix. With a shared joint list, binding is direct and parts that share a material can be merged (§6).

### 2. Builds: bone offsets and compensated uniform scales
Clips never write bone scale and only write the pelvis position, so a build can safely change the skeleton's bone positions and scales.

| Build change | How |
|---|---|
| Shoulder width | scale the clavicles' x offsets |
| Chest | uniform scale on `spine_03`, with its children's offsets kept and their scale compensated |
| Limb girth | uniform scale `s` on upperarm, lowerarm, thigh and calf; each child gets `1/s` scale and its original offset, so the limb gets thicker but not longer |
| Head | uniform scale on `Head` |
| Height | the root's uniform scale, folded into today's ±4% variety, at most ±8% in total |

Only *uniform* scales are used, compensated at the child, so there is no shear.

Leg lengths don't change, so feet stay on the ground in every clip; the pelvis height in clips stays valid. A dev check in the spike measures the lowest foot point over every clip for each build.

Builds are data (a small table in `looks.json`: `slim`, `normal`, `bulky`, `stout`), and the look check limits their factors.

- **Rejected: separate base bodies per build.** The Superhero, Regular and Teen bodies would need outfits fitted to each, which the packs don't provide.

### 3. Team-colour masks from colour variants
Each outfit ships 3 colour variants of one base-colour texture. In those variants, only the dyeable cloth changes.

`bake_parts.py` writes a mask per outfit:
- `m = clamp(maxdiff(variantA, variantB, variantC) · k)`;
- slightly blurred;
- 256 px, greyscale JPEG.

In the shader, added next to the rim-light hook: `color = mix(base, teamColor · luma(base) · 1.6, mask · strength)`. This keeps the cloth's shading and pattern and changes only its hue.

The recipe's colour variant picks the base texture, and both teams share it. This removes the Red alternative textures and the code-made trims; the rim light stays.

- **Rejected: hand-painted masks.** That's 12 outfits of manual work, and it doesn't scale to new outfits.

### 4. Weapons and items
Items are rigid parts with an attach bone and an offset in `parts.json`:
- main hand → `hand_r` (bow → `hand_l`);
- off hand → `lowerarm_l` (shield);
- back → `spine_03` (quiver, sheath).

**Source:** CC0 low-poly weapons, Quaternius or KayKit, chosen during intake for style match. Grip offsets are measured once per item in the spike and stored in `parts.json`.

The code-made sword, bow and staff, and `addTrims()`, are removed.

### 5. Recipes in looks.json
```json
"warrior": {
  "recipes": [
    { "build": "bulky", "head": "Head_Superhero_M", "hair": ["Hair_Short", "Hair_Beard"],
      "outfit": { "torso": "Knight_Torso", "arms": "Knight_Arms", "legs": "Knight_Legs", "feet": "Knight_Feet", "extra": ["Knight_Helmet"] },
      "variant": 1, "main": "Sword_A", "off": "Shield_Round", "team": 0.9 },
    ...
  ],
  "anims": { ... }, "skills": { ... }
}
```
`models: [...]` becomes `recipes: [...]`. The n-th unit of a class picks recipe n mod the list, as today's model choice does. Hair colour, tint and the height variety stay seeded per unit.

### 6. Draw calls and loading
- **Merging:** after assembly, skinned parts that share a material (one texture set per outfit variant) are merged with `mergeGeometries`, which works because the joint list is shared. The result is about 3–4 draw calls per unit (outfit, head and skin, hair, items) instead of 8–10.
- **Loading:** parts load lazily through one cached GLTF loader, per recipe actually used. Until a unit's parts arrive, it renders as today's capsule.
- **Portraits:** `portraits.ts` renders the assembled character, after assembly.

### 7. Budget
- **Textures:** 1024 px for outfits, 512 px for hair and items, JPEG q85; normal and ORM maps kept at 512 px or dropped where the look test shows no difference.
- **Limits:** the library at most 10 MB, and the default army's parts at most 3 MB. `looks_test.go` sums the files of every part named by any recipe and checks the total.

### 8. Sources
The free Standard editions go into `~/fun/assets-src/outfits`, `basechars` and `items`, each with its licence file. The user is asked before each download, with file names and sizes: outfits 280 MB zip, base characters 122 MB zip, items to be named during intake.

## Determinism

The sim is unchanged, so `sim.Version`, golden replays and the snapshot layout stay the same. `pnpm test` must pass untouched, and a battle in the browser must still be verified by the server.

## Risks / Trade-offs

- **[Risk]** v2 outfit parts may not share bone names or rest pose with our rig, in spite of the "compatible" claim. → Mitigation: task 1 is a spike that assembles one outfit on our skeleton and plays every clip. Stop and report if a real retarget is needed.
- **[Risk]** Build scaling shows seams between parts (for example a helmet over a scaled head) or clips through clothing. → Mitigation: items attach to bones and scale with them. Builds are kept moderate, and the spike checks each build with each outfit's widest clip.
- **[Risk]** A mask from variant differences catches non-cloth details that also change between variants. → Mitigation: there's a per-outfit threshold in the bake, and masks are reviewed visually. A recipe's `team` strength can soften an outfit.
- **[Risk]** Download size. → Mitigation: the budget test, lazy loading per recipe, and merged textures.
- **[Trade-off]** Runtime assembly costs some milliseconds per unit at spawn (cloning plus merging). This happens once per battle and is hidden behind the existing capsule fallback.

## Migration Plan

The old assets (`web/public/assets/units/`, `bake_units.py` and `pack_units.py`) are removed in the same PR once all three classes use recipes. To roll back, revert the PR.

## Open Questions

- Exact recipe contents (which outfit for each class, which hair) are chosen during intake from what the packs contain, with screenshots in the PR. This doesn't change the approach or the tasks.
