# Design

## Context

The scaffold has a toy deterministic sim (`sim/`), its WASM bridge (`wasm/`), a Connect server that re-runs battles (`server/`) and a placeholder R3F scene (`web/`). The real rules are in `../web3d-demo/battle.html`, and the prototype inventory is summarised below. These points from it matter for the port:
- **Randomness:** `Math.random` is used in 7 rules (damage roll, execute, frost, cooldown jitter, initial retarget and cooldown phase, sweep goal). Map, Red army, formation and skills use seeded mulberry32.
- **Timestep:** a variable frame `dt`, and the game-speed keys scale it. Only `demo.sim` uses a fixed 1/30 s.
- **Update order:** comes from the order models finish loading. It affects movement, separation and tie-breaks.
- **Swing timing:** melee swing time comes from the GLTF animation clip length.
- **Yuka:** used only for archer flee direction. Separation, A* and string-pulling are hand-written.
- **Obstacles:** rocks, trees and ponds are continuous circles, with a 0.5 m nav grid built from them.

## Goals / Non-Goals

**Goals:**
- One Go implementation of the rules, bit-identical between WASM and native.
- The prototype's feel kept: same numbers, same behaviours.
- Fast enough to run a full battle on the server in well under 50 ms.

**Non-Goals:**
- Reproducing the prototype's exact outcomes. It isn't deterministic itself, so we can't compare against it.
- Optimising for large armies (still 5 vs 5).

## Decisions

**Fixed-point everywhere (Q48.16 `sim.Fx`).**
- Positions, speeds, health and damage are all `Fx`.
- Angles are only needed for flank cones and facing. We keep facing as a unit vector and test the cone with a dot product against cos 60° = 0.5, so the sim needs no trig at all.
- Blink, kite and formation candidate directions come from fixed tables of unit vectors.
- *Alternative:* float64 avoiding FMA and trig. Rejected because one compiler change could silently break old replays.

**20 Hz fixed tick, rendering interpolates.** The prototype's 30 Hz (in `demo.sim`) was arbitrary. 20 Hz halves the cost of server verification, and units move at most 0.8 m per tick even when dashing at 16 m/s. Durations become tick counts: 0.3 s stagger = 6 ticks, 1.3 s cooldown = 26 ticks. The web interpolates between the last two snapshots.

**One seeded RNG stream per battle (SplitMix64), drawn in a fixed order.**
- Map, Red army, formation and skills each get a derived sub-seed: `mix(seed, tag)`. Adding a new map feature then doesn't shift the combat rolls.
- Combat draws come from a separate combat stream, in unit order.
- *Alternative:* a stream per unit. Rejected as unnecessary at 10 units.

**Deterministic update order.**
- Units are stored in a slice ordered Blue then Red, each in input order.
- Each step loops over it by index, separation goes pair by pair in index order, and ties always go to the lower index.
- No Go maps are iterated in logic.

**Animation-independent timing.** Melee swing time becomes a data value per class: `min(cooldown × 0.9, 0.9 s)`. The prototype's clip length is about 0.9 s. The renderer stretches its animation to fit, not the other way round.

**Replacing Yuka's FleeBehavior.** The escape direction is the normalised sum of `(self − threat) / distance` over threats within 3.5 m. That's equivalent in effect, and integer-only. The fallback when blocked is the same table of offset directions.

**Pathfinding.**
- The prototype's 8-connected A* without corner cutting, with octile heuristic, wet cost ×10/7 and string-pulling, ported to integer costs (straight 10, diagonal 14, wet ×10/7, rounded).
- The open set is a binary heap, ordered by (f, h, cell index) for stable ties.
- `fzipp/astar` was considered but rejected: we need tie-break control and water cost, and the code is small.

**Data, not code.** `STATS`, `MATCHUP`, `SKILLS` and terrain constants live in one `sim/data.go` table set. The proposal keeps the prototype numbers.

**Line of fire uses the sight test (deliberate change from the prototype).** The prototype checked shots against walkable cells, including the target's own cell. A unit standing in a rock's padded no-walk cells could then never be shot, and on seeds 90 and 657 two archers walked into each other until the tick limit. Shots now use the same test as sight: rocks and trees block, padding and water don't.

**Draw rules (deliberate change from the prototype).** A mutual wipe, or reaching the tick limit (3 minutes, 3,600 ticks), is a draw. The prototype calls a mutual wipe a Red win and has no limit. A server-checked game needs a guaranteed end.

**Wire formats.**
- `BattleInput`: version, seed, per-team units (class, skills, x/z as raw `Fx` int64, or a formation preset), and Red skill count. Positions are raw fixed-point so there's no float rounding on the wire.
- `BattleOutcome`: winner (Blue, Red or draw), ticks, survivors (index, health raw), and checksum.
- The checksum is FNV-1a over the final state, as in the scaffold.

**WASM API.**
- `create(inputBytes)`: protobuf-encoded `BattleInput`, so the browser and the server parse the same bytes.
- `step(n)` → a compact `Int32Array` snapshot: per unit x, z, facing, health, state; plus projectiles and events (hits, skills, deaths) for floating numbers.
- `outcome()`.
- The worker drives it and posts snapshots to the main thread.

**How determinism is tested.**
- **Golden replays:** about 20 hand-picked inputs (all formations, every skill, water-heavy maps, a mutual wipe, the tick limit), with recorded outcomes in `sim/testdata/golden.json`.
- **Native vs WASM:** `pnpm test:go` runs them natively, and `pnpm test:wasm` runs the same test under `go_js_wasm_exec`. Both must match the file.
- **Regenerating:** a `-update` flag rewrites the file, and a test forbids changing it without bumping `sim.Version`, by storing the version in the file.
- **Fuzzing:** a fuzz test plays random valid inputs natively twice and checks the outcomes are identical, which catches map-order and hidden state.
- **Lint:** a test fails if `sim/` imports `math`, `math/rand`, `time` or `sync`.

## Risks / Trade-offs

- [Fixed-point overflow in squared distances] → Coordinates are bounded at ±14 m, so a squared distance ≤ 2 × (28·65536)² ≈ 6.7e12 fits in int64. Add a debug assertion in tests.
- [Feel drifts from the prototype: 20 Hz, no trig, integer A*] → Side-by-side check in the debug viewer against the prototype on the same seed (visual only), and tune constants in data.
- [WASM binary size, about 2–3 MB] → Acceptable for now. TinyGo can be revisited later, but must pass the same golden test.
- [Snapshot traffic between worker and main thread] → An `Int32Array` with transfer stays tiny at 10 units.

## Migration Plan

The toy sim and `run` API are replaced in place. `sim.Version` goes to `0.2.0`. There's no stored data to migrate.

## Open Questions

- Exact tick limit (3 min proposed): a balance call that can change later with a version bump.
