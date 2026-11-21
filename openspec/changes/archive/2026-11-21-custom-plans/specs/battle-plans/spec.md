# Spec Delta

## ADDED Requirements

### Requirement: Custom plans
A side MAY fight by a custom plan instead of a built-in one: per role (melee, ranged, caster) a stance chosen from the chip vocabulary and at most one edge, optionally split into a left and a right group with an opening stance, one trigger and a second stance each. A custom plan SHALL change units exactly as a built-in plan with the same stances and effects would.

#### Scenario: Archers keep away
- **WHEN** Blue's custom plan gives ranged units "kite on" and "keep distance: far"
- **THEN** Blue's archers kite and keep 8 m from their target, and every other role fights by its own temperament

#### Scenario: Empty custom plan
- **WHEN** a custom plan leaves every chip on "own" with no edge and no groups
- **THEN** the battle plays out exactly as under Instinct

### Requirement: Chip vocabulary
A custom plan's stance SHALL be built only from chips, each with "own" (the unit's own value) as the default: target (nearest, wounded, ranged first, strongest, focus, focus within reach, protect), engage (charge, line, hold, flank), stay near (the melee, behind the front line; or ranged and casters, to guard them) and kite (ranged only: on, off).

#### Scenario: Unknown chip
- **WHEN** a custom plan carries a target rule or distance outside the vocabulary
- **THEN** the battle input is rejected as invalid

### Requirement: Distance chips
Distances in a custom plan SHALL be coarse steps: hold trigger close (6 m), mid (8 m) or far (10 m); keep distance close (5 m), mid (7 m) or far (8 m); leash short (10 m) or long (14 m); fall back below 50% or below 30% health.

#### Scenario: Far keep
- **WHEN** a custom plan sets ranged units' keep distance to far
- **THEN** they step back from targets closer than 8 m

### Requirement: Custom group triggers
A custom group's trigger SHALL be one of: after 5, 10, 15 or 20 seconds; the first ally falls; the other group is fighting.

#### Scenario: Unknown trigger
- **WHEN** a custom plan's group carries a trigger after 7 seconds
- **THEN** the battle input is rejected as invalid

### Requirement: Edge menu
Each role of a custom plan MAY take one edge from a fixed menu; every edge SHALL pair an advantage with a cost and SHALL only be offered to roles it suits. A group's second stance SHALL NOT take an edge. Edges are part of the rules: changing the menu changes the sim version.

#### Scenario: Ranged-only edge on melee
- **WHEN** a custom plan gives its melee units an edge meant for ranged units
- **THEN** the battle input is rejected as invalid

#### Scenario: Edge on a second stance
- **WHEN** a custom plan's group gives its second stance an edge
- **THEN** the battle input is rejected as invalid

### Requirement: Custom plans are checked like built-in ones
A custom plan SHALL only be accepted if every stance it gives is a valid temperament for every unit type with each of its skills (no kiting melee, no holding without a trigger distance) and its triggers are valid. The server SHALL check a locked army's custom plan the same way, and re-running the battle SHALL give the same result.

#### Scenario: Holding without a trigger
- **WHEN** a custom plan sets engage "hold" for ranged units and leaves the hold trigger on "own", and some ranged unit type has no trigger of its own
- **THEN** the battle input is rejected as invalid

#### Scenario: Fair verdict
- **WHEN** Blue locks a valid custom plan and plays the battle
- **THEN** the server's re-run gives the same outcome and verifies it as fair

### Requirement: No dominant custom plan
For every playable faction, with the same army on both sides, no custom plan SHALL win 55% or more of its battles against every built-in plan, and no edge on its own SHALL win more than 60% against the same army fighting by Instinct.

#### Scenario: Plan search
- **WHEN** the balance tool searches custom plans (random and hill-climbed) against the nine built-in plans for each playable faction
- **THEN** the best one found loses to or ties at least one built-in plan
