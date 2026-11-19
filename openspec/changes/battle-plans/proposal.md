# Proposal

## Why

Every unit fights only by its own temperament, so battles between the same
armies play out much the same way, and the player has no say in how their army
fights. Battle plans give each side an army-wide stance chosen before the
battle. This is step 1 of `docs/battle-plans-roadmap.md`: the rules in the sim,
which the plan picker (step 2), tactics (step 4), custom plans (step 5) and
free-text orders (step 6) all build on.

## What Changes

- **Battle plans:** a fixed set of built-in plans, each changing the
  temperament of the side's melee, ranged and caster units: Instinct (no
  change, the default), Shield wall, All-out charge, Hunt the archers, Focus
  fire and Skirmish.
- **Plan in the army:** each army in the battle input carries a plan (0 =
  Instinct). The input check rejects an unknown plan. The protobuf `Army` gets
  a `plan` field, so the server stores and replays Blue's plan.
- **Order of changes to a temperament:** roster, then the side's plan, then
  learned skills.
- **Red's plan:** when Red's army is generated, Red's plan is drawn at random
  among all plans from the combat seed. The combat seed is revealed only after
  Blue's army is locked, so the plan can't be known during the draft.
- **BREAKING (rules):** `RulesVersion` goes to 0.9.0. Generated Red armies now
  get a plan, so recorded goldens with a generated Red and the Red baseline are
  re-recorded. A battle where both sides use Instinct plays out exactly as
  before.
- **Balance tool:** a plan-vs-plan win matrix.

Out of scope:

- Any client UI: the plan picker and showing Red's plan are step 2. Until
  then the client always sends Instinct.
- Tuning the plans for rock-paper-scissors balance (step 3). This change sets
  starting values and reports the matrix.
- Groups, triggers and phase switches (step 4), custom plans (step 5), and any
  language model (step 6).
- New target or engage rules: the built-in plans only use the existing ones.

## Capabilities

### New Capabilities
- `battle-plans`: army-wide battle plans: the built-in set, how a plan changes
  its side's units, plans in the battle input, and Red's hidden plan.

### Modified Capabilities
- `battle-rules`: the Temperament requirement now says a side's plan changes
  its units' temperament, between the roster and learned skills.

## Impact

- `sim/`: plan data (one file per plan), plan ids in `ids.go`, `Army.Plan`,
  input check, applying a plan at spawn (`temperOf`), Red's plan draw (a new
  RNG stream added last), `RulesVersion`, goldens and Red baseline.
- `proto/chill/v1/units.proto` and generated Go and TS: `Army.plan`.
- `pb/convert`: carry the plan between protobuf and the sim.
- `tools/balance`: plan-vs-plan matrix.
- Server and client code are otherwise unchanged. Old stored battles have no
  plan, which reads as Instinct.
