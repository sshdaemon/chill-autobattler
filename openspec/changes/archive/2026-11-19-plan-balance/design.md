# Design

## Context

See proposal.md for why. A plan's stance could only set temperament fields.
The effect system already has what trade-offs need: `StatBonus` (speed,
damage, range), `DamageTaken` with conditions (always, ranged, front,
standing), `Guard` and others, folded at spawn by `foldStats` and
`loadout.learn`.

## Goals / Non-Goals

**Goals:** each plan has a reason to pick it; no plan dominates in any
playable faction; the plan shape stays small enough for custom plans and
free-text orders later.

**Non-Goals:** equal plans for every roster; new effect kinds.

## Decisions

### A stance may carry effects

`Stance.Effects []Effect` is passed as extras to `foldStats` and
`loadout.learn`, after the unit's traits and learned skills, so stat bonuses
fold in a fixed order and replays never depend on it. `ValidatePlans` runs
each effect's own check against every roster entry (a ranged-only effect on
melee fails).

*Alternative:* plan-specific hooks in the battle loop. Rejected: effects
already cover it, and step 5's custom plans can reuse the same list.

### How the values were found

A temporary test played every plan pair per faction with sides swapped (240
battles per pair), and plans were changed one or two levers at a time:

1. Effects first set large (Focus fire +12% damage, Charge +15% damage taken)
   swung the matrix hard: damage and damage taken are strong levers, so the
   final values are small.
2. Shield wall stayed weak against Hunt the archers whatever its bonuses.
   Testing variants showed the Protect target rule was the cause: melee chased
   enemies near their own archers and lost the fight. With Nearest it became
   strong, so its front bonus dropped from 20–30% to 8%. A Guard effect
   (covering nearby archers) made almost no difference and was dropped.
3. Skirmish's melee holding and falling back left the archers alone; a line
   with a 10 m leash screens them instead.

### Final values

| Plan | Melee | Ranged | Caster |
|---|---|---|---|
| Shield wall | hold (8 m), nearest, −8% damage from the front | line, focus in reach, stay near melee | same as ranged |
| All-out charge | charge, nearest, +20% speed, +3% damage taken | same, no kiting | same as melee |
| Hunt the archers | charge, ranged first, +25% speed, +5% damage taken | focus in reach | focus in reach |
| Focus fire | line, focus, +10% damage | line, focus in reach, +10% damage | same as ranged |
| Skirmish | line, 10 m leash | kite, keep 8 m, +30% range, +15% speed | keep 8 m, +30% range |

## Determinism

Effects fold at spawn through the existing fixed-order paths; no new
randomness or floats. Goldens run natively and under WASM; the plan-0 cases
(Instinct) are unchanged apart from the version string, which hashes the plan
table.

## Risks / Trade-offs

- [Some plans are weak for some factions: Hunt for the Rus (17%), Instinct for the Varangians (34%)] → recorded in the findings; it reads as faction character (druzhynnyk shield their archers; huskarls left alone are reckless).
- [Balance shifts when units change] → the plan matrix is in the balance report; rerun it after unit changes.
