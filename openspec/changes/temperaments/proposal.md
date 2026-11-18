## Why

Roadmap Phase 3 (`docs/ROADMAP.md`): the nine planned units
(`design/ROSTER-PLAN.md`) behave differently: some charge, some hold the line,
some keep their distance or protect their archers. Phase 0 made behaviour data
with one target rule and one engage rule; this adds the rest of the plan's
temperament so unit sessions only fill in data.

## What Changes

- Target rules: wounded, ranged first, strongest, focus, protect (besides nearest).
- Engage rules: line (keeps pace with its allies), hold (waits for an enemy
  within its trigger distance or an ally fighting), besides charge.
- Temperament fields: trigger, leash, keep distance, fall back below a health %,
  stay near allies with a tag. All off at zero.
- No unit uses them yet: every golden outcome is unchanged (the fingerprint
  changes because the temperament data grew).

## Capabilities

### New Capabilities

### Modified Capabilities
- `battle-rules`: units fight by their temperament.

## Impact

`sim/temperament.go`, `sim/target_rules.go`, `sim/engage_rules.go`, `sim/ids.go`, `sim/unit.go`.
