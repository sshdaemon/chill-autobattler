# Tasks

## 1. Rules

- [x] 1.1 `Fx.Mul` rounds toward zero
- [x] 1.2 Pushes summed and applied together
- [x] 1.3 Coin flip per tick for which side acts first
- [x] 1.4 `Turn.For(team)` in kiting, stepping away and escape; `TurnToward` turns toward +x when the target is straight behind
- [x] 1.5 `ToCell`: an edge point goes to the cell nearer the centre line
- [x] 1.6 `RulesVersion` 0.14.0

## 2. Tests

- [x] 2.1 `TestMirrorImageBattles`, `TestToCellMirrors`, `TestMulIsOdd` (fail on the old code)
- [x] 2.2 Goldens re-recorded; mutual-wipe seeds 4171 and 4436; `TestMovementSmooth` still passes
- [x] 2.3 `pnpm test`

## 3. Measure

- [x] 3.1 `-exp maps`: no cell more than 3 standard errors from 50%; ablation of each fix
- [x] 3.2 Side bias over random maps and the plan matrix unchanged; findings in `docs/balance-findings.md`
