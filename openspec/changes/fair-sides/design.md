# Design

## Context

After mirrored maps, mirror matches still leaned on some maps. The method:
play a battle with the same army on both sides, Red's start stagger copied from
Blue's, and compare each Blue unit with its Red counterpart reflected, tick by
tick, until they first differ; fix that; repeat.

## Findings, in the order the traces hit them

1. **`Mul` floored** (`>> 16`). Negative products round away from zero, so
   Blue, walking toward −z, moved one raw unit further per step than Red.
   Diverged at tick 1. Now `a*b / One`, truncating like `Div` and `MulDiv`.
2. **Pushes in unit order.** Blue–Blue pairs were pushed before Blue–Red
   pairs, Red–Red after. Now all pushes are summed from the same positions and
   applied together.
3. **Blue thinks first.** With movement mirror-exact, two units reach a
   dash or a blow on the same tick; the first to act wins it. A fixed order gave
   Blue the tie; alternating by tick parity gave it to whoever owned the
   decisive tick (97% → 7% on one map). A coin flip per tick from the combat
   stream is fair over seeds and still deterministic per battle.
4. **Rotations tried in one order.** `KiteTurns` (kiting, stepping away,
   escaping) try +θ before −θ; reflection reverses rotation, so Red now uses
   `Turn.For(team)` (θ negated). `TurnToward` turning to face straight behind
   turns toward +x, which reflection keeps.
5. **Cell edges floored.** `ToCell` put both z = +8 and z = −8 in the cell on their +z side of the edge,
   which aren't mirror images; formation slots sit
   on edges, so near a rock one slot was blocked and nudged and its mirror
   wasn't. A point exactly on an edge now goes to the cell nearer the centre
   line on both halves.

**Ablation** (each fix removed alone, `-exp maps`, 400 per cell): removing (4)
leaves 3 cells off (worst 14 pts), (5) 2 cells (15 pts), (3) none in the mixed
armies but the pure-huskarl Charge case at 97%; (1) and (2) change little on
their own but are needed for exact mirror play, which the new test checks.

## Determinism

The coin flip draws from the existing combat stream (`b.rng`), once per tick
before units think, so a battle is still a pure function of its input; native
and WASM agree (goldens under 0.14.0, mutual-wipe seeds 4171 and 4436).
`TestMirrorImageBattles` (4 factions × 25 seeds, every plan and formation)
checks exact mirror play until first contact; `TestToCellMirrors` and
`TestMulIsOdd` pin the arithmetic. All three fail on the previous code.

## Risks / Trade-offs

- Every battle's outcome changes (new version); balance moves only by noise
  (side bias over random maps stays 50%; plan matrix checked).
- After contact a single battle isn't mirror-exact: the coin flip decides
  same-tick ties. That is the intended fairness: over seeds, not per battle.
