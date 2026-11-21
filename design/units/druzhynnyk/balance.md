# Balance: druzhynnyk

> Recorded before the Classic faction was retired (November 2026, OpenSpec
> `retire-classic`): rows and notes about the warrior, archer and mage are
> history; `pnpm balance:unit <name>` measures against today's roster.

1000 battles per row, fixed seeds (sim 0.7.0+602aad09). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× druzhynnyk vs 5× warrior | 29.0% | 25–75% | pass |
| 5× druzhynnyk vs 5× archer | 54.0% | 25–75% | pass |
| 5× druzhynnyk vs 5× mage | 29.8% | 25–75% | pass |
| 5× druzhynnyk vs 5× strilets | 42.5% | 25–75% | pass |
| 5× druzhynnyk vs 5× huslyar | 100.0% | 25–75% | **FAIL** |
| 5× druzhynnyk vs 5× huskarl | 58.3% | 25–75% | pass |
| 5× druzhynnyk vs 5× bowman | 60.7% | 25–75% | pass |
| 5× druzhynnyk vs 5× volva | 94.8% | 25–75% | **FAIL** |
| 5× druzhynnyk vs 5× skoutatos | 60.7% | 25–75% | pass |
| 5× druzhynnyk vs 5× toxotes | 93.2% | 25–75% | **FAIL** |
| 5× druzhynnyk vs 5× siphonarios | 60.9% | 25–75% | pass |
| standard army with one Warrior swapped for druzhynnyk | 41.9% (base 49.9%, -8.0 pts) | ±10 pts | pass |
| level-1 pick Shield wall | 49.6% (+8.3 pts vs its average) | ±10 pts | pass |
| level-1 pick Prince's guard | 35.5% (-5.8 pts vs its average) | ±10 pts | pass |
| level-1 pick Braced spear | 36.5% (-4.8 pts vs its average) | ±10 pts | pass |
| level-1 pick Heavy mail | 43.5% (+2.2 pts vs its average) | ±10 pts | pass |
| level-1 pick Axe and sword | 41.4% (+0.1 pts vs its average) | ±10 pts | pass |

## Notes

**Tuning.** One change from the roster plan: the leash went from 8 m to 14 m.
With the plan's 8 m (placeholder run, same seeds), the druzhynnyk lost almost
every head-to-head against ranged units: archer 0.9%, strilets 1.7%, bowman
1.6%, mage 2.1%, siphonarios 6.3%. In the Spread formation the wings start more
than 8 m from their allies' centre, so against kiters they kept walking back to
the centre and never reached the archers. Leash sweep against archer and
strilets (300 battles each): 10 m 15% and 10%, 12 m 44% and 30%, 14 m 56% and
41%, 16 m 65% and 42%, none 66% and 43%. 14 m keeps it a guard that doesn't
run off across the map, and lets its ×1.4 against ranged work.

**Rows still out of band** (all against placeholders that aren't done yet, and
the classic warrior lands in the same place):

| Opponent | Druzhynnyk | Warrior (500 battles) | Why |
|---|---|---|---|
| huslyar | 100% | 100% | weak alone by design (roster plan, review answer 2) |
| völva | 94.8% | 96% | placeholder; its curses and fear aren't tuned yet |
| toxotes | 93.2% | 72% | placeholder; a toxotes army has no melee to stay near and falls back at 30% health |

Lowering the druzhynnyk to fit these would push its warrior and mage rows
(29%, near the 25% floor) out of band instead. Re-check once those units are
balanced in their own sessions.

**Against the warrior (29%)**: the lab's warriors each carry one of the
warrior's strong skills (Swift, Strength, Dash, Executioner, Toughness), while
two druzhynnyk skills (Prince's guard, Braced spear) do nothing or little
in an all-melee mirror. Their stats are the same.
