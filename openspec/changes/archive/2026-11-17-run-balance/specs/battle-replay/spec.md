# Spec Delta

## MODIFIED Requirements

### Requirement: Battle input is complete
A battle input SHALL contain everything that affects the outcome: sim version, map seed, combat seed, Red's health percentage, and per team the units (class, skills) with deployment positions or a formation preset. The map seed SHALL decide the map and a generated Red army. The combat seed SHALL decide every random combat roll. A health percentage of 0 SHALL mean full health. Nothing outside the input (time, device, frame rate) SHALL affect the outcome.

#### Scenario: Frame rate does not matter
- **WHEN** the same battle input is played once at 30 fps and once at 144 fps in the browser
- **THEN** both runs produce the same outcome and checksum

#### Scenario: Red army comes from the seed
- **WHEN** a battle input lists no units for Red
- **THEN** Red's army, formation and skills are generated from the map seed, and the same map seed always gives the same Red army

#### Scenario: Combat seed changes only the fight
- **WHEN** the same input is played with two different combat seeds
- **THEN** the map and both armies' starting positions are identical, but the battles play out differently

#### Scenario: Reduced Red health
- **WHEN** a battle input sets Red's health to 85%
- **THEN** every Red unit starts with 85% of its maximum health, and the same input always gives the same outcome
