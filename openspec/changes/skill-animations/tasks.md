# Tasks

## 1. Clip intake

- [x] 1.1 Ask the user before downloading. Fetch the free editions of Quaternius Universal Animation Library 1 and 2 into `~/fun/assets-src/ual1` and `ual2`, with their licence files. Verify that both licence files say CC0.
- [x] 1.2 Spike: export one UAL attack clip, load it on a current warrior in the browser, and compare bone names and rest pose. Verify the clip plays without a distorted pose (screenshot). If it needs a retarget, stop and report.
- [x] 1.3 Write `tools/assets/ual_clips.py` (reads the UAL `.glb` directly, no Blender) that exports clips with trimming (scale and leaf tracks dropped, 4 decimals, 30 fps, linear keys removed). Verify that the clips we already had come out value for value the same.
- [x] 1.4 Choose the clips for each slot: warrior attack variants, an archer bow shot and an aim-up shot, mage cast variants, a two-handed Blast cast, a Dash leap, a Blink cast, extra hit reactions and deaths. Measure each attack's blow time. Verify by recording the choices and `hit` times in `web/src/units/looks.json`.

## 2. Looks and the look check

- [x] 2.1 Write `web/src/units/looks.json` (clips with `hit` and `loop`, plus the three units with models, animation set and all five skills each). Move `LOOKS` out of `models.ts`, and add TS types in `gamedata.ts`. Verify with `pnpm --filter web typecheck`.
- [x] 2.2 Change `pack_units.py` to keep exactly the clips named in `looks.json`. Rebuild `web/public/assets/clips.json`. Verify the file is within 1.5 MB raw and about 400 KB gzipped.
- [x] 2.3 Add `wasm/looks_test.go`: every roster entry has a look, skill keys equal the roster's skill ids, clips exist (attacks have `hit`), `fx` names exist in `fx.json`, and the clip budget holds. Add a failing fixture for each rule. Verify with `pnpm test:wasm`.

## 3. Each unit's skills for the client

- [x] 3.1 Add `chillSim.units()` (class and skill ids per unit, Red included) to `wasm/main.go`, with a test that compares it with `sim.New(...).Units` for a generated Red army. Verify with `pnpm test:wasm`.
- [x] 3.2 Call it after `create` in the worker or store, keep the list next to the snapshots, and show Red's skill names in the info panel only when Red is visible (as today, the count). Verify with the typecheck, and in the browser that the store has both teams' skills.

## 4. Animation from looks

- [x] 4.1 Replace `wanted()` in `Units.tsx` with a picker driven by looks: state clips, attack variants by hashed index and swing count, a skill's attack animation, the gap-closer's action while dashing, the escape action on EvBlink, cheer, and the runtime fallback with a single dev warning. Verify in the browser with `fxDemo`-style forced states and a real battle.
- [x] 4.2 Time attacks to the sim's hit (`rate = h / (0.45·S)`, clamped to 0.6–1.8, with a start offset when clamped; crossfade out when the swing ends). Verify with a dev overlay or log that the blow frame is within one frame of the EvHit or EvShot tick at 1× and 4×, for a normal and a Swift warrior.
- [x] 4.3 Update the `battle-presentation` behaviour in the browser: Dash leap, Blink cast, Blast and Arcing-shot attacks, warrior variants. Verify with screenshots of each, and check `benchFrame` frame time is no worse than on `main`.

## 5. Effect presets

- [x] 5.1 Write `web/src/scene/fx.json` with today's cues as presets (`hit`, `hit.flank`, `hit.strong`, `hit.magic`, `execute`, `death`, `dash`, `blink`, `frost`, `blast`, `shot.bolt`) and port `effects.ts` to read them, keeping pools, budget, scaling and the fog check. Verify that `window.fxDemo()` screenshots match `main` side by side.
- [x] 5.2 Make skill events use the skill's `fx` from the look (finisher, slow, splash, gap-closer, escape), and give Executioner its heavy-impact preset. Verify in the browser that a warrior's execute shows the new preset and an archer's Headshot keeps its own.

## 6. Docs and checks

- [x] 6.1 Update CLAUDE.md (looks, presets, `units()`, adding a unit = roster file + look) and the README layout. Verify by reading them against the code.
- [x] 6.2 Run `pnpm test` and play a battle with every unit type and skill. Verify that golden replays are untouched, the server reports the battle verified, and the clip budget test passes.

## Workflow follow-up

- Open a PR for review. Merge it when the user says so, then archive the change in a separate PR (auto-merged after green CI).
