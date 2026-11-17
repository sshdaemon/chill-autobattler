# Spec Delta

## ADDED Requirements

### Requirement: Graphics quality
The game SHALL offer a High and a Low graphics quality, remembered by the browser.
- **High** adds image-based lighting, soft shadows near the camera, ambient occlusion, bloom on magical effects, and colour grading.
- **Low** SHALL keep the scene readable with simpler lighting and no post-processing.

If High runs well below the display's own refresh rate (under 70% of it), the game SHALL switch to Low by itself and tell the player. Quality SHALL NOT affect the battle.

#### Scenario: Slow machine
- **WHEN** the frame rate stays under 70% of the display's refresh rate on High
- **THEN** the game switches to Low and shows a short note that it did

#### Scenario: Remembered choice
- **WHEN** the player picks Low and reloads
- **THEN** the game starts on Low
