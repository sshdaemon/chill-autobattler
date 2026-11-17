# Spec Delta

## MODIFIED Requirements

### Requirement: Start a battle
On request for an active run with no picks owed and no unfinished battle, the server SHALL create the run's next battle with a new random map seed and a new random combat seed. Red's skill count SHALL be the total skills of the run's veterans. The server SHALL store the battle and return the battle id, the map seed, Red's skill count and the deploy deadline. The combat seed SHALL NOT be returned at this point.

#### Scenario: New battle
- **WHEN** a client starts a battle for its active run whose veterans have 2 skills in total
- **THEN** it receives a battle id, a map seed, a Red skill count of 2 and a deploy deadline, but no combat seed

#### Scenario: Unknown sim version
- **WHEN** a client starts a battle for a sim version the server doesn't have
- **THEN** the server refuses with a reason naming both versions

#### Scenario: Seeds differ between battles
- **WHEN** two battles are started
- **THEN** they have different ids, map seeds and combat seeds

#### Scenario: Run not ready
- **WHEN** a client starts a battle for a run that is over, owes skill picks, or has an unfinished battle
- **THEN** the server refuses

### Requirement: Lock a deployment once
The server SHALL accept Blue's army for a battle once, before its deploy deadline (default 2 minutes after start). It SHALL validate the army against the battle rules and the run's veterans, store it, and only then reveal the combat seed.

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
