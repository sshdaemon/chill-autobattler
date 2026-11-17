# Balance findings (sim 0.3.0)

The numbers are from [`balance-report.md`](balance-report.md), generated with `pnpm balance`. This file says what they mean and lists the candidate fixes; nothing here has been changed in the game yet.

## 1. The sim is fair between sides ✅

With identical armies and formations, Blue wins 49.2% ± 2.2% and Red 50.6% ([Side bias](balance-report.md#side-bias)). There's no first-mover or side advantage.

## 2. The rock-paper-scissors is broken ❌

The concept intends warriors to beat archers, archers to beat mages, and mages to beat warriors (concept §4.2). Five against five, in a Spread formation ([Class matchups](balance-report.md#class-matchups)):

| | vs Warrior | vs Archer | vs Mage |
|---|---|---|---|
| Warrior | — | **100%** | 54% |
| Archer | **0%** | — | **100%** |
| Mage | 46% | **0%** | — |

- **Matchups are absolute:** two of the three end 100% to 0%, which is a hard lock rather than an advantage. Archers never beat warriors, and mages never beat archers.
- **Mages don't beat warriors** (46%, about even), despite the 125% and 75% multipliers. So mages have no matchup they win, and warriors have none they lose.

## 3. Warriors dominate compositions ❌

Against the game's random Red army, more warriors is simply better ([Compositions](balance-report.md#compositions)):

- **The top 4 mixes are warrior-heavy:** 5W wins 83%, 4W 1M 81%, 3W 2M 77% and 4W 1A 76%.
- **No warriors is a losing army:** every mix without warriors wins at most 24% on average (5A only 16%).
- **Formation matters a lot:** up to 10 points between the best and the average formation for the same mix (for example 0W 0A 5M: 24% average, 36% with 5-0). That's good: deployment is a real decision.

There are two likely reasons, both visible in the matchups:
1. Archers die to warriors before kiting pays off. Their 1.2 s retreat bursts with 2 s rests aren't enough against a 3.2 m/s warrior.
2. Mages hold their ground with 55 HP at 7 m range, so warriors reach them.

## 4. Two skills are far stronger than the rest ❌

Each figure is the change in win rate when one unit learns the skill, measured on the same seeds ([Skill impact](balance-report.md#skill-impact)):

| Strong | Δ | Weak (≈ noise, < 2 pts) | Δ |
|---|---|---|---|
| Headshot (archer, 15% instant kill) | **+17.8** | Long shot | +1.2 |
| Executioner (warrior, 25% instant kill) | **+16.8** | Blink | +1.2 |
| Blast | +8.3 | Frost | +1.1 |
| Dash | +8.0 | Swift | +0.9 |
| Toughness (warrior) | +7.6 | Arcing shot | +0.7 |

The concept already suspected this (§14): one instant-kill roll is worth more than any other skill by a factor of 2 or more. Five skills hardly matter at all. Picking them is a trap, which feels bad in a game where your veterans choose a skill after every win.

## 5. Battles are short and almost never draw ✅

Battles average 14.9 s, and 0.1% are draws (mutual wipes); none reached the 3-minute limit.

## Candidate fixes (to decide, then measure again)

1. **Soften the matchups:**
   - **Archers:** let them survive warriors longer, through longer kiting bursts, shorter rests, or a little more speed.
   - **Mages:** make them beat warriors, through more damage against them, a short knock-back, or slightly more health.
   - Target: each counter wins about 65–75%, not 100%.
2. **Rein in the instant kills:**
   - only against victims below a health threshold (for example 30% HP);
   - or a lower chance (Executioner 25% → 10%, Headshot 15% → 8%);
   - or a cooldown.
3. **Buff the dead-weight skills:**
   - Long shot and Swift: bigger numbers.
   - Frost: higher chance or a stronger slow.
   - Blink: shorter cooldown.
   - Arcing shot: also +damage, or ignore flanking cover.
4. **Re-run `pnpm balance`** after each change and compare. Every balance PR should show before and after numbers.
