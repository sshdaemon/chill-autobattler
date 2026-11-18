## ADDED Requirements

### Requirement: Factions not yet playable
A faction SHALL say whether it is playable. Only playable factions SHALL be offered to players when a run starts and drawn as Red's faction; at least one faction SHALL be playable. A faction whose units are still placeholders SHALL NOT be playable.

#### Scenario: Faction being built
- **WHEN** a player starts a run with a faction that isn't playable
- **THEN** the server rejects it

#### Scenario: Red never draws it
- **WHEN** battles are started with any map seed
- **THEN** Red's faction is always a playable one

### Requirement: Planned units exist as placeholders
The units of the roster plan SHALL exist in the roster with their allocated ids, numbers, abilities and temperaments, and with placeholder looks, before unit work starts, so that unit work never allocates ids.

#### Scenario: Placeholder look
- **WHEN** the unit lab shows a planned unit before its unit work
- **THEN** it appears with the look borrowed from the classic unit of its role
