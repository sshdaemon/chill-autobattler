# Auto-battle: game concept

Status: concept taken from the throwaway prototype in `battle.html` (November 2026). Numbers updated by balance pass 1 (sim 0.4.0), measured with `pnpm balance`; see `docs/balance-findings.md`.
Every number here is a prototype value: a starting point for balancing, not a final value.

## 1. The idea in one paragraph

This is a web-based 3D game with two activities: a shared **global map** (postponed) and **battles**. A battle is an **indirect-control auto-battle**. You choose your army, place it on the field, and then watch your units fight on their own. Warcraft III-style 1v1 comes first; many vs many comes later. The skill is in **what you bring and where you put it**, not in clicking during the fight. Battles are chained into a **run**. You keep playing while you keep winning, survivors become veterans who learn skills, and the enemy grows to match you. The run ends at the first loss, and your score goes on a scoreboard.

## 2. Core loop

```
New run (random name)
  └─► Choose army (5 units; veterans locked in)
        └─► Deploy (15 s: formation or drag units)
              └─► Fight (automatic)
                    ├─ Win  ─► score + each survivor picks a skill ─► next match (new map)
                    └─ Loss ─► run over ─► scoreboard ─► new run
```

- **One match:** army picker → deployment → short countdown → fight → result.
- **Run:** a chain of matches. Only winning continues it.
- **Page refresh or Restart** starts a brand-new run with a new name. A run can't be resumed after a refresh. That stops players from reloading their way out of a loss.

## 3. Match structure

| Phase | What happens | Numbers |
|---|---|---|
| Army picker | Pick 5 units from the classes. Veterans from the last match are locked in and can't be dismissed. | Army size 5 |
| Deployment | Arrange units inside your zone with a formation preset or by dragging. Red also deploys. Fog hides the enemy. | 15 s (skip with Ready); zone 20 m wide × 6 m deep |
| Countdown | Short "1… Fight!" banner | ~1 s |
| Fight | Units act on their own until one side is wiped out | — |
| Result | The winner is shown. Damage stops the moment the match is decided, so arrows still in flight don't count. | — |

## 4. Classes

In the code each class is an entry of the unit roster (`sim/unit_warrior.go`, `unit_archer.go`, `unit_mage.go`); those files are the source of truth for the numbers below.

### 4.1 Base stats

| | ⚔️ Warrior | 🏹 Archer | ✨ Mage |
|---|---|---|---|
| Role | Sturdy melee | Fragile ranged; backs off from melee | Mid-range caster; holds ground |
| Health | 100 | 55 | 55 |
| Damage per hit | 10–16 | 9–14 | 14–20 |
| Attack cooldown | 1.3 s | 1.7 s | 2.2 s |
| ≈ Base DPS | ~10 | ~6.8 | ~7.7 |
| Range | 1.5 m | 9 m | 7 m |
| Speed | 3.2 m/s | 3.6 m/s | 3.0 m/s |
| Sight | 7 m | 11 m | 9 m |
| Attack | Sword swing (hit lands 45% into the animation) | Arrow: arcs, homes in, 18 m/s; a miss sticks in the ground | Magic bolt: straight, homes in, 14 m/s; a miss fizzles |
| Wind-up | up to the cooldown | 0.6 s | 0.8 s |
| Special behaviour | — | Kites away from melee | No kiting |

### 4.2 Class matchups (rock-paper-scissors)

Damage multiplier, **attacker → victim**. It stacks with every other multiplier.

| Attacker ↓ / Victim → | Warrior | Archer | Mage |
|---|---|---|---|
| ⚔️ Warrior | 100% | **145%** | 95% |
| 🏹 Archer | 100% | 100% | 100% |
| ✨ Mage | **135%** | 100% | 100% |

How to read it:
- Each counter should win about 65–75% five against five, not 100%. Measured: warriors beat archers 65%, archers beat mages 69%, mages beat warriors 65%.
- Archers win against mages through range (9 m against 7 m) and kiting, not through a multiplier.
- Same-class hits are 100%.

