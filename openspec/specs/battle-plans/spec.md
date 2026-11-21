# battle-plans Specification

## Purpose
Battle plans are army-wide stances a side picks before a battle. A plan changes
how the side's melee, ranged and caster units fight, so the same armies can
fight in different ways.

## Requirements

### Requirement: Built-in plans
The game SHALL have these built-in battle plans:
- **Instinct:** no change to any unit. It is the default.
- **Shield wall:** melee units hold, meet the nearest enemy and take less damage from the front; ranged and caster units focus on enemies within reach and stay behind the melee.
- **All-out charge:** every unit charges the nearest enemy faster than usual, ranged units don't kite, and everyone takes a little more damage.
- **Hunt the archers:** melee units charge the enemy's ranged and caster units first, faster than usual but taking a little more damage; ranged and caster units focus on enemies within reach.
- **Focus fire:** every unit advances at its allies' pace, attacks the enemy most allies are already fighting, and hits harder.
- **Skirmish:** melee units screen the back line without chasing far; ranged units kite, keep their distance, move faster and shoot further; casters keep their distance and reach further.

#### Scenario: Instinct changes nothing
- **WHEN** both sides use Instinct
- **THEN** the battle plays out exactly as it would without plans

### Requirement: Plan stances
A plan SHALL set some temperament fields for each role (melee, ranged, caster). Fields a plan leaves unset SHALL keep the unit's own value. A plan SHALL only produce temperaments that are valid for the unit.

#### Scenario: No kiting melee
- **WHEN** a plan would make a melee unit kite, or make a unit hold without a trigger distance
- **THEN** the game refuses to start and the tests fail

#### Scenario: Charge overrides a holding unit
- **WHEN** a side with a unit whose own temperament holds uses All-out charge
- **THEN** that unit charges the nearest enemy at once

#### Scenario: Fields the plan leaves alone
- **WHEN** a unit that stays near its melee allies with no target is in a side whose plan doesn't set that field
- **THEN** it still stays near its melee allies

### Requirement: Order of changes
A unit's temperament SHALL be built from its roster entry, then changed by its side's plan, then by its learned abilities.

#### Scenario: A skill on top of a plan
- **WHEN** a unit with an ability that makes it target ranged units first is in a side using Focus fire
- **THEN** it targets ranged units first

### Requirement: Plans in the battle input
Each army in the battle input SHALL carry a plan; an army without one SHALL use Instinct. The input check SHALL reject an unknown plan. Replays SHALL carry each side's plan, so the server's re-run gives the same result. Plans SHALL be part of the rules: a change to the plans SHALL change the sim version.

#### Scenario: Unknown plan
- **WHEN** a battle input names a plan that doesn't exist
- **THEN** the input is rejected and the battle isn't played

#### Scenario: Replay with plans
- **WHEN** the server re-runs a battle where Blue used Skirmish
- **THEN** it gets the same result as the client

### Requirement: Red's plan
When Red's army is generated, Red's plan SHALL be drawn at random among all built-in plans, from the combat seed, so it is decided only once Blue's army is locked. The same combat seed SHALL always give the same plan. When the input gives Red's army, Red SHALL use the plan given with it.

#### Scenario: Hidden during the draft
- **WHEN** a player knows the battle's map seed but not its combat seed
- **THEN** they can't tell which plan Red will use

#### Scenario: Same seed, same plan
- **WHEN** two battles have the same combat seed and a generated Red
- **THEN** Red uses the same plan in both

### Requirement: Plans don't stall
Any two plans, on any two armies, SHALL meet and fight before the time limit.

#### Scenario: Shield wall against Shield wall
- **WHEN** both sides use Shield wall
- **THEN** the armies meet and fight before the time limit

### Requirement: Plan effects
A plan's stance for a role MAY grant effects to that role's units for the battle, as if learned, applied after the unit's own traits and skills. Every effect a plan grants SHALL be valid for every unit it reaches.

#### Scenario: Charge is faster
- **WHEN** a side uses All-out charge
- **THEN** its units move faster than the same units under Instinct

#### Scenario: Invalid plan effect
- **WHEN** a plan grants a ranged-only effect to melee units
- **THEN** the game refuses to start and the tests fail

