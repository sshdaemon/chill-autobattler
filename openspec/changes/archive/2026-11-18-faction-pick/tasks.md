## 1. Rules

- [x] 1.1 Faction registry, Classic faction, validation (`sim/faction.go`)
- [x] 1.2 `Input.BlueFaction`, Blue's units checked; Red's faction from the seed, Red drawn from it
- [x] 1.3 `RulesVersion` 0.7.0; version fingerprint covers factions; goldens (outcomes unchanged)

## 2. Server

- [x] 2.1 Proto fields; wire and convert decoders
- [x] 2.2 Runs and battles store the faction (migration 0006); `StartRun` checks it

## 3. Client

- [x] 3.1 `wasm data()` publishes factions
- [x] 3.2 Faction screen with lore and units, coming-soon factions; picker limited to the faction
