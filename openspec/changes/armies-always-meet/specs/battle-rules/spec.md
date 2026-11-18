## MODIFIED Requirements

### Requirement: Temperament
Each unit type SHALL have a temperament that decides how it fights on its own: a target rule (nearest, wounded, ranged first, strongest, focus or protect; each falls back to the nearest enemy), an engage rule (charge, line or hold), and optional fields: a trigger distance for holding, a leash, a distance to keep, a health % below which it falls back behind its allies, and a kind of ally to stay near with no target. Fields left at zero SHALL change nothing. Abilities MAY change a unit's temperament. Temperaments SHALL never make two armies wait for each other forever: any two armies of any factions SHALL meet and fight.

#### Scenario: Hold
- **WHEN** a holding unit with a 4 m trigger sees an enemy 6 m away and no ally is fighting
- **THEN** it stays where it is, for at most 4 seconds

#### Scenario: Hold with nothing in sight
- **WHEN** a holding unit sees no enemy
- **THEN** it marches like a line

#### Scenario: Mirror match
- **WHEN** two armies of the same faction, both of holding or line units with supports behind them, start a battle
- **THEN** they advance, meet and fight before the time limit

#### Scenario: Wounded first
- **WHEN** a unit with the wounded rule has a healthy enemy and a badly wounded one within 4 m
- **THEN** it targets the wounded one

#### Scenario: Fall back
- **WHEN** a unit that falls back below 50% health drops to a third of its health with an enemy in sight and allies behind
- **THEN** it moves toward its allies instead of fighting
