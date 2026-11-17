# Tasks

## 1. Particles

- [x] 1.1 Pooled GPU particles (glow and soft pools, ring buffer, partial uploads) and the Blast shockwave rings; verify no per-frame allocation and the budget held

## 2. Event cues

- [x] 2.1 `effects.ts`: hit (flank, strong, mage), execute, death, dash, blink, frost, blast and shot cues, hidden under fog, rate-limited by playback speed; verify each event in a real match
- [x] 2.2 Continuous cues: running dust, wading splashes, dash trail, frost motes, and projectile trails

## 3. Death and camera

- [x] 3.1 Bodies fade into the ground (per-unit cloned materials)
- [x] 3.2 Camera shake on executes and on the last kill

## 4. Wrap-up

- [x] 4.1 Verify the GPU time per frame on High in a busy fight (timer queries), screenshots, `pnpm test`, the web build, and a full match verified fair
