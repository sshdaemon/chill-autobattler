# Spec Delta

## ADDED Requirements

### Requirement: Battle plan pick
During deployment the deployment bar SHALL offer every battle plan with a short description, with the current one marked. The chosen plan SHALL go into Blue's army when it is locked. The plan SHALL stay chosen for the next matches of the run; a new run SHALL start on Instinct.

#### Scenario: Pick a plan
- **WHEN** the player chooses Shield wall and presses Ready
- **THEN** Blue's locked army fights by Shield wall, and the server verifies the battle as fair

#### Scenario: Next match
- **WHEN** a player who chose Skirmish wins and goes to the next match of the run
- **THEN** Skirmish is still chosen at deployment

#### Scenario: New run
- **WHEN** the player starts a new run
- **THEN** Instinct is chosen

### Requirement: Red's plan revealed at the fight
Red's plan SHALL NOT be shown before Blue's army is locked. When the fight starts, the "Fight!" banner SHALL name Red's plan, and during the fight both sides' plans SHALL be visible.

#### Scenario: During deployment
- **WHEN** the player is deploying
- **THEN** nothing on screen tells Red's plan

#### Scenario: Fight starts
- **WHEN** the deployment is locked and Red drew All-out charge
- **THEN** the banner reads "Fight!" with "Red: All-out charge", and the status line shows Blue's and Red's plans
