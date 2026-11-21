# Design

## Context

`GenerateObstacles` placed ponds, rocks and trees anywhere they fit, so a map
could crowd one side's approach. The new map experiment showed per-map side
edges of up to ±49 points with the same army on both sides. Formations put a
preset's first flank at −x for both teams (`FormationSlots`), and Blue starts at
+z, Red at −z.

## Decisions

**Reflection across z = 0, not a 180° rotation.** Formations are laid out by
world x for both teams, so a reflection gives each unit the same terrain as
its counterpart; a rotation would swap the flanks' terrain.

**Generate in Red's half and reflect** each feature. A rock or tree within
`R + gap/2` of the centre line is snapped onto it and placed once (its own
reflection). A pond whose blobs reach their reflections is one pond (same id),
else the reflection is a new pond. Clearance is checked for the feature and its
reflection together, so the result obeys the old spacing rules. Counts are per
half (1–2 ponds, 2–3 rocks, 2–4 trees), which keeps a map's obstacle count
near the old 2–3 / 3–6 / 3–7.

**Red decides in Blue's frame.** The grid breaks ties by scan order (A*'s
neighbour order, `NearestFree`'s rings), which isn't symmetric in z, so on a
mirrored map Blue and Red went round the same rock on different sides.
`NewMap` sets `Map.Mirrored` when every obstacle has its reflection;
`Battle.findPath`, `freeSpot` and `reachable` then reflect Red's question,
answer it, and reflect the answer back. Hand-made test maps that aren't
mirrored work as before.

**Measured** (`-exp maps`, 10 maps × 4 factions, 400 battles per cell,
same army on both sides):

| | Before | Mirrored maps | + Red in Blue's frame |
|---|---|---|---|
| Mean gap from 50% | 20.0 pts | 8.0 pts | 5.2 pts |
| Cells > 3 SE from 50% | 28 / 40 | 10 / 40 | 7 / 40 |

**Tried and dropped:** alternating which side acts first each tick. Units
act in order and see earlier units' blows, but alternating didn't change the
leftover cases, so the order isn't their cause.

## Determinism

Generation and the reflections are exact Fx negation; the map stays a pure
function of the seed. Goldens are re-recorded under 0.13.0 (new mutual-wipe
seeds 12 and 703) and run natively and under WASM. `TestMapRules` checks every
obstacle has its mirror image; `TestMirroredMapChoices` checks Red's paths and
free spots are Blue's reflected on 40 generated maps.

## Risks / Trade-offs

- **Less variety:** maps look symmetric. The client picks rock and tree models
  by position, so reflected pairs still look different.
- **Leftover lean** for Varangian huskarls under the wounded and focus target
  rules on some maps (up to 85% on one map). Not explained by turn order; a
  follow-up should trace it (likely fixed-point rounding or a rotation that
  turns one way first).
- **Matchups still vary by map** (an archer army likes open ground): that is
  the map's character, the same for both sides, and the player sees the map
  before drafting.
