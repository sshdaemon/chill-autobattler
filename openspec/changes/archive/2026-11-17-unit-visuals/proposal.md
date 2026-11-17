# Proposal

## Why

Battles still show coloured capsules. The prototype (`../web3d-demo`) already has the game's look: baked Quaternius characters per class and team, a shared animation library, weapons, and arrows and bolts. Bringing them over makes battles readable at a glance: who is attacking, who is hurt, who is dead. That readability is the point of an auto-battler you watch.

## What Changes

- **Character models:**
  - Ship the 8 baked looks × Blue/Red from the prototype under `web/public/assets/units/`:
    - warriors: Mirko, Vesna, Radim;
    - archers: Bogdan, Lada;
    - mages: Zoran, Milena, Velimir, with the purple robes and hat baked in.
  - Each unit gets a look from its class and its place in the army.
- **Smaller assets:** Blue and Red share one geometry file per look (they were byte-identical), and the roughness PNGs become JPEGs. A pack script, `tools/assets/pack_units.py`, makes this reproducible from the Blender bake output.
- **Animations:**
  - The prototype's shared `clips.json`, trimmed to the clips used, drives each unit from the sim's snapshot flags: idle, jog, sword attack or spell, hit reaction, death, and a victory cheer at the end.
  - An attack animation is stretched to the sim's swing time. Walking slows in water and under frost, and a dash plays the jog faster.
- **Weapons and projectiles:** sword, bow and staff (from the prototype) are attached to the hand bones. Arrows fly as arrows and point along their flight; bolts are glowing orbs.
- **Unit decorations:** a team-coloured ground ring, the health bar, veterans' stars above it, units sinking a little in water, and the dead sinking into the ground after a few seconds.
- **The Blender bake script** reads the Quaternius source packs from a configurable path (`QUATERNIUS_DIR`), so they don't have to live in this repo.

## Out of scope

- Fog of war during the fight, portraits, and the Warcraft III camera (a later change).
- Water and shore rendering beyond the current discs.
- New art. Real mage and noble outfits remain an open item from the concept.
- Any sim, server or rule change. Rendering never decides anything.

## Capabilities

### New Capabilities

- `battle-presentation`: how the browser shows a battle. Covers unit models by class and team, animations driven by sim state, weapons, projectiles, health bars and stars, and the guarantee that presentation never affects the outcome.

### Modified Capabilities

None.

## Impact

- `web/public/assets/units/` (17.2 MB, down from 30 MB) and `web/public/assets/clips.json` (0.7 MB, down from 1.4 MB).
- `web/src/scene/`: a model loader and cache, `Unit` rendering with an `AnimationMixer`, weapons, and projectile meshes. The capsule rendering is removed.
- `tools/assets/pack_units.py` (new); `tools/blender/bake_units.py` (the source path becomes configurable).
- The README asset section, plus the CC0 credit to Quaternius.
