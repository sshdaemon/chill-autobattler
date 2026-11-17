import { useEffect, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

// Pooled GPU particles. Each pool is one instanced mesh of camera-facing
// quads in a ring buffer: spawning writes the next slot, the shader animates
// it (p = p0 + v·t + ½·g·t²) and fades it by age. Nothing is allocated per
// frame. Presentation only.

export type PoolKind = 'glow' | 'soft';
const SIZES: Record<PoolKind, number> = { glow: 1200, soft: 800 };

type Pool = {
  geo: THREE.InstancedBufferGeometry;
  attrs: Record<'p0' | 'vel' | 'col', THREE.InstancedBufferAttribute> & Record<'birth' | 'life' | 'size' | 'grav' | 'streak', THREE.InstancedBufferAttribute>;
  next: number;
  size: number;
  lo: number; // dirty range this frame (slots), or lo > hi when clean
  hi: number;
  wrapped: boolean;
};

const pools: Partial<Record<PoolKind, Pool>> = {};
let slow = 1; // dev only: window.fxSlow(0.05) runs particles in slow motion
export const now = () => (performance.now() / 1000) * slow;
if (import.meta.env.DEV) Object.assign(window, { fxSlow: (k: number) => { slow = k; }, fxPools: pools, fxSpawn: (k: PoolKind, s: Spawn) => spawn(k, s) });

export type Spawn = {
  x: number; y: number; z: number;
  vx?: number; vy?: number; vz?: number;
  life: number; size: number;
  r: number; g: number; b: number;
  grav?: number; // m/s², negative pulls down
  streak?: number; // > 0 stretches the quad along its velocity (sparks)
};

export function spawn(kind: PoolKind, s: Spawn) {
  const p = pools[kind];
  if (!p) return;
  const i = p.next;
  p.next = (p.next + 1) % p.size;
  if (p.next === 0) p.wrapped = true;
  const a = p.attrs;
  a.p0.setXYZ(i, s.x, s.y, s.z);
  a.vel.setXYZ(i, s.vx ?? 0, s.vy ?? 0, s.vz ?? 0);
  a.col.setXYZ(i, s.r, s.g, s.b);
  a.birth.setX(i, now());
  a.life.setX(i, s.life);
  a.size.setX(i, s.size);
  a.grav.setX(i, s.grav ?? 0);
  a.streak.setX(i, s.streak ?? 0);
  p.lo = Math.min(p.lo, i);
  p.hi = Math.max(p.hi, i);
}

const VERT = /* glsl */ `
  attribute vec3 p0, vel, col;
  attribute float birth, life, size, grav, streak;
  uniform float uTime;
  varying vec2 vUv;
  varying vec3 vCol;
  varying float vFade;
  void main() {
    float t = uTime - birth;
    if (t < 0.0 || t > life) { gl_Position = vec4(2.0, 2.0, 2.0, 1.0); return; }
    vec3 v = vel + vec3(0.0, grav * t, 0.0);
    vec3 p = p0 + vel * t + vec3(0.0, 0.5 * grav * t * t, 0.0);
    vec4 mv = viewMatrix * vec4(p, 1.0);
    float k = t / life;
    float s = size * (0.6 + 0.4 * (1.0 - k));
    vec2 off = position.xy * s;
    if (streak > 0.0) {
      vec2 dir = (viewMatrix * vec4(v, 0.0)).xy;
      float len = length(dir);
      if (len > 1e-4) {
        dir /= len;
        vec2 side = vec2(dir.y, -dir.x); // keeps the quad's winding (x→side, y→dir), or it gets culled
        off = dir * position.y * s * (1.0 + streak * len * 0.06) + side * position.x * s * 0.35;
      }
    }
    mv.xy += off;
    gl_Position = projectionMatrix * mv;
    vUv = position.xy + 0.5;
    vCol = col;
    vFade = smoothstep(0.0, 0.06, k) * (1.0 - k) * (1.0 - k);
  }`;

const FRAG = /* glsl */ `
  varying vec2 vUv;
  varying vec3 vCol;
  varying float vFade;
  uniform float uSoft;
  void main() {
    float d = length(vUv * 2.0 - 1.0);
    float a = uSoft > 0.5 ? smoothstep(1.0, 0.2, d) * 0.55 : smoothstep(1.0, 0.0, d);
    if (a * vFade < 0.003) discard;
    gl_FragColor = vec4(vCol * (uSoft > 0.5 ? 1.0 : a * vFade), uSoft > 0.5 ? a * vFade : 1.0);
    #include <tonemapping_fragment>
    #include <colorspace_fragment>
  }`;

function makePool(kind: PoolKind) {
  const size = SIZES[kind];
  const geo = new THREE.InstancedBufferGeometry();
  const quad = new THREE.PlaneGeometry(1, 1);
  geo.index = quad.index;
  geo.setAttribute('position', quad.getAttribute('position'));
  const attr = (n: number, fill = 0) => {
    const a = new THREE.InstancedBufferAttribute(new Float32Array(size * n).fill(fill), n);
    a.setUsage(THREE.DynamicDrawUsage);
    return a;
  };
  const attrs = {
    p0: attr(3), vel: attr(3), col: attr(3), birth: attr(1, -1e6), life: attr(1), size: attr(1), grav: attr(1), streak: attr(1),
  };
  for (const [k, a] of Object.entries(attrs)) geo.setAttribute(k, a);
  geo.instanceCount = size;
  const soft = kind === 'soft';
  const mat = new THREE.ShaderMaterial({
    vertexShader: VERT, fragmentShader: FRAG,
    uniforms: { uTime: { value: 0 }, uSoft: { value: soft ? 1 : 0 } },
    transparent: true, depthWrite: false,
    blending: soft ? THREE.NormalBlending : THREE.AdditiveBlending,
  });
  const mesh = new THREE.Mesh(geo, mat);
  mesh.frustumCulled = false;
  mesh.renderOrder = soft ? 7 : 8;
  const pool: Pool = { geo, attrs, next: 0, size, lo: size, hi: -1, wrapped: false };
  return { mesh, mat, pool };
}

// Uploads only the slots written since the last frame.
function flush(p: Pool) {
  if (p.lo > p.hi) return;
  for (const a of Object.values(p.attrs)) {
    a.clearUpdateRanges();
    if (!p.wrapped) a.addUpdateRange(p.lo * a.itemSize, (p.hi - p.lo + 1) * a.itemSize);
    a.needsUpdate = true;
  }
  p.lo = p.size; p.hi = -1; p.wrapped = false;
}

// Blast shockwaves: a few pooled rings that grow and fade on the ground.
const RINGS = 8, RING_LIFE = 0.45;
const rings: { mesh: THREE.Mesh; born: number }[] = [];
let nextRing = 0;
export function shockwave(x: number, z: number) {
  const r = rings[nextRing];
  if (!r) return;
  nextRing = (nextRing + 1) % RINGS;
  r.mesh.position.set(x, 0.09, z);
  r.born = now();
}

export function Particles() {
  const glow = useMemo(() => makePool('glow'), []);
  const soft = useMemo(() => makePool('soft'), []);
  const { ringGroup, ringList } = useMemo(() => {
    const g = new THREE.Group(), list: typeof rings = [];
    const geo = new THREE.RingGeometry(0.82, 1, 48);
    for (let i = 0; i < RINGS; i++) {
      const mesh = new THREE.Mesh(geo, new THREE.MeshBasicMaterial({ color: new THREE.Color(3, 1.7, 0.7), transparent: true, opacity: 0, depthWrite: false, blending: THREE.AdditiveBlending, toneMapped: false }));
      mesh.rotation.x = -Math.PI / 2;
      mesh.visible = false;
      mesh.renderOrder = 8;
      g.add(mesh);
      list.push({ mesh, born: -1e6 });
    }
    return { ringGroup: g, ringList: list };
  }, []);

  useEffect(() => {
    // registered in an effect (not the memo), so a remount registers again
    pools.glow = glow.pool;
    pools.soft = soft.pool;
    rings.splice(0, rings.length, ...ringList);
    return () => { delete pools.glow; delete pools.soft; rings.length = 0; };
  }, [glow, soft, ringList]);

  useFrame(() => {
    const t = now();
    glow.mat.uniforms.uTime.value = t;
    soft.mat.uniforms.uTime.value = t;
    flush(glow.pool);
    flush(soft.pool);
    for (const r of rings) {
      const k = (t - r.born) / RING_LIFE;
      r.mesh.visible = k >= 0 && k < 1;
      if (!r.mesh.visible) continue;
      const s = 0.3 + 2.7 * (1 - (1 - k) * (1 - k));
      r.mesh.scale.setScalar(s);
      (r.mesh.material as THREE.MeshBasicMaterial).opacity = (1 - k) * 0.9;
    }
  });

  return (
    <>
      <primitive object={soft.mesh} />
      <primitive object={glow.mesh} />
      <primitive object={ringGroup} />
    </>
  );
}
