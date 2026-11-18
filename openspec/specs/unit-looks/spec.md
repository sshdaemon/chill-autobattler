# unit-looks Specification

## Purpose
Defines how each unit type looks and moves on screen: its models, its animation set and the animation and effect preset for each of its skills. It is kept as presentation data matched to the unit roster by name, so a new unit gets its look from data and never affects the battle.

## Requirements

### Requirement: Every roster entry has a look
Each unit-roster entry SHALL have a look, matched by roster name. A look SHALL hold one or more character recipes and the unit's animation set: idle, move, one or more attack variants, hit reaction, death and cheer.

#### Scenario: Warrior look
- **WHEN** the client loads the looks
- **THEN** the warrior's look lists at least two recipes, an idle, a move, at least two attack variants, a hit reaction, a death and a cheer

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

### Requirement: Characters are assembled from recipes
A recipe SHALL name a body build, a part for each body slot (head, hair, torso, arms, legs, feet), optional accessory parts, an outfit colour variant, a main-hand weapon and an optional off-hand or back item. The client SHALL assemble a unit from shared parts by its recipe at runtime. The n-th unit of a class in a team SHALL use recipe n (mod the list), stable across reloads.

#### Scenario: New combination without a re-bake
- **WHEN** a recipe is changed to put the ranger hood on a knight outfit
- **THEN** the next battle shows that combination with no asset rebuild

#### Scenario: Two warriors
- **WHEN** Blue has two warriors and the warrior look has two recipes
- **THEN** the two warriors are assembled from different recipes

### Requirement: Body builds shape the body
A recipe's build (slim, normal, bulky or stout) SHALL change chest and limb thickness, shoulder width, head size and height on top of every animation. A build SHALL keep the unit's overall height within ±8% of the base body, its width within the sim's unit footprint, and its feet on the ground in every clip.

#### Scenario: Bulky warrior
- **WHEN** a warrior uses the bulky build
- **THEN** it is visibly broader in shoulders and arms than a slim archer, and its attack and death animations still play without feet sinking or floating

### Requirement: Team colours come from cloth masks
Each outfit texture SHALL have a mask of its dyeable cloth, derived from where its colour variants differ (leaving out leather that recolours between variants). Units SHALL show their team colour on that cloth, at the strength the recipe gives, instead of separate team textures or code-made trims.

#### Scenario: Same recipe, two teams
- **WHEN** Blue and Red both field a warrior with the same recipe
- **THEN** both use the same textures, and the cloth is blue on one and red on the other while skin, leather and metal look the same

### Requirement: Weapons are modelled parts
Main-hand weapons, off-hand items and back items SHALL be modelled parts attached to the matching hand or spine bone, named by the recipe.

#### Scenario: Knight with shield
- **WHEN** a warrior recipe names a sword and a shield
- **THEN** the unit holds the sword in its right hand and the shield on its left arm through every animation

### Requirement: Parts stay within budget and are checked
The parts library SHALL stay within 10 MB in total, and the parts of every unit type's first recipe together within 4.5 MB. The test suite SHALL fail if a recipe names a build, part, colour variant or item the parts manifest lacks, if a slot holds a part of another slot, or if the library exceeds its budget.

#### Scenario: Missing part
- **WHEN** a recipe names a helmet the parts manifest lacks
- **THEN** the look check names the unit, the recipe and the part, and the tests fail

### Requirement: Move clips know their speed
Every walk, move and sprint clip a look names SHALL record the ground speed it covers when played at 1× on our characters (how fast a planted foot moves back relative to the body). A look SHALL name a walk and a move clip, and MAY name a sprint clip. The look check SHALL fail if a named move clip has no speed.

#### Scenario: Missing speed
- **WHEN** a look names a move clip without a recorded speed
- **THEN** the look check names the clip and the tests fail

### Requirement: Shields are strapped to the forearm
Shields SHALL be mounted along the forearm, face out from the back of the hand and far enough from the arm that the hand and forearm stay behind the shield in every clip, for every build.

#### Scenario: Shield idle
- **WHEN** a heavy warrior with a round shield plays its shield idle
- **THEN** the shield covers the body facing forward and no part of the fist shows through its face

### Requirement: Capes are simulated cloth
Capes SHALL hang from the shoulders as cloth that swings under gravity and the unit's motion, drapes over the back and doesn't pass through the torso or legs. They SHALL settle over the pose for still renders and reset when the unit teleports.

#### Scenario: Running
- **WHEN** a warrior with a cape jogs for two seconds and stops
- **THEN** the cape trails behind while jogging and swings back to hang over its back after stopping

#### Scenario: Teleport
- **WHEN** a unit wearing a cape is moved more than 0.8 m in one frame
- **THEN** the cape re-hangs from the new position instead of stretching across the gap

### Requirement: Units stand on the ground
Each assembled character SHALL be lifted so that its lowest point standing in its idle pose touches the ground, whatever its outfit, sex and build, because the clips' pelvis height belongs to the animation pack's shorter-legged mannequin.

#### Scenario: Standing and walking
- **WHEN** any unit stands, walks or attacks on dry ground
- **THEN** its soles rest on the ground within 2 cm instead of sinking into it
