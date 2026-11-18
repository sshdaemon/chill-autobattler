## MODIFIED Requirements

### Requirement: Skills
Units SHALL apply the skills (abilities) of their roster entry, each made of effects (unit-roster) and learned at most once per unit. A unit in a battle SHALL have 1 to 3 abilities, one per level; an input with a unit outside that range SHALL be rejected. For the first three units the abilities are the 15 in docs/GAME_CONCEPT.md §10.1. Chance-based effects (such as Executioner, Headshot and Frost) SHALL roll the battle's seeded random generator. Instant kills SHALL only be possible against victims below 40% of their maximum health.

#### Scenario: Executioner
- **WHEN** a warrior with Executioner hits a victim below 40% health and its roll succeeds
- **THEN** the victim dies regardless of remaining health

#### Scenario: No execute at full health
- **WHEN** a warrior with Executioner hits a victim at full health
- **THEN** the hit deals normal damage, whatever the roll

#### Scenario: Blink
- **WHEN** a melee enemy comes within 2.5 m of a mage with Blink off cooldown
- **THEN** the mage teleports about 8 m away from the threats to a free spot, and Blink goes on an 8 s cooldown

#### Scenario: Levels
- **WHEN** a battle input has a unit with no ability, or with four
- **THEN** the input is rejected

### Requirement: Red army
When an input leaves Red's army empty, Red SHALL get 5 units of seeded random classes, a seeded random formation preset, one seeded random ability each (level 1), and then as many more abilities in total as the given count, spread at random among its units up to level 3. All of this SHALL come from the seed.

#### Scenario: Matching skills
- **WHEN** a battle input asks for 3 Red skills
- **THEN** Red's units have exactly 8 abilities in total (one each, plus 3), with no unit holding the same ability twice or more than 3
