# Design

## Context

See proposal.md for the motivation. Today:

- **`sim/data.go`** holds `ClassStats[3]`, a 3×3 `Matchup` table and `SkillsOf[3][5]` (skill id, name, one `Pct`, one `Value`).
- **Behaviour is spread through the code:**
  - `newUnit` applies stat bonuses by skill id (`pct(SkillHealth)`…, Swift's attack speed, Lob's damage).
  - `think` checks `SkillBlink` and `SkillDash`.
  - `damage` checks `SkillExecute`.
  - `updateShots` checks `SkillFrost` and `SkillBlast`.
  - `shoot` sets the bolt from `u.Class == Mage`.
  - Blink and kite read the victim's `Stats.Ranged`.
- **Effect tuning lives in global constants:** `DashSpeed`, `DashCooldown`, `BlinkPanic`, `BlinkDist`, `BlastRadius`, `FrostTicks`, `FrostSpeed`.
- **Other places that use these tables:**
  - `RedArmy` draws classes with `IntN(numClasses)` and skills from `SkillsOf`.
  - `wasm/main.go data()` serialises the tables.
  - `server/run` validates picks with `SkillDefOf`.
  - `pb/convert` and `pb/wire` map the proto enums to `sim.Class` and `sim.Skill` by number.
  - `tools/balance` loops over classes.

**Constraints:**
- The sim must stay deterministic (fixed point only, no map iteration, seeded RNG streams).
- Every recorded golden replay must reproduce bit for bit, so the order of RNG draws and of `MulDiv` roundings must not change.

## Goals / Non-Goals

**Goals:**
- One place describes a unit: its roster entry.
- A fixed set of effect kinds with parameters, applied at fixed hook points.
- Bit-identical outcomes, proven by tests.
- A cheap hot loop: no per-tick scanning of skill lists.

**Non-Goals:**
- A scripting language or data-defined behaviour trees.
- Per-unit AI changes, such as new targeting logic.
- Renaming the wire enums.

## Decisions

### 1. Roster entries are Go tables, one file per unit
`sim/roster.go` defines the types and `var Roster = []UnitDef{...}`, built from `sim/unit_warrior.go`, `unit_archer.go` and `unit_mage.go`. Each of those files holds one `UnitDef` literal.

- **Rejected: embedded JSON or YAML.** The values would need a decimal parser into `Fx` (floats are banned), runtime loading in WASM, and checks the compiler gives us for free.
- **Why Go tables:** `Ratio(145, 100)` literals are exact. A future `/new-unit` step can write a Go file as easily as a JSON file, and the type checker catches typos.

The id type stays `Class` (now meaning "roster id"), so `pb/convert`, `pb/wire` and the proto enum numbers don't change. A rename can come later.

### 2. Effects are a closed set of tagged structs
```
type Effect struct {
    Kind   EffectKind // StatBonus, Finisher, SlowOnHit, SplashOnHit, GapCloser, Escape, OverCover
    Stat   StatKind   // for StatBonus: HP, Damage, Range, Speed, AttackSpeed
    Pct    int        // bonus %, chance %, splash %, attack-speed factor %
    Value  Fx         // reach, distance, radius
    Ticks  int        // cooldown or duration
    Factor Fx         // e.g. slow speed factor
}
type SkillDef struct { Skill Skill; Name string; Effects []Effect }
```

The constants that tune effects move into effect parameters, with today's values:
- Dash reach 10 m, cooldown 200 ticks, speed 16 m/s;
- Blink panic 2.5 m, distance 8 m, cooldown 8 s;
- Frost 100%, 60 ticks, ×0.6;
- Blast 35%, radius 2 m.

Rules that apply to every unit stay global: the 40% finisher threshold, flank ×2 and stagger.

- **Rejected: an interface or closure per skill.** Behaviour hidden in code can't be published to the client or checked as data, and the order effects run in becomes whatever order the skills were registered.

### 3. A loadout is compiled at spawn; hooks read it in a fixed order
`newUnit` folds the unit's learned skills into a flat `loadout`:
- final stats;
- `finisherPct`;
- `slow`, `splash`, `gapCloser` and `escape` (each nil, or its parameters);
- `overCover`.

The hooks then read fields instead of skill ids:
- **tick:** escape, then swing, then kite, then gap-closer. This is today's order in `think`.
- **hit:** damage roll, flank, matchup, then finisher. This is today's order in `damage`.
- **projectile lands:** damage, then slow roll, then splash. This is today's order in `updateShots`.

The order is fixed by code, by effect kind, and never by skill-list order. That keeps RNG draws identical to today.

