## MODIFIED Requirements

### Requirement: Formation presets
The deployment bar SHALL offer the presets Fully spread, 3-2, 2-3, 4-1, 1-4, 0-5 and 5-0. Each preset SHALL show a mini-diagram of where the player's own units would stand, marking melee, ranged and caster units differently. Choosing a preset SHALL move Blue's units there.

#### Scenario: Pick a preset
- **WHEN** the player chooses 5-0
- **THEN** all five Blue units move to the left flank, and the 5-0 button is marked as selected

#### Scenario: Marks by role
- **WHEN** a Rus army of two druzhynnyky, two striltsi and a huslyar looks at the presets
- **THEN** each diagram marks two melee, two ranged and one caster spot

### Requirement: Faction screen
A new run SHALL open on a faction screen before the first army picker. Each playable faction SHALL show its name, colours, lore and its units (label, icon and blurb); choosing one SHALL start the run as that faction. Factions not playable yet MAY be shown with their planned units and SHALL NOT be selectable. Retired factions SHALL NOT be shown.

#### Scenario: Pick a faction
- **WHEN** a player with no active run opens the game and chooses Rus
- **THEN** a run starts as Rus and the army picker opens

#### Scenario: Coming soon
- **WHEN** the faction screen shows a faction that isn't playable yet
- **THEN** its card shows its planned units and its button is disabled

#### Scenario: Classic is gone
- **WHEN** the faction screen opens
- **THEN** it offers Rus, Varangians and Romans, and no Classic card

### Requirement: Draft
Before each battle of a run, after the battle is started, both sides SHALL draft five units in turns of 1, 2, 2, 2, 2 and 1 picks, alternating, the side picking first decided by a coin flip from the battle's map seed. Blue's veterans SHALL fill Blue's first picks by themselves. Every other Blue pick SHALL be a unit type of the run's faction with one of its abilities as its level-1 ability. Each Blue turn SHALL last at most 15 seconds; when it runs out, the turn's remaining picks SHALL be made at random. Red's picks SHALL be its generated army, revealed one unit at a time in a seeded order. The drafted army SHALL then go to deployment. Reloading during a draft SHALL draft the same battle again, with the same order and Red picks.

#### Scenario: Turn order
- **WHEN** the coin flip says Red picks first
- **THEN** Red picks 1, Blue 2, Red 2, Blue 2, Red 2 and Blue 1

#### Scenario: Veterans first
- **WHEN** the run has two veterans
- **THEN** they fill Blue's first two picks by themselves, and the player drafts the other three

#### Scenario: Time runs out
- **WHEN** the player makes no pick in a 15-second turn
- **THEN** that turn's picks are made at random and the draft goes on

#### Scenario: A pick
- **WHEN** a Varangian player picks a volva with one of her abilities
- **THEN** Blue's army gains a level-1 volva whose ability is that one
