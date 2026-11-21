# Proposal

## Why

Players can pick one of nine built-in plans and tactics, but can't say what
they actually want ("archers keep their distance, melee hold until the enemy
is close, right flank goes round"). Step 5 of `docs/battle-plans-roadmap.md`:
let the player build a plan from chips. It is also the ground Step 6 stands
on: whatever a language model makes of free text will be shown back as these
chips for the player to check and fix.

## What Changes

- **Custom plans** built from a small chip vocabulary, per role (melee,
  ranged, caster): target rule, engage rule, hold trigger, keep distance,
  leash, fall back, stay near, kite, and one **edge** from a fixed menu.
  Every distance is a coarse step (close, mid, far), not raw metres.
- **Tactics too:** a "split into groups" switch adds a left and a right group,
  each with an opening stance, one trigger (after 5/10/15/20 s, first ally
  falls, other group fighting) and a second stance, like the built-in tactics.
- **Edges, one per role, from a fixed menu:** each is an edge with a cost,
  balanced on its own (braced shields, reckless speed, drilled blows, long
  reach and a few more). Players can't stack free bonuses; the second stance
  of a group can't add one.
- **In the battle input:** `Army.custom_plan`, used instead of the plan id.
  The sim builds a plan from the chips and checks it with the same rules as
  the built-in ones; the server checks it like any other input. Red still
  draws a built-in plan.
- **Editor:** an Edit button next to the plan picker opens a chip panel over
  the field during deployment, starting from the chosen plan. The deploy
  countdown pauses while it is open (the server's 3-minute window still
  applies). The custom plan is remembered for the run.
- **Balance:** each edge alone, and a search over random and hill-climbed
  custom plans, against the nine built-in plans: no custom plan beats them
  all.
- **BREAKING (rules):** chip tables and edges are rules; `RulesVersion` bumps.

Out of scope:

- Free text and language models (Step 6).
- Saving several named custom plans, or sharing them.
- New behaviour (target or engage rules, triggers) beyond what exists.
- Red using custom plans.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-plans`: custom plans, the chip vocabulary, the edge menu, their
  checks and balance goal.
- `match-flow`: the plan editor at deployment.

## Impact

- `proto/chill/v1/units.proto`: `CustomPlan` and its messages; regenerated
  `pb/` and `web/src/gen/`.
- `pb/convert`, `pb/wire` (decode the custom plan; the fuzz test covers it).
- `sim/`: chip tables, edge registry (`edge_<name>.go`, ids in `ids.go`),
  `CustomPlan` → `PlanDef`, validation, fingerprint, `Army.Custom`.
- `wasm/main.go`: chip vocabulary and edges in `data()`, presets as chips,
  `checkPlan(json)`.
- `web/`: plan editor panel, store, countdown pause, words for chips and edges.
- `tools/balance`: edge check and custom-plan search.
- `server/`: tests only (the army is stored as protobuf already).
