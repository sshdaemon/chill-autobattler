# Spec Delta

## MODIFIED Requirements

### Requirement: Battle plan pick
During deployment the deployment bar SHALL offer every battle plan with a short description, with the current one marked; tactics SHALL be offered in their own row. The chosen plan SHALL go into Blue's army when it is locked. The plan SHALL stay chosen for the next matches of the run; a new run SHALL start on Instinct.

#### Scenario: Pick a plan
- **WHEN** the player chooses Shield wall and presses Ready
- **THEN** Blue's locked army fights by Shield wall, and the server verifies the battle as fair

#### Scenario: Pick a tactic
- **WHEN** the player chooses Hammer and anvil and presses Ready
- **THEN** Blue's locked army fights by Hammer and anvil, and the server verifies the battle as fair

#### Scenario: Next match
- **WHEN** a player who chose Skirmish wins and goes to the next match of the run
- **THEN** Skirmish is still chosen at deployment

#### Scenario: New run
- **WHEN** the player starts a new run
- **THEN** Instinct is chosen

## ADDED Requirements

### Requirement: Groups shown at deployment
While a tactic is chosen, the deployment view SHALL mark which group each of Blue's units is in, as the battle will split them, and SHALL update the marks when a unit is dragged or the formation changes.

#### Scenario: Drag across
- **WHEN** a tactic is chosen and the player drags a unit from the left half of the army to the right
- **THEN** its mark changes from the left group to the right group

#### Scenario: No tactic
- **WHEN** the chosen plan is not a tactic
- **THEN** no group marks are shown
