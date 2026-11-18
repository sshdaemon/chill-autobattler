# Spec Delta

## MODIFIED Requirements

### Requirement: Classes and stats
Each unit SHALL be an entry of the unit roster (see unit-roster), with the health, damage range, attack cooldown, range, speed and sight of that entry, adjusted by its skills. The roster's first entries are the warrior, archer and mage of docs/GAME_CONCEPT.md §4.1.

#### Scenario: Toughness
- **WHEN** a warrior with Toughness spawns
- **THEN** its health is 110

### Requirement: Damage
A hit SHALL deal a random roll in the attacker's damage range, multiplied by:
- the matchup multiplier from the attacker's bonuses against the victim's tags (unit-roster; for the first three units these equal §4.2),
- ×2 when the hit comes from outside the victim's ±60° front cone (for projectiles, measured from where they were fired),
- the attacker's damage skills.

#### Scenario: Flanked archer
- **WHEN** a warrior hits an archer from behind
- **THEN** the damage is the roll × 1.45 (matchup) × 2 (flank)

### Requirement: Skills
Units SHALL apply the skills of their roster entry, each made of effects (unit-roster) and learned at most once per unit; for the first three units these are the 15 skills in docs/GAME_CONCEPT.md §10.1. Chance-based effects (such as Executioner, Headshot and Frost) SHALL roll the battle's seeded random generator. Instant kills SHALL only be possible against victims below 40% of their maximum health.

#### Scenario: Executioner
- **WHEN** a warrior with Executioner hits a victim below 40% health and its roll succeeds
- **THEN** the victim dies regardless of remaining health

#### Scenario: No execute at full health
- **WHEN** a warrior with Executioner hits a victim at full health
- **THEN** the hit deals normal damage, whatever the roll

#### Scenario: Blink
- **WHEN** a melee enemy comes within 2.5 m of a mage with Blink off cooldown
- **THEN** the mage teleports about 8 m away from the threats to a free spot, and Blink goes on an 8 s cooldown
