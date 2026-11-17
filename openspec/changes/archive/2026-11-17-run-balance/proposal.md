# Proposal

## Why

The balance lab can now simulate whole runs (veterans, skill picks, Red scaled to the veterans' skills). With sim 0.4.0 it shows two things:
- **No snowballing:** the win rate per match stays flat (about 55–64%) from match 1 to 7+, because Red's skill matching cancels the veterans' growth. Counter-picking Red would only shorten runs by about 12%; that's the concept's open question, answered **no**.
- **Runs end almost at once:** 1.3 wins on average, median 1, and only 11% reach match 5. A run is a chain you keep while winning, so a first match that is a coin flip makes the loop feel like single matches.

## What Changes

- **Early matches are easier.** Red's units start at 85% health in match 1, 90% in match 2, 95% in match 3, then 100%. Measured over 1,000 runs (2W 2A 1M, best picks):

  | | Today | With the ramp |
  |---|---|---|
  | Win rate in matches 1 / 2 / 3 | 54 / 55 / 63% | 88 / 82 / 72% |
  | Win rate in later matches | about 63% | about 64% |
  | Average wins | 1.33 | 3.05 |
  | Reach match 5 | 11.5% | 33.4% |

- **New input field.** `BattleInput.red_health_pct` (0 = full, as before). The sim scales Red's maximum and starting health.
  - Existing inputs play exactly as before, so `sim.Version` stays 0.4.0 and the golden replays are unchanged; one golden case is added with 85%.
- **The server sets it.** Red's health comes from the run's match number and is stored per battle (migration 0003). Verification always uses the stored value, never the client's.
- **The client** passes the issued value into its input, and the run bar shows "Red at 85%" while the ramp applies.
- **The lab** gains run simulation (`Runs`, `RunCurves`) in the full report, plus the findings for runs.

## Out of scope

- Red counter-picking (measured, rejected; see the findings).
- Changing veteran rules or skill counts.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `runs`: adds "Early matches are easier".
- `battle-replay`: the battle input includes Red's health percentage.
- `battle-sessions`: starting a battle also issues Red's health percentage.

## Impact

- `sim/army.go`, `sim/battle.go`: `Input.RedHealthPct`, validation, the health scaling.
- `proto`: `BattleInput.red_health_pct`, `StartBattleResponse.red_health_pct`, `CurrentBattle.red_health_pct`; plus `pb/convert`.
- `server`: migration 0003 (`battles.red_health_pct`), `run.RedHealthFor(match)`, battle start/lock/verify.
- `web`: the input carries it; the run bar shows it.
- `tools/balance`: run experiments in the report; `docs/balance-findings.md`; concept §10.2.
