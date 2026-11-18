# Design

## Context

See proposal.md for the motivation. Today:

- **Characters:** `models.ts` loads one baked glTF per look and team (`/assets/units/<Look>_<team>.gltf`), all skinned to the same Quaternius skeleton. Animation clips are shared, loaded once from `/assets/clips.json`.
  - That file holds the 7 clips listed in `pack_units.py KEEP_CLIPS` and weighs about 650 KB raw (93 KB per clip).
  - It comes from a 12-clip export in the prototype's folder, not from the full library.
- **Clip choice:** `Units.tsx wanted(flags, cls)` picks the clip from snapshot flags, hard-coding "class 0 → `Sword_Attack`, else `Spell_Simple_Shoot`". Attacks are stretched so the whole clip fills the sim's swing.
- **Effects:** `effects.ts` turns sim events into particles with one hand-written branch per event kind.
- **What the client knows:** the snapshot has each unit's class and a skill *count*, not which skills. Blue's skills are known from the run; Red's are not.
- **Tests:** the web has no test runner, only `tsc`. Go tests run natively and under WASM (`wasm/` runs under node with file access).

## Goals / Non-Goals

**Goals:**
- Choosing animations and effects becomes data: looks and presets, which a future `/new-unit` can write.
- A clear clip budget.
- The blow lands with the sim's hit for every clip.
- Today's effects look the same after the move to presets.

**Non-Goals:**
- Changing the rig, retargeting, or IK.
- Layered or additive animation beyond today's aim and flinch.
- Anything that feeds back into the sim.

## Decisions

### 1. Presentation data lives in the web, not in the sim roster
The looks go in `web/src/units/looks.json`, matched to the roster by `name`. They hold models, the animation set, skill animations and effect presets.

- **Rejected: adding presentation strings to `sim.UnitDef`.** The sim stays purely battle logic, and step 3 will add a lot more presentation data (bodies, parts, colours) that has no place in a deterministic module.
- **Cost:** a unit is two files (a roster `.go` and a look entry). The look check (§6) keeps them in sync.

Shape, with the skill ids from the roster:
```json
{
  "clips": { "Sword_Attack": { "hit": 0.62 }, "Sword_Regular_C": { "hit": 0.45 }, "Jog_Fwd_Loop": { "loop": true }, ... },
  "units": {
    "warrior": {
      "models": ["Mirko", "Vesna", "Radim"],
      "anims": { "idle": "Sword_Idle", "move": "Jog_Fwd_Loop", "attack": ["Sword_Attack", "Sword_Regular_C"],
                 "hit": ["Hit_Chest", "Hit_Head"], "death": ["Death01"], "cheer": "Yes" },
      "skills": { "4": { "action": "Jump_Start", "fx": "dash" }, "5": { "fx": "execute" }, "1": {}, "2": {}, "3": {} }
    }, ...
  }
}
```

### 2. Clip intake and budget
- **Sources:** Quaternius Universal Animation Library 1 and 2, standard (free, CC0) editions, downloaded into `~/fun/assets-src/ual1`, `ual2` with their licence files. The files are named to the user before downloading.
- **Same rig:** both packs use the Universal Base Characters rig, the same one our characters are baked on. If a pack's bone names differ, a name map is applied in the bake step, not retargeting.
- **Clip list:** the bake script (`tools/blender/bake_units.py`, extended with a `clips` mode) exports the full clip list to a staging JSON. `pack_units.py` then keeps exactly the clips named in `looks.json`, replacing its hard-coded `KEEP_CLIPS`.
- **Trimming:**
  - drop tracks that are constant across the clip;
  - drop scale tracks;
  - round values to 4 significant decimals;
  - resample to 30 fps.
- **Budget:** `clips.json` at most 1.5 MB raw, about 400 KB gzipped, checked by the look test. Expected size is about 25–30 clips.

### 3. Picking the clip
`wanted()` becomes `pick(look, unitSkills, flags, prevFlags)`, which returns a clip and a rate:

