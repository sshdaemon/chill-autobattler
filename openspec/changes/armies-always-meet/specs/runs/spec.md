## ADDED Requirements

### Requirement: Surrender
A player SHALL be able to give up their active run. It SHALL end as surrendered with its score kept on the scoreboard, and the client SHALL return to the faction screen. A run that is already over SHALL NOT be surrendered.

#### Scenario: Surrender mid-run
- **WHEN** a player surrenders their active run in match 3
- **THEN** the run is over (surrendered) with its score kept, and reloading shows the faction screen

#### Scenario: Already over
- **WHEN** a player surrenders a run that is over
- **THEN** the server refuses
