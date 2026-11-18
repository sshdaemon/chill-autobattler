## Why

Roadmap Phase 3, step 3.3: unit ids must not be taken by parallel unit
sessions, and the client indexes units by id, so ids can't have gaps. All nine
planned units are therefore created at once, as working placeholders, before
the unit sessions start (Phase 4).

## What Changes

- Unit ids 3–11 (druzhynnyk, strilets, huslyar; huskarl, bowman, völva;
  skoutatos, toxotes, siphonarios) and faction ids 2–4 (Rus, Varangians,
  Romans) allocated in `sim/ids.go`.
- Nine roster entries with the roster plan's stats, five abilities and
  temperament (traits for the huslyar's song and the siphonarios' cone).
- Factions get a `Playable` flag: only playable factions can be picked and
  drawn for Red. The three new factions aren't playable yet; classic battles
  are unchanged.
- Placeholder looks borrowed from the classic unit of the same role
  (`ui.placeholder`); empty clip bundles.
- The faction screen lists factions being built with their units from the sim.
- The balance lab can mix any units (a lab-only faction); its standard pool is
  the playable factions' units.

## Capabilities

### New Capabilities

### Modified Capabilities
- `unit-roster`: planned units exist as placeholders; factions can be not yet playable.

## Impact

`sim/unit_*.go`, `sim/faction_*.go`, `sim/ids.go`, `sim/roster.go` (Red pool),
`server/run` (playable check), `wasm data()`, `web/src/units/looks/*.json`,
`web/src/ui/FactionPick.tsx`, `tools/balance`.
