# Proposal

## Why

Every unit is one of three classes (warrior, archer, mage), and their behaviour and skills are written directly into the sim's code. Examples are `u.Class == Mage`, `if u.Has(SkillBlink)` and a 3×3 matchup table. Every new unit or skill therefore means editing the core sim logic, which is why units feel generic and new ones are slow to add. This is step 1 of the unique-units roadmap. It moves units and skills into data, so that later steps (animations, a modular look, a `/new-unit` workflow) can add a unit mostly by writing a definition.

## What Changes

- **A unit roster.** Each unit type is a definition in the sim:
  - id and name;
  - base stats;
  - role traits (melee or ranged, kites, projectile type);
  - matchup tags;
  - its five learnable skills.
  
  Warrior, archer and mage become the first three entries, with the same numbers as today.
- **Skills built from effects.** Each skill is a list of effect blocks with parameters, taken from a fixed set:
  - stat bonuses (health, damage, range, speed, attack speed);
  - finisher (execute);
  - slow on hit (frost);
  - splash on hit (blast);
  - gap-closer (dash);
  - escape teleport (blink);
  - shooting over cover (lob).
  
  The sim runs effects at fixed hook points, not through checks for specific skills.
- **Matchups from tags.** The damage multiplier comes from the attacker's bonuses against the victim's tags (for example, warrior ×1.45 against `ranged`), not from a class-by-class table.
- **Checks on the roster.** A test rejects bad definitions: duplicate ids, wrong skill counts, missing projectile speed on a ranged unit, unknown tags. Another test builds a test-only unit from existing effects without new sim code, to prove that a new unit is data only.
- **Roster for the client.** `chillSim.data()` in WASM returns the roster: units, skills with their effects and parameters, and tags. Server skill-pick checks read the roster.
- **No rule change.** Every battle plays out exactly as before. All golden replays must reproduce bit for bit, natively and under WASM, and `sim.Version` stays 0.4.0.

## Out of scope

- New units or skills for players. This change only re-expresses the existing three classes and fifteen skills.
- Animations, effects presets, models and the modular look (steps 2 and 3), and the `/new-unit` workflow (step 4).
- Changing the wire format. The proto enums `UnitClass` and `Skill` keep their numbers, and the snapshot layout is unchanged.
- Balance changes and changes to how Red's army is generated. Red still draws from the same three units in the same order.
- Loading unit definitions from files at runtime, or user-made units.

## Capabilities

### New Capabilities
- `unit-roster`: unit types and skills defined as data (stats, traits, tags, skills made of effect blocks), how the roster is checked, and how it is exposed to the client and server.

### Modified Capabilities
- `battle-rules`: "Classes and stats", "Damage" and "Skills" are restated in terms of the roster: units come from roster entries, matchups come from tags, and skills come from effects. The numbers and outcomes stay the same.

## Impact

- **`sim/`:**
  - `data.go` (class and skill tables move into roster definitions, one file per unit);
  - `unit.go` (stat setup, dash and blink become effects);
  - `combat.go` (execute, frost, blast and lob become hit hooks; the bolt comes from the unit's projectile type);
  - `army.go` (Red picks from the roster's Red pool, `SkillDefOf`);
  - new roster validation and data-only unit tests.
- **`wasm/main.go`:** `data()` returns the roster; its test is updated.
- **`server/run/run.go`:** skill-pick validation goes through the roster.
- **`tools/balance`:** reads classes and skills from the roster.
- **No changes to** the proto, the snapshot or the web renderer, and no new dependencies.
