# Spec Delta

## ADDED Requirements

### Requirement: Shields are strapped to the forearm
Shields SHALL be mounted along the forearm, face out from the back of the hand and far enough from the arm that the hand and forearm stay behind the shield in every clip, for every build.

#### Scenario: Shield idle
- **WHEN** a heavy warrior with a round shield plays its shield idle
- **THEN** the shield covers the body facing forward and no part of the fist shows through its face

### Requirement: Capes are simulated cloth
Capes SHALL hang from the shoulders as cloth that swings under gravity and the unit's motion, drapes over the back and doesn't pass through the torso or legs. They SHALL settle over the pose for still renders and reset when the unit teleports.

#### Scenario: Running
- **WHEN** a warrior with a cape jogs for two seconds and stops
- **THEN** the cape trails behind while jogging and swings back to hang over its back after stopping

#### Scenario: Teleport
- **WHEN** a unit wearing a cape is moved more than 0.8 m in one frame
- **THEN** the cape re-hangs from the new position instead of stretching across the gap

### Requirement: Units stand on the ground
Each assembled character SHALL be lifted so that its lowest point standing in its idle pose touches the ground, whatever its outfit, sex and build, because the clips' pelvis height belongs to the animation pack's shorter-legged mannequin.

#### Scenario: Standing and walking
- **WHEN** any unit stands, walks or attacks on dry ground
- **THEN** its soles rest on the ground within 2 cm instead of sinking into it
