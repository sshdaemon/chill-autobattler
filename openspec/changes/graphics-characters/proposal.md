# Proposal

## Why

Units are hard to tell apart on the darker battlefield: Blue and Red differ mostly by their rings and health bars, and every archer of a team looks like a copy of the others. Their animations also feel stiff: archers jog forwards while backing away, hits barely register, and units don't turn toward what they shoot. This is step 4 of `docs/graphics-roadmap.md`: the class, team and state of any unit should read at a glance.

## What Changes

- **Team rim light:** a soft blue or red edge glow on every unit, so teams read against the ground. The selected unit gets a brighter gold rim.
- **Team trims,** made in code with no downloads:
  - warriors: pauldrons in the team colour;
  - archers: a quiver with a team-coloured band, and a team scarf;
  - mages: a team-coloured mantle.
- **Variety without new bakes:**
  - each unit gets its own hair colour, a slight tint to its clothes and a slightly different height;
  - all picked from a seed of team, class and position in the army, so an army looks the same after a reload.
- **Animation:**
  - smoother cross-fades between moves;
  - archers jog backwards while kiting instead of moonwalking;
  - archers and mages turn their upper body toward their target;
  - a short flinch away from the attacker when hit.

## Out of scope

- **New baked looks:** each one adds 1.1–2.6 MB to the download, so it's left for a separate decision.
- **Paid or commissioned characters.**
- Anything in the sim or on the server.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-presentation`: team readability (rim light, trims) and the animation details.

## Impact

- `web/src/scene/models.ts`: per-unit materials with the rim shader, variety, and trims.
- `web/src/scene/Units.tsx`: selection rim, backward jog, upper-body aim, flinch; the death fade uses the unit's own materials.
- `web/src/scene/effects.ts`: records hits for the flinch.
