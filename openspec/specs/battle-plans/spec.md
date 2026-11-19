# battle-plans Specification

## Purpose
Battle plans are army-wide stances a side picks before a battle. A plan changes
how the side's melee, ranged and caster units fight, so the same armies can
fight in different ways.

## Requirements

### Requirement: Built-in plans
The game SHALL have these built-in battle plans:
- **Instinct:** no change to any unit. It is the default.
- **Shield wall:** melee units hold and protect their ranged and caster allies; ranged and caster units focus on enemies within reach and stay behind the melee.
- **All-out charge:** every unit charges the nearest enemy, and ranged units don't kite.
- **Hunt the archers:** melee units charge the enemy's ranged and caster units first; ranged and caster units focus on enemies within reach.
- **Focus fire:** every unit advances at its allies' pace and attacks the enemy most allies are already fighting.
- **Skirmish:** melee units hold and fall back when badly wounded; ranged units kite and keep their distance.

#### Scenario: Instinct changes nothing
- **WHEN** both sides use Instinct
- **THEN** the battle plays out exactly as it would without plans

### Requirement: Plan stances
A plan SHALL set some temperament fields for each role (melee, ranged, caster). Fields a plan leaves unset SHALL keep the unit's own value. A plan SHALL only produce temperaments that are valid for the unit.

#### Scenario: No kiting melee
- **WHEN** a plan would make a melee unit kite, or make a unit hold without a trigger distance
- **THEN** the game refuses to start and the tests fail

#### Scenario: Charge overrides a holding unit
- **WHEN** a side with a unit whose own temperament holds uses All-out charge
- **THEN** that unit charges the nearest enemy at once

#### Scenario: Fields the plan leaves alone
- **WHEN** a unit that stays near its melee allies with no target is in a side whose plan doesn't set that field
- **THEN** it still stays near its melee allies

### Requirement: Order of changes
A unit's temperament SHALL be built from its roster entry, then changed by its side's plan, then by its learned abilities.

#### Scenario: A skill on top of a plan
- **WHEN** a unit with an ability that makes it target ranged units first is in a side using Focus fire
- **THEN** it targets ranged units first

### Requirement: Plans in the battle input
Each army in the battle input SHALL carry a plan; an army without one SHALL use Instinct. The input check SHALL reject an unknown plan. Replays SHALL carry each side's plan, so the server's re-run gives the same result. Plans SHALL be part of the rules: a change to the plans SHALL change the sim version.

#### Scenario: Unknown plan
- **WHEN** a battle input names a plan that doesn't exist
- **THEN** the input is rejected and the battle isn't played

#### Scenario: Replay with plans
- **WHEN** the server re-runs a battle where Blue used Skirmish
- **THEN** it gets the same result as the client

### Requirement: Red's plan
When Red's army is generated, Red's plan SHALL be drawn at random among all built-in plans, from the combat seed, so it is decided only once Blue's army is locked. The same combat seed SHALL always give the same plan. When the input gives Red's army, Red SHALL use the plan given with it.

#### Scenario: Hidden during the draft
- **WHEN** a player knows the battle's map seed but not its combat seed
- **THEN** they can't tell which plan Red will use

#### Scenario: Same seed, same plan
- **WHEN** two battles have the same combat seed and a generated Red
- **THEN** Red uses the same plan in both

### Requirement: Plans don't stall
Any two plans, on any two armies, SHALL meet and fight before the time limit.

#### Scenario: Shield wall against Shield wall
- **WHEN** both sides use Shield wall
- **THEN** the armies meet and fight before the time limit
