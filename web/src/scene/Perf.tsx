import { useFrame, useThree } from '@react-three/fiber';
import { PerformanceMonitor } from '@react-three/drei';
import { useBattle } from '../store.ts';
import { fps } from './perfStats.ts';

let frames = 0, since = performance.now();
const mountedAt = performance.now();

function FpsCounter() {
  const gl = useThree((s) => s.gl), scene = useThree((s) => s.scene), camera = useThree((s) => s.camera);
  // dev only: ms per frame rendered directly, independent of the page's frame loop
  // (embedded and background panes throttle requestAnimationFrame)
  const advance = useThree((s) => s.advance);
  if (import.meta.env.DEV) Object.assign(window, {
    benchScene: scene,
    battle: useBattle,
    // ms per full frame (frame loop subscribers, so post-processing too), synced with the GPU
    benchFrame: (n = 20) => {
      const ctx = gl.getContext(), px = new Uint8Array(4);
      const sync = () => ctx.readPixels(0, 0, 1, 1, ctx.RGBA, ctx.UNSIGNED_BYTE, px);
      advance(performance.now()); sync();
      const t = performance.now();
      for (let i = 0; i < n; i++) { advance(performance.now()); sync(); }
      return (performance.now() - t) / n;
    },
  });
  if (import.meta.env.DEV) (window as unknown as { benchRender: (n?: number) => number }).benchRender = (n = 20) => {
    const ctx = gl.getContext(), px = new Uint8Array(4);
    const sync = () => ctx.readPixels(0, 0, 1, 1, ctx.RGBA, ctx.UNSIGNED_BYTE, px); // waits for the GPU
    gl.render(scene, camera); sync();
    const t = performance.now();
    for (let i = 0; i < n; i++) { gl.render(scene, camera); sync(); }
    return (performance.now() - t) / n;
  };
  useFrame(() => {
    frames++;
    const now = performance.now();
    if (now - since >= 1000) { fps.value = Math.round((frames * 1000) / (now - since)); frames = 0; since = now; }
  });
  return null;
}

// Drops to Low once if High falls well below the display's own refresh rate
// (under 70% of it). Relative, not a fixed 45 fps: browsers that cap the
// frame rate (battery saver, embedded panes) aren't slow because of High.
export function Perf() {
  return (
    <>
      <FpsCounter />
      <PerformanceMonitor bounds={(refresh) => [refresh * 0.7, refresh]} flipflops={2}
        onDecline={({ fps: rate }) => {
          // not while warming up (shader compiles) or hidden (browsers throttle background pages).
          // A few fps means the page is throttled (a hidden pane that still says "visible"), not a slow GPU.
          if (performance.now() - mountedAt < 10_000 || document.visibilityState !== 'visible' || rate < 5) return;
          const s = useBattle.getState();
          if (s.quality !== 'high') return;
          // first render fewer pixels (keeps every effect), only then drop to Low
          if (!s.lowDpr) useBattle.setState({ lowDpr: true });
          else s.setQuality('low', true);
        }} />
    </>
  );
}
