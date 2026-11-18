## Why

Roadmap Phase 1, step 1.2 (`docs/ROADMAP.md`): a run is played as one
faction, picked at its start on a screen that says what the faction is about
and which units it has. Factions are what the Dnipro-road setting
(`design/SETTING.md`) builds on, and what the draft (step 1.3) drafts from.

## What Changes

- **Factions are sim data:** each faction lists the unit types it fields; one
  file per faction (`sim/faction_<name>.go`), ids in `sim/ids.go`. One
  faction for now, **Classic** (warrior, archer, mage).
- **BREAKING (rules, `RulesVersion` 0.7.0):** a battle input names Blue's
  faction, and Blue's units must come from it. Red's faction is drawn from the
  map seed (any faction, mirrors allowed) and Red's generated army comes from
  that faction's units. No outcome changes with one faction.
- **Server:** a run is started with a faction and keeps it; each battle stores
  the faction it checks Blue against. `StartRun` rejects unknown factions.
- **Client:** a faction screen opens every new run (lore, colours, units);
  factions not playable yet (Rus, Varangians, Romans) appear as "coming soon"
  with their planned units. The army picker shows only the faction's units.
- Lore and looks: `web/src/factions/<name>.json`, matched by name.

## Capabilities

### New Capabilities

### Modified Capabilities
- `runs`: a run plays one faction, chosen when it starts.
- `battle-rules`: Blue's units come from its faction; Red's faction from the seed.
- `match-flow`: a faction screen before the first picker; the picker shows the faction's units.

## Impact

`sim/faction*.go`, `sim/army.go`, `proto` (`BattleInput.blue_faction`,
`Run.faction`, `StartRunRequest.faction`), `pb/wire`, `pb/convert`,
`server/run`, `server/battle`, migration 0006, `wasm data()`, `web` (store,
`ui/FactionPick.tsx`, `factions/`).
