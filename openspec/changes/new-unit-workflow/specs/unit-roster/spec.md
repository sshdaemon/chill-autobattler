# Spec Delta

## ADDED Requirements

### Requirement: Adding a unit
A new roster entry outside Red's pool SHALL NOT change any existing battle outcome and SHALL need no sim version bump. Adding an entry to Red's pool, or changing an existing entry's numbers or effects, SHALL be a rule change with a version bump. Unit and skill ids SHALL be allocated as the next free number and never reused, and clients and the server SHALL accept them without a protocol change.

#### Scenario: New unit, old replays
- **WHEN** a new unit is added to the roster outside Red's pool
- **THEN** every recorded golden replay still reproduces under the same sim version

#### Scenario: New ids on the wire
- **WHEN** a client sends an army with the new unit's id and one of its new skill ids
- **THEN** the server accepts it if the roster knows them, and rejects ids the roster lacks

#### Scenario: Joining Red's pool
- **WHEN** a unit is added to Red's pool
- **THEN** the change bumps the sim version and re-records the golden replays
