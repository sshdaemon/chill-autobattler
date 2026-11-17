// Sim events and unit states → particles, shockwaves and camera shake.
// Presentation only: called with each snapshot (like sound and floating
// numbers) and per unit per frame; nothing here feeds back into the battle.
import { EF, EV, UF, type SimEvent, type Snapshot, type UnitSnap } from '../sim/snapshot.ts';
import { shockwave, spawn, type PoolKind, type Spawn } from './Particles.tsx';
import { addShake } from './shake.ts';

// Recent hits per unit (direction away from the attacker), for the flinch in Units.
export const flinches = new Map<number, { dx: number; dz: number; t: number }>();

export type View = { fogOn: boolean; fogged: boolean; speed: number; high: boolean };
export const viewOf = (s: { fogOn: boolean; phase: string; speed: number; quality: string }): View =>
  ({ fogOn: s.fogOn, fogged: s.phase === 'countdown' || s.phase === 'fighting', speed: s.speed, high: s.quality === 'high' });

const CLS_MAGE = 2;
const SNAP_BUDGET = 160; // particles per snapshot, at most

let budget = SNAP_BUDGET;
let scale = 1; // fewer particles at faster playback and on Low
const rnd = (a: number, b: number) => a + Math.random() * (b - a);

function emit(kind: PoolKind, n: number, f: (i: number) => Spawn) {
  const count = Math.min(budget, Math.max(1, Math.round(n * scale)));
  budget -= count;
  for (let i = 0; i < count; i++) spawn(kind, f(i));
}

export const visible = (u: UnitSnap, v: View) => u.team === 0 || !v.fogOn || !v.fogged || (u.flags & UF.seen) !== 0;

// Sparks flying from the attacker's side.
function sparks(u: UnitSnap, from: UnitSnap | undefined, n: number, col: [number, number, number], speed = 4) {
  let dx = 0, dz = 0;
  if (from) { dx = u.x - from.x; dz = u.z - from.z; const l = Math.hypot(dx, dz) || 1; dx /= l; dz /= l; }
  emit('glow', n, () => {
    const a = rnd(0, Math.PI * 2), s = rnd(0.4, 1) * speed;
    return {
      x: u.x + rnd(-0.1, 0.1), y: rnd(0.9, 1.3), z: u.z + rnd(-0.1, 0.1),
      vx: Math.cos(a) * s * 0.55 + dx * s, vy: rnd(0.6, 2.4), vz: Math.sin(a) * s * 0.55 + dz * s,
      life: rnd(0.25, 0.45), size: rnd(0.22, 0.34), r: col[0], g: col[1], b: col[2], grav: -9, streak: 1,
    };
  });
}

function burst(kind: PoolKind, x: number, y: number, z: number, n: number, col: [number, number, number], speed: number, size: number, life: number, grav = 0) {
  emit(kind, n, () => {
    const a = rnd(0, Math.PI * 2), e = rnd(-0.3, 1), s = rnd(0.5, 1) * speed;
    return { x, y, z, vx: Math.cos(a) * s, vy: e * s, vz: Math.sin(a) * s, life: rnd(0.6, 1) * life, size: rnd(0.7, 1.2) * size, r: col[0], g: col[1], b: col[2], grav };
  });
}

function dust(x: number, z: number, n: number, size = 0.8, up = 0.6) {
  emit('soft', n, () => ({
    x: x + rnd(-0.2, 0.2), y: 0.1, z: z + rnd(-0.2, 0.2),
    vx: rnd(-0.5, 0.5), vy: rnd(0.2, up), vz: rnd(-0.5, 0.5),
    life: rnd(0.5, 0.9), size: rnd(0.7, 1.2) * size, r: 0.55, g: 0.47, b: 0.36, grav: -0.3,
  }));
}

function splash(x: number, z: number, n: number) {
  emit('soft', n, () => ({
    x: x + rnd(-0.25, 0.25), y: 0.05, z: z + rnd(-0.25, 0.25),
    vx: rnd(-0.9, 0.9), vy: rnd(1.2, 2.4), vz: rnd(-0.9, 0.9),
    life: rnd(0.35, 0.6), size: rnd(0.2, 0.32), r: 0.85, g: 0.93, b: 1, grav: -9,
  }));
}

