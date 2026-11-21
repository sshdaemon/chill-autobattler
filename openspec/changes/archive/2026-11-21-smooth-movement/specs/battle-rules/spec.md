# Spec Delta

## ADDED Requirements

### Requirement: Steady movement
A unit SHALL NOT flip between two moves on alternate ticks at the edge of a distance it reacts to: once it turns back to its ward, falls back, or steps away to keep its distance, it SHALL carry that move on past the edge before changing its mind. A unit's path SHALL be found again when its goal moves more than 1 m. A goal outside the arena or inside a rock or tree SHALL be moved to the nearest spot a unit can stand on.

#### Scenario: Bodyguard at the edge of its reach
- **WHEN** a bodyguard gets just over 4 m from its ward and turns back
- **THEN** it heads for its ward until it is within 2 m, then advances again

#### Scenario: Keeping distance
- **WHEN** a unit that keeps 5 m from its target starts stepping back during its attack's cooldown
- **THEN** it steps back until it is 5 m away, not stopping at 4.5 m

#### Scenario: Fall-back spot past the edge
- **WHEN** a wounded unit's fall-back spot lies outside the arena
- **THEN** it falls back to the nearest spot inside the arena and stops there

#### Scenario: Few shakes and stalls
- **WHEN** 600 varied battles are played
- **THEN** fewer than 2% of walking ticks reverse the previous step, and units spend at most 1,500 unit-ticks in all wanting to move for 1.5 s while getting less than half a metre

### Requirement: Bodyguards and followers
A bodyguard SHALL NOT turn back for a ward that itself stays behind the front line, since that ward follows the melee anyway.

#### Scenario: Shield wall with bodyguards
- **WHEN** a Rus army fights by Shield wall, so its archers stay behind the melee while its druzhynnyk guard the archers
- **THEN** the druzhynnyk and the archers advance together instead of pacing on the spot

### Requirement: Wounded units wait out of sight
A unit that falls back when wounded SHALL, once its target is out of its sight, wait where it is, facing the target, instead of walking back toward it, as long as it has living allies.

#### Scenario: Waits
- **WHEN** a huslyar below half health has fallen back out of sight of its target
- **THEN** it stays there facing the target until the target comes into sight again or its allies are gone
