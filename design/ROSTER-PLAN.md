# Roster plan: nine units for the Dnipro road

Status: **proposal for review** (Phase 2 of `docs/ROADMAP.md`). Nothing is built
from this until it is approved. Lore and looks: `design/SETTING.md`. All numbers
are starting points for the balance lab, not final values.

Rules this plan assumes (Phase 1 of the roadmap):
- Each unit type has **five abilities**. A unit learns up to **three**: one at
  level 1 (picked in the draft), one more after each won battle (levels 2 and 3).
- Factions draft five units each from their own three types.

## Overview

| | Melee | Ranged | Special |
|---|---|---|---|
| **Kyivan Rus**: steady, protect each other | Druzhynnyk | Strilets | Huslyar (support) |
| **Varangians**: reckless, hit hard | Huskarl | Bowman | Völva (curses, fear) |
| **Romans**: disciplined, armour, fire | Skoutatos | Toxotes | Siphonarios (Greek fire) |

Matchups keep today's triangle so the draft has answers: melee beats ranged,
ranged beats specials through range, specials (fire, fear) beat massed melee.
The huslyar is the exception: weak alone, strong as part of an army.

## Behaviour

Today every unit runs to the nearest enemy (archers also back away from melee).
The plan makes behaviour part of a unit's data, a **temperament** with these fields:

| Field | Meaning |
|---|---|
| **Engage** | *charge*: go for the enemy at once; *line*: advance at the pace of the allies around it, keeping spacing; *hold*: wait in place until an enemy comes within its trigger distance or an ally nearby is attacked |
| **Leash** | how far it chases away from its allies before turning back (metres; none = chases anywhere) |
| **Keep** | the distance it tries to keep from enemies (0 for melee; ranged units back away when closer) |
| **Fall back** | below this % health it retreats behind its allies (none = fights to the end) |
| **Target** | *nearest*, *wounded* (lowest health in reach), *ranged first* (archers and specials), *strongest* (highest damage), *focus* (what its allies are already hitting), *protect* (the enemy closest to its own ranged and special units) |
| **Stay near** | keeps close to a kind of ally: *ranged* (bodyguard), *front line* (behind the melee) |

Today's three units become three temperaments with identical behaviour, and the
golden replays prove nothing changed (roadmap step 0.4).

## Kyivan Rus

### Druzhynnyk (melee): the prince's retainer
- **Role:** balanced, armoured, protects the archers and the singer.
- **Stats:** like today's warrior: 100 HP, 10–16 damage every 1.3 s, 1.5 m, 3.2 m/s.
- **Temperament:** *line*, leash 8 m, target *protect*, stays near *ranged*. Never falls back.
- **Strong against** ranged ×1.4.

| Ability | Effect |
|---|---|
| Shield wall | −30% damage from arrows and bolts while standing still |
| Prince's guard | Takes 30% of the damage aimed at ranged and special allies within 2.5 m |
| Braced spear | An enemy that dashes at him is stopped short and takes a counter-blow (×1.5 damage) |
| Heavy mail | +20% health, −5% speed |
| Axe and sword | +25% attack speed |

### Strilets (ranged): archer of the druzhyna
- **Role:** mobile archer, good at shooting while retreating.
- **Stats:** like today's archer: 55 HP, 9–14 every 1.7 s, 9 m, 3.6 m/s.
- **Temperament:** *line*, keeps 7 m, target *nearest*. Falls back below 30% health.

| Ability | Effect |
|---|---|
| Long shot | +25% range |
| Steppe shooting | Keeps shooting while backing away from melee (no pause to aim) |
| Bodkin points | +30% damage against melee (armour-piercing heads) |
| Arcing shot | Shoots over rocks and trees |
| Quick draw | +25% attack speed |

### Huslyar (special, support): the singer with the husli
- **Role:** makes the army around him stronger; almost harmless alone.
- **Stats:** 60 HP, a knife for 4–6 damage (so a lone huslyar can still finish a
  battle), 3.0 m/s. **Base trait:** a song, allies within 6 m deal +10% damage.
- **Temperament:** *line*, stays near *front line* about 3 m behind it, keeps 5 m
  from enemies, falls back below 50% health. Target *nearest* (knife only when cornered).
- **No matchup bonus.**

| Ability | Effect |
|---|---|
| Song of glory | Allies within the song attack 15% faster |
| Tale of the bohatyrs | Allies within the song take 15% less damage |
| Lament for the fallen | When an ally dies within the song, the allies there deal +30% damage for 5 s |
| Mocking song | Enemies within 5 m attack 15% slower |
| Loud voice | Song reach +50% (6 m → 9 m) |

## Varangians

### Huskarl (melee): sworn household warrior with a Dane axe
- **Role:** shock troops: big blows, less armour.
- **Stats:** 85 HP, 14–20 every 1.5 s, 1.7 m (long axe), 3.3 m/s.
- **Temperament:** *charge*, no leash, target *wounded* among enemies within 4 m,
  otherwise *nearest*. Never falls back.
- **Strong against** ranged ×1.45.

