// Sim events and unit states → particles, shockwaves and camera shake.
// Presentation only: called with each snapshot (like sound and floating
// numbers) and per unit per frame; nothing here feeds back into the battle.
// Event cues are named presets (fx.json); a skill's look can name its own
// preset for the events it causes (units/looks.json).
import { EF, EV, UF, type SimEvent, type Snapshot, type UnitSnap } from '../sim/snapshot.ts';
import type { UnitLoadout } from '../gamedata.ts';
import { skillWith } from '../units/looks.ts';
import { shockwave, spawn, type PoolKind, type Spawn } from './Particles.tsx';
import { addShake } from './shake.ts';
import presetsJson from './fx.json';

type Emitter = {
  type: 'sparks' | 'burst' | 'dust' | 'ring' | 'shake';
  pool?: PoolKind; n?: number; y?: number; forward?: number; color?: [number, number, number];
  speed?: number; size?: number; life?: number; grav?: number; up?: number; amount?: number;
};
export const PRESETS = presetsJson as unknown as Record<string, Emitter[]>;

// Skill actions started by an event this tick (dash, blink), for Units to play.
export const skillActions = new Map<number, { clip: string; rate: number }>();

// Recent hits per unit (direction away from the attacker), for the flinch in Units.
export const flinches = new Map<number, { dx: number; dz: number; t: number }>();

export type View = { fogOn: boolean; fogged: boolean; speed: number; high: boolean; units: UnitLoadout[]; bolt: (cls: number) => boolean };
export const viewOf = (s: { fogOn: boolean; phase: string; speed: number; quality: string; units: UnitLoadout[]; data?: { classes: { projectile?: string }[] } }): View =>
  ({ fogOn: s.fogOn, fogged: s.phase === 'countdown' || s.phase === 'fighting', speed: s.speed, high: s.quality === 'high', units: s.units,
    bolt: (cls) => s.data?.classes[cls]?.projectile === 'bolt' });
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

// Plays a preset at a unit; `from` is where the blow came from (sparks fly away from it).
export function play(name: string, u: UnitSnap, from?: UnitSnap) {
  const list = PRESETS[name];
  if (!list) {
    if (import.meta.env.DEV) console.warn(`no effect preset ${name}`);
    return;
  }
  for (const e of list) {
    const col = e.color ?? [1, 1, 1];
    switch (e.type) {
      case 'sparks': sparks(u, from, e.n ?? 8, col, e.speed ?? 4); break;
      case 'burst': burst(e.pool ?? 'glow', u.x + u.fx * (e.forward ?? 0), e.y ?? 1, u.z + u.fz * (e.forward ?? 0), e.n ?? 1, col, e.speed ?? 0, e.size ?? 0.4, e.life ?? 0.5, e.grav ?? 0); break;
      case 'dust': dust(u.x, u.z, e.n ?? 8, e.size ?? 0.8, e.up ?? 0.6); break;
      case 'ring': shockwave(u.x, u.z); break;
      case 'shake': addShake(e.amount ?? 0.2); break;
    }
  }
}

// The preset a skill names for an effect kind, else the event's default.
function presetFor(units: UnitLoadout[], unit: number | undefined, kind: string, fallback: string) {
  const l = unit !== undefined ? units[unit] : undefined;
  return (l && skillWith(l.cls, l.skills, kind)?.look.fx) || fallback;
}

export function onSnapshot(snap: Snapshot, events: SimEvent[], v: View) {
  budget = SNAP_BUDGET;
  scale = (v.high ? 1 : 0.5) / Math.sqrt(v.speed);
  for (const e of events) {
    const u = snap.units[e.unit], o = e.other >= 0 ? snap.units[e.other] : undefined;
    if (!u) continue;
    if (e.kind === EV.dash || e.kind === EV.blink) { // animation state is kept for hidden units too
      const l = v.units[e.unit], s = l && skillWith(l.cls, l.skills, e.kind === EV.dash ? 'gapCloser' : 'escape');
      if (s?.look.action) skillActions.set(e.unit, { clip: s.look.action, rate: s.look.rate ?? 1 });
    }
    if (!visible(u, v)) continue;
    switch (e.kind) {
      case EV.hit: {
        if (o) { const l = Math.hypot(u.x - o.x, u.z - o.z) || 1; flinches.set(e.unit, { dx: (u.x - o.x) / l, dz: (u.z - o.z) / l, t: performance.now() / 1000 }); }
        const kind = e.flags & EF.flank ? '.flank' : e.flags & EF.strong ? '.strong' : '';
        play((o && v.bolt(o.cls) ? 'hit.magic' : 'hit') + kind, u, o);
        break;
      }
      case EV.execute: play(presetFor(v.units, e.other, 'finisher', 'execute'), u, o); break;
      case EV.death: play('death', u); break;
      case EV.dash: play(presetFor(v.units, e.unit, 'gapCloser', 'dash'), u); break;
      case EV.blink: play(presetFor(v.units, e.unit, 'escape', 'blink'), u); break;
      case EV.frost: play(presetFor(v.units, e.other, 'slowOnHit', 'frost'), u, o); break;
      case EV.blast: play(presetFor(v.units, e.other, 'splashOnHit', 'blast'), u, o); break;
      case EV.shot: if (v.bolt(u.cls)) play('shot.bolt', u); break;
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
      const v: View = { fogOn: false, fogged: false, speed: 1, high: true, units: [], bolt: (cls) => cls === 2 };
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
