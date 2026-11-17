# Spec Delta

## ADDED Requirements

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
