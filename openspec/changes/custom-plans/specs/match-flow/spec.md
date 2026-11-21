# Spec Delta

## MODIFIED Requirements

### Requirement: Battle plan pick
During deployment the deployment bar SHALL offer every battle plan with a short description, with the current one marked; tactics SHALL be offered in their own row, and the player's custom plan, once made, SHALL be offered as "Custom". The chosen plan SHALL go into Blue's army when it is locked. The plan SHALL stay chosen for the next matches of the run; a new run SHALL start on Instinct.

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
- **THEN** Instinct is chosen and there is no custom plan

## ADDED Requirements

### Requirement: Plan editor
During deployment an Edit button next to the plan picker SHALL open a chip panel over the field: one row of chips per role, a "split into groups" switch that adds Left and Right tabs with a before and after stance and a trigger each, and a line saying in words what the chosen chips do. Invalid combinations SHALL be shown and SHALL NOT be saved. Saving SHALL choose the custom plan; it SHALL be remembered for the run.

#### Scenario: Invalid chips
- **WHEN** the player sets ranged units to hold with no trigger distance where a ranged unit type has none of its own
- **THEN** the panel shows what is wrong and won't save until it is fixed

#### Scenario: Remembered
- **WHEN** a player who saved a custom plan wins and goes to the next match
- **THEN** "Custom" is still chosen and the editor shows the same chips

### Requirement: Editor starts from the chosen plan
The editor SHALL open on the chosen plan's behaviour: its stances as the nearest chips, and for each role the menu edge that matches its effects, if any.

#### Scenario: Edit from a preset
- **WHEN** the player has Skirmish chosen and opens the editor
- **THEN** the ranged row shows kite on and keep distance far with the long-reach edge

### Requirement: Editing pauses the countdown
While the editor is open the deploy countdown SHALL pause; the server's deploy window still applies.

#### Scenario: Countdown pauses
- **WHEN** the player opens the editor with 9 seconds of deployment left
- **THEN** the countdown stays at 9 seconds until the panel is closed
