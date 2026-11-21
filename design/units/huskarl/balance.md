# Balance: huskarl

> Recorded before the Classic faction was retired (November 2026, OpenSpec
> `retire-classic`): rows and notes about the warrior, archer and mage are
> history; `pnpm balance:unit <name>` measures against today's roster.

1000 battles per row, fixed seeds (sim 0.8.0+60bad52c). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× huskarl vs 5× warrior | 53.1% | 25–75% | pass |
| 5× huskarl vs 5× archer | 65.8% | 25–75% | pass |
| 5× huskarl vs 5× mage | 28.5% | 25–75% | pass |
| 5× huskarl vs 5× druzhynnyk | 70.3% | 25–75% | pass |
| 5× huskarl vs 5× strilets | 41.1% | 25–75% | pass |
| 5× huskarl vs 5× huslyar | 100.0% | 25–75% | **FAIL** |
| 5× huskarl vs 5× bowman | 67.5% | 25–75% | pass |
| 5× huskarl vs 5× volva | 91.8% | 25–75% | **FAIL** |
| 5× huskarl vs 5× skoutatos | 64.5% | 25–75% | pass |
| 5× huskarl vs 5× toxotes | 83.7% | 25–75% | **FAIL** |
| 5× huskarl vs 5× siphonarios | 57.0% | 25–75% | pass |
| standard army with one Warrior swapped for huskarl | 63.7% (base 65.6%, -1.9 pts) | ±10 pts | pass |
| level-1 pick Dane axe sweep | 63.9% (+0.3 pts vs its average) | ±10 pts | pass |
| level-1 pick Charge | 64.0% (+0.4 pts vs its average) | ±10 pts | pass |
| level-1 pick Battle fury | 63.4% (-0.2 pts vs its average) | ±10 pts | pass |
| level-1 pick Shield-breaker | 64.3% (+0.7 pts vs its average) | ±10 pts | pass |
| level-1 pick Executioner | 62.2% (-1.4 pts vs its average) | ±10 pts | pass |

## Notes

**Tuning** (from the roster plan's 85 HP, 3.3 m/s; damage, cooldown, reach and
skills unchanged):

| Round | Change | Warrior | Mage | Strilets | Druzhynnyk | Why |
|---|---|---|---|---|---|---|
| 0 | plan numbers | 21.4% | 4.2% | 19.4% | 43.5% | kited and out-traded: 85 HP vs the warrior's 100 |
| 1 | HP 100, speed 3.6 | 67.2% | 48.9% | 48.5% | 83.2% | overshot: archer, bowman, skoutatos, druzhynnyk above 75% |
| 2 | HP 85, speed 3.6 | 27.1% | 8.0% | 28.0% | 44.5% | speed fixes the kiters; the mage still burns it down |
| 3 | HP 92 | 47.1% | 22.7% | 38.1% | 64.8% | mage just under the band |
| 4 | HP 94 (kept) | 53.1% | 28.5% | 41.1% | 70.3% | all bands met except the three below |

Final: 94 HP (+11%), 3.6 m/s (+9%). Still lighter than the druzhynnyk (100 HP,
120 with Heavy mail) and still the hardest hitter of the melee, so "big blows,
less armour" holds; it beats the druzhynnyk (70%, its brief says the huskarl
should) and loses to the mage (28.5%: specials beat massed melee).

**Rows still out of band** (none is about the huskarl's numbers):

| Opponent | Huskarl | Why |
|---|---|---|
| huslyar | 100% | weak alone by design (roster plan, review answer 2); every unit lands here |
| völva | 91.8% | placeholder in its own session; druzhynnyk saw 94.8% |
| toxotes | 83.7% | placeholder; druzhynnyk saw 93.2% |

Re-check once those units are done. Shield-breaker's curse lasts 4 s (the
plan's number). Executioner uses the sim's fixed 40% health threshold (the
plan said 30%; the threshold isn't a per-effect number).
