# battle-presentation Specification

## Purpose
How the browser shows a battle: unit models, animations, weapons and projectiles drawn from the sim's state. It is presentation only, and never affects what happens in the battle.

## Requirements

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

### Requirement: Fog of war
During the fight, the arena SHALL be darkened wherever none of Blue's living units can see, using the battle's own sight rules (sight range, and rocks and trees blocking sight). Red units SHALL be drawn, with their health bars, only while Blue's team vision sees them. During deployment only Blue's zone SHALL be lit. A debug key SHALL toggle the fog.

#### Scenario: Enemy behind a rock
- **WHEN** a Red unit stands behind a rock that blocks every Blue unit's view of it
- **THEN** that Red unit and its health bar are not drawn, and its spot is darkened

#### Scenario: Spotted by a teammate
- **WHEN** any Blue unit sees a Red unit
- **THEN** the Red unit is drawn for the player

### Requirement: Portraits
Each unit SHALL have a portrait card: Blue's along the bottom-left, Red's along the bottom-right. Each card shows the unit's face, health bar and number, and veteran stars. A dead unit's card SHALL be greyed with ✝. A Red unit Blue can't currently see SHALL show "?" instead of its health.

#### Scenario: Unit dies
- **WHEN** a unit dies
- **THEN** its portrait turns grey and shows ✝

#### Scenario: Unseen enemy
- **WHEN** a Red unit is outside Blue's vision
- **THEN** its portrait shows "?" and no health number

### Requirement: Unit info panel
Selecting a unit (by clicking it or its portrait) SHALL show a panel with its name, stars, class, health, current status, kills, damage dealt, damage taken from each class, and skills. A Red unit's skills SHALL show only as a count. Information about an unseen Red unit SHALL NOT update while it is unseen.

#### Scenario: Selected archer retreats
- **WHEN** the player selects an archer that is backing away from a warrior
- **THEN** the panel's status reads "Falling back from melee"

### Requirement: RTS camera
The camera SHALL look down at a fixed tilt. It SHALL pan with WASD, the arrow keys or the pointer at the screen edge, and zoom with the mouse wheel within limits. F SHALL follow the selected unit until the player pans, Space SHALL jump to it, and double-clicking a portrait SHALL centre on its unit.

#### Scenario: Follow
- **WHEN** a unit is selected and the player presses F
- **THEN** the camera keeps that unit centred as it moves, until the player pans

### Requirement: Minimap
During deployment and the fight, a minimap SHALL show the whole arena from above:
- obstacles and water;
- Blue's living units;
- the Red units Blue's vision currently sees;
- a marker for the camera's view.

Clicking the minimap SHALL move the camera to that spot.

#### Scenario: Hidden enemy
- **WHEN** a Red unit is outside Blue's vision during the fight
- **THEN** the minimap doesn't show it

#### Scenario: Jump
- **WHEN** the player clicks the minimap near Red's edge
- **THEN** the camera moves to look at that spot

### Requirement: Sound effects
Battle events SHALL play short sound effects: hits (melee, arrow, bolt), deaths, skills (Blink, Dash, Frost), the "Fight!" start, and the result (victory or defeat). A mute toggle SHALL silence them, and the choice SHALL be remembered by the browser. Sounds SHALL never affect the battle.

#### Scenario: Muted
- **WHEN** the player has muted the game and reloads
- **THEN** battles play without sound until unmuted

### Requirement: Graphics quality
The game SHALL offer a High and a Low graphics quality, remembered by the browser.
- **High** adds environment lighting, soft shadows, bloom, a colour grade, anti-aliasing and grass.
- **Low** SHALL stay readable without grass or post-processing.

If High runs under 70% of the display's refresh rate, the game SHALL first render fewer pixels, then switch to Low and tell the player. Quality SHALL NOT affect the battle.

#### Scenario: Slow machine
- **WHEN** the frame rate stays under 70% of the display's refresh rate on High, even at reduced pixel density
- **THEN** the game switches to Low and shows a short note that it did

#### Scenario: Remembered choice
- **WHEN** the player picks Low and reloads
- **THEN** the game starts on Low

### Requirement: Battlefield matches the sim
Every rock, tree and pond SHALL be drawn where the sim places it, with a visible footprint matching the sim's circle. Pond circles that overlap SHALL read as one pond. A map SHALL look the same every time it is shown. The battlefield's look SHALL NOT affect the battle.

#### Scenario: Rock footprint
- **WHEN** the sim places a rock of radius r
- **THEN** the rock model is centred on it and its base spans about r, never visibly smaller than the blocked area

#### Scenario: Same map, same look
- **WHEN** the same map seed is shown twice
- **THEN** each obstacle has the same model variant, rotation and size both times

#### Scenario: Wading
- **WHEN** a unit is in water
- **THEN** ripples spread around it

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

### Requirement: Units read at a glance
Each unit SHALL show its team through a team-coloured rim light and team-coloured trims, in addition to its ring and health bar. The selected unit SHALL stand out with a gold rim. Units of the same class SHALL differ in hair, clothing tint and height, stable for a given army. None of this SHALL affect the battle.

#### Scenario: Team at a glance
- **WHEN** Blue and Red units fight side by side
- **THEN** each one's team shows on its outline and its trims, even without looking at rings or bars

#### Scenario: Same army, same looks
- **WHEN** the same army is shown again after a reload
- **THEN** each unit has the same hair colour, tint and height

### Requirement: Animation reads the action
Archers SHALL jog backwards while kiting. Archers and mages SHALL turn their upper body toward their target. A hit unit SHALL flinch away from its attacker.

#### Scenario: Kiting
- **WHEN** an archer backs away from a warrior while facing it
- **THEN** its legs jog backwards rather than forwards

#### Scenario: Flinch
- **WHEN** a visible unit is hit from its left
- **THEN** its upper body briefly tilts to the right
