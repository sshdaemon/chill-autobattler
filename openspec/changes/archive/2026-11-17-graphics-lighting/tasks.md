# Tasks

## 1. Roadmap

- [x] 1.1 `docs/graphics-roadmap.md` with the 4 steps, budgets and asset rules; verify it's linked from the README

## 2. Lighting

- [x] 2.1 Code-generated sky, environment lighting (Sky + Lightformers baked once), a sun matching the sky; verify in the browser (before/after)
- [x] 2.2 Shadows follow the camera target with texel snapping, plus a softer blur radius on High; verify that panning doesn't shimmer and shadows are sharper near the camera

## 3. Post-processing

- [x] 3.1 The `EffectComposer` with N8AO, Bloom (threshold), ToneMapping, a grade and a vignette, plus an optional TiltShift (dev toggle, off); verify that units stay readable, the fog still works, and orbs and bolts glow

## 4. Quality

- [x] 4.1 Quality in the store (localStorage, default High), the dev panel selector and fps readout, and `PerformanceMonitor` auto-lowering with a note; verify that switching works live, the choice persists after a reload, and the fps on High is reported

## 5. Wrap-up

- [x] 5.1 README (graphics roadmap link, quality setting); verify `pnpm test`, the web build, and a full match verified fair

## Workflow follow-up

- Archive the change once merged (auto-merge after green CI).
