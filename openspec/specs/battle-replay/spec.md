# battle-replay Specification

## Purpose
Defines what a battle replay is (its input and outcome) and guarantees that the same replay plays out identically in the browser and on the server, so the server can verify a claimed result.

## Requirements

### Requirement: Battle input is complete
A battle input SHALL contain everything that affects the outcome: sim version, map seed, combat seed, Red's health percentage, and per team the units (class, skills) with deployment positions or a formation preset. The map seed SHALL decide the map and a generated Red army. The combat seed SHALL decide every random combat roll. A health percentage of 0 SHALL mean full health. Nothing outside the input (time, device, frame rate) SHALL affect the outcome.

#### Scenario: Frame rate does not matter
- **WHEN** the same battle input is played once at 30 fps and once at 144 fps in the browser
- **THEN** both runs produce the same outcome and checksum

#### Scenario: Red army comes from the seed
- **WHEN** a battle input lists no units for Red
- **THEN** Red's army, formation and skills are generated from the map seed, and the same map seed always gives the same Red army

#### Scenario: Combat seed changes only the fight
- **WHEN** the same input is played with two different combat seeds
- **THEN** the map and both armies' starting positions are identical, but the battles play out differently

#### Scenario: Reduced Red health
- **WHEN** a battle input sets Red's health to 85%
- **THEN** every Red unit starts with 85% of its maximum health, and the same input always gives the same outcome

### Requirement: Outcome is deterministic across platforms
Playing the same battle input SHALL produce a bit-identical outcome (winner, tick count, surviving units with health, state checksum) in the WASM build and in the native server build.

#### Scenario: Golden replays match on both builds
- **WHEN** the golden replay set is played natively and under WASM
- **THEN** every replay's outcome equals the recorded golden outcome on both

### Requirement: Battles always end
A battle SHALL end when one team has no living units, or at a tick limit, which counts as a draw.

#### Scenario: Tick limit reached
- **WHEN** both teams still have living units at the tick limit
- **THEN** the outcome is a draw with the limit as its tick count

### Requirement: Replays are versioned
Every battle input SHALL carry the sim version that played it. The server SHALL verify only with that exact version, and SHALL reject a version it doesn't have rather than verify with different rules.

#### Scenario: Unknown version
- **WHEN** the server receives a battle input whose sim version it does not have
- **THEN** it answers "not fair" with a reason naming both versions, and does not run the battle

### Requirement: Server verifies a claimed outcome
The server SHALL verify a claimed outcome against the battle input it stored for that battle (see battle-sessions), re-running it natively and reporting it fair only if winner, tick count and checksum all match the client's claim.

#### Scenario: Honest claim
- **WHEN** a client submits the outcome its WASM sim produced for a battle it locked
- **THEN** the server reports it fair and returns its own outcome

#### Scenario: Tampered claim
- **WHEN** the claimed winner or checksum differs from the server's re-run
- **THEN** the server reports it not fair with a reason, and returns its own outcome

#### Scenario: Tampered input
- **WHEN** a client tries to lock an army that puts a unit outside its team's deployment zone, gives a unit a skill twice, or has more than 5 units per team
- **THEN** the server rejects the army as invalid, and no battle with that army can be verified
