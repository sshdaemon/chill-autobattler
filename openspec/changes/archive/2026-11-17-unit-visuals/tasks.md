# Tasks

## 1. Assets

- [x] 1.1 `tools/assets/pack_units.py`: share Blue/Red geometry, convert roughness to JPEG, trim clips; make `bake_units.py` read `QUATERNIUS_DIR`; verify the packed folder is about half the original size (17.2 MB, down from 30 MB) and every `.gltf` references existing files
- [x] 1.2 Pack the prototype's baked units into `web/public/assets/units/` (+ `clips.json`); verify all 16 load in the browser

## 2. Sim export

- [x] 2.1 Snapshot unit record gains `skills` (11 ints); `data()` gains each class's swing time; update the TS decoder; verify the sim and WASM tests (layout size) and that the golden replays still pass unchanged

## 3. Scene

- [x] 3.1 Model cache + per-unit clone, look assignment, weapons on hand bones, team ring; capsule fallback while loading; verify in the browser that both armies show the right looks and weapons
- [x] 3.2 AnimationMixer per unit driven by snapshot flags (idle/jog/attack stretched to swing time/hit/death/cheer; wet and frost slow the jog; speed scales the mixer); verify in the browser at 1× and 4×
- [x] 3.3 Arrow and bolt meshes along their flight; stars above health bars; wet sink; dead sink after 3 s; verify in the browser

## 4. Docs

- [x] 4.1 README (assets, pack/bake steps, Quaternius CC0 credit) and CLAUDE.md; verify `pnpm test` passes

## Workflow follow-up

- Archive the change once merged (auto-merge after green CI).
