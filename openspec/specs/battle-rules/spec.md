# battle-rules Specification

## Purpose
Defines how a battle plays out tick by tick once both armies are deployed: classes, targeting, movement, attacks, skills and how a battle ends. Balance numbers come from docs/GAME_CONCEPT.md and live in data tables.

## Requirements

### Requirement: Fixed tick
The battle SHALL advance in fixed ticks of 1/20 s. All durations (cooldowns, wind-ups, stagger, slow, skill cooldowns) SHALL be counted in ticks.

#### Scenario: Rendering speed does not change the battle
- **WHEN** the browser renders at any frame rate or plays at 1×, 2× or 4× speed
- **THEN** the battle advances by the same sequence of ticks and ends with the same outcome

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

### Requirement: Shared team vision and targeting
A team SHALL see an enemy if any of its living units has it within sight and not blocked by rocks or trees. Units SHALL target the nearest enemy their team sees, and SHALL re-check targets every 0.5 s.

#### Scenario: Spotted by a teammate
- **WHEN** an archer's target is behind a rock for the archer, but a teammate sees it
- **THEN** the target stays targetable, and the archer moves to get a clear shot

#### Scenario: Nothing in sight
- **WHEN** a unit's team sees no enemy
- **THEN** the unit checks the last seen position if there is one, and otherwise marches toward the enemy edge, then sweeps random spots in the enemy half

### Requirement: Movement
Units SHALL path around rocks and trees on the 0.5 m grid, where a water step costs about 1.43× a dry one. Units SHALL push apart when closer than 0.9 m.

#### Scenario: Pond detour
- **WHEN** a pond lies between a unit and its goal and walking around it is about as quick
- **THEN** the unit walks around instead of wading

### Requirement: Attacks and projectiles
Melee hits SHALL land partway through the swing. Archers and mages SHALL need a clear line of fire, and SHALL fire homing projectiles (arrows at 18 m/s, bolts at 14 m/s) that hit if the target is still alive when they arrive.

#### Scenario: Target dies before the arrow lands
- **WHEN** an arrow's target dies while the arrow is in flight
- **THEN** the arrow deals no damage

### Requirement: Stagger
A non-lethal hit SHALL stagger the victim for 0.3 s, unless the victim is mid-attack.

#### Scenario: Interrupted walk
- **WHEN** a walking warrior takes a non-lethal hit
- **THEN** it does not move or attack for 0.3 s

### Requirement: Archer kiting
An archer SHALL retreat from melee enemies closer than 4.5 m, in bursts of at most 2 s followed by at least 1.2 s of standing and shooting. It SHALL back away while still facing its attackers, at full speed. A cornered archer SHALL stand and shoot.

#### Scenario: Warrior closes in
- **WHEN** a warrior comes within 4.5 m of an archer that is not resting
- **THEN** the archer steps away from it, still facing it, to a free spot it can walk to in a straight line

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

### Requirement: Red army
When an input leaves Red's army empty, Red SHALL get 5 units of seeded random classes, a seeded random formation preset, and as many skills in total as the given count, spread at random among its units. All of this SHALL come from the seed.

#### Scenario: Matching skills
- **WHEN** a battle input asks for 3 Red skills
- **THEN** Red's units have exactly 3 skills in total, with no unit holding the same skill twice

### Requirement: Formations
Formation presets (Spread, 3-2, 2-3, 4-1, 1-4, 0-5, 5-0) SHALL place a team's units inside its zone, melee in front and ranged behind. Any slot that falls on a blocked cell SHALL move to the nearest free cell.

#### Scenario: Preset 5-0
- **WHEN** Blue uses 5-0 with 2 warriors, 2 archers and 1 mage
- **THEN** all five stand on the left flank, the two warriors in the front row

### Requirement: Battle end
The battle SHALL end on the tick when one team has no living units, and that team loses. If both teams lose their last units on the same tick, or the tick limit is reached, the battle SHALL be a draw. Damage SHALL stop once the battle has ended.

#### Scenario: Mutual wipe
- **WHEN** the last Blue and the last Red unit die on the same tick
- **THEN** the outcome is a draw