| Unit state | Clip |
|---|---|
| dead | `death[i]` |
| staggered | `hit[i]` |
| swinging | the skill's attack animation if the unit holds a skill that has one (first in roster skill order); otherwise `attack[i]` |
| dashing | the gap-closer skill's action animation, held or looped for the dash's length; otherwise `move` at ×2.2 (today's behaviour) |
| moving | `move`, with today's wet, slow and kite handling |
| cheering | `cheer` |
| anything else | `idle` |

**Variants:** `i` comes from `(unit index, swing count)` hashed, so a replay looks the same each time. The swing count rises on each rising edge of `UF.swinging`.

**Event actions:** for one-shot skills that act through an event (Blink → the escape skill's `action`), the clip is started on the event for that unit. It plays once, then the unit returns to its state clip.

### 4. Lining the blow up with the hit
For an attack clip with natural length `D` and blow at `h` seconds, and the sim's swing `S` (from `data().classes[c].swing`, after Swift), the blow must land at `0.45·S`:

- `rate = h / (0.45·S)`, clamped to `[0.6, 1.8]`.
- If the rate is clamped, the clip starts at an offset so the blow still lands on time, as far as the offset allows.
- After the swing ends, the clip crossfades to the state clip (0.12 s), even if the clip hasn't finished.

The blow times (`hit`) are measured once per clip during intake, at the frame of the weapon's furthest reach, and stored in `looks.json`.

### 5. Effect presets as data
`web/src/scene/fx.json` maps a preset name to a list of emitters:
```json
{ "execute": [ { "pool": "glow", "n": 22, "at": "victim", "y": 1.1, "color": [3,0.5,0.4], "speed": 3.5, "size": 0.4, "life": 0.55, "grav": -4 },
               { "pool": "glow", "n": 1, "at": "victim", "y": 1.1, "color": [3,0.6,0.5], "size": 1.4, "life": 0.25 },
               { "shake": 0.25 } ] }
```

- **Emitter shapes:** `burst`, `sparks` (directional, streaked) and `dust`, plus `ring` (shockwave) and `shake`.
- **The engine stays the same:** `effects.ts` keeps the pools, the per-snapshot budget, the speed and quality scaling and the fog check. It only reads emitter parameters from presets instead of hard-coded branches.
- **Choosing presets:** event kinds map to default presets (`hit`, `hit.flank`, `hit.strong`, `hit.magic`, `death`, `shot.bolt`). A skill's `fx` in the look overrides the preset for the events that skill causes:
  - EvExecute → the attacker's finisher skill;
  - EvFrost and EvBlast → the shooter's slow and splash skills;
  - EvDash and EvBlink → the unit's gap-closer and escape skills.
- **Keeping the look:** porting today's branches is one-to-one. `window.fxDemo()` before and after is compared in screenshots.

### 6. Each unit's skills, and the look check
- **`chillSim.units()`:** a new WASM call made after `create`. It returns `[{cls, skills: [ids]}]` for every unit in index order, Red's generated army included. The store keeps it beside the snapshots. There are no snapshot or proto changes. Red's skills could already be computed by the client from the seed, so nothing becomes visible that wasn't before.
- **`wasm/looks_test.go`:** runs under node and reads `../web/src/units/looks.json`, `../web/src/scene/fx.json` and `../web/public/assets/clips.json`. It checks:
  - every roster entry has a look;
  - each look's `skills` keys equal its unit's five skill ids;
  - every clip named exists in `clips.json` and has a `hit` if it is used as an attack;
  - every `fx` exists in `fx.json`;
  - `clips.json` stays within the budget.

  The check sits in `wasm/` because that package is the bridge between sim and web.
- **Fallback at runtime:** a missing look or clip falls back to the warrior's or archer's set, by the roster's `ranged` trait. In dev it logs one warning.

## Determinism

The sim and its rules don't change. `units()` only reads `Battle.Units` after `New`. Golden replays, `sim.Version` and the snapshot layout stay the same. `pnpm test` (native and WASM goldens) must pass untouched, and the battle checksum of a sample battle in the browser must match the server's verification, as today.

## Risks / Trade-offs

- **[Risk]** The UAL rig's bone names or rest pose differ from the baked characters. → Mitigation: first task is a one-clip spike that loads a UAL clip on a current unit and checks it in the browser. A bone-name map in the bake step if needed. Stop and report if a real retarget is required.
- **[Risk]** No suitable bow animation in the free editions. → Mitigation: archers keep `Spell_Simple_Shoot` as their attack, and Arcing shot uses the closest "aim up" clip. The look check still passes.
- **[Risk]** Clip download size. → Mitigation: the budget is enforced by test, with track trimming. If it's over, drop variants before adding streaming.
- **[Trade-off]** Two files per unit (roster + look) instead of one. This is accepted for keeping the sim pure, and the look check catches drift.
- **[Trade-off]** Red's skill animations reveal which skills a Red unit has once you can see it. This matches what the effects already reveal (frost, blast) and is fine.

## Migration Plan

This is a presentation-only change, with no data migration. It ships as one PR. To roll back, revert it; old `clips.json` comes back with it.

## Open Questions

- The exact clip names chosen for each slot. They are picked during intake from what the free editions contain and recorded in `looks.json`. This doesn't change the approach or the tasks.
