# Design

## Context

- `App.tsx` sets up a `Canvas` with a flat background colour, a hemisphere light and one directional light, whose shadow box covers the whole arena (±18 m, 2048² shadow map).
- The renderer uses ACES tone mapping by default (R3F).
- Fog of war is a transparent plane at y = 0.06. Health bars and stars are billboards and HTML.

## Goals / Non-Goals

**Goals:**
- a clearly richer, more grounded look;
- the High budget: 60 fps on a mid-range laptop;
- a safe fallback.

**Non-Goals:** new assets, terrain, water, particles, characters (later steps).

## Decisions

**A code-generated sky, not an HDRI file.**
- three's `Sky` (Preetham model) draws the dome, and drei's `<Environment>` renders a light probe from children: a `Sky` plus a few `Lightformer`s (a warm key, a cool fill, a ground bounce).
- It's baked once (`frames={1}`), costs nothing per frame, and needs no CDN or download.
- *Alternative:* an HDRI from Poly Haven (CC0) looks a bit richer but adds about 1–4 MB and an asset pipeline. It's left for step 2, where textures arrive anyway.

**The sun's shadows follow the camera.**
- The directional light and its target move with `camView` (the RTS camera's look-at point).
- The orthographic shadow box shrinks to about ±14 m around it on High (2048²), so the same shadow map covers a smaller area and is about twice as sharp.
- Shadow positions are snapped to texel-size steps so edges don't shimmer while panning.
- Softer edges come from the shadow blur radius on High. three r187 removed `PCFSoftShadowMap` (it now falls back to PCF), so the radius on the standard PCF map is used. drei's `<SoftShadows>` (PCSS) patches shader chunks globally and can't be toggled per quality setting without recompiling every material.

**Post-processing** with an `EffectComposer` (High only):
- `N8AO`: quality "medium", radius about 1.2 m, half resolution. It's cheap and looks good at RTS distances.
- `Bloom`: luminance threshold about 1, so only emissive and un-tone-mapped materials glow (orbs, bolts, the selection ring and effects).
- `ToneMapping`: ACES, as before, but now inside the composer.
- `HueSaturation` and `BrightnessContrast` for a mild grade, and `Vignette`: subtle.
- `TiltShift2`: very light. It's a dev toggle and off by default, so it can be judged in playtests.

**Fog and UI.** The fog overlay is part of the scene, so AO and bloom apply to it naturally. Bloom's threshold keeps it from glowing. HTML (floating numbers, stars) sits above the canvas, unaffected.

**Quality and fallback.**
- The store holds `quality: 'high' | 'low'` (localStorage, default High) and an `autoLowered` flag.
- drei's `<PerformanceMonitor>` uses bounds relative to the measured display refresh: it declines below 70% of it. A fixed 45 fps wrongly downgraded browsers that cap at 30 fps; the embedded test pane measured about 33 fps on Low and about 31 on High. On decline it switches to Low once per session and shows a note in the dev panel.
- The dev panel shows a quality selector and an fps readout averaged over 1 s.

**How determinism is preserved and tested.** Rendering only. Nothing in `sim/`, the server or the wire changes, so the golden replays run unchanged. Checks happen in the browser: before/after screenshots, the fps readout on High, the switch to Low and back, and a full match verified fair.

## Risks / Trade-offs

- [Post-processing costs on weak GPUs] → It's High only, with the automatic fallback and the half-resolution AO.
- [Bloom can blur the readable bits] → A high threshold limits it to magical emissives; floating numbers are HTML, outside it.
- [The tilt-shift might hurt readability] → Off by default; it's an experiment for playtests.

## Pixel density

On Retina screens the canvas rendered at 2× density, so post-processing worked on 4× the pixels. The canvas is now capped: High uses `dpr={[1, 1.5]}`, Low uses `dpr={1}`. The composer uses 2× multisampling instead of 4×.
