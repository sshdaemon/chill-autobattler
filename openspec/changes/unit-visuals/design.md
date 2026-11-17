# Design

## Context

- The scene draws units as capsules from interpolated snapshots (`web/src/scene/Units.tsx`), and projectiles as instanced spheres.
- The prototype's baked units are glTF + `.bin` + shared textures (about 30 MB), made by `tools/blender/bake_units.py`.
- Its animations come from one `clips.json` (12 clips, 1.4 MB) built for the shared Quaternius skeleton.

## Goals / Non-Goals

**Goals:** the prototype's look, with animation driven purely by sim snapshots and assets at about half the size (31 MB → 18 MB).

**Non-Goals:** fog of war, portraits, the camera, new art, and any rule change (see proposal).

## Decisions

**Asset packing.** A pack script (`tools/assets/pack_units.py`, Python stdlib) takes the bake output and writes `web/public/assets/units/`:
- **Shared geometry:** the Blue and Red `.bin` of a look are byte-identical (checked with md5), so it keeps one `<Look>.bin` and points both `.gltf` files at it.
- **Textures:** PNG roughness maps become JPEG (macOS `sips`; a dev-time tool only), and their uris are updated.
- **Clips:** `clips.json` is copied with only the clips used: Sword_Idle, Jog_Fwd_Loop, Sword_Attack, Spell_Simple_Shoot, Hit_Chest, Death01, Yes.
- **Committed normally:** the result is committed as ordinary files, not Git LFS. At about 18 MB (17.2 MB of units plus 0.7 MB of clips), that keeps cloning and CI simple.

**Look assignment.**
- Each class has a list of looks: warrior → Mirko, Vesna, Radim; archer → Bogdan, Lada; mage → Zoran, Milena, Velimir.
- The n-th unit of a class in a team (in snapshot order) gets look n modulo the list. The file is `<Look>_<blue|red>.gltf`.
- It's deterministic, so the same army looks the same after a reload.

**Loading and cloning.**
- Each look/team glTF is loaded once (a module cache of promises) and cloned per unit with `SkeletonUtils.clone`, so bones aren't shared.
- Until a model loads, the unit shows its old capsule.
- Loading happens in React suspense-free effects, so a slow load never blocks the sim worker. The battle keeps ticking (see the spec's "never changes the battle" requirement).

**Animation from snapshot flags.**
- Each unit has an `AnimationMixer` with actions built from the shared clips (`AnimationClip.parse`).
- Every frame, the wanted action is chosen in priority order: dead → Death01 (once, clamped); staggered → Hit_Chest (once); swinging → Sword_Attack or Spell_Simple_Shoot (once); dashing → Jog ×2.2; moving → Jog; otherwise Sword_Idle. When the battle is won, the winners' survivors play Yes on a loop.
- **Stretching the swing:** on the rising edge of the swinging flag, the attack action is restarted with `timeScale = clipLength / swingSeconds`. Swing seconds come from the sim's data export (`data().classes[c].swing`, added here), divided by the playback speed, so the blow lands with the sim's hit at 45%.
- **Water and frost:** the jog's time scale is ×0.7 when wet and ×0.6 when slowed (from the snapshot's wet and slowed flags), matching the sim's speeds.
- **Playback speed:** mixers advance by `delta × speed`, so 4× playback animates 4× faster.
- Cross-fades take 0.1–0.2 s, as in the prototype.

**Snapshot gains a skill count.**
- Unit records go from 10 to 11 ints, adding `skills`, the unit's skill count, for the stars.
- `sim/snapshot.go` and `web/src/sim/snapshot.ts` change together. A WASM test checks the layout constants match the decoder's expectations.
- This isn't part of the outcome or the checksum, so `sim.Version` and the golden replays are unaffected.

**Weapons** are the prototype's procedural meshes (sword, bow, staff with orb), attached to `hand_r` (sword, staff) or `hand_l` (bow) with the prototype's offsets.

**Projectiles.**
- Arrows are a shaft, tip and fletching group, oriented along the arc's tangent each frame; Arcing shot uses a higher arc (×0.22 instead of ×0.07 of the distance, as in the prototype).
- Bolts are a glowing orb (an inner and outer sphere, not tone-mapped).
- Both stay instanced or pooled per kind to avoid allocating each frame.

**Decorations.** Ring at the feet in the team colour; the health bar (existing); stars from the snapshot's skill count; wet units 0.12 m lower; dead units sink after 3 s.

**How determinism is preserved and tested.**
- Nothing in `sim/` that affects the outcome changes. `Snapshot` only gains an output field, so the golden replays and `sim.Version` stay.
- The WASM build already runs in the worker, separate from rendering. Asset loading happens on the main thread, and the worker's tick timer never waits for it.
- Tests: a sim test that the snapshot unit size matches the declared layout, and the existing golden replays both natively and under WASM.

## Risks / Trade-offs

- [18 MB first load] → The models are cached by the browser after the first match, and capsules show until they load. KTX2 or meshopt compression can come later if needed.
- [`sips` is macOS-only] → Packing is a rare dev step (after a re-bake), not part of CI.
- [Animations can drift slightly from the sim at high speed] → The swing is stretched to the sim's timing, and the blow is judged by the sim, never by the animation.
