## Context

Classic is faction 1 with warrior (class 0), archer (1) and mage (2). It is a
playable faction and Red draws it like any other. Beyond the faction itself,
the classic trio is load-bearing in places that aren't about Classic at all:

- the roster is indexed by id (`Roster[c].ID == c`, checked by
  `ValidateRoster`), so removing ids 0–2 breaks the indexing;
- `wasm data().classes` is indexed by class id, and the client reads
  `data.classes[cls]` in about 15 places;
- about 190 test references across `sim`, `pb`, `wasm` and `server` use
  warrior, archer and mage as "a melee", "a ranged" and "a caster";
- the balance tool's side, matchups, comps, skills, runs and curves
  experiments are built on `mix(w, a, m)`, `bestOrder` and `counters` of the
  classic trio;
- the client falls back to the warrior's or archer's look for units with no
  look, and the formation mini-diagram marks classes 0 and 1 specially;
- the proto `UnitClass` enum names the three units.

## Goals / Non-Goals

**Goals:** Classic and its units are gone from the game, data, tests, tools,
art and docs. Nothing that remains assumes classes 0–2 exist. Ids stay
reserved. Classic players' active runs end cleanly.

**Non-Goals:** rebalancing factions; renumbering ids; editing archived
OpenSpec changes or dated balance history.

## Decisions

### Retired ids are holes in the roster, not renumbered

Ids are never reused (`sim/ids.go`), and replays, stored armies and the wire
format all carry them. `Warrior`, `Archer`, `Mage` and `FactionClassic`
become a comment block of retired ids in `ids.go`, with
`retiredClasses = []Class{0, 1, 2}` as the one list the checks read.

`Roster` keeps "index = id": `register` pads missing ids with zero entries,
and `validateEntries` accepts an empty entry exactly at a retired id (and
rejects one anywhere else, and a real entry at a retired id). `RosterOf`
returns nil for a hole, so every existing "unknown unit" path rejects a
retired id with no new code. Loops over the roster (`RedPool`, `Pool`, data(),
validation, the balance tool) use a new `Units()` iterator that skips holes,
and nothing ranges over `Roster` directly any more.

*Alternative:* a dense roster with a lookup table by id. It would touch every
`Roster[c]` and change the data() contract for no gain; holes are cheaper and
keep the indexing the client already relies on.

### data() keeps index = id, with `null` at retired slots

`classes` stays indexed by class id; retired slots are `null`. The TS type
becomes `(ClassInfo | null)[]`, so the compiler finds every reader. Readers
that only ever see live ids (an army's units, a faction's units) assert
non-null through one helper, `classOf(data, cls)`.

### The proto enum reserves the classic values

`UNIT_CLASS_WARRIOR/ARCHER/MAGE` become `reserved 1 to 3;` plus their
reserved names. The enum keeps only `UNSPECIFIED`, and its comment says a
unit's wire value is its class id + 1. Proto3 enums are open, so no wire
change; tests that used the named values use `UnitClass(c + 1)`.

### Runs end by migration; battles need nothing extra

Migration `0007_retire_classic.sql` ends active Classic runs:
`state = 'over'`, `end_reason = 'faction retired'`, `current_battle_id =
NULL`, `ended_at` = now, where `faction = 1 AND state != 'over'`. It follows
0004 and 0005. Battles of those runs are not touched: the version bump
already refuses replays from the old sim, as for any rule change. The
`DEFAULT 1` on `runs.faction` and `battles.blue_faction` stays, because
SQLite can't change a column default in place and the server always writes
the value. The migration's comment notes this. The client's run-over screen
gets wording for the new end reason.

### Roles replace classic classes in the client

A unit's role is read from its tags (melee, ranged, caster), as the plan
editor already does. The formation mini-diagram marks melee as a filled
circle, ranged as a ring and casters as a diamond (today's three marks,
keyed by role, not class id). The placeholder look is the first unit of the
same role in roster order that has a look, so druzhynnyk, strilets or
huslyar. It warns once in dev, as now.

