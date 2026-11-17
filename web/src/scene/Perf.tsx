import { useFrame } from '@react-three/fiber';
import { PerformanceMonitor } from '@react-three/drei';
import { useBattle } from '../store.ts';
import { fps } from './perfStats.ts';

let frames = 0, since = performance.now();
const mountedAt = performance.now();

function FpsCounter() {
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
        onDecline={() => {
          // not while warming up (shader compiles) or hidden (browsers throttle background pages)
          if (performance.now() - mountedAt < 10_000 || document.visibilityState !== 'visible') return;
          const s = useBattle.getState();
          if (s.quality === 'high') s.setQuality('low', true);
        }} />
    </>
  );
}