## 5. Combat rules

- **Flanking:** a hit that comes from outside the victim's front cone (±60° of where it faces) does **×2 damage**. For arrows and bolts, the direction is measured from where they were fired. Positioning and formation therefore matter a lot.
- **Damage formula:**
  `random(min..max) × flank(×2 or ×1) × matchup × skill damage bonus`
  Instant-kill skills then override the result (see §8).
- **Hit reaction:** a non-lethal hit staggers the victim for 0.3 s unless it is mid-attack.
- **Targeting:**
  - Units lock on only to enemies their team **can see** (see §6).
  - They pick the nearest visible enemy and re-check targets and paths every 0.5 s.
- **Ranged line of fire:** archers and mages need a clear walkable line to the target, so rocks and trees block shots unless the unit has Arcing shot. Water never blocks shots.
- **Archer kiting:**
  - **Trigger:** when a melee enemy comes within 4.5 m, the archer backs away in about 2.5 m steps, **still facing** its attackers, at full speed. Several threats are combined into one escape direction.
  - **Limits:** retreats come in **bursts** of at most 2 s, followed by at least 1.2 s standing and shooting. Archers are faster than warriors, so without this limit they would run away forever.
  - **Cornered:** a cornered archer stands and shoots.
- **Movement:**
  - **Pathing:** A* on a 0.5 m grid, with string-pulling for natural paths.
  - **Spacing:** units push apart when closer than 0.9 m (unit radius 0.35 m).
- **Without a target:** a unit that sees no enemy marches roughly straight toward the enemy's edge. If it loses sight of an enemy, it checks the last seen position. Once it reaches the enemy edge, it sweeps random spots in the enemy half.

## 6. Vision and fog of war

- Each unit sees as far as its **sight** value. Rocks and trees block sight; water doesn't.
- **Team vision is shared:** if any unit on a team sees an enemy, the whole team can target it.
- **Fog of war** for the player:
  - Unseen areas are darkened, and enemy units in them are hidden.
  - Enemy portraits show "?" and hidden health.
  - The fog stays fully closed during deployment, so you can't scout Red's setup.
- **Debug:** V toggles the fog off.

## 7. Battlefield

- **Arena:** 28 × 28 m with a low stone border. Blue deploys at the bottom, Red at the top.
- **A new random map every match**, built from a seed. The same seed gives the same map, which helps with replays and balance tests.
- **Obstacles** (spaced at least 1.6 m apart and 2 m from the border, so there is always a way around; deployment zones stay clear):

| Feature | Count | Size | Walk | Sight | Shots |
|---|---|---|---|---|---|
| Rocks | 3–6 | radius 0.9–1.9 m | blocked | blocked | blocked (Arcing shot goes over) |
| Trees | 3–7 | radius 0.8 m | blocked | blocked | blocked (Arcing shot goes over) |
| Water ponds | 2–3 ponds, 2–4 blobs each | 1.6–2.6 m blobs, middle band only | **70% speed** | clear | clear |

- **Walls removed:** long block walls were tried in the prototype and taken out. They funnelled the fights too hard.
- **Water:**
  - Units wade at 70% speed with slowed animation and sink slightly.
  - Pathfinding treats a water step as about 1.43× the cost of a dry one. Units go around a pond when that's about as quick and wade across when the detour is long.
  - Arrows that miss over water vanish.

## 8. Deployment and formations

- **Zone:** your units can be placed only inside your own zone (20 × 6 m).
- **Presets:** these are left–right flank splits as seen on screen. In each flank block, melee stands in front and ranged behind, and classes are shared out so both flanks get a mix.

| Preset | Meaning |
|---|---|
| Fully spread | One wide staggered line across the whole zone |
| 3-2, 2-3, 4-1, 1-4 | Units on the left flank – units on the right flank |
| 0-5, 5-0 | Everyone on one flank |
| Custom | Dragging any unit switches to Custom |

