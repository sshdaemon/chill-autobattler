# Spec Delta

## MODIFIED Requirements

### Requirement: Map is generated from the seed
The obstacles SHALL be generated only from the battle seed, and the map SHALL be mirrored across the centre line: every obstacle's reflection (z → −z) is on the map too. Each feature is made in Red's half and reflected; a rock or tree near the centre line sits on it, once, and a pond that reaches its reflection is one pond:
- 1–2 water ponds of 2–4 round blobs each, with their reflections, in the middle band.
- 2–3 rocks with radius 0.9–1.9 m.
- 2–4 trees with radius 0.8 m.

Obstacles SHALL keep 1.6 m clearance from each other and 2 m from the border, and SHALL stay out of the middle of each deployment zone (|x| < 6 m, |z| > 7.5 m). Ponds SHALL stay in the middle band (|z| + r < 6.5 m).

#### Scenario: Same seed, same map
- **WHEN** a map is generated twice from the same seed, natively or in WASM
- **THEN** both maps have identical obstacles (positions and radii, exactly)

#### Scenario: Different seeds
- **WHEN** maps are generated from seeds 1 and 2
- **THEN** the obstacle layouts differ

#### Scenario: There is always a way through
- **WHEN** a map is generated for any of 1,000 test seeds
- **THEN** every free cell of Blue's zone can reach every free cell of Red's zone by walking

#### Scenario: Mirrored
- **WHEN** a map is generated for any seed
- **THEN** for every obstacle at (x, z) there is one of the same kind and size at (x, −z)

## ADDED Requirements

### Requirement: Same choices for both sides
On a mirrored map, wherever the grid chooses between equally good options for a unit (a path round either side of a rock, the nearest free spot), Red's choice SHALL be the mirror image of the choice Blue would make in the mirrored position.

#### Scenario: Round a rock
- **WHEN** a Blue unit and a Red unit start at mirror-image positions and head for mirror-image goals
- **THEN** their paths are mirror images of each other

### Requirement: Maps don't pick a side
With the same army on both sides, no side SHALL gain a large edge from the map alone.

#### Scenario: Map experiment
- **WHEN** the balance tool plays each playable faction's army against itself on 10 fixed maps
- **THEN** Blue's win rate averages within 6 points of 50% across all cells
