# Proposal

## Why

Units now differ in data (the unit roster), but on screen they all move the same way. There are only 7 clips:
- one idle and one jog;
- one sword attack for every melee unit;
- one "spell shot" for every ranged unit, including archers;
- one hit reaction, one death and one cheer.

Skills barely show: a dash is just a faster jog, and Blast looks like any other bolt. The effects for each skill are hand-written branches in `effects.ts`. This is step 2 of the unique-units roadmap. It gives each unit its own set of animations and each skill its own animation and effect, all chosen by data, so that a new unit (step 4) gets its movement and effects without new rendering code.

## What Changes

- **Animation library.** Bring in clips from Quaternius's Universal Animation Library 1 and 2 (CC0, the same rig as our characters, so no retargeting): attack variants, a bow shot, heavy blows, leaps, casts, more hit reactions and deaths. Bake them into `clips.json`, within a download budget.
- **Unit looks (presentation data).** A look file per roster entry, matched by roster name. It holds:
  - the unit's models;
  - its animation set: idle, move, attack variants, hit, death, cheer;
  - for each skill, an action animation, a replacement attack animation and an effect preset.
  
  The models move here from `models.ts`.
- **Clip timing.** Each attack clip records when its blow lands. Playback is timed so that moment matches the sim's hit (45% into the swing), instead of stretching the whole clip.
- **Skill animations.**
  - **Action skills:** a gap-closer plays a leap-and-rush and an escape plays a cast-and-vanish.
  - **Skills that change every attack:** a unit with Blast or Arcing shot uses that skill's attack animation.
  - **Skills that change hits:** a finisher hit gets its own effect.
- **Effect presets as data.** Today's cues move from code into named presets (particles, rings, shake), with the same look. Sim events and skills choose presets by name.
- **Each unit's skills for the client.** A new browser-sim call lists every unit's class and learned skills, Red's included, so skill animations work for both teams. This doesn't change the snapshot or any rule.
- **Look check.** A test fails if a roster entry has no look, if a look misses one of its unit's skills, or if a look names a clip or preset that doesn't exist.
- **Proof.** Warriors get two or three attack variants and a leaping Dash. Archers get a bow shot and an arcing draw for Arcing shot. Mages get a cast-and-vanish Blink and a two-handed Blast cast. Executioner kills get a heavy-impact effect.

## Out of scope

- Any sim rule change. `sim.Version` stays 0.4.0, and golden replays are untouched.
- New models, body builds, parts or team-colour masks (step 3), and the `/new-unit` workflow (step 4).
- Creature or non-humanoid rigs.
- Sound changes per skill. The sounds stay keyed to events as today.
- Blending upper and lower body (for example attacking while running). Units don't do that in the sim.

## Capabilities

### New Capabilities
- `unit-looks`: presentation data for each roster entry: models, animation set, and an animation and effect preset for each skill. Also how looks are checked against the roster and what happens when something is missing.

### Modified Capabilities
- `battle-presentation`:
  - "Animation follows the sim": clips come from the unit's look, attack variants are timed to the sim's hit, and skills play their own animations.
  - "Combat effects": cues come from named presets that events and skills choose.

## Impact

- **`web/`:**
  - new `web/src/units/looks.json` and `web/src/scene/fx.json`;
  - `scene/models.ts` (models and clips come from looks);
  - `scene/Units.tsx` (state → clip from looks, hit-time alignment, skill actions);
  - `scene/effects.ts` (presets instead of branches);
  - `store.ts` (keeps each unit's skills);
  - `gamedata.ts` types.
- **Assets:** `web/public/assets/clips.json` grows. The budget is 1.5 MB raw and about 400 KB gzipped, with tracks trimmed and rounded. `tools/assets/pack_units.py` keeps the clips the looks name.
- **`wasm/`:** a new `units()` call (each unit's class and skill ids), plus a look-check test that reads `looks.json`, `fx.json` and `clips.json`.
- **Docs:** CLAUDE.md and the README describe looks and presets.
- **Source packs:** UAL 1 and 2 downloads go under `~/fun/assets-src`, outside the repo, with their licence files.
- **No changes to** `sim/` rules, the proto or the snapshot layout.
