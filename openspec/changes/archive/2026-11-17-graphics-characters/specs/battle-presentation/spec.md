# Spec Delta

## ADDED Requirements

### Requirement: Units read at a glance
Each unit SHALL show its team through a team-coloured rim light and team-coloured trims, in addition to its ring and health bar. The selected unit SHALL stand out with a gold rim. Units of the same class SHALL differ in hair, clothing tint and height, stable for a given army. None of this SHALL affect the battle.

#### Scenario: Team at a glance
- **WHEN** Blue and Red units fight side by side
- **THEN** each one's team shows on its outline and its trims, even without looking at rings or bars

#### Scenario: Same army, same looks
- **WHEN** the same army is shown again after a reload
- **THEN** each unit has the same hair colour, tint and height

### Requirement: Animation reads the action
Archers SHALL jog backwards while kiting. Archers and mages SHALL turn their upper body toward their target. A hit unit SHALL flinch away from its attacker.

#### Scenario: Kiting
- **WHEN** an archer backs away from a warrior while facing it
- **THEN** its legs jog backwards rather than forwards

#### Scenario: Flinch
- **WHEN** a visible unit is hit from its left
- **THEN** its upper body briefly tilts to the right
