# Tasks

## 1. Lab

- [x] 1.1 Run simulation in `tools/balance` (policies, Red modes, health curves); verify against the measured tables in the proposal

## 2. Sim and wire

- [x] 2.1 `Input.RedHealthPct` (validation 0–100; scales Red's max and current health at spawn); proto fields and `pb/convert`; verify a sim test (85% exactly, invalid values rejected), existing goldens unchanged plus one new 85% golden case, and `pnpm lint:proto`

## 3. Server

- [x] 3.1 Migration 0003 `battles.red_health_pct`; `run.RedHealthFor(match)`; StartBattle stores and returns it; GetRun's CurrentBattle carries it; VerifyBattle rebuilds with it; verify server tests (match 1 gets 85%, match 4 gets 0, and a claim played at full health for a match-1 battle is not fair)

## 4. Web

- [x] 4.1 The client input uses the issued `red_health_pct`; the run bar shows Red's strength while below 100%; verify typecheck and a browser match verified fair

## 5. Report and docs

- [x] 5.1 The full report includes run experiments; the findings gain a runs section; concept §10.2; verify `pnpm test` and a regenerated report

## Workflow follow-up

- Archive the change once merged (auto-merge after green CI).