**Stat bonuses** apply grouped by stat, in the fixed order HP, Damage, Range, Speed, AttackSpeed. Within one stat they apply in the entry's skill-list order. Each one is a separate `MulDiv`, so the rounding is unchanged. Today's archer applies Damage (Sharp arrows) before Lob, which is also the list order. Lob is declared as `OverCover` + `StatBonus{Damage, 10}`.

**Swift** is `StatBonus{Speed, 40}` + `StatBonus{AttackSpeed, 80}`. The second keeps today's integer `*80/100` on cooldown and swing length.

**Limit:** a unit may hold at most one effect of each non-stat kind. The roster check enforces it, so the loadout never has to merge two slows.

### 4. Traits and matchups
- **Traits:** `UnitDef` carries `Melee bool` (today's `!Ranged`), `Kite bool` and `Projectile` (Arrow or Bolt). Blink and kite count threats with `def.Melee`. `shoot` sets `Bolt` from `def.Projectile`.
- **Tags:** a `Tag` bitmask (`TagMelee`, `TagRanged`, `TagCaster`), with each tag declared once in the roster.
- **Matchup bonuses:** `Bonus []TagBonus{Tag, Mult}`, in order. The multiplier is that of the first bonus whose tag the victim has, else `One`.
  - Warrior: `ranged ×1.45`, `caster ×0.95`.
  - Mage: `melee ×1.35`.
  - Archer: none.
  - These give exactly the cells of today's table.

The `FlagStrong` and `FlagWeak` event flags still compare the multiplier with `One`.

### 5. Red's pool
Each `UnitDef` has `InRedPool bool`. `RedArmy` draws `IntN(len(pool))` over the pool in roster order. With today's three entries that is `IntN(3)`, the same draw. Adding a unit to the pool later changes Red's armies, so that change will need a version bump; the comment in the code says so.

### 6. Roster check
`ValidateRoster(r []UnitDef) error` checks the rules in the unit-roster spec and runs in a test. The sim does not check its own constant data at startup in production, so WASM size and start-up time stay unchanged.

The data-only test unit (a ranged "frost ranger" with gap-closer, slow on hit, splash, finisher and a stat bonus; slow and splash act on projectiles, so the roster check only allows them on ranged units) lives in a `_test.go` file and is swapped into `Roster` by a test helper that restores the original afterwards.

### 7. Client data
`wasm data()` keeps its current fields (`classes[].hp…`, `skills[].name/pct/value`, the `matchup` matrix computed from tags), so the web client doesn't change. It adds `traits`, `tags`, `bonus`, and `effects` for each skill. `pct` and `value` stay, filled from the skill's main effect, so the existing UI text still reads correctly.

## How determinism is preserved and tested

1. **First commit, before the refactor:** add golden cases recorded with today's code that use every skill. That includes:
   - each skill on its own;
   - Frost + Blast on one mage;
   - Sharp arrows + Lob on one archer;
   - Swift + Dash;
   - Executioner and Headshot at their thresholds;
   - Red armies with 0, 3 and 25 skills.
   
   Adding cases without a version bump is allowed.
2. **Refactor.** Then `pnpm test:go` and `pnpm test:wasm` must pass with no `-update` and `Version` still "0.4.0". Any drift in an RNG draw or rounding fails a case.
3. **Unit tests:**
   - the loadout of each of the 15 skills matches today's numbers (for example warrior + Toughness has 110 HP, archer + Sharp arrows + Lob has the same `DmgLo` and `DmgHi` as before);
   - the matchup matrix from tags equals the old table;
   - `RedArmy` gives the same output for a range of seeds.
4. **Static check:** `TestNoNondeterministicImports` still covers the new files. The roster is a slice, never a map, so iteration order is fixed.

## Risks / Trade-offs

- **[Risk]** A subtle change in order (for example applying Lob's damage before Sharp arrows) changes a rounding in a rare case. → Mitigation: golden cases that combine those skills, plus loadout unit tests that compare exact `Fx` values with the old formula.
- **[Risk]** The fixed effect set is too narrow for future units (heal, summon, shield). → Mitigation: adding an effect kind is a small, local sim change with its own tests and a version bump. The set grows as real units need it.
- **[Trade-off]** "At most one of each non-stat effect per unit" limits some combinations. It's acceptable for five-skill units, and the roster check makes the limit explicit.
- **[Trade-off]** Keeping the name `Class` for roster ids is slightly misleading; it avoids wire and code churn in this change.

## Migration Plan

This is a refactor only, with no data migration. The server, WASM and web ship together, and the version is unchanged, so stored replays still verify. To roll back, revert the PR.

## Open Questions

- Whether skill ids should stay globally shared across units (today `SkillDamage` is "Strength", "Sharp arrows" or "Power" depending on the unit) or become unique per unit when new units arrive. This doesn't affect this change, which keeps today's ids. Decide in step 4.