- **Red** picks a random preset (seeded by the map).
- The deployment bar is compact and see-through at the bottom of the screen. Each button shows a small diagram of the actual army mix: ● warrior, ○ archer, ◆ mage.

## 9. Army picker

- **Cards:** one per class, with stats, a blurb and a "Damage taken" line from the matchup table. Pick counts with − / +; a row of 5 slots fills up.
- **First match default:** 2 warriors, 2 archers, 1 mage. The last pick is remembered within the session.
- **Veterans:** they take slots first, shown in gold with their ★. You can't go below your veteran count in any class.
- **Red's army:** 5 units, each a random class, seeded by the map. The end banner reveals Red's mix and formation.

## 10. Progression: veterans and skills

- **Learning:** after a **won** match, every surviving unit learns **one new skill** for its class. The player picks it on the Victory screen, and Next match stays locked until every survivor has one.
- **Limit:** each skill can be learned once, so a unit maxes out at **5 skills**.
- **★ count:** a unit's ★ equals its number of skills. Stars show over its health bar, on its portrait and in its info panel.
- **Death:** a veteran who dies is gone for good, with all their skills.
- **Recruits:** new units always start with no skills.

### 10.1 Skill list

In the code each skill is a list of effects (stat bonus, finisher, slow on hit, splash on hit, gap-closer, escape, over cover) in its unit's roster file.

**⚔️ Warrior**

| Skill | Effect |
|---|---|
| ❤️ Toughness | +10% health |
| 👟 Swift | +40% speed, and attacks 20% faster |
| 💪 Strength | +15% damage |
| 💨 Dash | Rush up to 10 m at an enemy (16 m/s, stops at sword range); 10 s cooldown. Fires when an enemy is between range + 1.5 m and 10 m away with a clear path. |
| 💀 Executioner | 35% chance per hit to kill outright a victim below 40% health |

**🏹 Archer**

| Skill | Effect |
|---|---|
| 🎯 Long shot | +25% range |
| 🏹 Sharp arrows | +10% damage |
| 🌳 Arcing shot | Shoots over rocks and trees, with a higher arc (the team still has to see the target); +10% damage |
| 💀 Headshot | 25% chance per hit to kill outright a victim below 40% health |
| ❤️ Toughness | +20% health |

**✨ Mage**

| Skill | Effect |
|---|---|
| ❄️ Frost | Every bolt slows the target by 40% for 3 s (stacks with water) |
| 🔥 Power | +10% damage |
| 🌀 Blink | Teleports about 8 m away when a melee enemy gets within 2.5 m; 8 s cooldown |
| 💥 Blast | Bolts also hit other enemies within 2 m of the target for 35% damage |
| ❤️ Toughness | +20% health |

### 10.2 Enemy scaling

- **Red matches you:** in each match, Red gets **as many skills in total as your army has**.
- **Spread:** each skill goes to a random Red unit, as a random skill it doesn't have yet, seeded by the map. Your one ★★ veteran might face one ★★ enemy or two ★ enemies.
- **Fresh start:** if all your veterans die, Red is back to having no skills.
- **Easier first matches:** Red's units start at **85% health in match 1, 90% in match 2 and 95% in match 3**, then at full health. Without this, a run was about a coin flip from the first match (1.3 wins on average); with it, about 88/82/72% of those matches are won, then about 64% from match 4 on (3 wins on average).
- **No counter-picking:** Red's army stays seeded-random. The lab found no snowballing (the win rate per match is flat), so counter-picking would only shorten runs.
- **Why:** the run stays tense without a flat difficulty curve. The edge comes from *which* skills you chose and *keeping* units alive.

Earlier versions tried and dropped:
1. Every veteran got +10% HP and damage per match survived, and Red got nothing. Too easy.
2. Every Red unit got +10% per win. Too blunt.
3. Fewer Red units in the first matches (3 or 4 against your 5). That turned those matches into free wins (about 100%).

## 11. Scoring and scoreboard

| Event | Points |
|---|---|
| Won match | +100 |
| Each of your units alive at the end of a won match | +10 |

