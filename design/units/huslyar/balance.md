# Balance: huslyar

> Recorded before the Classic faction was retired (November 2026, OpenSpec
> `retire-classic`): rows and notes about the warrior, archer and mage are
> history; `pnpm balance:unit <name>` measures against today's roster.

The huslyar is a support unit, weak alone by design (brief and design
checkpoint, 2026-11-18). The standard new-unit bands measure a unit head to
head, five against five, so he fails them at 0% whatever his numbers. He is
measured inside an army instead, with these bands (agreed at the design
checkpoint):

| Band | Why |
|---|---|
| The standard army (2 warriors, 2 archers, 1 mage, 3-2) with its **mage** swapped for him: within ±10 pts of base | He takes the special slot, so the swap measures what he is worth there |
| His song (the trait) on vs off, same army: at least +10 pts | The song must be what he brings, not just a fifth body |
| Each ability as his level-1 pick: within ±10 pts of his average | No trap or must-pick song |
| 5× huslyar vs 5× each other unit: at most 25% | Weak alone stays true |

The army bands come from a local harness (`harness.go.txt` here: copy it to
`tools/balance/zz_huslyar_bands_test.go`, run
`go test ./tools/balance -run TestHuslyarBands -count=1 -v`, delete it). No
shared code changed. `pnpm balance:unit huslyar` still exits non-zero on its
head-to-head rows, on purpose.

## Army bands

2000 battles per army row, 500 per head-to-head side, fixed seeds (sim 0.7.0+06798761).

| Check | Result | Band | |
|---|---|---|---|
| standard army with its Mage swapped for huslyar | 41.8% (base 49.3%, -7.6 pts) | ±10 pts | pass |
| his song (trait) on vs off, same army | +17.9 pts (23.8% without it) | ≥ +10 pts | pass |
| level-1 pick Song of glory | 46.6% (+4.4 pts vs its average) | ±10 pts | pass |
| level-1 pick Tale of the bohatyrs | 47.3% (+5.1 pts vs its average) | ±10 pts | pass |
| level-1 pick Lament for the fallen | 38.2% (-4.0 pts vs its average) | ±10 pts | pass |
| level-1 pick Mocking song | 40.0% (-2.2 pts vs its average) | ±10 pts | pass |
| level-1 pick Loud voice | 39.0% (-3.2 pts vs its average) | ±10 pts | pass |
| 5× huslyar vs 5× warrior | 0.0% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× archer | 0.0% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× mage | 0.0% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× druzhynnyk | 0.0% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× strilets | 0.1% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× huskarl | 0.2% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× bowman | 0.0% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× volva | 0.2% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× skoutatos | 0.0% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× toxotes | 0.1% | ≤ 25% (weak alone) | pass |
| 5× huslyar vs 5× siphonarios | 0.0% | ≤ 25% (weak alone) | pass |
| (info) standard army with a Warrior swapped for huslyar | 51.9% (+2.6 pts) | none | pass |

## Tuning rounds

Starting point: the roster plan (60 HP, song +10% damage). The mage swap
gave 16% (−34 pts): the song was worth +3 pts, so he played as a weak fifth
body. A 4-unit army wins 0.7%, so a body that lives counts for a lot.

| Round | Change | Mage swap | Song on/off |
|---|---|---|---|
| design | 85 HP, song +20%, glory +20%, tale −20%, mocking −20% at 6 m | 29.8% (−19.5) | +12.2 |
| 1 | HP 85 → 105, song +20 → +25% | 41.2% (−8.7) | +17.1 |
| 2 | lament +30 → +38%, loud voice +50 → +65% (the two weakest picks) | 41.5% (−8.4) | +17.3 |
| 3 | HP 105 → 110 (the +30% limit) | 41.8% (−7.6) | +17.9 |

He is still the special slot's weakest pick against the classic mage (−7.6
pts), inside the band. More would mean going past +30% on HP or the song.
Swapped for a warrior instead, the army stays level (+1.6 to +2.6 pts): a
singer is worth about one fighter to an army, and nothing on his own.

## Standard bands (`pnpm balance:unit huslyar`)

1000 battles per row, fixed seeds (sim 0.7.0+06798761). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× huslyar vs 5× warrior | 0.0% | 25–75% | **FAIL** |
| 5× huslyar vs 5× archer | 0.0% | 25–75% | **FAIL** |
| 5× huslyar vs 5× mage | 0.0% | 25–75% | **FAIL** |
| 5× huslyar vs 5× druzhynnyk | 0.0% | 25–75% | **FAIL** |
| 5× huslyar vs 5× strilets | 0.0% | 25–75% | **FAIL** |
| 5× huslyar vs 5× huskarl | 0.1% | 25–75% | **FAIL** |
| 5× huslyar vs 5× bowman | 0.0% | 25–75% | **FAIL** |
| 5× huslyar vs 5× volva | 0.2% | 25–75% | **FAIL** |
| 5× huslyar vs 5× skoutatos | 0.0% | 25–75% | **FAIL** |
| 5× huslyar vs 5× toxotes | 0.4% | 25–75% | **FAIL** |
| 5× huslyar vs 5× siphonarios | 0.0% | 25–75% | **FAIL** |
| standard army with one Warrior swapped for huslyar | 51.5% (base 49.9%, +1.6 pts) | ±10 pts | pass |
| level-1 pick Song of glory | 57.3% (+7.0 pts vs its average) | ±10 pts | pass |
| level-1 pick Tale of the bohatyrs | 50.3% (+0.0 pts vs its average) | ±10 pts | pass |
| level-1 pick Lament for the fallen | 46.8% (-3.5 pts vs its average) | ±10 pts | pass |
| level-1 pick Mocking song | 47.0% (-3.3 pts vs its average) | ±10 pts | pass |
| level-1 pick Loud voice | 50.0% (-0.3 pts vs its average) | ±10 pts | pass |
