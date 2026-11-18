# unit-looks Specification

## Purpose
Defines how each unit type looks and moves on screen: its models, its animation set and the animation and effect preset for each of its skills. It is kept as presentation data matched to the unit roster by name, so a new unit gets its look from data and never affects the battle.

## Requirements

### Requirement: Every roster entry has a look
Each unit-roster entry SHALL have a look, matched by roster name. A look SHALL hold the unit's models and its animation set: idle, move, one or more attack variants, hit reaction, death and cheer.

#### Scenario: Warrior look
- **WHEN** the client loads the looks
- **THEN** the warrior's look lists its models, an idle, a move, at least two attack variants, a hit reaction, a death and a cheer

### Requirement: Skills name their animation and effect
A look SHALL give each of its unit's five skills at most one action animation (played when the skill acts, such as a gap-closer or escape), at most one attack animation (replacing the unit's attacks while it has the skill) and at most one effect preset. A skill with none of them SHALL show only its stat changes.

#### Scenario: Blast replaces the attack
- **WHEN** a mage with Blast attacks
- **THEN** it plays Blast's attack animation instead of its normal attack

#### Scenario: Stat-only skill
- **WHEN** a warrior has Toughness
- **THEN** its animations and effects are those of a warrior without it

### Requirement: Attack clips know when the blow lands
Each attack animation SHALL record the moment its blow lands. Playback SHALL be timed so that moment matches the sim's hit, and the clip SHALL NOT play faster than 1.8× or slower than 0.6× its natural speed.

#### Scenario: Slow heavy blow
- **WHEN** a warrior with Swift plays a heavy attack variant whose blow lands late in the clip
- **THEN** the blow lands on screen within one frame of the sim's hit, with the clip sped up by at most 1.8×

### Requirement: Looks are checked
The test suite SHALL fail if a roster entry has no look, if a look misses one of its unit's skills or has one the unit can't learn, or if a look names an animation or effect preset that doesn't exist.

#### Scenario: Missing skill
- **WHEN** a look omits one of its unit's skills
- **THEN** the look check names the unit and the skill, and the tests fail

### Requirement: Missing looks fall back
If the client meets a unit or skill with no look at runtime, it SHALL fall back to the warrior's or archer's animation set (by melee or ranged) and no skill animation, and SHALL keep playing.

#### Scenario: Unit without a look
- **WHEN** a battle contains a roster entry whose look is missing
- **THEN** the unit renders with the fallback animation set and the battle plays normally

### Requirement: Every unit's skills are known to the client
The browser sim SHALL tell the client each unit's class and learned skills for both teams when a battle is created. Presentation SHALL use this list and SHALL NOT change any battle outcome.

#### Scenario: Red's skills
- **WHEN** a battle starts against a Red army with 3 skills
- **THEN** the client knows which Red units hold which skills, and those units play their skill animations
