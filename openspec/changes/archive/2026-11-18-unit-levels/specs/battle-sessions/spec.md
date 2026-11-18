## MODIFIED Requirements

### Requirement: Start a battle
For an active run with no picks owed and no unfinished battle, the server SHALL create and store the run's next battle with new random map and combat seeds. Red's skill count SHALL be the veterans' abilities above level 1; Red's health % SHALL follow the match number (see runs). It SHALL return the battle id, map seed, Red's skill count and health %, and the deploy deadline, but NOT the combat seed.

#### Scenario: New battle
- **WHEN** a client starts a battle for its active run whose veterans have 2 abilities above level 1 in total
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
