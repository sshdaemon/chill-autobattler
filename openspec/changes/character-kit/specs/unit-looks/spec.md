# Spec Delta

## MODIFIED Requirements

### Requirement: Every roster entry has a look
Each unit-roster entry SHALL have a look, matched by roster name. A look SHALL hold one or more character recipes and the unit's animation set: idle, move, one or more attack variants, hit reaction, death and cheer.

#### Scenario: Warrior look
- **WHEN** the client loads the looks
- **THEN** the warrior's look lists at least two recipes, an idle, a move, at least two attack variants, a hit reaction, a death and a cheer

## ADDED Requirements

### Requirement: Characters are assembled from recipes
A recipe SHALL name a body build, a part for each body slot (head, hair, torso, arms, legs, feet), optional accessory parts, an outfit colour variant, a main-hand weapon and an optional off-hand or back item. The client SHALL assemble a unit from shared parts by its recipe at runtime. The n-th unit of a class in a team SHALL use recipe n (mod the list), stable across reloads.

#### Scenario: New combination without a re-bake
- **WHEN** a recipe is changed to put the ranger hood on a knight outfit
- **THEN** the next battle shows that combination with no asset rebuild

#### Scenario: Two warriors
- **WHEN** Blue has two warriors and the warrior look has two recipes
- **THEN** the two warriors are assembled from different recipes

### Requirement: Body builds shape the skeleton
A recipe's build (slim, normal, bulky or stout) SHALL scale the skeleton's shoulders, chest, limbs, head and height on top of every animation. A build SHALL keep the unit's overall height within ±8% of the base body, its width within the sim's unit footprint, and its feet on the ground in every clip.

#### Scenario: Bulky warrior
- **WHEN** a warrior uses the bulky build
- **THEN** it is visibly broader in shoulders and arms than a slim archer, and its attack and death animations still play without feet sinking or floating

### Requirement: Team colours come from cloth masks
Each outfit texture SHALL have a mask of its dyeable cloth, derived from where its colour variants differ. Units SHALL show their team colour on that cloth, at the strength the recipe gives, instead of separate team textures or code-made trims.

#### Scenario: Same recipe, two teams
- **WHEN** Blue and Red both field a warrior with the same recipe
- **THEN** both use the same textures, and the cloth is blue on one and red on the other while skin, leather and metal look the same

### Requirement: Weapons are modelled parts
Main-hand weapons, off-hand items and back items SHALL be modelled parts attached to the matching hand or spine bone, named by the recipe.

#### Scenario: Knight with shield
- **WHEN** a warrior recipe names a sword and a shield
- **THEN** the unit holds the sword in its right hand and the shield on its left arm through every animation

### Requirement: Parts stay within budget and are checked
The parts library SHALL stay within 10 MB in total, and the parts a default army uses within 3 MB. The test suite SHALL fail if a recipe names a build, part, colour variant or item the parts manifest lacks, if a slot holds a part of another slot, or if the library exceeds its budget.

#### Scenario: Missing part
- **WHEN** a recipe names a helmet the parts manifest lacks
- **THEN** the look check names the unit, the recipe and the part, and the tests fail
