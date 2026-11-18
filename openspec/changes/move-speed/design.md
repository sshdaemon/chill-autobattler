# Design

## Context
See proposal.md. The packs' root-motion exports (`*_RM.glb`) give 0.97, 5.36 and 8.25 m/s for walk, jog and sprint, but those are for the pack's mannequin. Our outfit skeleton has longer legs (pelvis 0.95 m vs 0.92 m), so on our characters a planted foot moves back at 1.04, 6.24 and 9.57 m/s. Those are the speeds that keep feet planted, and the ones recorded. They were measured in the browser by playing each clip on a character at rest and taking the median backward speed of a foot within 1.5 cm of its lowest point.

## Decisions

### Speed source
The ground speed comes from the two latest snapshots: `|b − a| / 0.05 s` of sim time. The mixer runs in game time × playback speed, so both are in sim seconds. It's smoothed with an exponential filter (time constant 0.12 s of sim time) so path corners don't jitter the legs.
- **Rejected: computing from stats and flags.** The sim's own positions already include Swift, water, frost and the sharp-turn slowdown, and stay right if rules change.

### Walk–jog blend, kept in step
For a speed `v` (in m/s, divided by the unit's size `s`), choose a blend `k ∈ [0, 1]` between walk (w) and jog (j). With stride `S(k) = lerp(Sw, Sj, k)` and cycle `D(k) = lerp(Dw, Dj, k)`:
- solve `S(k)/D(k) = v`, which has a closed form because both are linear in `k`;
- play walk at `Dw/D(k)` and jog at `Dj/D(k)`, so both finish a cycle together;
- weight them `1−k` and `k`;
- re-align walk to jog's phase every frame.

Below walk speed the walk plays alone at `v/0.97` (min 0.5×); above jog speed the jog plays alone, up to 1.4×. Leaving or entering the move state crossfades both clips together. Kiting uses negative time scales.
- **Rejected: one clip at `v / clipSpeed`.** At 3.2 m/s that plays the jog at 0.6×: no sliding, but slow-motion legs.

### Sprint for plain dashes
A unit dashing without a skill action plays the sprint clip at `min(1.6, v / 8.25)`. A dash is 12–16 m/s, so the legs still undershoot; at that speed it reads as a lunge, which is accepted.

## Result
An offline check (a character driven at a given speed with the same blend function, planted-foot speed over the ground measured at 240 Hz) gives 0–3% of the ground speed across unit types and builds, from 0.96 to 4.48 m/s. The old fixed jog gave 44–550%. Blends land at walk 1.1–1.27× and jog 0.76–0.89×.

## Risks / Trade-offs
- **[Risk]** Walk and jog have different arm poses, so a 50/50 blend may look mushy. → Mitigation: check in the preview and a battle. If it looks wrong, bias the blend towards jog.
