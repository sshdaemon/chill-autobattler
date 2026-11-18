## MODIFIED Requirements

### Requirement: Start and resume a run
A player SHALL have at most one active run. Starting a run SHALL name the faction it plays, which the run keeps to the end; an unknown faction SHALL be rejected. A new run SHALL get a random name made of an adjective, a noun and a number, starting at match 1 with no wins, no score and no veterans, and SHALL end any previous active run as lost. Reading the run SHALL return its current state and faction, so a client that reloads continues where it was.

#### Scenario: New run
- **WHEN** a player starts a run as the Classic faction
- **THEN** they get a run named like "Iron Fox 910" at match 1, with 0 wins, 0 points, no veterans and the Classic faction

#### Scenario: Restart abandons the old run
- **WHEN** a player with an active run starts another
- **THEN** the old run is over with its score kept, and the new one is active

#### Scenario: Reload mid-run
- **WHEN** a player reloads the page during match 3
- **THEN** reading the run returns match 3 with the same veterans, wins, score and faction

#### Scenario: Unknown faction
- **WHEN** a player starts a run with a faction that doesn't exist
- **THEN** the server rejects it, and any active run is unchanged
