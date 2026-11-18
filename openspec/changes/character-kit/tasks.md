# Tasks

## 1. Sources and spikes

- [ ] 1.1 Ask the user before downloading. Fetch the free editions of "Modular Character Outfits – Fantasy" (280 MB) and "Universal Base Characters" (122 MB), and choose a CC0 weapons and items pack, into `~/fun/assets-src/outfits`, `basechars` and `items`, with their licence files. Verify that every licence says CC0.
- [ ] 1.2 Assembly spike: export one outfit's parts, a base head and a hairstyle skinned to the canonical joint list. Assemble them on one skeleton in the browser and play every clip. Verify with screenshots of idle, an attack and death, and that bone names and rest pose match (or stop and report).
- [ ] 1.3 Builds spike: apply `slim`, `normal`, `bulky` and `stout` (§2 in design.md) to the spike character. Verify with screenshots side by side, and with a dev check that the lowest foot point stays within ±2 cm of the ground over every clip and the width stays within the unit footprint.
- [ ] 1.4 Mask spike: derive the cloth mask for that outfit from its 3 colour variants and recolour it in the shader. Verify with Blue and Red screenshots that skin, leather and metal are unchanged.

## 2. Parts library

- [ ] 2.1 Write `tools/blender/bake_parts.py`: every outfit part, head, hairstyle and item, with the canonical joint list (keeping `restore_rest`), reduced textures, team masks and `parts.json` (slot, file, material, variants, attach bone and offset). Verify that the bake runs headless and `parts.json` lists every part.
- [ ] 2.2 Measure grip offsets for each item and store them in `parts.json`. Verify with screenshots of each weapon in hand during idle and attack.
- [ ] 2.3 Pack into `web/public/assets/parts/`. Verify the total is at most 10 MB and the default army's parts at most 3 MB.

## 3. Assembly in the client

- [ ] 3.1 In `models.ts`, assemble a character from a recipe: one cloned canonical skeleton, parts bound by name, items attached, materials merged per texture set, mask and rim shader, hair and tint variety, the death fade. Verify with the typecheck and a battle in the browser.
- [ ] 3.2 Apply builds as data from `looks.json`, with the look check limiting their factors. Verify with screenshots of all builds mid-battle.
- [ ] 3.3 Switch `portraits.ts` to assembled characters. Verify the portrait strip shows faces for all classes.
- [ ] 3.4 Check draw calls and frame time with a full battle. Verify `renderer.info` shows at most 4 draw calls per unit and `benchFrame` is no worse than `main`.

## 4. Recipes

- [ ] 4.1 Write two or three recipes for each class (warrior: bulky, knight-like, sword and shield; archer: slim ranger, bow and quiver; mage: wizard robe and hat, staff), replacing `models` in `looks.json`. Verify with screenshots of all recipes on both teams, and a silhouette check of the three classes side by side.
- [ ] 4.2 Extend `wasm/looks_test.go`: recipes name existing parts, slots match, builds are within limits, and the parts budget holds. Add one failing case per rule. Verify with `pnpm test:wasm`.

## 5. Clean-up and docs

- [ ] 5.1 Remove `web/public/assets/units/`, `tools/blender/bake_units.py`, `tools/assets/pack_units.py`, the code-made weapons and `addTrims()`. Verify with the typecheck, and that nothing references them.
- [ ] 5.2 Update CLAUDE.md and the README (parts, recipes, builds, masks, budget, how to re-bake). Verify by reading them against the code.
- [ ] 5.3 Run `pnpm test` and play a battle with every class on both teams. Verify golden replays are untouched, the server reports the battle verified, and there are no console warnings.

## Workflow follow-up

- Open a PR for review. Merge it when the user says so, then archive the change in a separate PR (auto-merged after green CI).
