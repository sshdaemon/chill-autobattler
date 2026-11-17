// Camera shake: a trauma value (0..1) that decays; the offset grows with its
// square, so small shakes stay subtle. Read by RtsCamera each frame.
let trauma = 0;

export function addShake(amount: number) { trauma = Math.min(1, trauma + amount); }

const MAX = 0.35; // metres at full trauma

export function shakeOffset(dt: number, t: number): [number, number, number] {
  trauma = Math.max(0, trauma - dt * 2.5);
  const s = trauma * trauma * MAX;
  if (s === 0) return [0, 0, 0];
  return [Math.sin(t * 47.3) * s, Math.sin(t * 39.1 + 1.3) * s * 0.6, Math.sin(t * 53.7 + 2.1) * s];
}
