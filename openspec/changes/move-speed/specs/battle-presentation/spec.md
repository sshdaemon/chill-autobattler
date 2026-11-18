# Spec Delta

## MODIFIED Requirements

### Requirement: Animation follows the sim
Each unit's animation SHALL follow its state in the latest sim snapshot, using the animation set of its look (unit-looks):
- walking → its walk and move clips, blended and played so the stride covers the unit's ground speed; dashing → the gap-closer's action animation if it has one, otherwise its sprint clip at the dash's speed;
- attacking → one of its attack variants, or its skill's attack animation, timed so the blow lands with the sim's hit;
- staggered → hit reaction;
- dead → death, held on the last frame;
- otherwise → idle.

When the battle ends, the winning side's survivors SHALL cheer.

#### Scenario: A warrior swings
- **WHEN** a warrior starts an attack in the sim
- **THEN** one of its attack variants plays once, its blow landing as the sim's hit lands

#### Scenario: Death
- **WHEN** a unit dies
- **THEN** it plays its death animation, stays on the ground, and later sinks out of view

#### Scenario: Dash
- **WHEN** a warrior with Dash starts a dash
- **THEN** it plays Dash's leap-and-rush animation for the length of the dash

#### Scenario: Blink
- **WHEN** a mage blinks
- **THEN** it plays Blink's action animation where it lands

#### Scenario: Feet don't slide
- **WHEN** a unit walks at any speed the sim gives it (3.0–5.0 m/s on dry ground, slower in water, frost or a sharp turn)
- **THEN** its planted foot moves backwards at the unit's ground speed within 10%, so the foot stays put on the ground

#### Scenario: Slowing down
- **WHEN** a frozen archer moves at 60% of its speed
- **THEN** its legs shift towards a walk and slow down with it, rather than jogging in slow motion
