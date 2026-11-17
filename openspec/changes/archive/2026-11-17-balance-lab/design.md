# Design

## Context

`sim.Run(Input)` plays a whole battle deterministically. Red is generated from the map seed when `Input.Red` is empty, or given explicitly.

## Goals / Non-Goals

**Goals:** fast, reproducible, statistically meaningful numbers.

**Non-Goals:** UI, and simulating whole runs.

## Decisions

- **Reproducible:**
  - The seeds for each experiment are `base + i`, with a fixed base per experiment, so the report is identical on every run with the same `sim.Version`.
  - The combat seed is derived from the map seed (`seed * 7919 + 1`, as in the golden replays).
- **Parallel:** a worker pool with one goroutine per CPU. Goroutines are fine here: they live in the tool, never in `sim/`, and each battle is independent and deterministic.
- **Side bias:** both teams get the same explicit army and the same preset (it's mirrored by the zones), across all 7 presets and 3 mixes.
- **Matchups:** both teams are explicit (5×A against 5×B, `Spread` formation), so these are pure class numbers without Red's random army.
- **Compositions:** Blue is each of the 21 count mixes × 7 presets; Red is the game's seeded random army with no skills (match 1).
- **Skill impact:** the baseline is 2W 2A 1M in 3-2 against the seeded Red with no skills. The variant is the same, with the first unit of the skill's class learning it. Both run on the same seeds; the report gives win rate, Δ, and an approximate 95% interval.
- **Uncertainty:** the report prints the number of battles per row and a ±1.96·√(p(1−p)/n) margin, so small differences aren't over-read.
- **Output:** Markdown on stdout, with flags for the battle count (`-n`, default 2000) and the experiments to run.

**How determinism is preserved and tested:** the tool only reads `sim`, which is unchanged. A small test checks that two runs of the same experiment give identical results, which catches accidental shared state across goroutines.

## Risks / Trade-offs

- [The seeded Red mix differs from what a player will face later in a run] → Match 1 is the common case; run simulation is a later extension.
- [Run time] → About 70k battles at roughly 3 ms on 8 cores takes about 30 s.
