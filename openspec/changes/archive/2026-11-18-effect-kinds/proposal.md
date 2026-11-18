## Why

Roadmap Phase 3, step 3.2: the abilities in `design/ROSTER-PLAN.md` need effect
kinds the sim doesn't have. Phase 0 made each kind a plug-in; this adds the
hook points they need and the kinds themselves, before any unit uses them, so
unit sessions never touch shared code.

## What Changes

- **Modifiers** recomputed each tick (damage dealt, damage taken, attack speed,
  evasion, speed) from auras, timed buffs and the unit's own state; applied
  only when non-zero.
- **New hook points:** on hit (melee, shots, cone victims), aura, self,
  on hurt, intercept (guard), on death (rally), on charged (brace), strike
  (cone), kite-shoot; thrown bursts; burning ground zones; fear.
- **New effect kinds (ids 8–22):** damage taken (conditional), aura (allies or
  foes), rally, guard, brace, fury, curse, fear, burst, bonus against, focus,
  phalanx, burning ground, kite shoot, cone.
- **Traits:** effects every unit of a type has (a song, a cone attack).
- Slow and splash now work on melee and cone hits too; auras and other
  additive kinds may appear more than once on a unit.
- `StatAuraRadius`; new events for the client (miss, fear, burst, jet, burn,
  rally, guard, brace, curse).
- No unit uses them yet: every golden outcome is unchanged.

## Capabilities

### New Capabilities

### Modified Capabilities
- `battle-rules`: effect kinds and modifiers.

## Impact

`sim/mods.go`, `sim/effects.go`, `sim/effect_*.go`, `sim/combat.go`,
`sim/unit.go`, `sim/battle.go`, `sim/roster.go`, `wasm data()`, `web/src/gamedata.ts`.
