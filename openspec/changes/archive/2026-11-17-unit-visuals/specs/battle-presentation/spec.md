# Spec Delta

## Purpose

How the browser shows a battle: unit models, animations, weapons and projectiles drawn from the sim's state. It is presentation only, and never affects what happens in the battle.

## ADDED Requirements

### Requirement: Units look like their class and team
Each unit SHALL be drawn with a character model for its class (warrior, archer or mage) in its team's colours. Units of the same class in one army SHALL use different looks where the class has several. Warriors SHALL carry a sword, archers a bow, and mages a staff.

#### Scenario: Two warriors
- **WHEN** Blue has two warriors
- **THEN** they are drawn with two different warrior models in Blue's colours, each holding a sword

#### Scenario: Red's units
- **WHEN** Red has an archer
- **THEN** it is drawn with an archer model in Red's colours, holding a bow

### Requirement: Animation follows the sim
Each unit's animation SHALL follow its state in the latest sim snapshot:
- walking or dashing → jog;
- attacking → sword attack (melee) or spell shot (ranged), stretched to the attack's duration;
- staggered → hit reaction;
- dead → death, held on the last frame;
- otherwise → idle.

When the battle ends, the winning side's survivors SHALL cheer.

#### Scenario: A warrior swings
- **WHEN** a warrior starts an attack in the sim
- **THEN** its sword attack animation plays once, ending as the attack ends

#### Scenario: Death
- **WHEN** a unit dies
- **THEN** it plays its death animation, stays on the ground, and later sinks out of view

### Requirement: Projectiles look like what they are
Arrows SHALL be drawn as arrows pointing along their flight path, higher for Arcing shot. Magic bolts SHALL be drawn as glowing orbs flying straight.

#### Scenario: Arrow in flight
- **WHEN** an archer shoots
- **THEN** an arrow flies from the archer toward its target in an arc, nose first

### Requirement: Unit status is readable
Each living unit SHALL show a team-coloured ring at its feet and a health bar above it that shrinks with its health. Veterans SHALL show their stars above the health bar. Units standing in water SHALL be drawn slightly sunk.

#### Scenario: Veteran
- **WHEN** a unit with 2 skills is on the field
- **THEN** two stars show above its health bar

### Requirement: Presentation never changes the battle
Rendering, animation and asset loading SHALL NOT affect the battle's outcome. Battles SHALL play out identically whether models have loaded or not.

#### Scenario: Slow asset load
- **WHEN** the character models are still loading as the fight starts
- **THEN** the battle proceeds on schedule, and the models appear once loaded
