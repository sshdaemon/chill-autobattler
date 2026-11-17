# Balance findings

The current numbers are in [`balance-report.md`](balance-report.md) (generated with `pnpm balance`). This file says what they mean.

## Run balance (sim 0.4.0, change run-balance)

The lab now plays whole runs the way the server does: veterans first, then recruits; one new skill per survivor after each win; Red gets as many skills as the veterans have. See [Runs](balance-report.md#runs) in the report.

**1. No snowballing ✅.** Before the easing, the win rate per match stayed flat, about 55–64% from match 1 to match 7+, for every player policy tried. Red's skill matching cancels the veterans' growth. Picking the strongest skills helps a little (1.44 against 1.33 average wins). **Counter-picking Red** (half its units countering yours) only shortened runs (3.10 to 2.72 average wins, about 12%), so it was rejected.

**2. Runs ended almost at once ❌ → fixed.** At about 55–60% per match, runs averaged 1.3 wins (median 1), and only 11% reached match 5. Ways tried to ease the start:

| Lever | Matches 1 / 2 / 3 won | Average wins | Reach match 5 |
|---|---|---|---|
| None | 54 / 55 / 63% | 1.33 | 11.5% |
| Red 3 units, then 4 | 100 / 100 / 67% | 3.9 | 43% |
| Red 4 units in matches 1–2 | 100 / 100 / 66% | 3.7 | 40% |
| **Red health 85 / 90 / 95%** (chosen) | **88 / 82 / 72%** | **3.05** | **33%** |
| Red health 80 / 85 / 90 / 95% | 93 / 88 / 81% | 3.7 | 49% |

- **Fewer units makes early matches free wins.** One unit fewer decides a 5-unit fight outright.
- **Health is a smooth lever.** 85/90/95% gives a gentle ramp into the normal ~64% from match 4, and it's set and verified by the server per battle (`sim.RunRedHealth`).

---

## Balance pass 1: sim 0.3.0 → 0.4.0

| Measure | 0.3.0 | 0.4.0 | Target |
|---|---|---|---|
| Warriors vs archers | 100.0% | **63.4%** | 65–75% |
| Archers vs mages | 100.0% | **67.5%** | 65–75% |
| Mages vs warriors | 46.3% | **66.0%** | 65–75% |
| Best composition (average vs random Red) | 5W 82.8% | 2W 3M **56.2%** | no runaway mix |
| Worst composition | 5A 16.4% | 3W 2A **37.0%** | nothing hopeless |
| Mixes without warriors | ≤ 23.8% | **48.8–52.3%** | viable |
| Strongest skill | Headshot +17.8 | Arcing shot **+9.6** | ≤ ~10 |
| Weakest skill | Arcing shot +0.7 | Long shot **+3.1** | ≥ ~3 |
| Side bias (Blue wins with identical armies) | 49.2% | 47.8% ± 2.2% | ~50% |
| Average battle length | 14.9 s | 17.1 s | — |

**What changed** (details in `openspec/changes/archive/*-balance-pass-1/`):
- **Archers kite properly.** They back away facing the threat at full speed, start at 4.5 m, and retreat in 2 s bursts with 1.2 s rests. They also have 55 health and 3.6 m/s speed.
- **Softer matchup table.**
- **Instant kills only finish off the wounded** (victims below 40% health).
- **Weak skills buffed:** Swift also attacks faster, Arcing shot also hits harder, Frost always slows, Blink has an 8 s cooldown, and Long shot gives +25%.
- **Strong skills trimmed.**

**Still to watch:**
- **Side bias:** 47.8% is just inside its margin. It was 49.2% before, and no rule treats the sides differently, so it is likely noise. Re-check it with a larger sample in the next pass.
- **Compositions:** warrior-heavy mixes without mages are now the weakest (3W 2A 37%). That's the counter system working (warriors lose to mages), but keep an eye on it.
- **Runs:** veteran snowballing hasn't been measured yet; it's the next lab experiment.

---

## Baseline: sim 0.3.0 (before pass 1)

The sections below describe sim 0.3.0, the state that led to pass 1 (its report is kept as [`balance-report-0.3.0.md`](balance-report-0.3.0.md)).

## 1. The sim is fair between sides ✅

With identical armies and formations, Blue wins 49.2% ± 2.2% and Red 50.6% ([Side bias](balance-report-0.3.0.md#side-bias)). There's no first-mover or side advantage.

## 2. The rock-paper-scissors is broken ❌

The concept intends warriors to beat archers, archers to beat mages, and mages to beat warriors (concept §4.2). Five against five, in a Spread formation ([Class matchups](balance-report-0.3.0.md#class-matchups)):

| | vs Warrior | vs Archer | vs Mage |
|---|---|---|---|
| Warrior | — | **100%** | 54% |
| Archer | **0%** | — | **100%** |
| Mage | 46% | **0%** | — |

- **Matchups are absolute:** two of the three end 100% to 0%, which is a hard lock rather than an advantage. Archers never beat warriors, and mages never beat archers.
- **Mages don't beat warriors** (46%, about even), despite the 125% and 75% multipliers. So mages have no matchup they win, and warriors have none they lose.

## 3. Warriors dominate compositions ❌

Against the game's random Red army, more warriors is simply better ([Compositions](balance-report-0.3.0.md#compositions)):

- **The top 4 mixes are warrior-heavy:** 5W wins 83%, 4W 1M 81%, 3W 2M 77% and 4W 1A 76%.
- **No warriors is a losing army:** every mix without warriors wins at most 24% on average (5A only 16%).
- **Formation matters a lot:** up to 10 points between the best and the average formation for the same mix (for example 0W 0A 5M: 24% average, 36% with 5-0). That's good: deployment is a real decision.

There are two likely reasons, both visible in the matchups:
1. Archers die to warriors before kiting pays off. Their 1.2 s retreat bursts with 2 s rests aren't enough against a 3.2 m/s warrior.
2. Mages hold their ground with 55 HP at 7 m range, so warriors reach them.

## 4. Two skills are far stronger than the rest ❌

Each figure is the change in win rate when one unit learns the skill, measured on the same seeds ([Skill impact](balance-report-0.3.0.md#skill-impact)):

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
