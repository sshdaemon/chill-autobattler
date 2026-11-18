# unit-roster Specification

## Purpose
Defines the roster of unit types the battle can field: each unit's stats, traits, matchup tags and learnable skills, with skills built from a fixed set of effects, so a new unit can be added as data rather than as new battle logic.

## Requirements

### Requirement: Units come from the roster
Every unit in a battle SHALL be one entry of the roster. An entry has a stable numeric id, a name, base stats (health, damage range, attack cooldown, swing length, range, speed, sight, projectile speed for ranged units), traits (melee or ranged, kites), a projectile kind for ranged units, matchup tags and exactly five learnable skills.

#### Scenario: The first three entries
- **WHEN** the roster is listed
- **THEN** it starts with warrior (id 0), archer (id 1) and mage (id 2), with the stats and skills given in docs/GAME_CONCEPT.md §4.1 and §10.1

#### Scenario: Unknown unit in an input
- **WHEN** a battle input names a unit id that is not in the roster
- **THEN** the input is rejected as invalid and no battle is played

### Requirement: Skills are made of effects
Each skill SHALL be a name plus one or more effects from a fixed set, each with its own parameters: stat bonus (health, damage, range, speed, attack speed), finisher, slow on hit, splash on hit, gap-closer, escape teleport and shooting over cover. A unit's behaviour SHALL follow only from its roster entry and the effects of the skills it has.

#### Scenario: Same effect, different units
- **WHEN** two roster entries both have a skill with a finisher effect, with chances of 35% and 25%
- **THEN** each unit executes with its own chance, through the same rule

#### Scenario: A new unit needs no new rules
- **WHEN** a test adds a roster entry whose skills reuse existing effects with new parameters
- **THEN** battles with that unit run and follow those effects, with no change to the battle rules

### Requirement: Effects apply in a fixed order
When several effects trigger on the same event, they SHALL be applied in a fixed order by effect kind, not by the order skills were learned or listed. Within one hit, the order SHALL be damage, finisher, slow, then splash.

#### Scenario: Frost and Blast on one bolt
- **WHEN** a mage with both Frost and Blast lands a bolt
- **THEN** the victim takes damage, then rolls for the slow, then nearby allies of the victim take splash damage, whichever skill the mage learned first

### Requirement: Matchups come from tags
Each roster entry SHALL carry matchup tags (for example `melee`, `ranged`, `caster`) and a list of damage multipliers against tags. A hit's matchup multiplier SHALL be the attacker's multiplier for the first of the victim's tags that the attacker lists, or ×1 when it lists none of them.

#### Scenario: Warrior against archer
- **WHEN** a warrior hits an archer, which is tagged `ranged`
- **THEN** the matchup multiplier is ×1.45

#### Scenario: No listed tag
- **WHEN** an archer hits any unit
- **THEN** the matchup multiplier is ×1

### Requirement: The roster is checked
The build SHALL fail its tests if any roster entry is invalid: a duplicate unit or skill id, a skill count other than five, a ranged unit without projectile speed or kind, a tag that is not declared, a stat at or below zero, or an effect parameter outside its allowed range.

#### Scenario: Ranged unit without a projectile speed
- **WHEN** a roster entry is ranged but has no projectile speed
- **THEN** the roster check reports that entry and the test suite fails

### Requirement: The roster is published to the client
The browser build SHALL return the whole roster from its data call: units with stats, traits, tags and matchup multipliers, and skills with names, effects and parameters. The client SHALL take unit and skill data only from there.

#### Scenario: Picker reads the roster
- **WHEN** the web client asks the browser build for its data
- **THEN** it receives every roster entry with its five skills and their effects, in roster order

### Requirement: Moving to the roster keeps every outcome
Expressing the existing units and skills as roster data SHALL NOT change any battle outcome. Every recorded golden replay SHALL reproduce bit for bit, natively and under WASM, under the unchanged sim version.

#### Scenario: Golden replays after the move
- **WHEN** the golden replay tests run after the move, natively and under WASM
- **THEN** every recorded outcome matches exactly and the sim version is still 0.4.0
