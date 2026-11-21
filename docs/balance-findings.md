# Balance findings

The current numbers are in [`balance-report.md`](balance-report.md) (generated with `pnpm balance`). This file says what they mean.

## Smoother movement (sim 0.12.0, change smooth-movement)

Units no longer flip between two moves at the edge of a distance they react
to (a bodyguard's reach, a wounded unit's sight, a hit-and-run unit's keep
distance), bodyguards don't wait for archers that follow the front line, and
goals past the arena edge or inside rocks are moved to the nearest spot a unit
can stand on. Measured by `TestMovementSmooth` over 600 varied battles:

| | Before | After |
|---|---|---|
| Walking ticks that reverse the previous step (shaking) | 3.45% | 1.34% |
| Unit-ticks stalled (wanting to move 1.5 s, < 0.5 m progress) | 12,717 | 679 |
| Battles with such a stall | 25% | 23%, but short (199 episodes instead of 762) |

Balance moved little (`pnpm balance`, main against the branch): class
matchups by 1–5 pts, the Classic baseline 62.4% → 64.2% (noise ±1.1), runs a
little easier for the player (average wins 2.35–3.12 → 2.44–3.44). The plan
matrix (600 battles per pair) still has no dominant plan in any faction.

## Battle plans: tactics (sim 0.11.0, change battle-tactics)

Three tactics join the six plans (`docs/battle-plans-roadmap.md` step 4):
Hammer and anvil, Reserve, and Hold, then charge. Each splits the army into a
left and a right group with its own stance and one trigger. Rerun with
`go run ./tools/balance -exp plans`.

- **No plan dominates in any faction**, tactics included.
- **The tactics average 42–58%:** Hammer and anvil 49–58%, Reserve 50–56%,
  Hold, then charge 42–55% (weakest for the Varangians).
- **Holding archers didn't shoot:** a holding unit fights only once an enemy is
  within its trigger, and a 6 m trigger kept archers from firing at all. Holding
  tactics give ranged and casters a 10 m trigger.
- **Reserve's damage bonus is very sensitive:** +15% made it dominate Classic,
  none left it at 45–50%; +5% puts it at 50–56%. Its front melee use Nearest
  (Focus left the front too weak, at 38% for the Romans).
- **Hold, then charge waits 10 seconds, braced like a shield wall.** At 15
  seconds without the brace it averaged about 30% everywhere.
- The six plans' numbers barely moved; the known weak spots stay (Rus Hunt
  the archers 18.5%, Varangian Instinct 35%).

Row plan's win rate against the column plan; the same army (the faction's units 2-2-1, recruits) on both sides, sides swapped every other battle, 1000 battles per pair. Beats: 55% or more.

#### classic

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | hammer and anvil | reserve | hold, then charge | Avg | Beats |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 57.1% | 62.2% | 59.3% | 52.8% | 62.5% | 57.8% | 45.8% | 52.4% | 56.2% | 5 |
| **shield wall** | 42.9% | — | 54.9% | 58.5% | 30.1% | 47.9% | 46.1% | 38.1% | 39.1% | 44.7% | 1 |
| **all-out charge** | 37.8% | 45.1% | — | 65.5% | 46.7% | 56.5% | 45.6% | 37.2% | 44.9% | 47.4% | 2 |
| **hunt the archers** | 40.7% | 41.5% | 34.5% | — | 41.2% | 55.9% | 46.4% | 40.2% | 51.4% | 44.0% | 1 |
| **focus fire** | 47.2% | 69.8% | 53.3% | 58.8% | — | 65.5% | 63.8% | 53.8% | 60.6% | 59.1% | 5 |
| **skirmish** | 37.5% | 52.1% | 43.5% | 44.1% | 34.5% | — | 43.0% | 40.9% | 53.2% | 43.6% | 0 |
| **hammer and anvil** | 42.2% | 53.9% | 54.4% | 53.6% | 36.2% | 57.0% | — | 40.9% | 50.0% | 48.6% | 1 |
| **reserve** | 54.2% | 61.9% | 62.8% | 59.8% | 46.2% | 59.1% | 59.1% | — | 46.4% | 56.2% | 5 |
| **hold, then charge** | 47.6% | 60.9% | 55.1% | 48.6% | 39.4% | 46.8% | 50.0% | 53.6% | — | 50.3% | 2 |

