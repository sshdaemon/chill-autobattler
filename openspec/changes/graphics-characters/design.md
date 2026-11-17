# Design

## Per-unit materials

`makeCharacter` clones the look's materials for each unit. Programs are shared, so it costs memory, not draw calls. Each material gets a rim term through `onBeforeCompile`:

`emissive += uRim * uRimK * pow(1 - |dot(normal, view)|, 2.5)`

`uRim` and `uRimK` are per-unit uniforms:
- team blue `#4f8dff` or red `#ff5a4a`, at 0.55;
- selected: gold `#ffd23f`, at 1.1, with a slow pulse.

The death fade reuses these materials instead of cloning again.

## Variety

A seed from (team, class, nth) picks:
- **hair:** one of 8 natural colours, applied to every `MI_Hair*` material;
- **outfit:** a ±7% brightness and ±4% hue shift of the `MI_Peasant` / `MI_Ranger` colour;
- **height:** a 0.96–1.04 scale.

## Trims

The trims are built in code and attached like the weapons: positioned in the bind pose, then `bone.attach`. They use team-coloured `MeshStandardMaterial`s (metal for pauldrons, cloth for the rest).
- **Warrior:** half-sphere pauldrons on `upperarm_l` and `upperarm_r`.
- **Archer:** a quiver (a cylinder with a team band) on `spine_03`, at the back, and a scarf (a torus) on `neck_01`.
- **Mage:** a mantle (a flattened cone) on `neck_01`.

## Animation

- **Cross-fades:** 0.25 s between locomotion states; 0.12 s into attacks, hits and death.
- **Kiting** (the sim's `kiting` flag while moving): the jog clip plays with a negative time scale, so the feet move backwards.
- **Aim:** for archers and mages with a live target, after the mixer updates, `spine_02` turns about the vertical by the angle between the unit's facing and the direction to the target. It's clamped to ±55° and smoothed.
- **Flinch:** `effects.ts` records each visible hit as (victim, direction from the attacker, time). For 0.3 s the unit's `spine_01` tilts up to 12° away from the attacker, easing out.

## Budget

There are no new draw calls beyond the trims: 2–3 small meshes per unit, at most 30. The rim is one extra term in the existing shader.

## Notes from building it

- **Portraits:** rendered per unit (look, team, nth), so they show each unit's hair. They're drawn without the rim, which tinted faces; the portrait frame already shows the team.
- **Cost:** GPU time per frame on High stays at about 9.6–12 ms (Apple M3, 1536×1152).
