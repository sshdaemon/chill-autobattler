# Proposal

## Why

The concept's open balance questions (§14) can't be answered by playing a few matches:
- Are Executioner and Headshot too strong?
- Does one class or composition dominate?
- Is there a side or first-mover bias?
- Do runs snowball?

The sim is deterministic and fast (about 3 ms a battle), so thousands of seeded battles can answer these with numbers. Every later balance change should then come with before and after numbers.

## What Changes

- **A `balance` command** (`tools/balance`, a Go module in the workspace). It plays batches of seeded battles in parallel through the real `sim` and prints a Markdown report:
  - **Side bias:** identical armies on both sides, mirrored formations, Blue and Red win shares.
  - **Class matchups:** 5 of class A against 5 of class B.
  - **Compositions:** every 5-unit mix of warriors, archers and mages, each with the 7 formations, against the game's seeded Red army.
  - **Skill impact:** for each of the 15 skills, the change in win rate when one unit learns it, paired on the same seeds and armies.
  - **Battle length** and the draw rate.
- **`pnpm balance`** runs it, and its output is committed as `docs/balance-report.md` (the baseline for later changes).

## Out of scope

- Changing any balance numbers. That comes next, decided from this report.
- Simulating whole runs with veterans and Red scaling (a later extension).

## Capabilities

### New Capabilities

None. This is tooling only, so `skip_specs` is set.

### Modified Capabilities

None.

## Impact

- New `tools/balance` module (added to `go.work`), depending only on `chill/sim`.
- Root `package.json`: a `balance` script.
- `docs/balance-report.md` (generated) and `docs/balance-findings.md` (written: what the numbers mean).
