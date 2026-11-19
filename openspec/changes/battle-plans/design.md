# Design

## Context

A unit's temperament (`sim/temperament.go`) is built once at spawn by
`temperOf(def, kit)`: the roster entry's `Temper`, then each learned ability's
`temper` hook. Target and engage rules are plug-ins registered by id
(`ids.go`). Holding units already wake after `HoldPatience` (40 s), so two
holding armies still meet. Red's army, skills and formation come from the map
seed (`RedArmy`), and its faction from `RedFactionFor(seed)`. The combat seed
is revealed only after Blue's army is locked. The version string hashes the
roster and factions (`refreshVersion`), and each golden case records
`VersionFor` its own units.

## Goals / Non-Goals

**Goals:**
- Plans are data, one file per plan, registered by id like factions and
  effects, so adding a plan later is one new file.
- The plan shape is the one that custom plans (step 5) and free-text orders
  (step 6) will fill in: per-role overrides of existing temperament fields.
- Plan 0 (Instinct) is provably the old behaviour.

**Non-Goals:**
- No new target or engage rules, no per-group stances and no triggers.
- No client changes beyond regenerated protobuf types.

## Decisions

### A plan is per-role overrides, not a whole temperament

```go
type Stance struct {
	Target   TargetRule // 0: keep
	Engage   EngageRule // 0: keep
	Trigger  Fx         // 0: keep (set whenever Engage is Hold)
	Leash    Fx
	Keep     Fx
	FallBack int
	StayNear Tag
	Kite     Toggle     // KeepKite, KiteOn, KiteOff
}
type PlanDef struct {
	ID                    Plan
	Name                  string
	Melee, Ranged, Caster Stance
}
```

A zero field keeps the unit's own value. `Kite` needs three states, because
"off" must differ from "keep". A unit's role is caster if it has the caster
tag, else ranged if its stats are ranged, else melee.

*Alternative:* a plan replaces the whole temperament per role. Rejected: it
erases unit character (a völva's preferred target, a siphonarios staying near
the melee) and makes plans longer to write.

*Limit:* a plan can't clear a field to zero (for example, remove a unit's
`StayNear`). None of the built-in plans needs that. Step 5 can add it if
custom plans do.

### Order: roster, then plan, then abilities

`temperOf(def, plan, kit)` applies the side's stance for the unit's role before
the ability hooks. A skill that makes a unit hunt ranged units still works
under any plan, so leveling a unit keeps its value.

### Every plan is checked against every unit at start-up

An init check applies each plan to each roster entry and runs
`checkTemperament`. A bad combination (kiting melee, holding without a trigger)
fails the tests and `ValidateRoster`, not a battle. Each plan that sets
`Engage: Hold` also sets a `Trigger`.

### Plan ids and Instinct

`type Plan uint8` in `ids.go`, with explicit ids: `PlanInstinct = 0`,
`PlanShieldWall = 1`, `PlanCharge = 2`, `PlanHunt = 3`, `PlanFocusFire = 4`,
`PlanSkirmish = 5`. Instinct is registered like the others (an empty plan), so
"no plan" and "Instinct" are the same thing. This also makes the proto3 default
(0) mean Instinct, and old stored armies read as Instinct.

### Red's plan comes from the combat seed

In `New`, when Red's army is generated, Red's plan is
`Sub(CombatSeed, StreamRedPlan).IntN(len(Plans))`, over all plans including
Instinct. `StreamRedPlan` is added at the end of the stream list, so no other
stream's value changes. The battle keeps both resolved plans
(`Battle.Plans [2]Plan`) so step 2 can show Red's plan.

*Alternative:* draw from the map seed like Red's formation. Rejected: the
client knows the map seed during the draft, so a player could compute Red's
plan before locking their army. The roadmap's decision is that the strategy
isn't counter-picked.

`TestCombatSeedChangesOnlyTheFight` still holds: the map and Red's army don't
change with the combat seed. Only Red's plan, which is part of the fight, does.

### Version

`RulesVersion` goes to 0.9.0. `refreshVersion` and `VersionFor` also hash the
plan table, so any later change to a plan changes the version.

### Built-in plans (starting values, tuned in step 3)

| Plan | Melee | Ranged | Caster |
|---|---|---|---|
| Instinct | — | — | — |
| Shield wall | Hold, trigger 8 m, Protect | Line, FocusReach, stay near melee | Line, FocusReach, stay near melee |
| All-out charge | Charge, Nearest | Charge, Nearest, kite off | Charge, Nearest |
| Hunt the archers | Charge, RangedFirst | FocusReach | FocusReach |
| Focus fire | Line, Focus | Line, FocusReach | Line, FocusReach |
| Skirmish | Hold, trigger 6 m, fall back at 40% | kite on, keep 8 m | keep 8 m |

### Determinism

Plans are fixed Go tables in fixed-point, applied once at spawn. There is no
new randomness except one draw from a seeded stream, and no floats or map
iteration. Native and WASM agree through the existing golden replays, which
run on both. Plan 0 equivalence is checked directly (see Testing).

### Testing

- **Plan 0 is the old behaviour:** before the change, record each golden
  case's result, tick count and final-state hash. With Red's plan forced to
  Instinct by a test-only hook, every case must match. After that, goldens are
  re-recorded with real draws.
- **New golden cases:** one per plan, with explicit armies on both sides.
- **No stalls:** extend the mirror test to every plan pair over every faction
  pair with a few seeds, and require that no battle runs out of time. This
  covers Shield wall vs Shield wall and Skirmish vs Skirmish.
- **Validation:** an unknown plan in either army is rejected.
- **Red's draw:** same combat seed gives the same plan; the plans drawn over
  many seeds cover all six.
- **Order:** a unit with a ranged-first ability under Focus fire targets ranged
  units first.
- The mutual-wipe seeds in `golden_test.go` may need new values after
  re-recording, as happened when factions became playable.

## Risks / Trade-offs

- [One plan dominates, e.g. All-out charge] → the balance matrix shows it; step 3 tunes the values.
- [Golden churn: every golden with a generated Red changes] → the plan-0 check above proves the change is only the new draw.
- [Kite off plus a unit whose ability gives kite-shoot] → the start-up check covers each roster entry with each plan and each skill's temper hook.
- [Red's plan visible in replays during the run] → replays are shown only after the battle, when the combat seed is already public.

## Migration Plan

The server already rejects battles from another sim version, so a running
battle that started before the deploy fails to verify, as with any rules
change. Stored battles without a plan read as Instinct.
