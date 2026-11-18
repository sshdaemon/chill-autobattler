# match-flow Specification

## Purpose
The match as the player experiences it in the browser: choose an army, deploy it on a server-issued map, watch the fight, and see a verified result before the next match.

## Requirements

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

### Requirement: Deployment under fog
After the picker, the client SHALL start a server-issued battle and show its map with Blue deployed in its zone. Red's units SHALL NOT be shown during deployment.

#### Scenario: Deploying
- **WHEN** the player continues from the picker
- **THEN** the map appears with Blue's units in Blue's zone in the default formation, and no Red units are visible

### Requirement: Formation presets
The deployment bar SHALL offer the presets Fully spread, 3-2, 2-3, 4-1, 1-4, 0-5 and 5-0. Each preset SHALL show a mini-diagram of where the player's own units would stand, marking warriors, archers and mages differently. Choosing a preset SHALL move Blue's units there.

#### Scenario: Pick a preset
- **WHEN** the player chooses 5-0
- **THEN** all five Blue units move to the left flank, and the 5-0 button is marked as selected

### Requirement: Drag to place
During deployment the player SHALL be able to drag any of their own units to a new spot inside Blue's zone. Doing so SHALL switch the formation to Custom, and the unit stands on the nearest walkable spot.

#### Scenario: Drag a unit
- **WHEN** the player drags an archer to another spot in the zone
- **THEN** the archer moves there, the other units stay put, and Custom is marked as selected

#### Scenario: Drag outside the zone
- **WHEN** the player drags a unit beyond the zone's edge
- **THEN** the unit stops at the edge of the zone

### Requirement: Deployment timer
Deployment SHALL last at most 15 seconds, with a visible countdown. **Ready** SHALL end it early. When it ends, the client SHALL lock the army with the server and only then start the fight.

#### Scenario: Time runs out
- **WHEN** the player doesn't press Ready within 15 seconds
- **THEN** the current deployment is locked and the fight starts

### Requirement: Countdown and fight
After locking, the client SHALL show a short "Fight!" banner, then play the battle with the revealed combat seed. Red SHALL become visible, and the speed controls SHALL work during the fight.

#### Scenario: Fight starts
- **WHEN** the deployment is locked
- **THEN** a "Fight!" banner shows for about a second, then both armies move

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

### Requirement: Resume on reload
On load, the client SHALL ask the server for the player's active run and continue it: at the army picker, at the skill picks after a win, or by replaying a locked battle that hasn't been verified yet. Without an active run, it SHALL start a new one.

#### Scenario: Reload during skill picks
- **WHEN** the player reloads after a win before picking skills
- **THEN** the skill picks show again for the same survivors

#### Scenario: Reload during a fight
- **WHEN** the player reloads while a locked battle is playing
- **THEN** the battle plays again from its start with the same armies and seeds, and is verified when it ends
