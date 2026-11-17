# Design

## Context

See proposal.md and `docs/balance-findings.md`. Tuning was done with `tools/balance` in quick mode, one lever at a time:
- 300 battles per matchup cell;
- 200 per composition cell;
- 2000 paired battles per skill.

The final numbers are from the full report.

## Goals / Non-Goals

**Goals:**
- each counter wins about 65–75%, not 100%;
- no composition averages above about 60% against random Red;
- every skill adds about +3 to +10 points.

**Non-Goals:** run-level balance, new content.

## Decisions

**Fix kiting mechanically, not only with numbers.**
- Numbers alone failed: archer health 60 and speed 3.8 still lost 92% to warriors.
- A per-tick trace showed why. Kiting archers turned round first (the sharp-turn penalty cut them to 30% speed), so the distance to their warrior kept shrinking, and the retreat only began at 3.5 m.
- Backpedalling at full speed while facing the threat (`BackpedalSpeed = 1`, no turn penalty while kiting) plus starting at 4.5 m fixed the mechanics. The stat and matchup changes then set the strength.

**Instant kills finish off the wounded.**
- A threshold (victim below 40% health) keeps the skills' identity ("finisher") but stops them deleting full-health units.
- The higher chances keep the skills worth taking (+4.7 and +6.8 points).
- *Alternatives:* a lower flat chance (still swingy) or a cooldown (adds state).

**Make dead-weight skills do something.**
- **Swift:** speed alone doesn't help a melee unit, so it also attacks 20% faster (cooldown and swing ×0.8).
- **Arcing shot:** situational, so it also adds +10% damage.
- **Frost:** always slows.
- **Blink:** 8 s cooldown.
- **Long shot:** +25%.

**How determinism is preserved and tested.**
- This is a rule change, so `sim.Version` goes to 0.4.0 and the golden replays are regenerated (the version guard enforced the bump).
- The mutual-wipe golden cases move to seeds that still draw under the new rules (1069, 1710).
- Rule tests now pin the new numbers:
  - Executioner never fires on a full-health victim, and fires at about 35% on one at 38% health;
  - Frost always slows;
  - Swift's attack speed, Arcing shot's damage, and the 8 s Blink cooldown.
- The golden set runs natively and under WASM. The server tests still drive whole runs.

## Risks / Trade-offs

- [Balance is measured against random Red at match 1 only] → Run-level balance (veterans versus scaled Red) is the next lab experiment.
- [Kiting archers now face away from where they walk, which looks odd] → It's presentation only; a backward-jog animation can come later.

## Migration Plan

0.3.0 battles in progress when deploying 0.4.0 can't verify (the server rejects the version) and count as abandoned. There are no live players yet.
