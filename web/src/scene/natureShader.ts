// Shared bits for the battlefield shaders: value noise, and the ponds as a
// union of the sim's water circles. Presentation only.
import * as THREE from 'three';
import { OBSTACLE, type Obstacle } from '../sim/snapshot.ts';

export const MAX_PONDS = 16;

export const NOISE_GLSL = /* glsl */ `
float n_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }
float vnoise(vec2 p) {
  vec2 i = floor(p), f = fract(p);
  vec2 u = f * f * (3.0 - 2.0 * f);
  return mix(mix(n_hash(i), n_hash(i + vec2(1, 0)), u.x), mix(n_hash(i + vec2(0, 1)), n_hash(i + vec2(1, 1)), u.x), u.y);
}
float fbm(vec2 p) { return vnoise(p) * 0.6 + vnoise(p * 2.03 + 7.1) * 0.3 + vnoise(p * 4.1 + 3.7) * 0.1; }
`;

// pondDist: signed distance to the nearest pond edge (negative inside water)
export const POND_GLSL = /* glsl */ `
uniform vec3 uPonds[${MAX_PONDS}];
uniform int uPondCount;
float pondDist(vec2 p) {
  float d = 1e3;
  for (int i = 0; i < ${MAX_PONDS}; i++) {
    if (i >= uPondCount) break;
    d = min(d, length(p - uPonds[i].xy) - uPonds[i].z);
  }
  return d;
}
`;

// ponds as (x, z, r); the shaders use world (x, z) directly
export function pondUniforms(obstacles: Obstacle[]) {
  const ponds = obstacles.filter((o) => o.kind === OBSTACLE.water).slice(0, MAX_PONDS);
  const list = Array.from({ length: MAX_PONDS }, (_, i) => new THREE.Vector3(ponds[i]?.x ?? 0, ponds[i]?.z ?? 0, ponds[i]?.r ?? 0));
  return { uPonds: { value: list }, uPondCount: { value: ponds.length } };
}

// A small stable hash for picking model variants from a position.
export function hash2(x: number, z: number, salt = 0) {
  let h = Math.imul(Math.round(x * 64) ^ 0x9e3779b9, 0x85ebca6b) ^ Math.imul(Math.round(z * 64) + salt * 7919, 0xc2b2ae35);
  h ^= h >>> 15; h = Math.imul(h, 0x2c1b3c6d); h ^= h >>> 12;
  return (h >>> 0) / 4294967296;
}

// Seeded generator for scenery placement (mulberry32).
export function rng(seed: number) {
  return () => {
    seed = (seed + 0x6d2b79f5) | 0;
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}
