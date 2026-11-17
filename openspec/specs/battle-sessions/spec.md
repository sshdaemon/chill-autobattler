# battle-sessions Specification

## Purpose
Server-issued battles: the server fixes a battle's randomness before it is played and accepts a claimed outcome only for a battle it issued, so a client can't re-roll seeds or deployments until it finds a win.

## Requirements

### Requirement: Start a battle
On request, the server SHALL create a battle with a new random map seed and a new random combat seed, store it, and return the battle id, the map seed and the deploy deadline. The combat seed SHALL NOT be returned at this point.

#### Scenario: New battle
- **WHEN** a client starts a battle for the current sim version with a Red skill count of 2
- **THEN** it receives a battle id, a map seed and a deploy deadline, but no combat seed

#### Scenario: Unknown sim version
- **WHEN** a client starts a battle for a sim version the server doesn't have
- **THEN** the server refuses with a reason naming both versions

#### Scenario: Seeds differ between battles
- **WHEN** two battles are started
- **THEN** they have different ids, map seeds and combat seeds

### Requirement: Lock a deployment once
The server SHALL accept Blue's army for a battle once, before its deploy deadline (default 2 minutes after start). It SHALL validate the army against the battle rules, store it, and only then reveal the combat seed.

#### Scenario: Lock and reveal
- **WHEN** a client locks a valid army before the deadline
- **THEN** the server stores it and returns the combat seed

#### Scenario: Second lock
- **WHEN** a client tries to lock a different army for a battle that is already locked
- **THEN** the server refuses, and the stored army is unchanged

#### Scenario: Too late
- **WHEN** a client locks an army after the deploy deadline
- **THEN** the server refuses because the battle has expired

#### Scenario: Invalid army
- **WHEN** a client locks an army with six units or a unit outside its deployment zone
- **THEN** the server rejects it as invalid, and the battle stays unlocked

### Requirement: Verify by battle id
The server SHALL verify a claimed outcome only for a locked battle it issued, before its verify deadline (default 10 minutes after locking). It SHALL rebuild the battle input from what it stored (both seeds, Blue's army, Red's skill count, sim version), never from client-sent input, re-run it, and store the verdict.

#### Scenario: Honest claim
- **WHEN** a client claims the outcome its sim produced for a locked battle
- **THEN** the server reports it fair and returns its own outcome

#### Scenario: Claim for a re-rolled battle
- **WHEN** a client claims an outcome it got by playing with a different combat seed or army than the ones stored
- **THEN** the server reports it not fair

#### Scenario: Unknown or unlocked battle
- **WHEN** a client asks to verify a battle id the server never issued, or one that was never locked
- **THEN** the server refuses

### Requirement: Verify once
A battle's verdict SHALL be decided by its first verification. Later requests SHALL return the stored verdict and outcome, whatever they claim.

#### Scenario: Second claim
- **WHEN** a battle was verified as not fair and the client then sends the correct outcome
- **THEN** the server still answers not fair, with the stored outcome

### Requirement: Battles survive restarts
Issued battles, locked armies and verdicts SHALL be stored durably, so a server restart doesn't lose them or let them be re-issued.

#### Scenario: Restart between lock and verify
- **WHEN** the server restarts after a battle was locked
- **THEN** the battle can still be verified with the same seeds and army
