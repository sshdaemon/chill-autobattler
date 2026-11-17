# Spec Delta

## MODIFIED Requirements

### Requirement: Army picker
Before each match the player SHALL pick exactly 5 units from the classes. Each class card shows its stats and how much damage it takes from each class. The run's veterans SHALL already fill their slots, shown with their stars, and the player SHALL NOT be able to go below the number of veterans in any class. The first pick SHALL default to 2 warriors, 2 archers and 1 mage, and later picks SHALL default to the previous army, raised to cover the veterans.

#### Scenario: Incomplete army
- **WHEN** the player has picked 4 units
- **THEN** they can't continue to deployment, and the button says how many are missing

#### Scenario: Remembered army
- **WHEN** the player played a match with 1 warrior, 3 archers and 1 mage, then presses Next match
- **THEN** the picker starts at 1 warrior, 3 archers and 1 mage

#### Scenario: Veterans are kept
- **WHEN** the run has two veteran warriors
- **THEN** the picker shows them in gold with their stars, and the warrior count can't go below 2

### Requirement: Verified result
When the battle ends, the client SHALL show Victory, Defeat or Draw with the surviving units, and SHALL verify the battle with the server automatically. After a verified win it SHALL show the run's new score and ask for each owed skill pick before offering Next match. After anything else it SHALL show Run over with the top-10 scoreboard, the player's runs marked, and a New run button.

#### Scenario: Win
- **WHEN** Red's last unit dies
- **THEN** a Victory banner shows, the verdict appears once the server answers, and each surviving unit offers its class's skills to pick from

#### Scenario: Next match after picks
- **WHEN** every survivor has picked a skill
- **THEN** Next match becomes available and leads to the army picker with those veterans

#### Scenario: Defeat
- **WHEN** Blue's last unit dies
- **THEN** a Defeat banner shows, followed by Run over with the score, the wins and the top-10 scoreboard

#### Scenario: Server unreachable at the end
- **WHEN** verification fails because the server can't be reached
- **THEN** the result still shows, with a "couldn't verify" note and a retry button

## ADDED Requirements

### Requirement: Resume on reload
On load, the client SHALL ask the server for the player's active run and continue it: at the army picker, at the skill picks after a win, or by replaying a locked battle that hasn't been verified yet. Without an active run, it SHALL start a new one.

#### Scenario: Reload during skill picks
- **WHEN** the player reloads after a win before picking skills
- **THEN** the skill picks show again for the same survivors

#### Scenario: Reload during a fight
- **WHEN** the player reloads while a locked battle is playing
- **THEN** the battle plays again from its start with the same armies and seeds, and is verified when it ends
