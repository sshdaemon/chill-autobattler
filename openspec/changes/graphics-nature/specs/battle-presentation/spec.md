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
