## ADDED Requirements

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
- **WHEN** the player picks a mage with Blast
- **THEN** Blue's army gains a level-1 mage whose ability is Blast

## REMOVED Requirements

### Requirement: Army picker
**Reason**: Replaced by the draft: picks alternate with Red's and each includes the unit's level-1 ability.
**Migration**: Choose units in the draft (Requirement: Draft).
