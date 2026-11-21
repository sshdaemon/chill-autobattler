# Proposal

## Why

Mirrored maps (`mirrored-maps`, #95) cut the side edge a map gives from 20 to
5.2 points on average, but 7 of 40 faction × map cells still leaned, up to 85%
for Blue (Varangians on one map). On a mirrored map with the same army on both
sides, nothing should tell the sides apart except the dice. Tracing mirror
battles tick by tick against their mirror image found five places where the
rules treat Blue and Red differently.

## What Changes

- **Which side acts first is a coin flip each tick** (from the combat seed).
  Units act in turn and see what earlier units did this tick (a dash, a blow).
  On a mirrored map two units often reach the decisive moment on the same tick,
  and Blue always went first: five huskarls with Charge won 97% as Blue on one
  map.
- **Rotations are seen from each side.** Kiting, stepping away and escaping try
  turns in a fixed order (+θ first); for Red that order is reflected. A unit
  turning to face straight behind turns toward +x, the same for both sides.
- **A point on a grid cell edge goes to the cell nearer the centre line.** Cells
  were floored, so Blue's formation slot at z = 8 m and Red's at −8 m fell in
  cells that aren't mirror images; near a rock one counted as blocked and the
  other didn't.
- **Fixed-point multiplication rounds toward zero** (like division already did),
  so a mirrored value multiplies to the mirrored result.
- **Units push apart all at once:** every push is worked out from where the
  units stand, then applied, instead of pair by pair in unit order.
- A test checks that mirror-image armies on a mirrored map march, path and push
  as exact mirror images until first contact.
- **BREAKING (rules):** `RulesVersion` 0.14.0; goldens re-recorded.

Out of scope:

- Exact mirror play after contact: from then on the coin flip (and the dice)
  decide, which is fair over many battles but not mirror-exact in one.
- How much a map suits one army over another (the same for both sides).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-rules`: neither side acts first by rule; mirror-image movement.

## Impact

- `sim/battle.go` (turn order, pushes), `sim/fixed.go` (`Mul`, `TurnToward`,
  `Turn.For`), `sim/grid.go` (`ToCell`), `sim/unit.go`, `sim/temperament.go`,
  `sim/effect_escape.go`.
- New `sim/mirror_test.go`; goldens; mutual-wipe seeds.
- Depends on `mirrored-maps` (#95).
