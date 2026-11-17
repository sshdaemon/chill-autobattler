# Design

## Particle pools

`Particles.tsx` holds two `InstancedMesh` pools of camera-facing quads:
- **Glow:** additive, 1,200 particles on High and 600 on Low.
- **Soft:** normal blending, 800 on High and 400 on Low.

Each particle has instanced attributes: start position, velocity, birth time, life, size, colour and gravity. The vertex shader computes `p = p0 + v·t + ½·g·t²` and billboards the quad. The fragment shader draws a soft disc (or a streak for sparks, stretched along the velocity) and fades by age.

- **Spawning** writes the next ring-buffer slot and marks only the touched attribute range for upload (`addUpdateRange`).
- **No per-frame allocation;** dead particles just expire in the shader.

## Effects from events

`effects.ts` is called from the store with each new snapshot and its events, next to sound and floating numbers:
- It maps each event to spawns (counts and colours as in the proposal).
- It skips units hidden by fog: a Red unit without the `seen` flag while fog is on.
- **Budget:** at most 160 particles per snapshot, with counts scaled by `1/√speed`. At 4× playback, a hit spawns half as many sparks.

## Continuous cues

`Units.tsx` already runs per unit each frame. It calls `effects.feet(unit, dt)`, which keeps a per-unit timer:
- **Running:** a dust puff every 0.22 s, small and brown.
- **Wading:** a splash (droplets and a white puff) every 0.18 s instead.
- **Dashing:** dust every 0.05 s.
- **Slowed by Frost:** a pale-blue mote every 0.3 s.

`Shots.tsx` drops a trail particle at each projectile's position every frame: faint white for arrows, violet glow for bolts.

## Blast shockwave

`Particles.tsx` also keeps 8 pooled ring meshes. Each expands from 0.3 m to 3 m over 0.45 s with an additive orange-white edge.

## Death

When a unit has been dead for 1.2 s, its model's materials are cloned once, so other units sharing them aren't affected. Over 1.5 s they fade to transparent while the body sinks 0.35 m, then the model is hidden. The ring and health bar already hide on death.

## Camera shake

`shake.ts` keeps a trauma value, decaying at 2.5 per second:
- an execute adds 0.25;
- the last kill (the outcome arriving) adds 0.55.

`RtsCamera` offsets the camera by `trauma²` × 0.35 m along two sine-noise axes.

## Notes from building it

- **Sizes:** particle sizes are about 2.5–3× "real", so they read from the strategy camera 30 m away. Real-size sparks were under a pixel.
- **Spark quads:** stretched along their velocity. The side vector must keep the quad's winding, or the GPU culls it.
- **Cost:** GPU time per frame on High stayed at 12–14.5 ms in fights (Apple M3, 1536×1152). The particles themselves are within measurement noise.
