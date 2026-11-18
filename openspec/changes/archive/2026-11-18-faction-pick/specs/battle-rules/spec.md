## ADDED Requirements

### Requirement: Factions
A faction SHALL list the unit types it fields, with at least one in Red's pool. A battle input SHALL name Blue's faction, and every Blue unit SHALL be of a type that faction fields; any other input SHALL be rejected. Red's faction SHALL be drawn from the map seed among all factions, so it may be the same as Blue's.

#### Scenario: Unit outside the faction
- **WHEN** a battle input's Blue army has a unit type its faction doesn't field
- **THEN** the input is rejected

#### Scenario: Red's faction
- **WHEN** two battles share a map seed
- **THEN** Red plays the same faction in both

## MODIFIED Requirements

### Requirement: Red army
When an input leaves Red's army empty, Red SHALL get 5 units of seeded random types from its faction's units in Red's pool, a seeded random formation preset, one seeded random ability each (level 1), and then as many more abilities in total as the given count, spread at random among its units up to level 3. All of this SHALL come from the seed.

#### Scenario: Matching skills
- **WHEN** a battle input asks for 3 Red skills
- **THEN** Red's units have exactly 8 abilities in total (one each, plus 3), with no unit holding the same ability twice or more than 3

#### Scenario: Faction's units only
- **WHEN** Red's faction fields only archers and mages
- **THEN** Red's generated army has no warriors
