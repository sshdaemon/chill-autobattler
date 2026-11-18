## ADDED Requirements

### Requirement: Faction screen
A new run SHALL open on a faction screen before the first army picker. Each playable faction SHALL show its name, colours, lore and its units (label, icon and blurb); choosing one SHALL start the run as that faction. Factions not playable yet MAY be shown with their planned units and SHALL NOT be selectable.

#### Scenario: Pick a faction
- **WHEN** a player with no active run opens the game and chooses Classic
- **THEN** a run starts as Classic and the army picker opens

#### Scenario: Coming soon
- **WHEN** the faction screen shows a faction that isn't playable yet
- **THEN** its card shows its planned units and its button is disabled

## MODIFIED Requirements

### Requirement: Army picker
Before each match the player SHALL pick exactly 5 units from the unit types of the run's faction. Each card shows the type's label, icon and blurb (from its look), stats and damage taken from each type. Veterans SHALL fill their slots, shown with stars, and no type can go below its veterans. The first pick SHALL default to 2 warriors, 2 archers and 1 mage where the faction fields them; later picks SHALL default to the previous army, raised to cover the veterans; types the faction doesn't field SHALL start at 0.

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
- **WHEN** the run's faction fields a fourth unit type
- **THEN** the picker shows four cards, each with its own label, icon, blurb and damage-taken line, and armies can include the new type
