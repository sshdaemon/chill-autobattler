## Why

Roadmap Phase 1, step 1.3 (`docs/ROADMAP.md`): instead of picking an army on
its own, each side drafts before every battle, in turns, so you answer what
the other side picks. Each pick is a unit and its level-1 ability
(`unit-levels`), from the run's faction (`faction-pick`).

## What Changes

- **Draft before every battle**, after the battle is started (its map seed
  fixes Red's army): a coin flip from the seed decides who picks first; turns
  of 1-2-2-2-2-1 picks (five each); 15 s per turn, the turn's remaining picks
  random when it runs out.
- Blue's veterans fill its first picks automatically; every other Blue pick is
  a unit type of the faction and its level-1 ability.
- Red's picks are its generated army (`RedArmy`), revealed one by one in a
  seeded order.
- The sim publishes the draft (`sim.Draft`, `sim.DraftTurns`; wasm `draft()`),
  so client and server agree without a round trip; the server keeps checking
  only the locked army. No `DraftPick` RPC in single player.
- The army picker is removed; the drafted army goes to deployment as before.
- The deploy window grows to 3 minutes (draft plus deployment).
- No rule change for battles (`RulesVersion` unchanged).

## Capabilities

### New Capabilities

### Modified Capabilities
- `match-flow`: the draft replaces the army picker.
- `battle-sessions`: deploy deadline 3 minutes after start.
- `battle-rules`: the draft order and Red's reveal order come from the map seed.

## Impact

`sim/draft.go`, `wasm` (`draft()`), `server/battle` (deploy window),
`web` (`draft.ts`, `ui/Draft.tsx`, store phases, worker), `ui/Picker.tsx` removed.
