# Proposal

## Why

A battle plan gives the whole army one stance for the whole battle. Real
battle moves need two parts of the army doing different things and changing
course when something happens: one flank holds while the other swings round,
a reserve waits until the line breaks. This is step 4 of
`docs/battle-plans-roadmap.md`.

## What Changes

- **Tactics are plans with groups.** A plan may split the army into two
  groups, left and right, each with its own stance and **one trigger** that
  switches the group to a second stance for the rest of the battle. Tactics
  are picked like plans and travel in the same `Army.plan` field, so nothing
  changes in the proto, the server or the wire decoder.
- **Groups come from starting positions:** the units on the left half of the
  army (the side of a formation's first flank) and those on the right, so formations and
  dragged units both work.
- **Triggers** (one plug-in each, ids in `sim/ids.go`): a set time passes, the
  first ally falls, the other group is fighting.
- **A holding unit in a group is woken by its own group's fight**, not the
  whole army's, so a reserve stays back.
- **A trigger switches behaviour only:** target, engage, leash, kite and the
  other temperament fields. The effects a plan grants (its edge and cost) are
  fixed at spawn.
- **A new engage rule, Flank:** with no target, the unit swings wide round the
  enemy's side instead of walking straight at it, so a group can go round.
- **Three built-in tactics:**
  - **Hammer and anvil:** the left group holds as a shield wall; the right
    group goes round the flank and charges once the left is fighting.
  - **Reserve:** the left group advances in line; the right group holds,
    without being drawn in by the left group's fight, and joins in when the
    first ally falls.
  - **Hold, then charge:** the army holds, braced, for 10 seconds, letting the enemy
    come, then charges the nearest enemy.
- Red draws among all plans, tactics included, hidden until the fight as
  today.
- **Client:** the plan picker shows tactics in their own row; while a tactic
  is chosen, the deployment view marks which group each unit is in, using the
  sim's own split.
- **BREAKING (rules):** `RulesVersion` goes to 0.11.0; Red's plan draw now
  covers nine plans, so goldens with generated Reds are re-recorded.

Out of scope:

- Choosing groups, stances or triggers freely (step 5's chip editor).
- More than two groups or more than one trigger per group.
- Triggers that change stat effects mid-battle.
- Triggers on enemy distance or group health (step 5 can add them as more
  trigger plug-ins).
- A banner or sound when a trigger fires.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-plans`: plans may have groups and triggers; three built-in tactics;
  the Flank engage rule.
- `match-flow`: the plan pick offers tactics and shows groups at deployment.

## Impact

- `sim/`: `PlanDef` gains groups, triggers as plug-ins, the Flank engage rule,
  three `plan_<name>.go` tactic files, ids in `ids.go`, a mid-battle
  temperament switch, `RulesVersion` 0.11.0, re-recorded goldens.
- `wasm/main.go`: `data().plans` marks tactics; units report their group.
- `web/`: plan picker row for tactics, group markers at deployment, words for
  the tactics.
- `tools/balance`: the plan matrix covers the tactics.
- Docs: roadmap status and balance findings.
