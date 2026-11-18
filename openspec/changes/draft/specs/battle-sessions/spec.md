## MODIFIED Requirements

### Requirement: Lock a deployment once
The server SHALL accept Blue's army for a battle once, before its deploy deadline (default 3 minutes after start, room for the draft and deployment). It SHALL validate the army against the battle rules and the run's veterans, store it, and only then reveal the combat seed.

#### Scenario: Lock and reveal
- **WHEN** a client locks a valid army with the run's veterans before the deadline
- **THEN** the server stores it and returns the combat seed

#### Scenario: Second lock
- **WHEN** a client tries to lock a different army for a battle that is already locked
- **THEN** the server refuses, and the stored army is unchanged

#### Scenario: Too late
- **WHEN** a client locks an army after the deploy deadline
- **THEN** the server refuses because the battle has expired

#### Scenario: Invalid army
- **WHEN** a client locks an army with six units, a unit outside its deployment zone, or without the run's veterans
- **THEN** the server rejects it as invalid, and the battle stays unlocked
