## MODIFIED Requirements

### Requirement: Units come from the roster
Every unit in a battle SHALL be one entry of the roster. An entry has a stable numeric id, a name, base stats (health, damage range, attack cooldown, swing length, range, speed, sight, projectile speed for ranged units), traits (melee or ranged, kites), a projectile kind for ranged units, matchup tags and exactly five learnable skills. A retired unit's id SHALL stay reserved, with no entry, and SHALL never be given to another unit.

#### Scenario: The first three entries
- **WHEN** the roster is listed
- **THEN** ids 0, 1 and 2 (the retired warrior, archer and mage) have no entry, and the Rus, Varangian and Roman units follow at ids 3 to 11

#### Scenario: Unknown unit in an input
- **WHEN** a battle input names a unit id that is not in the roster
- **THEN** the input is rejected as invalid and no battle is played

#### Scenario: Retired unit in an input
- **WHEN** a battle input names unit id 0, the retired warrior
- **THEN** the input is rejected as invalid, as for any unknown unit

### Requirement: Planned units exist as placeholders
The units of the roster plan SHALL exist in the roster with their allocated ids, numbers, abilities and temperaments, and with placeholder looks, before unit work starts, so that unit work never allocates ids.

#### Scenario: Placeholder look
- **WHEN** the unit lab shows a unit that has no look of its own yet
- **THEN** it appears with a stock look for its role (melee, ranged or caster), and the dev build warns once

## REMOVED Requirements

### Requirement: Effects apply in a fixed order
**Reason**: Its scenario was about the retired mage; restated without it as "Hit effects apply in a fixed order".
**Migration**: None; the rule is unchanged.

### Requirement: Matchups come from tags
**Reason**: Its scenarios were about the retired warrior and archer; restated with Rus units as "Matchup multipliers come from tags".
**Migration**: None; the rule is unchanged.

## ADDED Requirements

### Requirement: Hit effects apply in a fixed order
When several effects trigger on the same event, they SHALL be applied in a fixed order by effect kind, not by the order skills were learned or listed. Within one hit, the order SHALL be damage, finisher, slow, then splash.

#### Scenario: Slow and splash on one hit
- **WHEN** a unit whose skills give it both slow on hit and splash on hit lands a hit
- **THEN** the victim takes damage, then rolls for the slow, then nearby allies of the victim take splash damage, whichever skill it learned first

### Requirement: Matchup multipliers come from tags
Each roster entry SHALL carry matchup tags (for example `melee`, `ranged`, `caster`) and a list of damage multipliers against tags. A hit's matchup multiplier SHALL be the attacker's multiplier for the first of the victim's tags that the attacker lists, or ×1 when it lists none of them.

#### Scenario: Druzhynnyk against strilets
- **WHEN** a druzhynnyk hits a strilets, which is tagged `ranged`
- **THEN** the matchup multiplier is ×1.40

#### Scenario: No listed tag
- **WHEN** a strilets, which lists no multipliers, hits any unit
- **THEN** the matchup multiplier is ×1

### Requirement: Retiring a faction
A faction MAY be retired. Its id and the ids of units only it fielded SHALL stay reserved and never be reused. A retired faction SHALL NOT be offered, drawn for Red or accepted in a battle input or a new run; a run already playing it SHALL end. Retiring SHALL be a rule change with a sim version bump and re-recorded golden replays.

#### Scenario: Classic retired
- **WHEN** a battle input names faction 1, the retired Classic, for Blue
- **THEN** the input is rejected as an unknown faction

#### Scenario: Red after Classic
- **WHEN** battles are started with any map seed
- **THEN** Red's faction is always Rus, Varangians or Romans
