# Spec Delta

## ADDED Requirements

### Requirement: Early matches are easier
Red's units SHALL start with 85% of their health in match 1 of a run, 90% in match 2, 95% in match 3, and full health from match 4 on. The server SHALL set this for each battle it starts, and SHALL verify the battle with the value it set.

#### Scenario: First match
- **WHEN** a run starts its first battle
- **THEN** every Red unit in that battle starts at 85% of its normal maximum health

#### Scenario: Later matches
- **WHEN** a run starts its fourth or later battle
- **THEN** Red's units start at full health
