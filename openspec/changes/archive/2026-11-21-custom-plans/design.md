# Design

## Context

A built-in plan is a `PlanDef` (`sim/plan.go`): a `Stance` per role (target,
engage, trigger, leash, keep, fall back, stay near, kite, effects) and, for a
tactic, two `GroupDef`s (opening, trigger, then). Zero fields keep the unit's
own value. `Army.Plan` is a `uint8` id; the battle looks plans up with
`PlanOf`. The battle input is protobuf; the server stores the locked `Army`
as protobuf and re-runs it; the browser decodes with `pb/wire`. The deploy
countdown (15 s) is client-only; the server allows 3 minutes from battle start
to lock.

## Goals / Non-Goals

**Goals:** a chip vocabulary that maps one-to-one onto `PlanDef`; edges that
can't stack into free buffs; the same checks as built-in plans; an editor that
fits deployment; a data shape a language model can fill in Step 6.

**Non-Goals:** named or shared plans; new rules; Red custom plans.

## Decisions

**Chips are a small, closed data type in the sim**, mirrored in proto:

```go
type Step uint8   // 0 own, 1 close/short, 2 mid/long, 3 far
type Chips struct {
    Target   TargetRule // 0 own
    Engage   EngageRule // 0 own
    Trigger  Step       // hold trigger: 6 / 8 / 10 m
    Keep     Step       // 5 / 7 / 8 m
    Leash    Step       // 10 / 14 m
    FallBack uint8      // 0 own, 1 below 50%, 2 below 30%
    StayNear uint8      // 0 own, 1 the melee, 2 ranged and casters
    Kite     Toggle
    Edge     Edge       // 0 none; ids in ids.go
}
type CustomGroup struct {
    Opening, Then [3]Chips // melee, ranged, caster
    When CustomTrigger    // kind and seconds (5/10/15/20)
}
type CustomPlan struct {
    Roles  [3]Chips
    Groups []CustomGroup // none, or left and right
}
func (c *CustomPlan) Def() (PlanDef, error)
```

`Def()` maps steps to metres through a table in `data.go` (the presets' own
values: triggers 6/8/10, keep 5/7/8, leash 10/14) and edges to their effects,
then the result goes through the same validation as built-in plans
(`validatePlan`, split out of `ValidatePlans`). A raw-number custom plan was
the alternative; coarse steps keep it balanceable and give a language model a
small, closed set of words.

**One plan definition per side in the battle.** `Battle` keeps
`planDefs [2]*PlanDef` (from `PlanOf` or `Custom.Def()`), and the code that
used `PlanOf(b.Plans[team])` (triggers, spawning) uses it. `Battle.Plans`
reports `PlanCustom` (255, reserved in `ids.go`) for a custom side, so the
client shows "Custom".

**Input:** proto `Army.custom_plan` (`CustomPlan { Roles roles = 1; repeated
Group groups = 2; }`, enums as uint32 like the plan id). With `custom_plan`
set, `plan` must be 0. `pb/convert` and `pb/wire` build `sim.Army.Custom`;
the wire fuzz test generates random custom plans. The server needs no new
code: `LockDeployment` validates through the sim, the stored army bytes carry
the custom plan, and `VerifyBattle` re-runs it.

**The edge menu** (`sim/edge_<name>.go`, registered like plans; ids in
`ids.go`). Each pairs an advantage with a cost and names the roles it suits.
Starting values, tuned in the balance pass:

| Edge | Roles | Advantage | Cost |
|---|---|---|---|
| Braced shields | melee | 7% less damage from the front | −9% damage |
| Reckless | all | +25% speed, +3% damage | takes 6% more damage |
| Drilled | melee | +10% damage | takes 8% more damage |
| Long reach | ranged, caster | +20% range | −5% damage, −5% health |
| Light feet | ranged | +20% speed | −4% health |

These are the values after the balance pass. Two lessons from it: speed is
worth little as an advantage or a cost, and a damage-taken cost barely touches
a back line, so Drilled is melee only (on archers it was a free +10% damage,
and the search's best plans stacked it on every role).

Built-in plans keep their own effects (their costs are partly behaviour, like
Shield wall giving up the first move). The editor maps a preset's effects to
the menu edge with the same advantage when it starts from a preset, so an
edited preset pays the menu's cost.

**Rules version.** The step table and the edge menu go into
`plansFingerprint`, so changing a step or an edge changes `sim.Version`.
`RulesVersion` bumps for the new input field and code.

**WASM:** `data().chips` (targets, engages, steps in metres, fall back, stay
near, triggers, edges with roles and effects) and `data().plans[i].chips` (each
built-in plan as its nearest chips, for the editor to start from);
`chillSim.checkPlan(json)` returns "" or the first error, so the editor can
say what's wrong before the lock. The JSON shape is `CustomPlan`'s; Step 6's
model output targets the same shape.

**Editor** (`web/src/ui/PlanEditor.tsx`): opened by Edit in the deploy bar.
Tabs: Army, and with "split into groups" on, Left and Right, each with Before
/ After and a trigger. One row per role; chips are buttons; a sentence under
each row says what it does (words in `web/src/plans.ts`). The store keeps
`custom?: CustomPlan` and `plan = CUSTOM (-1)` when chosen; `blueArmy` sends
`customPlan` then. Opening pauses the 15 s countdown (the server's 3-minute
window bounds it: the panel shows the time left on it when under a minute).
The custom plan is kept for the run in the store's resume state; a new run
clears it.

**Balance tool:** `-exp edges` (each edge, on the roles it suits, with
Instinct behaviour, against plain Instinct, every playable faction) and
`-exp custom` (per faction: 300 random custom plans against the nine built-in
plans at 100 battles per pair, the 5 best hill-climbed one chip at a time for
40 steps at 200 per pair, the best checked at 1000 per pair). Goal: no edge
over 60% alone, and the best custom plan doesn't beat every built-in one.

## Determinism

Chips map to fixed Fx tables; `Def()` is pure; the battle runs the resulting
`PlanDef` exactly like a built-in one. Golden cases: a custom plan per side
(one army-wide, one with groups and each trigger kind), native and WASM.
Tests: chips → `PlanDef` equals the hand-written equivalent; an empty custom
plan plays exactly like Instinct; invalid chips, edges on the wrong role or on
a second stance, and unknown triggers are rejected (sim, convert, wire,
server `LockDeployment`); a server round trip verifies fair.

## Risks / Trade-offs

- **A combination nobody tested** might beat the presets: the search is the
  guard; anything it finds is fixed by tuning edges or steps, not by
  forbidding chips.
- **Editor under time pressure:** the countdown pauses, but the server window
  doesn't; three minutes is ample, and the plan is kept for the run.
- **Proto growth:** a few small messages; `pb/wire` must keep matching
  `pb/convert` (the fuzz test).
