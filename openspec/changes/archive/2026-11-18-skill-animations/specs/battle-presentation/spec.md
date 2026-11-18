# Spec Delta

## MODIFIED Requirements

### Requirement: Animation follows the sim
Each unit's animation SHALL follow its state in the latest sim snapshot, using the animation set of its look (unit-looks):
- walking → move; dashing → the gap-closer's action animation if it has one, otherwise a faster move;
- attacking → one of its attack variants, or its skill's attack animation, timed so the blow lands with the sim's hit;
- staggered → hit reaction;
- dead → death, held on the last frame;
- otherwise → idle.

When the battle ends, the winning side's survivors SHALL cheer.

#### Scenario: A warrior swings
- **WHEN** a warrior starts an attack in the sim
- **THEN** one of its attack variants plays once, its blow landing as the sim's hit lands

#### Scenario: Death
- **WHEN** a unit dies
- **THEN** it plays its death animation, stays on the ground, and later sinks out of view

#### Scenario: Dash
- **WHEN** a warrior with Dash starts a dash
- **THEN** it plays Dash's leap-and-rush animation for the length of the dash

#### Scenario: Blink
- **WHEN** a mage blinks
- **THEN** it plays Blink's action animation where it lands

### Requirement: Combat effects
Every sim event SHALL have a visible cue for units the player can see: hits, flank hits, executes, deaths, dashes, Blink, Frost, Blast and shots. Each cue SHALL be a named effect preset (particles, rings, camera shake), chosen by the event and, for skill events, by the skill's preset in the unit's look. Running and wading units SHALL kick up dust or splashes. Effects SHALL NOT hide units or floating numbers, SHALL be limited at faster playback, and SHALL NOT affect the battle.

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

#### Scenario: Execute
- **WHEN** a warrior's Executioner kills a visible enemy
- **THEN** the effect preset named by Executioner plays (a heavy red burst and a short shake) on top of the hit sparks
