## MODIFIED Requirements

### Requirement: Veterans join every match
In a match of a run, Blue's army SHALL contain every veteran of the run, with exactly its class and abilities, plus recruits at level 1 with exactly one ability of their class each, up to 5 units. The server SHALL reject any other army.

#### Scenario: Missing veteran
- **WHEN** a player with a veteran archer locks an army without that archer
- **THEN** the server rejects the army

#### Scenario: Recruit with a skill
- **WHEN** a player locks an army in which a new recruit has two abilities (level 2)
- **THEN** the server rejects the army

#### Scenario: Recruit without an ability
- **WHEN** a player locks an army in which a new recruit has no ability
- **THEN** the server rejects the army

### Requirement: Red matches the veterans
For each match, the server SHALL give Red as many abilities above level 1 in total as Blue's veterans have above level 1, spread from the battle's map seed. A run with no veterans above level 1 SHALL face Red at level 1.

#### Scenario: Two-star army
- **WHEN** the run's veterans are one unit at level 3 and one at level 1
- **THEN** the battle's Red army has exactly 2 abilities above level 1 in total

### Requirement: Skill picks after a win
After a win, each veteran below level 3 (fewer than 3 abilities) SHALL owe one ability pick: an ability of its class that it doesn't have. The next battle SHALL NOT start until every owed pick is made.

#### Scenario: Picks owed
- **WHEN** a run won with two survivors, one of them at level 3
- **THEN** exactly one pick is owed, and starting the next battle is refused until it is made

#### Scenario: Invalid pick
- **WHEN** a player picks an ability the veteran already has, or one its class doesn't offer
- **THEN** the server rejects the pick, and nothing changes
