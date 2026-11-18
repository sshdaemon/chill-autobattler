# Proposal

## Why

Steps 1–3 made a unit fully describable as data:
- a roster file (stats, traits, tags, five skills built from effects);
- a look entry (recipe, animations, skill animations and effect presets).

Writing those by hand is still slow and easy to get wrong. A new unit also has to be balanced against the others, and the UI still assumes exactly three classes. This is step 4 of the unique-units roadmap. A `/new-unit` workflow turns a short brief and a few reference pictures into a complete, balanced, previewable unit in a PR, which is how the game gets its first truly unique units.

## What Changes

- **A brief per unit.** `design/units/<name>/brief.md` follows a template: concept, role, how it should feel, skill ideas, and anything to avoid. Reference pictures go in `design/units/<name>/refs/`, which is git-ignored: they stay local and are never committed or published, because reference art is often someone else's work.
- **A `/new-unit` Claude Code skill** (`.claude/skills/new-unit/`) that runs the whole flow, with two checkpoints where the user decides:
  1. **Design.** It reads the brief and pictures, then proposes the unit: role, base stats, traits, tags and matchup bonuses, five skills mapped to existing effect kinds, and a look sketch (recipe, animations, effects). It waits for the user's approval. If a skill needs behaviour no effect kind covers, it stops and proposes a separate sim change instead of bending the design.
  2. **Build and balance.** It writes the roster file, look entry, effect presets and UI text. It runs the balance check and tunes only numbers within limits until the unit lands in the target band. It renders previews, runs the tests and opens a PR with the balance report and screenshots. The user reviews and merges.
- **Unit balance check.** A new `pnpm balance:unit <name>` runs seeded battles comparing the new unit with every existing one, in mixed armies and per skill, and reports whether it lands in the target bands.
- **Unit preview.** A dev-only page (`?preview=<name>`) shows the unit's recipe on both teams, with idle, move, each attack, each skill action and effect, death and portrait, for screenshots.
- **Any number of unit types in the UI.** Labels, icons and blurbs come from each unit's look, not fixed three-entry arrays. The picker, info panel, result screen and portraits work for any roster size. The few places that branch on class numbers use roster traits instead (ranged, projectile).
- **Adding units without a rule change.** A new roster entry outside Red's pool doesn't change any existing battle, so it needs no `sim.Version` bump. Adding it to Red's pool does. New unit and skill ids need no proto edit, because proto3 enums are open and the sim's roster is the check.

## Out of scope

- New effect kinds (heal, summon, shield, aura…). The skill detects the need and stops; each is its own OpenSpec change.
- Creatures or non-humanoid rigs, and new outfit parts. The skill combines the existing parts library, and may only add simple items modelled from primitives in the bake (as the dry run did with a spear).
- Generating meshes or textures from the reference pictures.
- Letting players create units in the game.
- The first actual new unit. That is the first run of `/new-unit`, from the user's own brief, in its own PR.

## Capabilities

### New Capabilities
- `unit-authoring`: how a new unit goes from a brief to a PR. Covers the brief format and where references live, the two user checkpoints, the balance targets, the preview, what the PR contains, and when to stop for a sim change.

### Modified Capabilities
- `unit-roster`: adds a requirement on adding units. Units outside Red's pool need no version bump; changing the pool needs one. Unit and skill ids are allocated and published so that clients and server accept new ids without a proto change.
- `match-flow`: the army picker shows every roster entry the player may field, each with label, icon and blurb from its look, instead of a fixed set of three.

## Impact

- **New files:**
  - `.claude/skills/new-unit/SKILL.md` (plus a brief template and a checklist);
  - `design/units/_template/brief.md`;
  - `.gitignore` entry for `design/units/*/refs/`.
- **`tools/balance`:** a `unit` mode (`pnpm balance:unit <name>`) generalised from the fixed 3×3 tables to the roster.
- **`web/`:**
  - `gamedata.ts` loses `CLASS_ICON`, `CLASS_LABEL` and `CLASS_BLURB`;
  - `looks.json` gains `ui: { label, icon, blurb }` per unit;
  - the picker, info panel, overlays and portraits read from looks;
  - `store.ts` and `Units.tsx` stop branching on class numbers;
  - a dev-only preview page.
- **`wasm/looks_test.go`:** every look has UI text and an icon.
- **`sim/`:** no rule change. The roster check gains id allocation rules, and a golden-case template is documented for new units.
- **Docs:** CLAUDE.md and the README describe the workflow.
