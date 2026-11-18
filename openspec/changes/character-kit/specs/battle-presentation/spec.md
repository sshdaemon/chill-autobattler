# Spec Delta

## MODIFIED Requirements

### Requirement: Units look like their class and team
Each unit SHALL be drawn as a character assembled from one of its class's recipes (unit-looks), with its team's colour on its cloth. Units of the same class in one army SHALL use different recipes where the class has several. Warriors SHALL carry a sword, archers a bow, and mages a staff, as modelled weapons.

#### Scenario: Two warriors
- **WHEN** Blue has two warriors
- **THEN** they are drawn from two different warrior recipes with Blue's colour on their cloth, each holding a sword

#### Scenario: Red's units
- **WHEN** Red has an archer
- **THEN** it is drawn from an archer recipe with Red's colour on its cloth, holding a bow

### Requirement: Units read at a glance
Each unit SHALL show its team through a team-coloured rim light and team-coloured cloth, in addition to its ring and health bar. The selected unit SHALL stand out with a gold rim. Units of the same class SHALL differ in recipe (build, outfit, weapon), hair, clothing tint and height, stable for a given army. None of this SHALL affect the battle.

#### Scenario: Team at a glance
- **WHEN** Blue and Red units fight side by side
- **THEN** each one's team shows on its outline and its cloth, even without looking at rings or bars

#### Scenario: Same army, same looks
- **WHEN** the same army is shown again after a reload
- **THEN** each unit has the same recipe, hair colour, tint and height

#### Scenario: Classes by silhouette
- **WHEN** a warrior, an archer and a mage stand side by side
- **THEN** they can be told apart by silhouette alone: a broad warrior with a shield, a slim archer with a bow and quiver, a mage with a robe, hat and staff