export function onSnapshot(snap: Snapshot, events: SimEvent[], v: View) {
  budget = SNAP_BUDGET;
  scale = (v.high ? 1 : 0.5) / Math.sqrt(v.speed);
  for (const e of events) {
    const u = snap.units[e.unit], o = e.other >= 0 ? snap.units[e.other] : undefined;
    if (!u || !visible(u, v)) continue;
    switch (e.kind) {
      case EV.hit: {
        if (o) { const l = Math.hypot(u.x - o.x, u.z - o.z) || 1; flinches.set(e.unit, { dx: (u.x - o.x) / l, dz: (u.z - o.z) / l, t: performance.now() / 1000 }); }
        const flank = e.flags & EF.flank, strong = e.flags & EF.strong;
        const col: [number, number, number] = o?.cls === CLS_MAGE ? [1.6, 0.9, 2.6] : flank ? [3, 2.2, 0.6] : strong ? [3, 1.4, 0.4] : [2.2, 1.9, 1.5];
        sparks(u, o, flank ? 18 : strong ? 12 : 8, col, flank ? 5.5 : 4);
        if (flank) burst('glow', u.x, 1.15, u.z, 1, [3, 2.2, 0.8], 0, 0.9, 0.18);
        break;
      }
      case EV.execute:
        burst('glow', u.x, 1.1, u.z, 22, [3, 0.5, 0.4], 3.5, 0.4, 0.55, -4);
        burst('glow', u.x, 1.1, u.z, 1, [3, 0.6, 0.5], 0, 1.4, 0.25);
        addShake(0.25);
        break;
      case EV.death:
        dust(u.x, u.z, 10, 1.1, 0.9);
        break;
      case EV.dash:
        dust(u.x, u.z, 8, 1, 0.8);
        break;
      case EV.blink:
        burst('glow', u.x, 1, u.z, 18, [1.8, 1.1, 3], 2.5, 0.36, 0.5);
        burst('glow', u.x, 1, u.z, 1, [1.6, 1, 3], 0, 1.6, 0.3);
        break;
      case EV.frost:
        burst('glow', u.x, 1.1, u.z, 12, [1.2, 2.2, 3], 2, 0.3, 0.6, -3);
        break;
      case EV.blast:
        burst('glow', u.x, 0.9, u.z, 20, [3, 1.5, 0.5], 4, 0.38, 0.5, -2);
        shockwave(u.x, u.z);
        break;
      case EV.shot:
        if (u.cls === CLS_MAGE) burst('glow', u.x + u.fx * 0.5, 1.5, u.z + u.fz * 0.5, 6, [1.8, 1.2, 3], 0.8, 0.4, 0.4);
        break;
    }
  }
}

// Per unit per frame: dust when running, splashes when wading, a dust trail
// while dashing, ice motes while slowed.
const timers = new Map<number, number>();
export function feet(index: number, x: number, z: number, flags: number, dt: number, v: View, u: UnitSnap) {
  if (flags & UF.dead || !visible(u, v)) return;
  const wet = flags & UF.wet, dashing = flags & UF.dashing, moving = flags & UF.moving, slowed = flags & UF.slowed;
  if (!moving && !dashing && !slowed) return;
  const every = dashing ? 0.05 : wet ? 0.18 : moving ? 0.22 : 0.3;
  const t = (timers.get(index) ?? Math.random() * every) - dt * v.speed;
  if (t > 0) { timers.set(index, t); return; }
  timers.set(index, every / (v.high ? 1 : 0.5));
  budget = Math.max(budget, 4); // continuous cues aren't starved by a busy snapshot
  scale = v.high ? 1 : 0.5;
  if (wet && (moving || dashing)) splash(x, z, 3);
  else if (moving || dashing) dust(x, z, dashing ? 2 : 1, dashing ? 0.9 : 0.65, 0.4);
  if (slowed) emit('glow', 1, () => ({ x: x + rnd(-0.3, 0.3), y: rnd(0.4, 1.6), z: z + rnd(-0.3, 0.3), vy: 0.35, life: 0.8, size: 0.2, r: 1.1, g: 1.9, b: 2.6 }));
}

// Projectile trails, dropped every frame at each shot's position.
export function trail(x: number, y: number, z: number, bolt: boolean, v: View) {
  if (!v.high && Math.random() < 0.5) return;
  spawn(bolt ? 'glow' : 'soft', bolt
    ? { x, y, z, life: 0.3, size: 0.32, r: 1.5, g: 1, b: 2.6 }
    : { x, y, z, life: 0.25, size: 0.12, r: 0.92, g: 0.92, b: 0.9 });
}

// The last kill: a stronger shake.
export function lastKill() { addShake(0.55); }

// Dev only: every cue side by side across the arena, for checking the look.
if (import.meta.env.DEV) {
  Object.assign(window, {
    fxDemo: () => {
      const v: View = { fogOn: false, fogged: false, speed: 1, high: true };
      const unit = (x: number, cls = 0): UnitSnap => ({ team: 0, cls, x, z: 4, fx: 0, fz: -1, hp: 1, maxHp: 1, flags: 0, target: -1, skills: 0, kills: 0, dealt: 0 });
      const units = [-9, -6, -3, 0, 3, 6, 9].map((x) => unit(x)).concat([unit(-9, 2), unit(-6, 2), unit(3, 2)]);
      const ev = (kind: number, u: number, other = -1, flags = 0): SimEvent => ({ kind, unit: u, other, amount: 10, flags });
      const snap = { tick: 0, result: 0, units, shots: [], events: [] } as Snapshot;
      units[7].z = units[8].z = units[9].z = 6;
      onSnapshot(snap, [ev(EV.hit, 0, 7), ev(EV.hit, 1, 8, EF.flank), ev(EV.execute, 2, 0), ev(EV.blink, 3), ev(EV.frost, 4, 9), ev(EV.blast, 5, 9), ev(EV.death, 6), ev(EV.shot, 9)], v);
      for (let i = 0; i < 6; i++) { feet(100, -9, 8, UF.moving, 1, v, unit(-9)); feet(101, 0, 8, UF.moving | UF.wet, 1, v, unit(0)); feet(102, 9, 8, UF.dashing, 1, v, unit(9)); }
    },
  });
}
