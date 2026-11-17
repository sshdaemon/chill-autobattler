# Spec Delta

## ADDED Requirements

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
