# Proposal

## Why

The web app is still a debug panel: you press New map, pick a formation from a dropdown and press Fight. The game concept (§2, §3, §8, §9) describes a real match: choose your army, deploy it under fog with presets or by dragging, a short countdown, the fight, and a result. The server-verified battle flow is now in place, so this is the next step toward something playable. Launching the game also takes two terminals today.

## What Changes

- **Army picker:** pick 5 units from warrior, archer and mage cards. Each card shows stats and the damage-taken line from the matchup table. The default is 2/2/1, and the last pick is remembered.
- **Deployment:**
  - The server-issued map is shown, with Blue in its zone and Red hidden under fog.
  - A bottom bar has the formation presets, each with a mini-diagram of your actual army mix.
  - You can drag your own units anywhere in your zone, which switches to Custom.
  - A 15 s timer runs; **Ready** skips it.
- **Countdown and fight:** locking the army reveals the combat seed. A short "Fight!" banner follows, then the battle plays. 1×/2×/4× speed stays.
- **Result:** a Victory, Defeat or Draw banner with survivors. The battle is verified with the server automatically and the verdict shown. **Next match** returns to the army picker with the last army.
- **Single source of balance data:** the WASM build exposes class stats, skills, matchups and formation slots, so the UI never copies numbers from `sim/data.go`.
- **One-command local launch:** `pnpm start` builds the WASM and runs the Go server and the web app together, stopping both on Ctrl-C.

## Out of scope

- Runs, veterans, skill picking, Red scaling, scoring and the scoreboard (the next change).
- Real character models and animations, fog of war *during* the fight, portraits, and the Warcraft III camera (the visuals change). Red is simply hidden while deploying.
- Sim or server rule changes. `sim.Version` stays 0.3.0, except for the read-only data and formation-slot export from the WASM bridge.

## Capabilities

### New Capabilities

- `match-flow`: the player-facing match sequence in the web client: army picker, deployment (presets, dragging, fog, timer), countdown, fight, and result with automatic verification.

### Modified Capabilities

None. The battle rules, replay and session specs are unchanged.

## Impact

- `wasm/`: new read-only `data()` (stats, skills, matchups as JSON) and `formationSlots(formation, classes)`.
- `web/`: a new phase machine in the store; new `Picker`, `DeployBar`, `Banner` and `Result` components; drag in the scene; Red hidden while deploying; debug controls moved into a small dev panel.
- Root `package.json`: a `start` script.
- README: how to play and run.
