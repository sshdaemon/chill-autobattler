# Design

## Context

Plans (`sim/plan.go`) are data: per role a `Stance` that overrides
temperament fields and may grant effects, applied at spawn after the roster
entry and before learned abilities (`temperOf`). A unit's temperament never
changes after spawn today. `Army.Plan` is a `uint8` id; Red draws its plan
from the combat seed (`RedPlanFor`, `IntN(len(Plans))`). Holding units wake on
`teamFighting`, which looks at the whole army. Formation flanks sit at x ±5.5
(`flankX`), and players may drag units anywhere in their zone.

## Goals / Non-Goals

**Goals:**
- Two groups, each with an opening stance, one trigger and a second stance,
  as plain data in the plan table.
- No change to the battle input, the proto, the server or `pb/wire`.
- Deterministic, tested natively and under WASM like every rule.

**Non-Goals:**
- Free choice of groups, stances or triggers (step 5).
- Effects that change mid-battle.

## Decisions

**A tactic is a plan.** `PlanDef` gains `Groups [2]GroupDef` (left, right) and
is a tactic when either group is set. Alternatives: a separate `Army.Tactic`
field on top of the plan. That means a proto and wire change, a second picker
and plan × tactic combinations to balance; one list keeps the input, Red's
draw and the server as they are. Step 5's editor can still split them later.

```go
type Roles struct{ Melee, Ranged, Caster Stance }

type GroupDef struct {
    Opening Roles   // over the plan's own roles; effects allowed (fixed at spawn)
    When    Trigger // its kind and number (seconds)
    Then    Roles   // after the trigger; behaviour fields only
}
```

`PlanDef` embeds `Roles` (so `p.Melee` and friends keep working). A unit's
stance is built in layers: roster → plan role stance → its group's opening
role stance (or, once switched, the group's `Then`) → learned abilities.
Effects are the plan role's plus the opening's, given at spawn.

**Groups by position.** At `New`, after positions are set, each team's
units are split at the team's mean x (`GroupsOf`): x ≤ mean is group 0. Both
teams use −x as "left", because `FormationSlots` puts a preset's first flank
at −x for both teams; so a Red in 3-2 has its three-unit flank as group 0,
like Blue. This matches the formation flanks (the two clusters at ±5.5) and
follows drags. The split is computed once and stored as `Unit.Group`. For a
plan without groups every unit is group 0.

**Triggers are plug-ins**, like target and engage rules: `trigger_<name>.go`
registers under an id in `ids.go`:

| Id | Name | Fires when |
|---|---|---|
| 1 | after | `Tick` ≥ N seconds |
| 2 | ally falls | any unit of its army is dead |
| 3 | other group fighting | a unit of the other group is within reach of its target (the `teamFighting` test) |

`Trigger{Kind, Seconds}`; kind 0 means none (the group never switches).
Triggers are checked once per tick, at the start of `Step` before units
think, for each team in order and each group in order; a group that fires is
marked switched for good and each of its living units gets its temperament
rebuilt with `temperOf` from the stored layers and its kit. Unit-level
counters (`held`, `lineSet`) are reset so the new stance starts clean. The
check reads only battle state, so it's deterministic.

**Group-scoped waking.** `teamFighting(u)` only counts allies in `u.Group`.
Without a tactic everyone is group 0, so today's battles are unchanged (the
goldens prove it before the version bump). `HoldPatience` still guarantees a
holding group goes in eventually, so tactics can't stall.

**Flank engage rule** (`engage_flank.go`, id 4). With no target, the unit
moves toward a point beside the enemy army: the enemy's centre (from what
its team sees, else the far side of the field as `advance` does) shifted
sideways by `FlankWide` (8 m, in `data.go`) toward the unit's own side of the
field, until it is level with that centre; then it advances like a charge.
With a target it fights normally (no `engage` hook). Pathfinding and
obstacles are handled by `moveToward` as for any move.

**Validation** (`ValidatePlans`): for each tactic, every (role, group, stage)
stance must give a valid temperament on every roster entry with each single
skill, as today; `Then` stances may not carry effects; each trigger's kind
must be registered and its number in range (after: 1–120 s); Instinct must
stay empty (no groups).

**Built-in tactics** (ids 6–8 in `ids.go`, one file each):

| Tactic | Left (opening) | Right (opening) | Right trigger → then | Left trigger → then |
|---|---|---|---|---|
| Hammer and anvil | Shield wall stances (incl. its front-damage effect for melee) | melee Flank + Nearest, +15% speed; ranged Focus within reach | other group fighting → melee Charge, Nearest | none |
| Reserve | Line, Focus | Hold (trigger 6 m), Nearest | ally falls → Charge, Nearest | none |
| Hold, then charge | Hold (trigger 6 m) | Hold (trigger 6 m) | after 10 s → Charge, Nearest | after 10 s → Charge, Nearest |

After the balance pass (`docs/balance-findings.md`): holding ranged and
casters get a 10 m trigger so they shoot what comes into reach; Reserve's
front melee use Nearest and its reserve gets +5% damage; Hold, then charge
waits 10 s with the melee braced (8% less damage from the front).

**Red's draw** stays `IntN(len(Plans))`, now over nine plans. The stream is
unchanged; only the draw's range changes, which changes which plan a generated
Red gets for a given seed. That's a rule change: `RulesVersion` 0.11.0 and
the golden files are re-recorded.

**Client.** `data().plans` gains `tactic: bool`; `DeployBar` shows a
"Tactics" row below the plans with the same buttons and description line.
Group marks: the preview battle already runs in the worker from Blue's army;
`units()` reports each unit's group, and the deploy view draws a small L/R
tag under each Blue unit while a tactic is chosen. The split is the sim's, so
TS never re-implements it. Words for the tactics go in `web/src/plans.ts`.

## Determinism

Everything new is integer/Fx data read at spawn or from battle state once
per tick in a fixed order (teams, then groups, then unit index). Tests:
- Golden replays: one case per tactic (Blue on the tactic against an
  explicit Red on another tactic), run natively and under WASM by
  `pnpm test` like the others; every existing case still matches its
  outcome with Red forced to the old draw before re-recording, which shows
  the group-scoped waking changes nothing without tactics.
- `TestPlansFight` covers all nine plans against each other for every faction
  pair: no stall.
- Unit tests: group split (3-2 formation, a dragged unit, the centre-line
  tie, Red mirrored); each trigger fires on its condition and only once;
  effects don't change on switch; a `Then` with effects is rejected; Reserve's
  right group keeps holding while only the left fights.

## Risks / Trade-offs

- **Flank paths on cramped maps:** a wide swing may run into ponds or rocks.
  `moveToward` paths round them; if a map leaves no room, the unit simply
  takes the long way or meets an enemy first. Watch it in the browser.
- **Tactics vs plans balance:** nine plans make a 36-pair matrix per
  faction. The goal stays: nothing beats everything.
- **Uneven groups:** an army dragged all to one side splits at its own centre
  anyway, so both groups have units (unless only one unit lives).
- **Red's draw changes for every seed**, so old replays of a generated Red
  stop matching. Replays carry `sim.Version`, so they're refused rather than
  misread.
