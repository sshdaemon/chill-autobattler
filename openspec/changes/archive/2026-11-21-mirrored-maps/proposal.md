# Proposal

## Why

The fully random map decides too many battles. A new experiment
(`go run ./tools/balance -exp maps`) fixes 10 maps and plays the same armies
on each, changing only the combat rolls. With the same army on both sides,
Blue's win rate per map ranged from 1% to 95% (28 of 40 faction × map cells
more than 3 standard errors from 50%; on average 20 points off). Over many
random maps the sides are fair (49–51%), so each battle's map hands one side an
edge by luck: the player is always Blue.

## What Changes

- **Mirrored maps.** The map is generated in Red's half and reflected across
  the centre line (z → −z), so both sides face the same field. A rock or tree
  close to the centre line sits on it, once; a pond that reaches its
  reflection becomes one pond. Counts per half: 1–2 ponds, 2–3 rocks, 2–4
  trees, so a whole map has about as many obstacles as before.
- **Red paths in Blue's frame.** On a mirrored map, Red's path searches and
  free-spot snaps are reflected, answered as Blue's would be, and reflected
  back, so every tie the grid breaks (which side of a rock) falls the same way
  for both sides.
- **The map experiment** joins the balance tool (`-exp maps`).
- **BREAKING (rules):** every map changes; `RulesVersion` 0.13.0; goldens
  re-recorded.

Out of scope:

- The leftover side lean for some Varangian armies on some maps (5 of 40
  cells; the huskarl under target rules other than nearest). Its cause is not
  found yet; a follow-up.
- How much a map favours one faction's army over another's: on a mirrored map
  both sides get the same terrain, so this is the map's character, and players
  see the map before they draft.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-map`: maps are mirrored; Red's pathing choices mirror Blue's.

## Impact

- `sim/mapgen.go`, `sim/grid.go` (`Map.Mirrored`), `sim/battle.go`
  (`freeSpot`, `reachable`, `findPath` per team), `sim/unit.go`,
  `sim/engage_flank.go`.
- `tools/balance/maps.go`, `tools/balance/main.go`.
- Tests: map rules, `TestMirroredMapChoices`, goldens; the guard test no
  longer depends on a particular map.
- The client draws obstacles from the sim, so it needs no change.
