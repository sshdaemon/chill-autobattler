## Why

Classic (warrior, archer, mage) was kept "while the three historical factions
are built, then retired" (`web/src/factions/classic.json`). Rus, Varangians and
Romans are all playable now, so Classic is a fourth faction from a different
game: it dilutes the setting, Red still draws it, and its units anchor tests,
tools and docs that should be about the real roster.

## What Changes

- **BREAKING (rules, `RulesVersion` bump):** the Classic faction and its three
  units leave the sim. Red's faction is drawn from Rus, Varangians and Romans
  only, so Red's armies change for most seeds. Golden replays are re-recorded.
- **Ids stay retired, never reused:** `Faction` 1 and `Class` 0, 1, 2 become
  reserved in `sim/ids.go`. The roster keeps its "id = position" indexing
  with empty slots at 0–2; an input naming a retired unit or faction is
  rejected as unknown. The proto `UnitClass` values for warrior, archer and
  mage become `reserved`.
- **Runs:** active Classic runs end ("faction retired", score kept) through a
  migration; `StartRun` rejects faction 1 as unknown. Finished Classic runs
  stay on the scoreboard.
- **Client:** the Classic lore file, the warrior, archer and mage looks and
  the art only they use (part, weapon and clip files) are deleted. The
  placeholder-look fallback and the formation mini-diagram stop naming
  classic classes and use roles (melee, ranged, caster) instead.
- **Tests:** every test built on warrior, archer or mage moves to a faction
  unit that carries the effect or role under test.
- **Balance tool:** experiments hard-wired to the classic trio (side,
  matchups, comps, skills, runs, curves) run per playable faction on its
  three units (melee, ranged, caster).
- **Docs and design:** `docs/GAME_CONCEPT.md`, `docs/ROADMAP.md`,
  `CLAUDE.md`, `design/units/warrior/` and units' balance notes that compare
  against classic units are updated or removed; balance history keeps its
  dated numbers with a note that Classic is retired.

Out of scope: rebalancing the three factions (beyond re-checking that the
balance tool's per-faction tables still pass), new factions or units,
reworking the archived OpenSpec changes (history stays as written), and step 6
of the battle-plans roadmap.

## Capabilities

### New Capabilities

### Modified Capabilities
- `unit-roster`: the roster no longer starts with warrior, archer and mage;
  retired ids stay reserved and are rejected; matchup and effect-order
  scenarios use faction units; placeholder looks borrow by role.
- `match-flow`: the formation mini-diagram marks units by role; the faction
  and draft scenarios name real factions and units.
- `runs`: Classic runs end on upgrade; the scenarios name a real faction.

## Impact

`sim/` (ids, roster, faction and unit files, every test using classic units,
golden replays, `RulesVersion`), `proto/chill/v1/units.proto` and generated
code, `pb/convert` and `pb/wire` tests, `wasm/` (data() with retired roster
slots, tests), `server/` (migration 0007, run and battle tests), `web/`
(factions, looks, parts and clips, `DeployBar`, `units/looks.ts`, `gamedata`
types), `tools/balance/`, docs and `design/units/`.