### Requirement: No dominant plan
For every playable faction, with the same army on both sides, no plan SHALL win 55% or more of its battles against every other plan.

#### Scenario: Plan matrix
- **WHEN** the balance tool plays every plan against every other for each playable faction
- **THEN** no plan beats all the others

### Requirement: Tactics
A plan MAY be a tactic: it splits the army into a left and a right group, gives each group its own stance, and gives each group one trigger and a second stance. Tactics SHALL be picked and sent like any other plan, and Red SHALL draw among all plans, tactics included.

#### Scenario: Tactic in the input
- **WHEN** Blue's army carries the id of Hammer and anvil
- **THEN** the battle runs it as a tactic, and the server's re-run of the battle gives the same result

#### Scenario: Red draws a tactic
- **WHEN** many combat seeds are drawn
- **THEN** Red's plans include each tactic

### Requirement: Groups
A tactic's groups SHALL be decided once, at the start of the battle, from the units' starting positions: units on the formation's left side of the army's centre (the side of a preset's first flank, the same for both teams) form the left group and the others the right group; a unit exactly on the centre line joins the left group.

#### Scenario: Flank formation
- **WHEN** Blue deploys in 3-2 with a tactic
- **THEN** the three units of the left flank form the left group and the two of the right flank the right group

#### Scenario: Dragged unit
- **WHEN** the player drags a left-flank unit to the far right before the battle
- **THEN** that unit is in the right group

### Requirement: Triggers
Each group of a tactic SHALL switch to its second stance, once and for the rest of the battle, when its trigger fires. The triggers SHALL be: a set time has passed since the battle started; the first unit of its army has fallen; a unit of the other group is fighting an enemy within reach.

#### Scenario: Timed switch
- **WHEN** the army uses Hold, then charge
- **THEN** for the first 10 seconds its units hold (they stop once they see the enemy and wait for it to come close), and after that they charge the nearest enemy

#### Scenario: Reserve joins
- **WHEN** the army uses Reserve and its first unit falls
- **THEN** the right group, which was waiting, goes in

#### Scenario: Switch is final
- **WHEN** a group has switched to its second stance
- **THEN** it stays in that stance for the rest of the battle

### Requirement: Switch changes behaviour only
A trigger SHALL change only how the group's units pick targets and move. The effects a tactic grants SHALL be given at the start of the battle from the group's first stance and SHALL NOT change when the trigger fires. A unit's learned abilities SHALL still apply on top of its second stance.

#### Scenario: Effects stay
- **WHEN** a group whose first stance grants extra speed switches to a stance without it
- **THEN** its units keep the extra speed

#### Scenario: Effects in a second stance
- **WHEN** a tactic's second stance grants an effect
- **THEN** the game refuses to start and the tests fail

### Requirement: Flank engage rule
A unit with the Flank engage rule and no target SHALL move round the side of the enemy army instead of straight at it, and SHALL fight normally once it has a target.

#### Scenario: Hammer goes round
- **WHEN** the army uses Hammer and anvil
- **THEN** its right group's melee reach the enemy from the side, further out than they would under Instinct

### Requirement: Built-in tactics
The game SHALL have these built-in tactics:
- **Hammer and anvil:** the left group holds as a shield wall; the right group flanks and charges the nearest enemy once the left group is fighting.
- **Reserve:** the left group advances in line; the right group holds, ignoring the left group's fight, and charges once the first ally falls.
- **Hold, then charge:** both groups hold, braced, for 10 seconds, then charge the nearest enemy.

#### Scenario: Tactics listed
- **WHEN** the game lists its plans
- **THEN** Hammer and anvil, Reserve and Hold, then charge are among them, marked as tactics

### Requirement: Tactics don't stall
Any two plans, tactics included, SHALL meet and fight before the battle's time limit, for every pair of factions.

#### Scenario: Two reserves
- **WHEN** both sides use Reserve
- **THEN** the battle ends before the time limit

### Requirement: Holding groups wait for their own group
In a tactic, a holding unit SHALL be woken by an enemy within its trigger distance, by a unit of its own group fighting, or by running out of patience, but not by the other group's fight. Without a tactic, the whole army counts as one group.

#### Scenario: Reserve stays back
- **WHEN** the army uses Reserve and its left group starts fighting while no ally has fallen
- **THEN** the right group keeps holding
