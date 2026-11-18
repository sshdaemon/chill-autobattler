# match-flow Specification

## Purpose
The match as the player experiences it in the browser: choose an army, deploy it on a server-issued map, watch the fight, and see a verified result before the next match.

## Requirements

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

### Requirement: Faction screen
A new run SHALL open on a faction screen before the first army picker. Each playable faction SHALL show its name, colours, lore and its units (label, icon and blurb); choosing one SHALL start the run as that faction. Factions not playable yet MAY be shown with their planned units and SHALL NOT be selectable.

#### Scenario: Pick a faction
- **WHEN** a player with no active run opens the game and chooses Classic
- **THEN** a run starts as Classic and the army picker opens

#### Scenario: Coming soon
- **WHEN** the faction screen shows a faction that isn't playable yet
- **THEN** its card shows its planned units and its button is disabled

### Requirement: Draft
Before each battle of a run, after the battle is started, both sides SHALL draft five units in turns of 1, 2, 2, 2, 2 and 1 picks, alternating, the side picking first decided by a coin flip from the battle's map seed. Blue's veterans SHALL fill Blue's first picks by themselves. Every other Blue pick SHALL be a unit type of the run's faction with one of its abilities as its level-1 ability. Each Blue turn SHALL last at most 15 seconds; when it runs out, the turn's remaining picks SHALL be made at random. Red's picks SHALL be its generated army, revealed one unit at a time in a seeded order. The drafted army SHALL then go to deployment. Reloading during a draft SHALL draft the same battle again, with the same order and Red picks.

#### Scenario: Turn order
- **WHEN** the coin flip says Red picks first
- **THEN** Red picks 1, Blue 2, Red 2, Blue 2, Red 2 and Blue 1

#### Scenario: Veterans first
- **WHEN** the run has two veterans
- **THEN** they fill Blue's first two picks by themselves, and the player drafts the other three

#### Scenario: Time runs out
- **WHEN** the player makes no pick in a 15-second turn
- **THEN** that turn's picks are made at random and the draft goes on

#### Scenario: A pick
- **WHEN** the player picks a mage with Blast
- **THEN** Blue's army gains a level-1 mage whose ability is Blast
