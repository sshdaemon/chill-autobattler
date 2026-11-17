# runs Specification

## Purpose
Runs chain matches together: survivors become veterans who learn skills, Red grows to match them, and the run ends at the first loss with a score. The server decides all of it, so a client can't invent veterans, weaken Red or hide a loss.

## Requirements

### Requirement: Anonymous player
The server SHALL issue a random secret token to identify a player, and SHALL store only a hash of it. Every run request SHALL carry the token, and requests without a known token SHALL be refused.

#### Scenario: First visit
- **WHEN** a browser without a token asks for one
- **THEN** it receives a new token, and later requests carrying it are recognised as the same player

#### Scenario: Missing or unknown token
- **WHEN** a request for a run arrives without a token or with an unknown one
- **THEN** the server refuses it as unauthenticated

### Requirement: Start and resume a run
A player SHALL have at most one active run. Starting a run SHALL give it a random name made of an adjective, a noun and a number, starting at match 1 with no wins, no score and no veterans, and SHALL end any previous active run as lost. Reading the run SHALL return its current state, so a client that reloads continues where it was.

#### Scenario: New run
- **WHEN** a player starts a run
- **THEN** they get a run named like "Iron Fox 910" at match 1, with 0 wins, 0 points and no veterans

#### Scenario: Restart abandons the old run
- **WHEN** a player with an active run starts another
- **THEN** the old run is over with its score kept, and the new one is active

#### Scenario: Reload mid-run
- **WHEN** a player reloads the page during match 3
- **THEN** reading the run returns match 3 with the same veterans, wins and score

### Requirement: Veterans join every match
In a match of a run, Blue's army SHALL contain every veteran of the run, with exactly its class and skills, plus recruits without skills, up to 5 units. The server SHALL reject any other army.

#### Scenario: Missing veteran
- **WHEN** a player with a veteran archer locks an army without that archer
- **THEN** the server rejects the army

#### Scenario: Recruit with a skill
- **WHEN** a player locks an army in which a new recruit has a skill
- **THEN** the server rejects the army

### Requirement: Red matches the veterans
For each match, the server SHALL give Red as many skills in total as Blue's veterans have, spread from the battle's map seed. A run with no veterans SHALL face Red with no skills.

#### Scenario: Two-star army
- **WHEN** the run's veterans have 2 skills in total
- **THEN** the battle's Red army has exactly 2 skills in total

### Requirement: Winning a match
When a battle of the run is verified fair and Blue won, the run SHALL gain 100 points plus 10 per surviving Blue unit and one win, and SHALL move to the next match. Every surviving Blue unit SHALL become a veteran, and veterans that died SHALL be gone.

#### Scenario: Win with three survivors
- **WHEN** Blue wins with 3 units alive
- **THEN** the run gains 130 points and 1 win, and its veterans are exactly those 3 units

### Requirement: Skill picks after a win
After a win, each veteran with fewer than 5 skills SHALL owe one skill pick: a skill of its class that it doesn't have. The next battle SHALL NOT start until every owed pick is made.

#### Scenario: Picks owed
- **WHEN** a run won with two survivors, one of them with 5 skills
- **THEN** exactly one pick is owed, and starting the next battle is refused until it is made

#### Scenario: Invalid pick
- **WHEN** a player picks a skill the veteran already has, or one from another class
- **THEN** the server rejects the pick, and nothing changes

### Requirement: Losing ends the run
A battle that isn't a fair win SHALL end the run: a loss, a draw, or a claim that isn't fair. A finished run SHALL keep its score and SHALL NOT start more battles.

#### Scenario: Defeat
- **WHEN** a battle of the run is verified and Red won
- **THEN** the run is over with its score, and starting another battle in it is refused

### Requirement: Abandoned battles are losses
If a run's current battle passes its deploy deadline without being locked, or its verify deadline without a verdict, the run SHALL end as lost.

#### Scenario: Tab closed mid-fight
- **WHEN** a player locks a deployment, closes the tab, and returns after the verify deadline
- **THEN** reading the run shows it as over, lost

### Requirement: Scoreboard
The server SHALL list the 10 highest-scoring finished runs, best first (ties go to the earlier finish), with name, score and wins, and SHALL mark which of them belong to the requesting player.

#### Scenario: Run over
- **WHEN** a player's run ends with the 4th best score of all time
- **THEN** the scoreboard shows it 4th and marks it as theirs

### Requirement: Early matches are easier
Red's units SHALL start with 85% of their health in match 1 of a run, 90% in match 2, 95% in match 3, and full health from match 4 on. The server SHALL set this for each battle it starts, and SHALL verify the battle with the value it set.

#### Scenario: First match
- **WHEN** a run starts its first battle
- **THEN** every Red unit in that battle starts at 85% of its normal maximum health

#### Scenario: Later matches
- **WHEN** a run starts its fourth or later battle
- **THEN** Red's units start at full health
