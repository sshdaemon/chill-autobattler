# Spec Delta

## ADDED Requirements

### Requirement: Neither side acts first by rule
Units act one after another within a tick and see what earlier units did. Which side's units act first SHALL be drawn each tick from the battle's combat seed, so neither side wins ties by rule.

#### Scenario: Same-tick charge
- **WHEN** two mirror-image units on a mirrored map dash at each other on the same tick, over many combat seeds
- **THEN** each side wins about half of those battles

### Requirement: Mirror-image movement
On a mirrored map, armies that are mirror images of each other SHALL march, path, turn and push apart as exact mirror images until the first contact (a dash, a swing or a blow). Rotations SHALL be tried in mirrored order for Red, a point on a grid cell edge SHALL count as in the cell nearer the centre line, and units SHALL push apart all at once rather than in unit order.

#### Scenario: Before contact
- **WHEN** both sides field the same army with the same plan and formation on a mirrored map, and start in step
- **THEN** every Red unit stands and faces exactly as its Blue counterpart's reflection until the first contact

### Requirement: Fair sides on every map
With the same army on both sides, Blue's win rate on any single map SHALL be within 3 standard errors of 50%.

#### Scenario: Map experiment
- **WHEN** the balance tool plays each playable faction's army against itself on its 10 fixed maps, 400 battles per cell
- **THEN** no cell is more than 3 standard errors from 50%
