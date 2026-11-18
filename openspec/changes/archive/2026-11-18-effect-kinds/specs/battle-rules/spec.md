## ADDED Requirements

### Requirement: Modifiers
Each tick, every living unit's modifiers (damage dealt, damage taken, attack speed, evasion, speed) SHALL be recomputed from its timed buffs, the auras of units around it and its own conditional effects, and SHALL apply to its blows, the blows it takes, its attack cooldown and its movement. A unit with no modifiers SHALL play exactly as without them.

#### Scenario: A song
- **WHEN** a unit with a damage aura of +10% within 6 m stands 3 m from an ally
- **THEN** that ally's blows deal 10% more, and an ally 9 m away is unaffected

#### Scenario: Dodging
- **WHEN** a unit with 10% evasion is hit
- **THEN** with a seeded 10% chance the blow misses entirely

### Requirement: Effect kinds of the roster plan
The sim SHALL offer these effect kinds for unit abilities and traits: conditional damage taken (always, against ranged, from the front, while standing), aura on allies or foes, rally on an ally's death, guard (taking a share of blows aimed at tagged allies), brace (a counter-blow against a dash), fury (below a health %), curse (the victim takes more damage for a while), fear (nearby enemies flee for a while), burst (a thrown ×factor attack with optional splash), bonus against tags, focus (more damage when allies share the target), phalanx (per nearby ally of the same type), burning ground (fire that damages over time), shooting while backing away, and a cone blow. Slow and splash SHALL apply to melee and cone hits as well as projectiles.

#### Scenario: Shield wall
- **WHEN** a unit with damage taken −30% against ranged is hit by an arrow
- **THEN** it takes 70% of the damage, and full damage from a sword

#### Scenario: Cone
- **WHEN** a unit with a cone blow strikes a target with another enemy beside it and one behind it
- **THEN** both enemies in front take the blow and the one behind does not