- **Run name:** each run gets a random name made of an adjective, a noun and a number (for example "Iron Fox 910"). A new run always gets a new name.
- **Run over:** a loss shows the run's total, wins and the **top 10** scoreboard with your row highlighted. The prototype keeps the board in the browser; the real game needs a server-side leaderboard.
- **Top bar during a run:** run name · match number · wins · points · Red's skill count.

## 12. UI and feedback

- **Portraits:**
  - Blue along the bottom-left, Red along the bottom-right.
  - Each shows a health bar, number and ★ badge; dead units are greyed with ✝, and unseen enemies show "?".
  - Click to select, double-click to centre the camera.
- **Info panel** for the selected unit:
  - Name, ★ and class.
  - Health and current status, for example "Casting at Mirko (Red)", "Falling back from melee", "wading, 70% speed" or "slowed by frost".
  - Skills, stats, kills, damage dealt and the "Takes from" matchup row.
- **Floating numbers:**
  - Plain white for a normal hit.
  - Yellow with "x2" for a flank hit.
  - Orange with ▲ for a strong matchup, grey with ▼ for a resisted one.
- **Skill labels:** short labels float up for skill moments ("💀 Execute", "💨 Dash", "🌀 Blink", "❄️ Slowed"). Blink and Blast also draw a ring on the ground.
- **Camera:** Warcraft III style with a fixed tilt. Pan with WASD, the arrow keys or the screen edge; zoom with the wheel; F follows the selected unit.
- **Debug keys:** 1/2/4 sets game speed, P shows paths, V toggles fog, R starts a new run.

## 13. Look and assets (prototype)

- **Engine:** Three.js, with Yuka for steering.
- **Characters:** Quaternius CC0 modular characters and animation library, all on one shared skeleton.
  - Warriors are peasant models with a sword.
  - Archers are rangers with a bow.
  - Mages are peasant models tinted purple, with a pointed hat and a staff with a glowing orb.
- **Team colours:** Red uses alternate textures, plus coloured ground rings and health bars.
- **Real game needs:** proper mage and noble outfits (Quaternius paid tier or custom) and a consistent medieval style that is less cartoonish.

## 14. Open questions and ideas for the real game

**Balance**
- ~~Executioner and Headshot are very strong.~~ Done in balance pass 1: they now only finish off victims below 40% health.
- ~~Snowballing: do runs get easier as veterans gain skills?~~ Measured: no, the win rate per match stays flat.
- ~~Red's skills are fully random. Should they counter your army instead?~~ Measured in the run lab: no snowballing, so no (Red stays random).
- Should healing between matches stay full (as now) or be partial?

**Content**
- More classes: healer, shield bearer, cavalry, siege.
- Recruit pool and economy: currently recruits are free. An economy could cost gold from points or limit choices.
- More terrain: bridges, high ground (range or sight bonus), bushes (hide but don't block), destructible cover.
- Spells or orders the player can trigger mid-battle (one per match) while keeping indirect control.

**Modes and scale**
- PvP: two players' runs or saved armies meet (asynchronous, like a "ghost" army).
- Many vs many: larger armies, squads, morale.
- Connection to the global map (postponed): where runs and veterans come from, and what winning gives you on the map.

**Technical**
- The simulation should be **deterministic and seeded** (map, Red army, formation and skills already are) so the server can run or verify battles and the client just replays them. Random rolls in combat (damage, execute, frost) need a seeded random generator too.
- The scoreboard and run state must be server-side to prevent cheating.
- Keep balance numbers in data (as `STATS`, `MATCHUP`, `SKILLS`, `ABIL` and `WATER` are now), not in code.

## 15. Prototype reference

- **Prototype:** `web3d-demo/battle.html`, served by `python3 -m http.server 8123 --directory web3d-demo`.
- **Map seeds:** `?seed=123` replays a map.
- **Test helpers:**
  - `demo.sim(seconds)` fast-forwards a fight without rendering.
  - `demo.counters.pairs` gives per-matchup damage statistics for balance checks.
