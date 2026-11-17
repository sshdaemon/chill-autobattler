# Proposal

## Why

The balance lab (#16, `docs/balance-findings.md`) measured sim 0.3.0 and found four problems:
- **The counters were absolute locks.** Warriors beat archers 100–0, archers beat mages 100–0, and mages beat nobody.
- **Warriors dominated.** 5W won 83% against random Red; every mix without warriors won ≤ 24%.
- **Instant kills were overpowered.** Headshot and Executioner each added about 17 win-rate points.
- **Five skills did almost nothing** (under 2 points).

The game's core decisions (army mix, formation, skill picks) only matter if none of them is a lock.

## What Changes

- **Archers kite properly.** The trace showed kiting archers losing ground: they turned their back (the sharp-turn penalty applied) and only started retreating at 3.5 m. Now they:
  - back away facing their attackers at full speed;
  - start at 4.5 m;
  - retreat for up to 2 s, then rest for 1.2 s (was 1.2 s / 2 s).
- **Archer stats:** health 45 → 55, speed 3.4 → 3.6 m/s.
- **Matchups** (attacker → victim):

  | Matchup | Was | Now |
  |---|---|---|
  | Warrior → archer | 150% | 145% |
  | Warrior → mage | 100% | 95% |
  | Archer → warrior | 75% | 100% |
  | Archer → mage | 125% | 100% |
  | Mage → warrior | 125% | 135% |
  | Mage → archer | 75% | 100% |

- **Instant kills** (Executioner, Headshot) only work on victims below 40% health. Their chances go up to compensate: Executioner 25% → 35%, Headshot 15% → 25%.
- **Skill numbers:**

  | Skill | Was | Now |
  |---|---|---|
  | Toughness (warrior) | +15% | +10% |
  | Strength | +10% | +15% |
  | Swift | +25% speed | +40% speed, and attacks 20% faster |
  | Long shot | +10% | +25% |
  | Sharp arrows | +15% | +10% |
  | Arcing shot | — | also +10% damage |
  | Frost | 50% chance | always slows |
  | Power | +15% | +10% |
  | Blink cooldown | 15 s | 8 s |
  | Blast splash | 50% | 35% |

- `sim.Version` 0.3.0 → **0.4.0**, with regenerated golden replays. Old battles keep verifying with 0.3.0 only if that build is kept; there are no stored live battles yet.
- `docs/GAME_CONCEPT.md` is updated to the new numbers. `docs/balance-report.md` is regenerated, and `docs/balance-findings.md` gains before and after.

## Out of scope

- New classes or skills, Red counter-picking, and run-level balance (snowballing) — later passes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `battle-rules`: the archer-kiting numbers and backpedalling, the instant-kill health threshold, and the scenario numbers (Toughness 110, matchup 1.45, Blink 8 s).

## Impact

- `sim/data.go` (numbers), `sim/unit.go` (backpedal; Swift attack speed; Arcing shot damage), `sim/combat.go` (execute threshold).
- Sim tests updated to the new numbers; golden replays regenerated for 0.4.0.
- Docs: the concept, the balance report and the findings.
