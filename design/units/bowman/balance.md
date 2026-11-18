# Balance: bowman

1000 battles per row, fixed seeds (sim 0.8.0+1d8f4dda). Bands: openspec unit-authoring.

| Check | Result | Band | |
|---|---|---|---|
| 5× bowman vs 5× warrior | 29.9% | 25–75% | pass |
| 5× bowman vs 5× archer | 48.4% | 25–75% | pass |
| 5× bowman vs 5× mage | 70.2% | 25–75% | pass |
| 5× bowman vs 5× druzhynnyk | 39.0% | 25–75% | pass |
| 5× bowman vs 5× strilets | 66.1% | 25–75% | pass |
| 5× bowman vs 5× huslyar | 100.0% | 25–75% | **FAIL** |
| 5× bowman vs 5× huskarl | 62.8% | 25–75% | pass |
| 5× bowman vs 5× volva | 100.0% | 25–75% | **FAIL** |
| 5× bowman vs 5× skoutatos | 100.0% | 25–75% | **FAIL** |
| 5× bowman vs 5× toxotes | 100.0% | 25–75% | **FAIL** |
| 5× bowman vs 5× siphonarios | 100.0% | 25–75% | **FAIL** |
| standard army with one Archer swapped for bowman | 65.5% (base 65.6%, -0.1 pts) | ±10 pts | pass |
| level-1 pick Throwing spears | 74.3% (+9.1 pts vs its average) | ±10 pts | pass |
| level-1 pick Hunter | 66.0% (+0.8 pts vs its average) | ±10 pts | pass |
| level-1 pick Yew longbow | 62.0% (-3.2 pts vs its average) | ±10 pts | pass |
| level-1 pick Arcing shot | 60.5% (-4.7 pts vs its average) | ±10 pts | pass |
| level-1 pick Sea-hardened | 63.2% (-2.0 pts vs its average) | ±10 pts | pass |

## Failed bands and why

The failures are all against units whose head-to-head is ~100% for every
ranged unit, the plain archer and the strilets included (`design/units/strilets/balance.md`):

- **Huslyar:** weak alone by design (roster plan).
- **Völva, skoutatos, toxotes, siphonarios:** placeholders that their unit
  sessions have not tuned yet.

Re-run this once those units are balanced. Against the finished units
(warrior, archer, mage, druzhynnyk, strilets) and the huskarl placeholder,
every band passes, as do the swap and all five skills.

## Tuning rounds

The starting point was the placeholder (roster-plan numbers, falls back
below 25%): warrior 30%, archer 38%, mage 67%, strilets 45%; Throwing spears
+13 pts as a pick (fail). Rounds 1–4 ran at 200 battles per row before the
design checkpoint; their numbers were approved there.

| Round | Change | Why | Warrior | Archer | Mage | Strilets | Swap | Spears pick |
|---|---|---|---|---|---|---|---|---|
| 1 | no fall-back | a ranged unit that retreats at low health stops shooting (strilets finding) | 36% | 83% | 94% | 91% | +2.0 | +13.2 |
| 2 | HP 60 → 52; spears ×2 every 8 s → ×1.6 every 10 s | too strong without the fall-back; spears over the skill band | 23% | 37% | 54% | 51% | −2.5 | +10.0 |
| 3 | HP 52 → 56, damage 9–13 → 8–12 (also tried HP 54, keep 5) | warrior under the band | 12% | 26% | 51% | 40% | −4.5 | +11.5 |
| 4 | HP 52, damage 9–14 | damage, not health, decides the warrior matchup; light kit, hits hard | 30% | 53% | 67% | 68% | −1.0 | +9.8 |
| final | round 4 at 1000 battles per row | the table above | 29.9% | 48.4% | 70.2% | 66.1% | −0.1 | +9.1 |

Final: 52 HP, 9–14 every 1.6 s, 8 m, 3.5 m/s, sight 10; charge, kite, keeps
4 m, target ranged first, no fall-back. Throwing spears ×1.6 up to 8 m every 10 s.
