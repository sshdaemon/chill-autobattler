// Decodes the sim's Int32 snapshots. Layout: sim/snapshot.go (keep in sync).
const FX = 65536;

export type UnitSnap = {
  team: number; cls: number; x: number; z: number; fx: number; fz: number;
  hp: number; maxHp: number; flags: number; target: number; skills: number; kills: number; dealt: number;
};
export type ShotSnap = { bolt: boolean; lob: boolean; sx: number; sz: number; ax: number; az: number; t: number; dur: number; victim: number };
export type SimEvent = { kind: number; unit: number; other: number; amount: number; flags: number };
export type Snapshot = { tick: number; result: number; units: UnitSnap[]; shots: ShotSnap[]; events: SimEvent[] };
export type Obstacle = { kind: number; x: number; z: number; r: number };

export const UF = { dead: 1, moving: 2, swinging: 4, wet: 8, slowed: 16, kiting: 32, dashing: 64, staggered: 128, seen: 256 };
export const EV = {
  hit: 1, death: 2, dash: 3, blink: 4, frost: 5, blast: 6, execute: 7, shot: 8,
  miss: 9, fear: 10, burst: 11, jet: 12, burn: 13, rally: 14, guard: 15, brace: 16, curse: 17, // miss, fear to rally and curse: roadmap 3.2, no presets yet
};
export const EF = { flank: 1, strong: 2, weak: 4 };
export const OBSTACLE = { water: 0, rock: 1, tree: 2 };
export const RESULT = { ongoing: 0, blue: 1, red: 2, draw: 3 };

const i32 = (buf: ArrayBuffer) => new Int32Array(buf);

export function decodeSnapshot(buf: ArrayBuffer): Snapshot {
  const a = i32(buf);
  let p = 5;
  const [tick, result, nu, ns, ne] = a;
  const units: UnitSnap[] = [];
  for (let i = 0; i < nu; i++, p += 13) {
    units.push({ team: a[p], cls: a[p + 1], x: a[p + 2] / FX, z: a[p + 3] / FX, fx: a[p + 4] / FX, fz: a[p + 5] / FX,
      hp: a[p + 6] / FX, maxHp: a[p + 7] / FX, flags: a[p + 8], target: a[p + 9], skills: a[p + 10], kills: a[p + 11], dealt: a[p + 12] / FX });
  }
  const shots: ShotSnap[] = [];
  for (let i = 0; i < ns; i++, p += 8) {
    shots.push({ bolt: !!(a[p] & 1), lob: !!(a[p] & 2), sx: a[p + 1] / FX, sz: a[p + 2] / FX, ax: a[p + 3] / FX, az: a[p + 4] / FX, t: a[p + 5], dur: a[p + 6], victim: a[p + 7] });
  }
  const events: SimEvent[] = [];
  for (let i = 0; i < ne; i++, p += 5) events.push({ kind: a[p], unit: a[p + 1], other: a[p + 2], amount: a[p + 3] / FX, flags: a[p + 4] });
  return { tick, result, units, shots, events };
}

export function decodeMap(buf: ArrayBuffer): Obstacle[] {
  const a = i32(buf), out: Obstacle[] = [];
  for (let p = 0; p < a.length; p += 4) out.push({ kind: a[p], x: a[p + 1] / FX, z: a[p + 2] / FX, r: a[p + 3] / FX });
  return out;
}