### Art goes with the looks, and a check keeps it clean

`looks/warrior.json`, `archer.json`, `mage.json` and `factions/classic.json`
are deleted. Part, item, texture and clip files that no remaining look,
`_shared.json` or arena file references are then deleted from
`web/public/assets`. `check-bundle.mjs` gains a check that fails the build on
any file under `assets/parts` or `assets/clips` that nothing references, so
the next retirement can't leave orphans either. The asset manifest and
service worker pick up the removals by themselves: removed paths drop out of
`sw.js`'s manifest and are pruned from caches on activate.

### Tests move to the unit that has what's under test

| Was | Becomes | Why |
|---|---|---|
| warrior as "a melee" | druzhynnyk (or huskarl where finisher, gap-closer or splash is tested) | melee, Rus, no-frills baseline |
| archer as "a ranged" | strilets | ranged, shoots over cover with Arcing shot |
| mage as "a caster" | huslyar, volva for escape, siphonarios for slow | caster role, effect owners |

The effect-order test (slow then splash on one hit) needs a unit with both,
which no faction unit has. It builds a test-only roster entry the way "A new
unit needs no new rules" already does, and restores the roster afterwards.
Tests that pin classic numbers from `docs/GAME_CONCEPT.md` §4.1 and §10.1 (for
example in `rules_test.go` and `roster_baseline_test.go`) are deleted if
another test covers the rule, or rewritten against a faction unit's own
numbers if none does. `faction_test.go` gains tests that faction 1 and
classes 0–2 are rejected, and that `ValidateRoster` refuses a hole at a
non-retired id and a unit at a retired one.

### The balance tool works per faction

`mix(w, a, m)` becomes `mix(f, melee, ranged, caster)` over a faction's
three units, in its `Units` order. Side bias, matchups, compositions and
skills print one table per playable faction. Runs and curves take Blue's
faction (every playable faction by default). The "best" skill policy ranks
each unit's skills from that faction's skills table, computed at the start
of the experiment, not from a hard-coded classic list. `counters` becomes a
role triangle (caster beats melee, melee beats ranged, ranged beats caster),
filled with Red's faction's units. Each faction has exactly one unit per
role; the tool checks this and fails loudly if a faction breaks it.

### Determinism

No mechanism changes: fixed-point maths, seeded streams and the turn coin
flip are untouched. The removal changes Red's faction draw
(`RedFactionFor` picks among three, not four) and the roster fingerprint, so
`RulesVersion` goes to 0.16.0. Golden replays are re-recorded with faction
armies and checked natively and under WASM (`TestGoldenReplays`, wasm
tests). The mutual-wipe seeds are found again with a Rus mix. The custom
plan and movement checks (`TestMovementSmooth`, mirror tests) run on faction
armies with their existing thresholds.

## Risks / Trade-offs

- **Coverage loss in rewritten tests** → each rewritten test is first run
  against a deliberately broken rule to confirm it still fails (as done for
  `TestGuardsAndFollowersAdvance`).
- **Thresholds tuned on classic armies** (movement reversals, side bias,
  custom-plan search) may sit differently on faction armies → re-measure and
  note the new numbers in `docs/balance-findings.md`. Change a threshold only
  with the measurement in the PR.
- **Orphan check false positives** (a file loaded by name from code, not
  from a look) → the check reads the same sources `battleFiles` does, plus
  an explicit allowlist in `check-bundle.mjs` for anything else.
- **Players lose an active Classic run** → the score is kept and the run-over
  screen says why.

## Migration Plan

One PR: sim, proto, server migration, client and tools ship together. The
version bump makes old clients' replays fail the version check, as with
every rule change. Rollback means reverting the PR. Ended runs stay ended,
which is acceptable for a game in development.

## Open Questions

None.
