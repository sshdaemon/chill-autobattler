# Balance: siphonarios

The siphonarios is a special: by the roster triangle he beats massed melee and
loses to ranged units and casters that outrange him. The bands agreed at the
design checkpoint (2026-11-19), the same as the völva's:

| Band | Why |
|---|---|
| 5× siphonarios vs 5× each **melee** unit: 25–75% | he must answer melee without hard-countering it |
| 5× siphonarios vs 5× ranged units and casters: no band | they are meant to beat him |
| The standard army (2 warriors, 2 archers, 1 mage, 3-2) with its **mage** swapped for him: within ±10 pts | he takes the special slot |
| Each ability as his level-1 pick, in that army: within ±10 pts of his average | no trap or must-pick |

The bands come from a local harness (`harness.go.txt` here: copy it to
`tools/balance/zz_siphonarios_bands_test.go`, run
`go test ./tools/balance -run TestSiphonariosBands -count=1 -v`, delete it). No
shared balance code changed. `pnpm balance:unit siphonarios` (below) still
exits non-zero on its ranged rows, on purpose.

## Result

2000 battles per army row, 1000 per head-to-head side, fixed seeds (sim 0.8.0+1b448215).
| Check | Result | Band | |
|---|---|---|---|
| 5× siphonarios vs 5× warrior (melee) | 68.0% | 25–75% | pass |
| 5× siphonarios vs 5× archer | 2.3% | none (ranged and casters may beat it) | pass |
| 5× siphonarios vs 5× mage | 33.0% | none (ranged and casters may beat it) | pass |
| 5× siphonarios vs 5× druzhynnyk (melee) | 48.3% | 25–75% | pass |
| 5× siphonarios vs 5× strilets | 1.5% | none (ranged and casters may beat it) | pass |
| 5× siphonarios vs 5× huslyar | 100.0% | none (ranged and casters may beat it) | pass |
| 5× siphonarios vs 5× huskarl (melee) | 28.1% | 25–75% | pass |
| 5× siphonarios vs 5× bowman | 3.1% | none (ranged and casters may beat it) | pass |
| 5× siphonarios vs 5× volva | 32.6% | none (ranged and casters may beat it) | pass |
| 5× siphonarios vs 5× skoutatos (melee) | 100.0% | 25–75% | **FAIL** |
| 5× siphonarios vs 5× toxotes | 57.8% | none (ranged and casters may beat it) | pass |
| standard army with its Mage swapped for siphonarios | 67.8% (base 71.9%, -4.1 pts) | ±10 pts | pass |
| level-1 pick Burning ground | 73.5% (+5.8 pts vs its average) | ±10 pts | pass |
| level-1 pick Fire pots | 73.3% (+5.6 pts vs its average) | ±10 pts | pass |
| level-1 pick Pressure | 65.7% (-2.0 pts vs its average) | ±10 pts | pass |
| level-1 pick Padded coat | 64.2% (-3.5 pts vs its average) | ±10 pts | pass |
| level-1 pick Fear of fire | 62.0% (-5.8 pts vs its average) | ±10 pts | pass |
| (info) standard army with an Archer swapped for siphonarios | 72.2% (+0.3 pts) | none | pass |

One melee row fails, against a **placeholder**: the skoutatos (holds with a
5 m leash) never walks into the jet for long, so any 5 m unit with range wins
100% (the strilets and völva sessions saw the same). Its own session is
redesigning it; re-run this row after it merges.

## Tuning rounds

Head-to-head vs warrior, druzhynnyk and huskarl (the finished melee units),
and the mage swap. 300 battles per row while tuning, 2000 per army row and 1000
per head-to-head side for the final table.

| Round | Change | vs warrior | vs druzhynnyk | vs huskarl | Mage swap |
|---|---|---|---|---|---|
| placeholder | 60 HP, 8–12, keep 3, fall back 40% | 54 | 45 | 40 | −17.7 |
| design | 70 HP, 9–13, keep 3, no fall-back | 97 | 97 | 88 | −4.7 |
| 1 | keep 3 → 0 (five jets backing off kite any melee to death; holding his ground he still jets the pack that reaches the line) | 57 | 39 | 21 | −3.8 |
| 2 | HP 70 → 76 (the huskarl's charge reached him too fast) | 71 | 52 | 29 | −2.3 |
| 3 | burning ground 25 → 22% per second, fear of fire ×0.65 → ×0.6 (the strongest and the weakest pick) | 68 | 48 | 28 | −4.1 |

Changes from the approved design: no keep distance (the plan's 3 m), HP 76
(+9%), burning ground 22%, fear of fire ×0.6. The huskarl stays his worst melee
matchup: it charges 10 m straight through the jet and finishes the wounded.

## `pnpm balance:unit siphonarios`

1000 battles per row, fixed seeds (sim 0.8.0+1b448215). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× siphonarios vs 5× warrior | 68.0% | 25–75% | pass |
| 5× siphonarios vs 5× archer | 2.3% | 25–75% | **FAIL** |
| 5× siphonarios vs 5× mage | 33.0% | 25–75% | pass |
| 5× siphonarios vs 5× druzhynnyk | 48.3% | 25–75% | pass |
| 5× siphonarios vs 5× strilets | 1.5% | 25–75% | **FAIL** |
| 5× siphonarios vs 5× huslyar | 100.0% | 25–75% | **FAIL** |
| 5× siphonarios vs 5× huskarl | 28.1% | 25–75% | pass |
| 5× siphonarios vs 5× bowman | 3.1% | 25–75% | **FAIL** |
| 5× siphonarios vs 5× volva | 32.6% | 25–75% | pass |
| 5× siphonarios vs 5× skoutatos | 100.0% | 25–75% | **FAIL** |
| 5× siphonarios vs 5× toxotes | 57.8% | 25–75% | pass |
| standard army with one Warrior swapped for siphonarios | 72.4% (base 71.0%, +1.4 pts) | ±10 pts | pass |
| level-1 pick Burning ground | 73.9% (+1.7 pts vs its average) | ±10 pts | pass |
| level-1 pick Fire pots | 78.8% (+6.6 pts vs its average) | ±10 pts | pass |
| level-1 pick Pressure | 70.3% (-1.9 pts vs its average) | ±10 pts | pass |
| level-1 pick Padded coat | 71.7% (-0.5 pts vs its average) | ±10 pts | pass |
| level-1 pick Fear of fire | 66.5% (-5.7 pts vs its average) | ±10 pts | pass |