#### rus

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | hammer and anvil | reserve | hold, then charge | Avg | Beats |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 37.4% | 30.1% | 77.5% | 33.2% | 57.0% | 39.0% | 40.9% | 43.4% | 44.8% | 2 |
| **shield wall** | 62.6% | — | 57.2% | 91.5% | 44.3% | 64.3% | 48.2% | 59.7% | 56.2% | 60.5% | 6 |
| **all-out charge** | 69.8% | 42.8% | — | 84.5% | 49.2% | 72.8% | 43.5% | 47.3% | 40.9% | 56.4% | 3 |
| **hunt the archers** | 22.5% | 8.5% | 15.5% | — | 20.6% | 19.8% | 20.9% | 18.1% | 22.3% | 18.5% | 0 |
| **focus fire** | 66.8% | 55.7% | 50.8% | 79.4% | — | 70.2% | 58.3% | 55.4% | 61.1% | 62.2% | 7 |
| **skirmish** | 43.0% | 35.7% | 27.2% | 80.2% | 29.8% | — | 38.0% | 38.8% | 36.9% | 41.2% | 1 |
| **hammer and anvil** | 61.0% | 51.8% | 56.5% | 79.0% | 41.7% | 62.1% | — | 55.1% | 58.1% | 58.2% | 6 |
| **reserve** | 59.1% | 40.3% | 52.7% | 81.8% | 44.6% | 61.3% | 44.9% | — | 41.9% | 53.3% | 3 |
| **hold, then charge** | 56.6% | 43.8% | 59.1% | 77.7% | 38.9% | 63.1% | 41.9% | 58.1% | — | 54.9% | 5 |

#### varangians

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | hammer and anvil | reserve | hold, then charge | Avg | Beats |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 25.0% | 32.3% | 31.5% | 43.6% | 37.8% | 37.5% | 30.0% | 40.6% | 34.8% | 0 |
| **shield wall** | 75.0% | — | 63.4% | 47.3% | 53.1% | 69.4% | 53.8% | 58.7% | 58.5% | 59.9% | 5 |
| **all-out charge** | 67.7% | 36.6% | — | 40.8% | 50.5% | 62.9% | 51.3% | 46.2% | 58.6% | 51.8% | 3 |
| **hunt the archers** | 68.5% | 52.7% | 59.2% | — | 62.3% | 74.6% | 65.2% | 58.4% | 69.5% | 63.8% | 7 |
| **focus fire** | 56.4% | 46.9% | 49.5% | 37.7% | — | 45.9% | 54.4% | 46.8% | 58.9% | 49.5% | 2 |
| **skirmish** | 62.3% | 30.6% | 37.1% | 25.4% | 54.1% | — | 47.7% | 49.1% | 66.0% | 46.5% | 2 |
| **hammer and anvil** | 62.5% | 46.2% | 48.7% | 34.8% | 45.6% | 52.3% | — | 44.2% | 57.1% | 48.9% | 2 |
| **reserve** | 70.0% | 41.3% | 53.8% | 41.6% | 53.2% | 50.9% | 55.8% | — | 56.0% | 52.8% | 3 |
| **hold, then charge** | 59.5% | 41.5% | 41.4% | 30.6% | 41.1% | 34.0% | 42.9% | 44.0% | — | 41.9% | 1 |

#### romans

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | hammer and anvil | reserve | hold, then charge | Avg | Beats |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 49.8% | 61.5% | 40.7% | 51.3% | 54.1% | 40.6% | 45.6% | 42.5% | 48.3% | 1 |
| **shield wall** | 50.2% | — | 60.5% | 50.9% | 52.8% | 48.9% | 39.0% | 43.4% | 38.6% | 48.0% | 1 |
| **all-out charge** | 38.5% | 39.5% | — | 45.5% | 56.7% | 47.1% | 30.6% | 35.0% | 34.8% | 41.0% | 1 |
| **hunt the archers** | 59.3% | 49.1% | 54.5% | — | 52.2% | 57.6% | 53.0% | 57.8% | 65.8% | 56.2% | 4 |
| **focus fire** | 48.7% | 47.2% | 43.3% | 47.8% | — | 52.8% | 49.0% | 54.0% | 57.4% | 50.0% | 1 |
| **skirmish** | 45.9% | 51.2% | 52.9% | 42.4% | 47.2% | — | 42.1% | 48.2% | 46.0% | 47.0% | 0 |
| **hammer and anvil** | 59.5% | 61.1% | 69.4% | 47.0% | 51.0% | 57.9% | — | 56.6% | 50.1% | 56.6% | 5 |
| **reserve** | 54.4% | 56.6% | 65.0% | 42.2% | 46.0% | 51.8% | 43.4% | — | 40.2% | 49.9% | 2 |
| **hold, then charge** | 57.5% | 61.5% | 65.1% | 34.2% | 42.6% | 54.0% | 49.9% | 59.9% | — | 53.1% | 4 |

