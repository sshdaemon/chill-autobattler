# Spec Delta

## ADDED Requirements

### Requirement: Combat effects
Every sim event SHALL have a visible cue for units the player can see: hits, flank hits, executes, deaths, dashes, Blink, Frost, Blast and shots. Running and wading units SHALL kick up dust or splashes. Effects SHALL NOT hide units or floating numbers, SHALL be limited at faster playback, and SHALL NOT affect the battle.

#### Scenario: Flank hit
- **WHEN** a unit lands a flank hit on a visible enemy
- **THEN** bigger gold sparks fly from the attacker's side, next to the "x2" number

#### Scenario: Hidden enemy
- **WHEN** a Red unit hidden by fog of war is hit by another hidden unit
- **THEN** no effect is shown there

#### Scenario: Death
- **WHEN** a unit dies
- **THEN** a dust puff rises and, after a moment, the body fades into the ground

#### Scenario: Last kill
- **WHEN** the battle's last unit falls
- **THEN** the camera shakes briefly
