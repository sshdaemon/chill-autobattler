# Spec Delta

## ADDED Requirements

### Requirement: Units can be inspected in a lab
In development builds, a unit lab SHALL show any unit type's recipes on either team, assembled as in battle, with a camera that orbits all the way round. It SHALL play, pause, loop, slow down and scrub any clip; preview the walk–jog blend at a chosen ground speed on a moving floor, with its foot sliding shown; fire each skill's animation and effect and any effect preset; and switch build and variety. It SHALL NOT be part of production builds.

#### Scenario: Look all around
- **WHEN** a developer opens the lab for the warrior and drags across the view
- **THEN** the camera orbits the warrior through 360°, and the cape, shield and helm are visible from every side

#### Scenario: Scrub an attack
- **WHEN** they pick Sword_Slash_A, pause it and drag the timeline
- **THEN** the pose follows the timeline frame by frame

#### Scenario: Walking speed
- **WHEN** they set the ground speed to the archer's speed while frozen (2.2 m/s)
- **THEN** the archer walks on the moving floor with the battle's walk–jog blend, and the readout shows the blend and a foot slide under 10%

#### Scenario: Not shipped
- **WHEN** the production bundle is built
- **THEN** it contains no unit lab code
