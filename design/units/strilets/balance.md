# Balance: strilets

1000 battles per row, fixed seeds (sim 0.7.0+05959fc2). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× strilets vs 5× warrior | 66.6% | 25–75% | pass |
| 5× strilets vs 5× archer | 40.0% | 25–75% | pass |
| 5× strilets vs 5× mage | 70.8% | 25–75% | pass |
| 5× strilets vs 5× druzhynnyk | 98.6% | 25–75% | **FAIL** |
| 5× strilets vs 5× huslyar | 100.0% | 25–75% | **FAIL** |
| 5× strilets vs 5× huskarl | 80.4% | 25–75% | **FAIL** |
| 5× strilets vs 5× bowman | 58.5% | 25–75% | pass |
| 5× strilets vs 5× volva | 100.0% | 25–75% | **FAIL** |
| 5× strilets vs 5× skoutatos | 100.0% | 25–75% | **FAIL** |
| 5× strilets vs 5× toxotes | 99.8% | 25–75% | **FAIL** |
| 5× strilets vs 5× siphonarios | 100.0% | 25–75% | **FAIL** |
| standard army with one Archer swapped for strilets | 49.9% (base 49.9%, +0.0 pts) | ±10 pts | pass |
| level-1 pick Long shot | 47.7% (-3.8 pts vs its average) | ±10 pts | pass |
| level-1 pick Steppe shooting | 47.0% (-4.5 pts vs its average) | ±10 pts | pass |
| level-1 pick Bodkin points | 53.2% (+1.7 pts vs its average) | ±10 pts | pass |
| level-1 pick Arcing shot | 53.5% (+2.0 pts vs its average) | ±10 pts | pass |
| level-1 pick Quick draw | 56.3% (+4.8 pts vs its average) | ±10 pts | pass |

## Failed bands and why

Every failure is against a Dnipro-road **placeholder** that its own unit
session has not designed yet. The classic archer, with nearly the same
numbers, posts the same results against them (`pnpm balance:unit archer`:
druzhynnyk 99%, huslyar 100%, völva 100%, skoutatos 100%, toxotes 100%,
siphonarios 100%, huskarl 72%). Those melee placeholders walk into arrows on
line/hold temperaments without being tuned, and the huslyar is meant to be
weak alone (roster plan). Re-run this check once those units are balanced,
before the Rus faction is made playable.

Against the finished units (warrior, archer, mage) and the bowman, every band
passes, as do the swap and all five skills.

## Tuning rounds

Starting point: the placeholder (roster plan numbers, falls back below 30%):
2.4% vs archer, 3.8% vs bowman. Test runs showed the fall-back is the cause:
a ranged unit that retreats at low health stops shooting. Dropped with the
user's approval at the design checkpoint.

| Round | Change | Why | Warrior | Archer | Mage | Swap | Skills spread |
|---|---|---|---|---|---|---|---|
| 1 | approved design: no fall-back, 3.8 m/s, Long shot +20%, Steppe shooting +10% speed, Arcing shot +10% dmg, Quick draw ×0.84 | design | 71.6% | 37.3% | 71.0% | +5.6 | −5.1..+6.5 |
| 2 | speed 3.8 → 3.6 | too strong against warriors for "melee beats ranged" | 64.2% | 34.0% | 70.3% | +1.0 | −5.5..+6.3 |
| 3 | HP 55 → 60 | archer duel should be close to even | 66.7% | 65.9% | 83.9% | +3.2 | −5.8..+7.2 |
| 4 | HP 60 → 57 | overshot: HP decides how many arrows a kill takes | 65.7% | 45.0% | 79.8% | +2.0 | −5.6..+6.6 |
| 5 | HP 57 → 56 | mage over the band | 64.8% | 39.1% | 74.8% | +1.0 | −5.7..+6.7 |
| 6 | range 9 → 8.5 m, HP 56 → 57 | mage 76.4% at the full run; less reach against fragile casters, HP back for the duel | 66.6% | 40.0% | 70.8% | +0.0 | −4.5..+4.8 |

Rounds 1–5 at 500 battles per row (`-n 1000`), round 6 and the table above at the default 1000.
Final: 57 HP, 9–14 every 1.7 s, 8.5 m, 3.6 m/s, sight 11; line, kite, keeps 7 m,
target nearest, no fall-back.
