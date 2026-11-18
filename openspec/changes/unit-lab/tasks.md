# Tasks

## 1. Lab
- [x] 1.1 Dev-only `?lab[=unit]` route and `UnitLab` (unit, recipe or lineup, team, build override, variety seed; orbit camera with turntable and framing). Verify the typecheck and that the production bundle has no lab code.
- [x] 1.2 Animation panel: the unit's clips and all clips; play, pause, loop, speed, timeline scrub. Verify in the browser.
- [x] 1.3 Locomotion: ground-speed slider and presets, moving floor, the battle's walk–jog blend, live blend and foot-slide readout. Verify the readout matches the earlier offline check (under 10%).
- [x] 1.4 Skills (action or attack, lean, effect) and any effect preset; portrait. Verify in the browser.
- [x] 1.5 Dev-HUD link and CLAUDE.md. Verify the link opens the lab.

## 2. Checks
- [x] 2.1 Run `pnpm test` and the production build. Verify everything is green and the lab is absent from `dist`.
