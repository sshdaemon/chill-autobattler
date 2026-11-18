# unit-authoring Specification

## Purpose
Defines how a new unit goes from a short brief and reference pictures to a complete, balanced and previewed unit in a pull request: the brief, the user's checkpoints, the balance targets, the preview and what must stop the workflow.

## Requirements

### Requirement: A brief describes the unit
Each new unit SHALL start from `design/units/<name>/brief.md`, following the template: concept, battlefield role, how it should feel, skill ideas and things to avoid. Reference pictures SHALL live in `design/units/<name>/refs/`, which SHALL be ignored by git and never committed, uploaded or published.

#### Scenario: References stay local
- **WHEN** a user adds three reference pictures to `design/units/spearman/refs/` and the workflow opens its PR
- **THEN** the PR contains the brief but none of the pictures

### Requirement: The user approves the design first
Before writing any game file, the workflow SHALL present the proposed unit: role, stats, traits, tags and matchup bonuses, five skills mapped to existing effect kinds with their numbers, and the look (recipe, animations, effect presets). It SHALL wait for the user's approval or changes.

#### Scenario: User changes a skill
- **WHEN** the user answers the design with "make the third skill a slow instead of a stun"
- **THEN** the workflow revises the design and asks again before writing files

### Requirement: Behaviour beyond existing effects stops the workflow
If the brief needs behaviour no existing effect kind covers, the workflow SHALL say which, and propose a separate change for a new effect kind. It SHALL NOT approximate the behaviour with unrelated effects without the user's agreement.

#### Scenario: A healer
- **WHEN** a brief asks for a unit that heals allies
- **THEN** the workflow reports that no heal effect exists, proposes a new-effect change, and writes no unit files until the user decides

### Requirement: New units are balanced against the roster
A new unit SHALL be checked with seeded battles against every existing unit and in mixed armies, and SHALL land in these bands:
- five of it against five of each other unit: 25–75% wins;
- a standard army with one unit swapped for it: within 10 points of the unswapped army's win rate;
- each of its skills: a 0 to +15 point gain over having no skill.

The workflow SHALL tune only the unit's numbers, within the roster limits, to reach the bands, and SHALL report the final results.

#### Scenario: Overtuned unit
- **WHEN** five of a new unit win 92% against five archers
- **THEN** the workflow lowers its numbers and re-runs until the result is within 25–75%, or reports that it could not

### Requirement: Every new unit is previewed
The workflow SHALL render the unit with the dev preview, on both teams, in idle, move, each attack variant, each skill action and effect, and death, plus its portrait. It SHALL include these screenshots in the PR.

#### Scenario: Preview in the PR
- **WHEN** the workflow opens the PR for a new unit
- **THEN** the PR shows the unit's screenshots on Blue and Red and its portrait

### Requirement: A new unit arrives as one reviewable PR
The workflow SHALL deliver the brief, roster entry, look entry, effect presets, UI text, a golden replay case using the unit and the balance report in one PR, with all tests passing. The user SHALL merge it.

#### Scenario: Tests in the PR
- **WHEN** the workflow opens the PR
- **THEN** the roster check, the look check, the golden replays and the typecheck pass, and the PR description holds the balance bands with their results

### Requirement: Units can be inspected in a lab
In development builds, a unit lab SHALL show any unit type's recipes on either team, assembled as in battle, with a camera that orbits all the way round. It SHALL play, pause, loop, slow down and scrub any clip; preview the walk–jog blend at a chosen ground speed on a moving floor, with its foot sliding shown; fire each skill's animation and effect and any effect preset; and switch build and variety. It SHALL NOT be part of production builds.

#### Scenario: Look all around
- **WHEN** a developer opens the lab for the warrior and drags across the view
- **THEN** the camera orbits the warrior through 360°, and the cape, shield and helm are visible from every side

#### Scenario: Scrub an attack
- **WHEN** they pick Sword_Slash_A, pause it and drag the timeline
- **THEN** the pose follows the timeline frame by frame

#### Scenario: Walking speed
- **WHEN** they set the ground speed to the archer's speed while frozen (2.2 m/s)
- **THEN** the archer walks on the moving floor with the battle's walk–jog blend, and the readout shows the blend and a foot slide under 10%

#### Scenario: Not shipped
- **WHEN** the production bundle is built
- **THEN** it contains no unit lab code