## Battle plans: balance pass (sim 0.10.0, change plan-balance)

Plans now trade an edge for a cost (`docs/battle-plans-roadmap.md` step 3;
values in `openspec/changes/archive/*-plan-balance/design.md`). Rerun with
`go run ./tools/balance -exp plans`; the matrix now swaps sides every other
battle, so the diagonal side bias is gone.

- **No plan dominates in any faction** (the goal): every plan loses to or ties
  with at least one other.
- **Every faction has at least four plans averaging 45% or better**, so the
  pick is a real choice. The Romans are tightest (45–55%).
- **Shield wall was held back by the Protect target rule**, not by its numbers:
  its melee chased enemies near their own archers. With Nearest it now beats
  Charge and Hunt for most factions and loses to Focus fire.
- **Known weak spots, kept as faction character:** Hunt the archers for the
  Rus (17%: druzhynnyk shield their archers, so hunters run into a wall) and
  Instinct for the Varangians (34%: huskarls left to themselves are reckless).
- Instinct stays the strongest Classic plan (59%); Classic is to be retired.

Row plan's win rate against the column plan; the same army (the faction's units 2-2-1, recruits) on both sides, sides swapped every other battle, 1000 battles per pair. Beats: 55% or more.

#### classic

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg | Beats |
|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 57.1% | 62.2% | 59.3% | 52.8% | 62.5% | 58.8% | 4 |
| **shield wall** | 42.9% | — | 54.9% | 58.5% | 30.1% | 47.9% | 46.8% | 1 |
| **all-out charge** | 37.8% | 45.1% | — | 65.5% | 46.7% | 56.5% | 50.3% | 2 |
| **hunt the archers** | 40.7% | 41.5% | 34.5% | — | 41.2% | 55.9% | 42.8% | 1 |
| **focus fire** | 47.2% | 69.8% | 53.3% | 58.8% | — | 65.5% | 58.9% | 3 |
| **skirmish** | 37.5% | 52.1% | 43.5% | 44.1% | 34.5% | — | 42.4% | 0 |

#### rus

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg | Beats |
|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 37.4% | 30.1% | 77.5% | 33.2% | 57.0% | 47.1% | 2 |
| **shield wall** | 62.6% | — | 57.2% | 91.5% | 44.3% | 64.3% | 64.0% | 4 |
| **all-out charge** | 69.8% | 42.8% | — | 84.5% | 49.2% | 72.8% | 63.8% | 3 |
| **hunt the archers** | 22.5% | 8.5% | 15.5% | — | 20.6% | 19.8% | 17.4% | 0 |
| **focus fire** | 66.8% | 55.7% | 50.8% | 79.4% | — | 70.2% | 64.6% | 4 |
| **skirmish** | 43.0% | 35.7% | 27.2% | 80.2% | 29.8% | — | 43.2% | 1 |

#### varangians

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg | Beats |
|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 25.0% | 32.3% | 31.5% | 43.6% | 37.8% | 34.0% | 0 |
| **shield wall** | 75.0% | — | 63.4% | 47.3% | 53.1% | 69.4% | 61.6% | 3 |
| **all-out charge** | 67.7% | 36.6% | — | 40.8% | 50.5% | 62.9% | 51.7% | 2 |
| **hunt the archers** | 68.5% | 52.7% | 59.2% | — | 62.3% | 74.6% | 63.5% | 4 |
| **focus fire** | 56.4% | 46.9% | 49.5% | 37.7% | — | 45.9% | 47.3% | 1 |
| **skirmish** | 62.3% | 30.6% | 37.1% | 25.4% | 54.1% | — | 41.9% | 1 |

#### romans

| Plan | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg | Beats |
|---|---|---|---|---|---|---|---|---|
| **instinct** | — | 49.8% | 61.5% | 40.7% | 51.3% | 54.1% | 51.5% | 1 |
| **shield wall** | 50.2% | — | 60.5% | 50.9% | 52.8% | 48.9% | 52.6% | 1 |
| **all-out charge** | 38.5% | 39.5% | — | 45.5% | 56.7% | 47.1% | 45.5% | 1 |
| **hunt the archers** | 59.3% | 49.1% | 54.5% | — | 52.2% | 57.6% | 54.5% | 2 |
| **focus fire** | 48.7% | 47.2% | 43.3% | 47.8% | — | 52.8% | 47.9% | 0 |
| **skirmish** | 45.9% | 51.2% | 52.9% | 42.4% | 47.2% | — | 47.9% | 0 |

## Battle plans: first matrix (sim 0.9.0, change battle-plans)