| Ability | Effect |
|---|---|
| Dane axe sweep | Blows also hit enemies within 1.5 m of the target for 40% |
| Charge | Rushes up to 10 m at an enemy (today's Dash) |
| Battle fury | +30% damage while below half health |
| Shield-breaker | Hits make the target take +15% damage for 4 s |
| Executioner | 25% chance per hit to kill outright a target below 30% health |

### Bowman (ranged): hunter-warrior with a yew bow and spears
- **Role:** aggressive archer who hunts the enemy's ranged and specials.
- **Stats:** 60 HP, 9–13 every 1.6 s, 8 m, 3.5 m/s.
- **Temperament:** *charge* to 6 m from its target, keeps 4 m (backs away only
  from melee that gets very close), target *ranged first*. Falls back below 25%.

| Ability | Effect |
|---|---|
| Throwing spears | Every 8 s, an enemy within 8 m gets a thrown spear for ×2 damage |
| Hunter | +20% damage against ranged and special units |
| Yew longbow | +20% range |
| Arcing shot | Shoots over rocks and trees |
| Sea-hardened | +20% health |

### Völva (special): the seeress
- **Role:** weakens and scares the enemy's front line.
- **Stats:** 55 HP, a curse at 7 m for 8–12 damage every 2.2 s, 3.0 m/s. Shown as
  a staff gesture and dark smoke on the target, not a projectile spell.
- **Temperament:** *hold* until an enemy is within 10 m or an ally is attacked,
  keeps 7 m, target *strongest*. Falls back below 40%.
- **Strong against** melee ×1.3.

| Ability | Effect |
|---|---|
| Curse | Her hits make the target take +20% damage for 5 s |
| Dread | Every 12 s, enemies within 4 m flee from her for 1.5 s |
| Foretelling | Allies within 6 m dodge 10% of blows |
| Slip away | When melee gets within 2.5 m, she runs 6 m away (8 s cooldown) |
| Staff of the seeress | +20% health |

## Romans

### Skoutatos (melee): heavy infantryman
- **Role:** the wall: slow, tough, holds the line.
- **Stats:** 115 HP, 9–14 every 1.4 s, 1.9 m (spear), 2.9 m/s.
- **Temperament:** *hold* until an enemy is within 9 m or an ally is attacked,
  then *line*; leash 5 m; target *nearest*. Never falls back.
- **Strong against** ranged ×1.35.

| Ability | Effect |
|---|---|
| Shield wall | −40% damage from the front |
| Phalanx | −8% damage taken for each other skoutatos within 2 m (up to 3) |
| Kontarion | +0.5 m reach |
| Drill | +20% attack speed |
| Klibanion | +20% health (lamellar armour) |

### Toxotes (ranged): archer of the line
- **Role:** disciplined archers who shoot together.
- **Stats:** 55 HP, 9–14 every 1.7 s, 9 m, 3.4 m/s.
- **Temperament:** *line*, stays near *front line* behind the skoutatoi, keeps 8 m,
  target *focus*. Falls back below 30%.

| Ability | Effect |
|---|---|
| Solenarion | +35% range, −20% damage (short darts through a guide tube) |
| Volley | +15% damage when two or more allies shoot the same target |
| Sharp arrows | +15% damage |
| Arcing shot | Shoots over rocks and trees |
| Quick draw | +25% attack speed |

### Siphonarios (special): the Greek fire operator
- **Role:** short-range area damage that punishes massed melee.
- **Stats:** 60 HP, a fire jet: a 5 m cone that hits everyone in it for 8–12
  every 2.4 s, 2.9 m/s.
- **Temperament:** *hold* behind the front line until two or more enemies are in
  the cone, or one is within 4 m; keeps 3 m from melee. Falls back below 40%.
- **Strong against** melee ×1.35.

| Ability | Effect |
|---|---|
| Burning ground | The jet leaves fire for 3 s that burns anyone standing in it (4 damage per second) |
| Fire pots | Every 10 s, throws a clay pot up to 8 m: 2 m splash |
| Pressure | +30% jet length |
| Padded coat | +20% health |
| Fear of fire | Melee units hit by the jet are slowed 30% for 2 s |

## What the rules need

Today's effect kinds: stat bonus, finisher, slow on hit, splash on hit,
gap-closer, escape, over cover. The plan needs these new ones (each one file
after roadmap step 0.3, built in Phase 3 before the units):

| New effect or trait | Used by | Size |
|---|---|---|
| Damage taken −% (from ranged / from the front / while standing) | druzhynnyk, skoutatos, huslyar | small |
| Aura: allies within a radius get a bonus | huslyar, völva | medium |
| Enemy aura: enemies within a radius get a malus | huslyar | small (after aura) |
| Rally: an ally's death gives nearby allies a bonus for a while | huslyar | small |
| Guard: take part of the damage aimed at nearby allies | druzhynnyk | medium |
| Brace: stop a dash and counter | druzhynnyk | small |
| Fury: bonus below a health % | huskarl | small |
| Vulnerable on hit: the target takes more damage for a while | huskarl, völva | small |
| Fear: enemies within a radius flee for a while | völva | medium |
| Burst attack: a stronger attack every N s with its own range (and optional splash) | bowman, siphonarios | medium |
| Bonus against tags | strilets, bowman | small |
| Focus bonus: more damage when allies share the target | toxotes | small |
| Formation bonus: per nearby ally of the same type | skoutatos | small |
| Burning ground: a zone that damages over time | siphonarios | medium |
| Shoot while moving | strilets | medium |
| Cone attack (base attack shape) | siphonarios | medium |
| Evasion and aura radius as stats; splash and slow for melee and cone hits | several | small |

Plus the temperament fields (roadmap step 0.4). New visual work per unit: looks,
clips (singing and strumming, curse gesture, siphon jet, spear throw), effect
presets (fire jet, burning ground, smoke, song notes or ripples).

## Questions for the review

1. Are three unit types per faction enough for a draft of five picks, or should
   each faction get a fourth (for example Rus spearmen, Varangian berserkir,
   Roman kataphraktoi on foot)?
2. Is the huslyar's support role right (weak alone), or should he deal real damage?
3. Too many new effect kinds (about fifteen)? Candidates to cut or simplify:
   Guard, Shoot while moving, Burning ground, Cone attack.
4. Should temperament be changeable by an ability (for example the bowman's
   Hunter switching target priority), or fixed per unit type?
