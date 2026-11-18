# Proposal

## Why

Units look generic and "small and skinny". Every unit is one of 8 pre-baked looks:
- built from only two outfits (peasant, ranger), all on the same slim body;
- mages are peasants with a purple tint and a code-made cone hat;
- weapons are code-made boxes and a torus.

Making a new look means editing a Blender script and re-baking, and each look ships its own textures, about 17 MB for 8 looks. This is step 3 of the unique-units roadmap. It replaces fixed looks with a kit: body builds, outfit parts, hair, weapons and team colours assembled from a recipe. Each unit gets its own silhouette, and a future `/new-unit` (step 4) can describe a new unit's appearance in data.

## What Changes

- **Parts library.** Bake Quaternius's "Modular Character Outfits – Fantasy" (12 outfits, 62 parts, 3 colour variants each), the "Universal Base Characters" heads and hairstyles, and a set of CC0 weapon and shield models into a shared parts library. Every part is skinned to our rig (or attached to a hand bone) with a manifest, and textures are reduced to a download budget.
- **Recipes instead of baked looks.** A look lists one or more recipes. A recipe gives:
  - a body build;
  - the outfit parts for each slot (head, hair, torso, arms, legs, feet, accessories);
  - a colour variant;
  - the main-hand weapon and an optional off-hand item (shield, quiver);
  - a team-colour strength.
  
  The client assembles each unit from the shared parts at runtime, so a new combination needs no re-bake.
- **Body builds.** Recipes pick a build (`slim`, `normal`, `bulky`, `stout`) that shapes the skeleton: shoulder width, chest, arm and leg thickness, head size and height. It applies on top of animations, within limits that keep the unit inside the sim's footprint and its feet on the ground.
- **Team colours from masks.** The bake step derives a mask for each outfit texture from where its three colour variants differ, which is the dyeable cloth. A shader recolours that area in the team colour. This replaces both the separate Red textures and the code-made trims.
- **Real weapons and shields.** These are modelled parts attached to hand bones, replacing the code-made sword, bow and staff. Recipes may add an off-hand shield or a back item.
- **The current classes as recipes.**
  - Warriors: a knight-like outfit on a bulky build, with sword and shield.
  - Archers: a ranger outfit on a slim build, with a quiver.
  - Mages: a wizard outfit with a robe and hat, holding a staff.
  
  Each class gets two or three recipes for variety, and hair and tint variety stay seeded as today.
- **Portraits** use the assembled characters.
- **Look check extended.** Every part, build, weapon and colour variant a recipe names must exist in the parts manifest, and the parts library must stay within its budget.

## Out of scope

- Any sim or rule change. `sim.Version` stays 0.4.0, and unit footprints in the sim are unchanged.
- New animations, or changes to the timing logic from step 2. Existing clips must keep working on every build.
- The `/new-unit` workflow (step 4), and creatures or non-humanoid rigs.
- A character editor UI for players.
- A bigger unit scale or a closer camera (tried before and rejected). Builds change proportions, not overall scale beyond a few percent.

## Capabilities

### New Capabilities
None: recipes, parts and builds extend `unit-looks`.

### Modified Capabilities
- `unit-looks`: "Every roster entry has a look" now holds recipes instead of model names. New requirements cover recipes, body builds, team-colour masks, weapons as parts, and the parts budget and check.
- `battle-presentation`:
  - "Units look like their class and team": units are assembled from recipes, with modelled weapons and team colours from masks.
  - "Units read at a glance": team-coloured cloth replaces code-made trims, and variety now includes build and outfit.

## Impact

- **Tools:**
  - new `tools/blender/bake_parts.py`, which exports every part, head, hair and weapon skinned to our rig, writes `parts.json` and the team masks, and reduces textures;
  - `tools/assets/pack_units.py` and `bake_units.py` retire after the switch;
  - source packs go in `~/fun/assets-src`, outside the repo.
- **Assets:** `web/public/assets/units/` (17 MB, 8 looks) is replaced by `web/public/assets/parts/` with a budget of 10 MB total and 3 MB on first load for the parts a default army uses.
- **`web/`:**
  - `scene/models.ts` (assembly from recipes, builds, mask shader, weapon attachment);
  - `scene/portraits.ts`;
  - `units/looks.json` and `looks.ts` (recipes);
  - `Units.tsx` (unchanged logic, new characters).
- **`wasm/looks_test.go`:** the recipe and parts checks.
- **Docs:** CLAUDE.md and the README.
- **No changes to** `sim/`, the proto or the snapshot.
