## ADDED Requirements

### Requirement: Draft order
The sim SHALL derive a battle's draft from its map seed alone: which side picks first (a fair coin flip) and the order Red's generated units are revealed in (a permutation of its five units). The same seed SHALL always give the same draft.

#### Scenario: Same seed, same draft
- **WHEN** the draft is derived twice from one map seed
- **THEN** both give the same first side and the same reveal order
