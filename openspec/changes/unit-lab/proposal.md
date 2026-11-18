# Proposal

## Why

There's no good way to look closely at a unit. The battle camera is far away. The `/new-unit` preview (`?preview=<unit>`) shows a fixed set of poses from one angle for screenshots. Judging a look, an animation, a build or the new walk–jog blend means playing whole battles. Designing and redesigning units needs a place to inspect them.

## What Changes

- **A dev-only unit lab** at `?lab` (or `?lab=<unit>`), linked from the dev HUD. Units are assembled by the same code as in battles. It offers:
  - **Pickers:** unit type, recipe (or the whole lineup), team, build (any build from the looks, to try combinations), and variety seed (hair, tint and height as in an army).
  - **360° view:** orbit, zoom and pan the camera around the unit, with an optional turntable spin and a choice of close-up or battle-distance framing.
  - **Animations:** every clip the unit uses (idle, walk, move, sprint, each attack, hit, death, cheer, skill actions) plus every other clip in `clips.json`. Each can play, pause and loop, with playback speed and a timeline to scrub.
  - **Locomotion test:** a ground-speed slider (0–6 m/s, with presets for the unit's speed, Swift, water and frost) that walks the unit on a moving floor through the battle blend, with a live readout of the blend and of the planted foot's sliding.
  - **Tuning:** live sliders for the walk and run numbers: the clip speeds, a walk/jog mix bias, the slowest walk and fastest jog and sprint, speed smoothing, and keeping feet on the floor and its smoothing. A 'wobble' mode varies the speed like turns in battle, and buttons copy or reset the values. The values are saved in the browser, and dev battles use them too, with a 'tuned' marker in the HUD.
  - **Skills and effects:** one button per skill (its action or attack animation, lean and effect preset), and any effect preset from `fx.json`.
  - **Portrait:** the unit's portrait as it appears in battle.
- The `/new-unit` preview (`?preview`) stays as it is for screenshots.

## Out of scope

- Editing looks or recipes from the lab: changes still go into `looks.json` and the bake.
- Shipping the lab to players (it's dev-only, like the preview).
- Battle sandboxes (scripted fights between chosen units).

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `unit-authoring`: adds the unit lab for inspecting units and their animations from any angle.

## Impact

- New `web/src/ui/UnitLab.tsx` (dev-only route in `main.tsx`);
- `scene/models.ts` (`makeCharacter` accepts a build override);
- a dev-HUD link in `ui/Overlays.tsx`;
- CLAUDE.md;
- no sim or asset changes.
