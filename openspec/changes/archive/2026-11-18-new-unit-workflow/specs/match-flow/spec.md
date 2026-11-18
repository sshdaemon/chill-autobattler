# Spec Delta

## MODIFIED Requirements

### Requirement: Army picker
Before each match the player SHALL pick exactly 5 units from every unit type in the roster. Each unit card shows its label, icon and blurb (from its look), its stats and how much damage it takes from each unit type. The run's veterans SHALL already fill their slots, shown with their stars, and the player SHALL NOT be able to go below the number of veterans of any type. The first pick SHALL default to 2 warriors, 2 archers and 1 mage, and later picks SHALL default to the previous army, raised to cover the veterans.

#### Scenario: Incomplete army
- **WHEN** the player has picked 4 units
- **THEN** they can't continue to deployment, and the button says how many are missing

#### Scenario: Remembered army
- **WHEN** the player played a match with 1 warrior, 3 archers and 1 mage, then presses Next match
- **THEN** the picker starts at 1 warrior, 3 archers and 1 mage

#### Scenario: Veterans are kept
- **WHEN** the run has two veteran warriors
- **THEN** the picker shows them in gold with their stars, and the warrior count can't go below 2

#### Scenario: A fourth unit type
- **WHEN** the roster has a fourth unit type
- **THEN** the picker shows four cards, each with its own label, icon, blurb and damage-taken line, and armies can include the new type
