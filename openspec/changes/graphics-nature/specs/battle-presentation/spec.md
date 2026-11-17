# Spec Delta

## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: Graphics quality
The game SHALL offer a High and a Low graphics quality, remembered by the browser.
- **High** adds image-based lighting, soft shadows near the camera, bloom on magical effects, colour grading, anti-aliasing and swaying grass.
- **Low** SHALL keep the scene readable with simpler lighting, no grass and no post-processing.

If High runs well below the display's own refresh rate (under 70% of it), the game SHALL first render fewer pixels, and if that is still too slow, switch to Low by itself and tell the player. Quality SHALL NOT affect the battle.

#### Scenario: Slow machine
- **WHEN** the frame rate stays under 70% of the display's refresh rate on High, even at reduced pixel density
- **THEN** the game switches to Low and shows a short note that it did

#### Scenario: Remembered choice
- **WHEN** the player picks Low and reloads
- **THEN** the game starts on Low
