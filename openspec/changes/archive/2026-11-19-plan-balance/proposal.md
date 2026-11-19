# Proposal

## Why

The first plan-vs-plan matrix (`docs/balance-findings.md`, sim 0.9.0) showed
no real choice: All-out charge and Instinct beat almost every plan, and
Skirmish lost almost everywhere. A plan could only change how units pick
targets and move, and the units' own temperaments are already well tuned, so
every plan made some units worse with nothing in return. This is step 3 of
`docs/battle-plans-roadmap.md`.

## What Changes

- **Plans get trade-offs:** a plan's stance for a role may grant effects for
  the battle, as if learned, applied after the unit's own traits and skills.
  Each plan gets an edge, and most a cost:
  - Shield wall: melee hold and take 8% less damage from the front; melee now
    meet the nearest enemy (the protect rule made them chase and lose).
  - All-out charge: +20% speed, takes 3% more damage.
  - Hunt the archers: melee +25% speed, take 5% more damage.
  - Focus fire: +10% damage.
  - Skirmish: melee screen the back line without chasing (line, 10 m leash,
    no longer hold and fall back); ranged +30% range and +15% speed; casters
    +30% range.
- **Balance goal:** with the same army on both sides, no plan beats every other
  plan in any playable faction.
- **Balance tool:** the plan matrix swaps sides every other battle and marks a
  dominating plan.
- **BREAKING (rules):** `RulesVersion` goes to 0.10.0; goldens with plans are
  re-recorded.
- Plan descriptions in the client mention each plan's edge and cost.

Out of scope:

- Making every plan good for every faction: some plans stay weak for some
  rosters (Hunt the archers for the Rus), which is faction character.
- New effect kinds, groups and triggers (step 4), custom plans (step 5).

## Capabilities

### New Capabilities

### Modified Capabilities
- `battle-plans`: plans may grant effects; the plan descriptions change; a new
  requirement that no plan dominates.

## Impact

- `sim/plan.go` (Stance.Effects, validation), `sim/effects.go` and
  `sim/effect_stat.go` (extra effects when learning and folding stats),
  `sim/unit.go`, the five plan files, `RulesVersion`, goldens.
- `tools/balance/plans.go`.
- `web/src/plans.ts` (descriptions).
