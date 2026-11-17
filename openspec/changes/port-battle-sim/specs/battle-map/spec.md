# Spec Delta

## Purpose

Generates the battlefield for a battle from its seed: arena, obstacles, water and deployment zones, plus how each kind of terrain affects walking, sight and shots.

## ADDED Requirements

### Requirement: Arena and deployment zones
The arena SHALL be 28 × 28 m, centred on the origin, with a border. Blue's deployment zone SHALL be x ∈ [-10, 10], z ∈ [7, 13], and Red's SHALL be x ∈ [-10, 10], z ∈ [-13, -7].

#### Scenario: Units stay inside the arena
- **WHEN** any unit moves, is pushed, dashes or blinks
- **THEN** its position stays within the arena minus the unit radius and border margin

### Requirement: Map is generated from the seed
The obstacles SHALL be generated only from the battle seed:
- 2–3 water ponds of 2–4 round blobs each, in the middle band.
- 3–6 rocks with radius 0.9–1.9 m.
- 3–7 trees with radius 0.8 m.

Obstacles SHALL keep 1.6 m clearance from each other and 2 m from the border, and SHALL stay out of the deployment zones.

#### Scenario: Same seed, same map
- **WHEN** a map is generated twice from the same seed, natively or in WASM
- **THEN** both maps have identical obstacles (positions and radii, exactly)

#### Scenario: Different seeds
- **WHEN** maps are generated from seeds 1 and 2
- **THEN** the obstacle layouts differ

#### Scenario: There is always a way through
- **WHEN** a map is generated for any of 1,000 test seeds
- **THEN** every free cell of Blue's zone can reach every free cell of Red's zone by walking

### Requirement: Terrain rules
Rocks and trees SHALL block walking, sight and straight shots. Water SHALL NOT block sight or shots, but SHALL slow units standing in it to 70% speed. Walkability SHALL be decided on a 0.5 m grid, where a cell is blocked if it lies within unit radius + 0.1 m of a rock or tree.

#### Scenario: Seeing across water
- **WHEN** an archer and an enemy have only water between them, within sight range
- **THEN** the archer can see and shoot the enemy

#### Scenario: Rock blocks sight
- **WHEN** a rock lies on the straight line between a unit and an enemy within its sight range
- **THEN** the unit cannot see that enemy by itself
