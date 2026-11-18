# Design

## Context

See proposal.md for the motivation. After steps 1–3, a unit is:
- `sim/unit_<name>.go`: a `UnitDef` with stats, traits, tags, matchup bonuses, five `SkillDef`s made of effects, and `InRedPool`;
- a look in `web/src/units/looks.json`: recipes from the parts library, an animation set, and per skill an action, attack, pitch and effect preset;
- possibly new presets in `web/src/scene/fx.json` and new clips in `looks.json` `clips`, packed by `tools/assets/ual_clips.py`.

`ValidateRoster` and `wasm/looks_test.go` check all of it.

**What still assumes three classes:**
- `gamedata.ts` `CLASS_ICON`, `CLASS_LABEL` and `CLASS_BLURB`, used by the picker, info panel, overlays and portraits;
- `store.ts` picks the shot sound with `cls === 2`;
- `Units.tsx` turns the upper body only for `cls !== 0`;
- `effects.ts` `fxDemo`;
- `tools/balance` uses `[3]` tables and the warrior, archer and mage constants.

**Wire format:** proto `UnitClass` and `Skill` are proto3 enums, which are open. Unknown numbers decode as plain numbers in Go and in protobuf-es. `pb/convert` and `pb/wire` map by number, and the sim's `Input.Validate` rejects ids the roster lacks.

## Goals / Non-Goals

**Goals:**
- One command, two user checkpoints, and a PR that is complete and reviewable.
- Balance by measurement, not guesswork.
- Nothing the user provided as a reference leaves their machine.

**Non-Goals:**
- Full autonomy. The design checkpoint is mandatory; the user decides what the unit is.
- New effect kinds, parts or rigs (see proposal.md, Out of scope).

## Decisions

### 1. The workflow is a project skill, not a script
`.claude/skills/new-unit/SKILL.md` holds the procedure. Design judgement is the agent's job: reading a brief and pictures, mapping ideas to effects, choosing parts and clips. A script couldn't do that. The skill leans on deterministic tools for everything measurable:
- `ValidateRoster` and the look check;
- `pnpm balance:unit`;
- the dev preview;
- `pnpm test`.

The skill ships with:
- `design/units/_template/brief.md`, the brief format;
- `checklist.md`, every file a unit touches, and the id rules;
- an **effect cookbook** in the skill: how common fantasy abilities map to the existing effect kinds, and which ones don't map. For example:
  - charge → gap-closer;
  - poison or chill → slow on hit;
  - cleave or explosion → splash on hit;
  - assassin → finisher;
  - teleport → escape;
  - lob → over cover;
  - heal, summon, shield, aura, stun → not available, so stop.

### 2. The two checkpoints
- **Design checkpoint.** The skill writes a design summary: a table of stats against the existing units, each skill with its effects and numbers, a recipe sketch, chosen clips and presets. It then asks the user (AskUserQuestion: approve / change / stop). Free-text changes loop back.
- **Build checkpoint.** After building and balancing, the skill opens the PR. The user's merge is the approval, under the project's usual rule (merge only when the user says so).

There's no checkpoint in between: tuning numbers to hit the bands is mechanical and recorded in the PR.

### 3. Balance: `pnpm balance:unit <name>`
A new mode in `tools/balance` (`-unit name -n 1000`), generalised from the fixed 3×3 code to `len(sim.Roster)`:

| Check | Battles | Band |
|---|---|---|
| **Head-to-head** | 5×new against 5×each other unit, Spread on both sides, `n` seeds | 25–75% wins |
| **Swap** | today's standard army (2W 2A 1M) against a generated Red, compared with the same army with one same-role unit swapped for the new one (melee for melee, ranged for ranged) | win rate within ±10 points |
| **Skills** | the new unit with each skill alone against without it, in the swap army | 0 to +15 points each |

It prints a table with pass or fail per band and exits non-zero on any fail.

**Tuning:** the skill changes only HP, damage, cooldown, range, speed and skill numbers, at most ±30% from the approved design. It makes at most 6 rounds. If the bands still fail, it reports why (for example "hard-countered by archers whatever its numbers") and lets the user decide.

The new unit isn't in Red's pool, so the existing Red-scaling and `docs/balance-report.md` results don't change. The report gains a per-unit section only when a unit is added to the pool.

### 4. Preview: a dev-only route
In dev builds only, `?preview=<name>` mounts a `UnitPreview` instead of the game. It shows:
- the unit's recipes on both teams;
- a strip of clips (idle, move, each attack variant, each skill action with its effect preset firing, death);
- the portrait.

It uses the same `makeCharacter` and `effects` code as battles. `window.previewShot(step)` advances to a named step for repeatable screenshots, and the skill takes screenshots through the browser pane. The screenshots are saved under `design/units/<name>/preview/` and committed: they're our own renders, not references.

### 5. UI from looks
- **Per unit:** `looks.json` gains `ui: { label, icon, blurb }` per unit (icon is an emoji or a short glyph).
- **Removed:** `gamedata.ts` loses the three arrays, and the components read `lookOf(cls).ui`.
- **Picker:** a flexible grid, so 3–6 cards fit at desktop width, wrapping on narrow screens.
- **Class-number branches:**
  - the bolt sound comes from `data.classes[cls].projectile === 'bolt'`;
  - upper-body aim comes from `ranged`;
  - `fxDemo` uses the roster.
- **Look check:** every look must have UI text.

### 6. Ids and versioning
- **Ids:** the roster check rejects a gap or a reuse in unit ids, which must equal their position. Skill ids must be unique across the roster unless two units share a skill's identity and wire number on purpose, as the existing Toughness, Damage and Execute do. New skills take the next free number. The checklist records the rule.
- **Versioning:** a new unit with `InRedPool: false` needs no version bump. The skill adds a golden case with the unit, which is allowed without a bump because recorded cases don't change. A later "add to Red's pool" change is separate and bumps the version.
- **Proto:** enums stay as they are, documenting the original three. A comment says that further ids come from the roster.

## Determinism

Making the UI and the balance tool roster-generic changes no rules. `sim.Version` stays the same and golden replays are untouched. Every unit the skill adds comes with a golden case and must pass native and WASM replays.

## Risks / Trade-offs

- **[Risk]** The agent maps a brief to effects that don't capture the fantasy. → Mitigation: the design checkpoint shows the mapping explicitly, the cookbook says what doesn't map, and the stop rule applies.
- **[Risk]** Balance bands are too loose or strict for interesting units, such as a glass cannon that is 80/20 against one class. → Mitigation: the bands are in the spec and can be argued per unit in the PR. The skill reports, never hides, a band it couldn't meet.
- **[Risk]** Reference pictures leak. → Mitigation: `refs/` is git-ignored, the skill never uploads or links them, and the PR template says so.
- **[Trade-off]** A new unit isn't in Red's pool, so players meet it only in their own armies until a pool change. That's deliberate: it keeps replays stable and the Red-scaling balance untouched.

## Migration Plan

The UI and balance tool become roster-generic, and the skill, template and preview are added. Nothing is migrated. To roll back, revert the PR.

## Open Questions

None that change the plan. The first unit's brief comes from the user after this lands.
