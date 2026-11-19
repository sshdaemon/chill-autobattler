# Spec Delta

## MODIFIED Requirements

### Requirement: Built-in plans
The game SHALL have these built-in battle plans:
- **Instinct:** no change to any unit. It is the default.
- **Shield wall:** melee units hold, meet the nearest enemy and take less damage from the front; ranged and caster units focus on enemies within reach and stay behind the melee.
- **All-out charge:** every unit charges the nearest enemy faster than usual, ranged units don't kite, and everyone takes a little more damage.
- **Hunt the archers:** melee units charge the enemy's ranged and caster units first, faster than usual but taking a little more damage; ranged and caster units focus on enemies within reach.
- **Focus fire:** every unit advances at its allies' pace, attacks the enemy most allies are already fighting, and hits harder.
- **Skirmish:** melee units screen the back line without chasing far; ranged units kite, keep their distance, move faster and shoot further; casters keep their distance and reach further.

#### Scenario: Instinct changes nothing
- **WHEN** both sides use Instinct
- **THEN** the battle plays out exactly as it would without plans

## ADDED Requirements

### Requirement: Plan effects
A plan's stance for a role MAY grant effects to that role's units for the battle, as if learned, applied after the unit's own traits and skills. Every effect a plan grants SHALL be valid for every unit it reaches.

#### Scenario: Charge is faster
- **WHEN** a side uses All-out charge
- **THEN** its units move faster than the same units under Instinct

#### Scenario: Invalid plan effect
- **WHEN** a plan grants a ranged-only effect to melee units
- **THEN** the game refuses to start and the tests fail

### Requirement: No dominant plan
For every playable faction, with the same army on both sides, no plan SHALL win 55% or more of its battles against every other plan.

#### Scenario: Plan matrix
- **WHEN** the balance tool plays every plan against every other for each playable faction
- **THEN** no plan beats all the others
