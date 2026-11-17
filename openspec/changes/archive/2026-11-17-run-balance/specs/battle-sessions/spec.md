# Spec Delta

## MODIFIED Requirements

### Requirement: Start a battle
On request for an active run with no picks owed and no unfinished battle, the server SHALL create the run's next battle with a new random map seed and a new random combat seed. Red's skill count SHALL be the total skills of the run's veterans, and Red's health percentage SHALL follow the run's match number (see runs). The server SHALL store the battle and return the battle id, the map seed, Red's skill count, Red's health percentage and the deploy deadline. The combat seed SHALL NOT be returned at this point.

#### Scenario: New battle
- **WHEN** a client starts a battle for its active run whose veterans have 2 skills in total
- **THEN** it receives a battle id, a map seed, a Red skill count of 2, a Red health percentage and a deploy deadline, but no combat seed

#### Scenario: Unknown sim version
- **WHEN** a client starts a battle for a sim version the server doesn't have
- **THEN** the server refuses with a reason naming both versions

#### Scenario: Seeds differ between battles
- **WHEN** two battles are started
- **THEN** they have different ids, map seeds and combat seeds

#### Scenario: Run not ready
- **WHEN** a client starts a battle for a run that is over, owes skill picks, or has an unfinished battle
- **THEN** the server refuses
