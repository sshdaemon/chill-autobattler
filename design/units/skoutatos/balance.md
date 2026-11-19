# Balance: skoutatos

1000 battles per row, fixed seeds (sim 0.8.0+28d5643d). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× skoutatos vs 5× warrior | 64.1% | 25–75% | pass |
| 5× skoutatos vs 5× archer | 51.2% | 25–75% | pass |
| 5× skoutatos vs 5× mage | 41.3% | 25–75% | pass |
| 5× skoutatos vs 5× druzhynnyk | 64.8% | 25–75% | pass |
| 5× skoutatos vs 5× strilets | 44.0% | 25–75% | pass |
| 5× skoutatos vs 5× huslyar | 100.0% | 25–75% | **FAIL** |
| 5× skoutatos vs 5× huskarl | 40.8% | 25–75% | pass |
| 5× skoutatos vs 5× bowman | 64.2% | 25–75% | pass |
| 5× skoutatos vs 5× volva | 57.3% | 25–75% | pass |
| 5× skoutatos vs 5× toxotes | 83.2% | 25–75% | **FAIL** |
| 5× skoutatos vs 5× siphonarios | 90.5% | 25–75% | **FAIL** |
| standard army with one Warrior swapped for skoutatos | 76.2% (base 71.0%, +5.2 pts) | ±10 pts | pass |
| level-1 pick Shield wall | 83.7% (+7.0 pts vs its average) | ±10 pts | pass |
| level-1 pick Phalanx | 72.6% (-4.1 pts vs its average) | ±10 pts | pass |
| level-1 pick Kontarion | 73.8% (-2.9 pts vs its average) | ±10 pts | pass |
| level-1 pick Drill | 74.6% (-2.1 pts vs its average) | ±10 pts | pass |
| level-1 pick Klibanion | 78.8% (+2.1 pts vs its average) | ±10 pts | pass |

## Notes

**Tuning** (from the roster plan's numbers; HP, damage, cooldown, reach and
speed unchanged):

| Round | Change | Warrior | Archer | Mage | Strilets | Huskarl | Army swap | Why |
|---|---|---|---|---|---|---|---|---|
| 0 | plan: hold, trigger 9 m, leash 5 m | 38.9% | 0.0% | 0.0% | 0.0% | 35.5% | −10.6 pts | it waits while archers shoot it from beyond 9 m, and the leash pulls it back to its centre |
| 1 | trigger 12 m, no leash (approved at the design checkpoint) | 61.9% | 52.1% | 45.7% | 44.5% | 41.2% | +5.7 pts | still waits for the enemy, but wakes before kiters are in range and keeps going |
| 2 | Shield wall −40% → −35% | 62.3% | 50.5% | 40.8% | 43.9% | 40.3% | +5.2 pts | Shield wall was the strongest pick (+8.1 pts) |
| 3 | Phalanx −8% within 2.5 m → −10% within 3 m | 64.1% | 51.2% | 41.3% | 44.0% | 40.8% | +5.2 pts | in the lab's spread line skoutatoi rarely stand within 2.5 m: five skoutatoi all with Phalanx beat 5 warriors 15% of the time (6% with the bonus never applying), 21% with this |

Variants tried at round 1 (400 battles each): *line* without a leash
(archer 60%, bowman 76%, toxotes 87%), *line* with a 12 m leash (archer 38%),
*charge* (archer 60%, bowman 78%: too strong, and not a wall).

The army-swap and level-1 rows put one skoutatos into the standard army, so
Phalanx (which needs other skoutatoi) can't show there; its -4.1 pts is the
lone-skoutatos case.

**Rows still out of band** (none is about the skoutatos's numbers):

| Opponent | Skoutatos | Why |
|---|---|---|
| huslyar | 100% | weak alone by design (roster plan, review answer 2); every unit lands here |
| toxotes | 83.2% | placeholder in its own session; the huskarl saw 83.7%, the druzhynnyk 93.2% |
| siphonarios | 90.5% | placeholder in its own session; its plan makes it the answer to massed melee |

Re-check once those two Roman units are done. The huskarl beats it (41%, as
both briefs say), and it beats the druzhynnyk (65%) and the warrior (64%):
an armoured line that stands its ground against plain infantry.
