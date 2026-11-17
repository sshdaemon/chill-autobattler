import { TICK_MS, useBattle } from '../store.ts';

// Where to draw things between the two latest sim ticks (0 = prev, 1 = curr).
export function alpha() {
  const { currAt, speed } = useBattle.getState();
  return Math.min(1, (performance.now() - currAt) / (TICK_MS / speed));
}
export const lerp = (a: number, b: number, k: number) => a + (b - a) * k;
