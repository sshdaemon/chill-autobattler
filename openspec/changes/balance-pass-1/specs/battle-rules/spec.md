# Spec Delta

## MODIFIED Requirements

### Requirement: Classes and stats
Each unit SHALL be a warrior, archer or mage, with the health, damage range, attack cooldown, range, speed and sight given in docs/GAME_CONCEPT.md §4.1, adjusted by its skills.

#### Scenario: Toughness
- **WHEN** a warrior with Toughness spawns
- **THEN** its health is 110

### Requirement: Damage
A hit SHALL deal a random roll in the attacker's damage range, multiplied by:
- the matchup multiplier (§4.2),
- ×2 when the hit comes from outside the victim's ±60° front cone (for projectiles, measured from where they were fired),
- the attacker's damage skills.

#### Scenario: Flanked archer
- **WHEN** a warrior hits an archer from behind
- **THEN** the damage is the roll × 1.45 (matchup) × 2 (flank)

### Requirement: Archer kiting
An archer SHALL retreat from melee enemies closer than 4.5 m, in bursts of at most 2 s followed by at least 1.2 s of standing and shooting. It SHALL back away while still facing its attackers, at full speed. A cornered archer SHALL stand and shoot.

#### Scenario: Warrior closes in
- **WHEN** a warrior comes within 4.5 m of an archer that is not resting
- **THEN** the archer steps away from it, still facing it, to a free spot it can walk to in a straight line

### Requirement: Skills
Units SHALL apply the 15 skills in docs/GAME_CONCEPT.md §10.1, each learned at most once per unit. Chance-based skills (Executioner, Headshot) SHALL roll the battle's seeded random generator. Instant kills SHALL only be possible against victims below 40% of their maximum health.

#### Scenario: Executioner
- **WHEN** a warrior with Executioner hits a victim below 40% health and its roll succeeds
- **THEN** the victim dies regardless of remaining health

#### Scenario: No execute at full health
- **WHEN** a warrior with Executioner hits a victim at full health
- **THEN** the hit deals normal damage, whatever the roll

#### Scenario: Blink
- **WHEN** a melee enemy comes within 2.5 m of a mage with Blink off cooldown
- **THEN** the mage teleports about 8 m away from the threats to a free spot, and Blink goes on an 8 s cooldown
