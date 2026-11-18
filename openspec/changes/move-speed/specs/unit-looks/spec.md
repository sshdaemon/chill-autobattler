# Spec Delta

## ADDED Requirements

### Requirement: Move clips know their speed
Every walk, move and sprint clip a look names SHALL record the ground speed it covers when played at 1× on our characters (how fast a planted foot moves back relative to the body). A look SHALL name a walk and a move clip, and MAY name a sprint clip. The look check SHALL fail if a named move clip has no speed.

#### Scenario: Missing speed
- **WHEN** a look names a move clip without a recorded speed
- **THEN** the look check names the clip and the tests fail
