# Balance: toxotes

> Recorded before the Classic faction was retired (November 2026, OpenSpec
> `retire-classic`): rows and notes about the warrior, archer and mage are
> history; `pnpm balance:unit <name>` measures against today's roster.

1000 battles per row, fixed seeds (sim 0.8.0+d12d9938). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× toxotes vs 5× warrior | 48.6% | 25–75% | pass |
| 5× toxotes vs 5× archer | 44.0% | 25–75% | pass |
| 5× toxotes vs 5× mage | 62.8% | 25–75% | pass |
| 5× toxotes vs 5× druzhynnyk | 37.9% | 25–75% | pass |
| 5× toxotes vs 5× strilets | 53.5% | 25–75% | pass |
| 5× toxotes vs 5× huslyar | 100.0% | 25–75% | **FAIL** |
| 5× toxotes vs 5× huskarl | 34.9% | 25–75% | pass |
| 5× toxotes vs 5× bowman | 39.9% | 25–75% | pass |
| 5× toxotes vs 5× volva | 92.7% | 25–75% | **FAIL** |
| 5× toxotes vs 5× skoutatos | 100.0% | 25–75% | **FAIL** |
| 5× toxotes vs 5× siphonarios | 99.9% | 25–75% | **FAIL** |
| standard army with one Archer swapped for toxotes | 71.9% (base 71.0%, +0.9 pts) | ±10 pts | pass |
| level-1 pick Solenarion | 66.1% (-6.2 pts vs its average) | ±10 pts | pass |
| level-1 pick Volley | 71.3% (-1.0 pts vs its average) | ±10 pts | pass |
| level-1 pick Sharp arrows | 76.0% (+3.7 pts vs its average) | ±10 pts | pass |
| level-1 pick Arcing shot | 70.5% (-1.8 pts vs its average) | ±10 pts | pass |
| level-1 pick Quick draw | 77.5% (+5.2 pts vs its average) | ±10 pts | pass |
## Failed bands and why

- **Völva (92.7%):** a 7 m caster that holds until an enemy comes within 10 m;
  any 9 m archer outranges her (her own balance.md: archer 94%, strilets 92%,
  "ranged and casters may beat her"). This is the roster's triangle (ranged
  beats specials through range). At 8 m range the toxotes still won 82% and
  lost 6–9 pts in every other row, so the range stays at 9 m.
- **Huslyar (100%):** weak alone by design (roster plan).
- **Skoutatos, siphonarios:** placeholders still being designed in parallel
  sessions; re-run this check before the Romans are made playable.

Against every finished unit (warrior, archer, mage, druzhynnyk, strilets,
huskarl, bowman), the swap and all five skills, every band passes.

## Tuning rounds

Starting point: the placeholder (roster plan: target focus, keeps 8 m, stays
near the front line, falls back below 30%): 0.2% vs archer, 0.2% vs strilets,
0.2% vs bowman, 25.8% vs warrior, swap −10.8 pts, at 250 battles per row.
Two causes, found in test runs: the fall-back (as for the strilets and bowman),
and the plain focus rule, which joins a shared target anywhere on the field,
so toxotai walk across into enemy fire. Without the fall-back and on the
nearest rule: archer 27%, warrior 18.5%. On a focus rule limited to enemies
within reach (range + 2 m): archer 37%, warrior 45%. The user approved that
rule (shared PR "Add a focus-within-reach target rule"), no fall-back, keeps 7 m.

| Round | Change | Why | Warrior | Archer | Mage | Volva | Swap | Skills spread |
|---|---|---|---|---|---|---|---|---|
| 1 | approved design: 57 HP, 9 m, 3.4 m/s, focusReach, keeps 7 m, no fall-back; Solenarion +30% range −10% dmg, Quick draw ×0.80 | design | 47.3% | 40.7% | 60.8% | 92.2% | +1.2 | −9.0..+5.8 |
| 2 | range 9 → 8.5 m | völva over the band | 42.8% | 40.7% | 57.8% | 88.3% | +1.3 | −7.7..+6.3 |
| 3 | range 8.5 → 8 m | völva still over | 38.8% | 32.2% | 51.7% | 82.2% | −2.3 | −5.5..+4.2 |
| 4 | range back to 9 m; Solenarion damage −10% → −5% | völva is a triangle hard counter (see above); Solenarion near the band edge | 49.0% | 44.5% | 64.8% | 93.5% | +0.3 | −6.2..+5.8 |

Rounds 1 at 500, 2–4 at 300 battles per row; the table above at the default 1000.
Final: 57 HP, 9–14 every 1.7 s, 9 m, 3.4 m/s, sight 11; line, kite, keeps 7 m,
target focus within reach, no fall-back.