Starting values from `docs/battle-plans-roadmap.md` step 1, before any tuning
(that is step 3). Rerun with `go run ./tools/balance -exp plans`.

- **All-out charge is the best or near-best plan for every faction** (62–73%
  on average) and never loses clearly to any plan. It dominates the Rus.
- **Instinct is strong too:** the units' own temperaments are already well
  tuned, so most plans are a step down from them. For a plan to be worth
  picking, it must beat Instinct against something.
- **Skirmish is the weakest plan everywhere** (23–33% on average) and only
  beats itself. Shield wall is weak except for the Rus and Varangians.
- **Hunt the archers depends on the faction:** the best Varangian plan (69%),
  the worst Rus one (29%).
- **No rock-paper-scissors yet:** Charge and Instinct beat almost every other
  plan. Step 3 should make the slow plans (Shield wall, Skirmish, Focus fire)
  beat Charge, e.g. by stronger holding and protecting, or by Charge
  breaking formation.

Blue win rate, rows = Blue's plan, columns = Red's plan; the same army (the faction's units 2-2-1, recruits) on both sides, 2000 battles per cell. A plan that wins every row cell dominates.

#### classic

| Blue \ Red | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg |
|---|---|---|---|---|---|---|---|
| **instinct** | 50.4% | 82.7% | 51.1% | 67.0% | 74.3% | 92.8% | 69.7% |
| **shield wall** | 19.6% | 49.5% | 18.9% | 39.9% | 34.1% | 59.0% | 36.8% |
| **all-out charge** | 49.6% | 83.2% | 51.1% | 70.0% | 72.4% | 92.8% | 69.8% |
| **hunt the archers** | 31.9% | 60.1% | 30.1% | 52.7% | 60.7% | 71.3% | 51.1% |
| **focus fire** | 27.5% | 65.8% | 30.5% | 40.2% | 52.1% | 67.7% | 47.3% |
| **skirmish** | 7.8% | 40.6% | 7.6% | 26.9% | 33.1% | 48.8% | 27.5% |

#### rus

| Blue \ Red | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg |
|---|---|---|---|---|---|---|---|
| **instinct** | 50.5% | 56.7% | 31.3% | 78.6% | 52.8% | 84.5% | 59.1% |
| **shield wall** | 42.8% | 43.7% | 23.6% | 77.3% | 42.2% | 72.0% | 50.3% |
| **all-out charge** | 68.5% | 75.3% | 51.7% | 82.1% | 67.8% | 94.8% | 73.4% |
| **hunt the archers** | 18.5% | 21.0% | 17.6% | 49.1% | 27.4% | 37.1% | 28.5% |
| **focus fire** | 50.2% | 59.2% | 31.8% | 71.8% | 50.2% | 72.2% | 55.9% |
| **skirmish** | 15.4% | 30.5% | 5.8% | 60.3% | 29.3% | 52.8% | 32.3% |

#### varangians

| Blue \ Red | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg |
|---|---|---|---|---|---|---|---|
| **instinct** | 50.2% | 48.9% | 29.0% | 31.6% | 54.0% | 84.1% | 49.6% |
| **shield wall** | 51.9% | 49.4% | 31.6% | 25.5% | 38.9% | 80.5% | 46.3% |
| **all-out charge** | 69.6% | 67.2% | 48.0% | 36.2% | 61.4% | 91.0% | 62.2% |
| **hunt the archers** | 72.0% | 73.3% | 56.8% | 51.1% | 72.0% | 90.0% | 69.2% |
| **focus fire** | 46.6% | 60.5% | 35.6% | 26.7% | 48.4% | 69.4% | 47.9% |
| **skirmish** | 18.1% | 21.3% | 10.1% | 10.7% | 30.6% | 49.6% | 23.4% |

#### romans

| Blue \ Red | instinct | shield wall | all-out charge | hunt the archers | focus fire | skirmish | Avg |
|---|---|---|---|---|---|---|---|
| **instinct** | 50.3% | 79.5% | 53.3% | 46.2% | 73.9% | 83.0% | 64.4% |
| **shield wall** | 18.6% | 49.5% | 16.8% | 29.0% | 60.1% | 48.0% | 37.0% |
| **all-out charge** | 47.2% | 85.4% | 50.7% | 49.1% | 79.5% | 86.8% | 66.5% |
| **hunt the archers** | 55.4% | 69.0% | 53.0% | 50.5% | 74.0% | 75.0% | 62.8% |
| **focus fire** | 26.7% | 42.4% | 21.0% | 26.2% | 50.2% | 60.9% | 37.9% |
| **skirmish** | 16.4% | 52.1% | 14.5% | 25.4% | 40.6% | 49.5% | 33.1% |

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
