import { useEffect, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { useBattle } from '../store.ts';
import { OBSTACLE, type Obstacle } from '../sim/snapshot.ts';
import { fogActive, fogTex } from './Fog.tsx';
import { NOISE_GLSL, rng } from './natureShader.ts';

const CLUMPS = 12000, AREA = 20; // clumps of 3 blades over ±20 m (the arena plus a margin)

// One clump: three tapered blades fanned around the origin (3 triangles each).
function clumpGeometry() {
  const pos: number[] = [], idx: number[] = [];
  for (let b = 0; b < 3; b++) {
    const a = (b / 3) * Math.PI + 0.3, cx = Math.cos(a), cz = Math.sin(a), ox = Math.cos(a * 2.3) * 0.06, oz = Math.sin(a * 2.3) * 0.06;
    const w = 0.035, base = pos.length / 3;
    // root left/right, middle left/right, tip
    pos.push(ox - cx * w, 0, oz - cz * w, ox + cx * w, 0, oz + cz * w,
      ox - cx * w * 0.6, 0.5, oz - cz * w * 0.6, ox + cx * w * 0.6, 0.5, oz + cz * w * 0.6, ox, 1, oz);
    idx.push(base, base + 1, base + 2, base + 1, base + 3, base + 2, base + 2, base + 3, base + 4);
  }
  const g = new THREE.BufferGeometry();
  g.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3));
  g.setIndex(idx);
  g.computeVertexNormals();
  // blades are lit as if facing up, so they shade like the ground below them
  const n = g.getAttribute('normal');
  for (let i = 0; i < n.count; i++) n.setXYZ(i, 0, 1, 0);
  return g;
}

function placements(obstacles: Obstacle[], zone: number[] | undefined) {
  const r = rng(77), out: THREE.Matrix4[] = [];
  const m = new THREE.Matrix4(), q = new THREE.Quaternion(), up = new THREE.Vector3(0, 1, 0);
  const blocked = (x: number, z: number) => obstacles.some((o) => {
    const pad = o.kind === OBSTACLE.water ? 0.1 : 0.2;
    return (x - o.x) ** 2 + (z - o.z) ** 2 < (o.r + pad) ** 2;
  });
  // Blue's zone and its mirror (Red's)
  const inZone = (x: number, z: number) => !!zone && x >= zone[0] && x <= zone[1] && ((z >= zone[2] && z <= zone[3]) || (-z >= zone[2] && -z <= zone[3]));
  let tries = 0;
  while (out.length < CLUMPS && tries++ < CLUMPS * 3) {
    const x = (r() * 2 - 1) * AREA, z = (r() * 2 - 1) * AREA;
    if (blocked(x, z)) continue;
    const edge = Math.max(Math.abs(x), Math.abs(z));
    if (edge > 13.9 && edge < 14.9) continue; // the wall
    if (inZone(x, z) && r() < 0.6) continue;
    const h = 0.22 + r() * 0.28;
    q.setFromAxisAngle(up, r() * Math.PI * 2);
    out.push(m.compose(new THREE.Vector3(x, 0, z), q, new THREE.Vector3(0.9 + r() * 0.4, h, 0.9 + r() * 0.4)).clone());
  }
  return out;
}

// Grass that sways in the wind (High only). It stands above the fog overlay,
// so it samples the fog texture and darkens itself the same way.
export function Grass() {
  const obstacles = useBattle((s) => s.obstacles);
  const zone = useBattle((s) => s.data?.deployZone);
  const geometry = useMemo(clumpGeometry, []);
  const uniforms = useMemo(() => ({ uTime: { value: 0 }, uFog: { value: fogTex }, uFogOn: { value: 0 } }), []);
  const material = useMemo(() => {
    const m = new THREE.MeshStandardMaterial({ side: THREE.DoubleSide, roughness: 0.9 });
    m.onBeforeCompile = (s) => {
      Object.assign(s.uniforms, uniforms);
      s.vertexShader = s.vertexShader
        .replace('#include <common>', `#include <common>
uniform float uTime;
varying float vH;
varying vec2 vXZ;
${NOISE_GLSL}`)
        .replace('#include <begin_vertex>', `#include <begin_vertex>
vec2 root = (modelMatrix * instanceMatrix * vec4(0.0, 0.0, 0.0, 1.0)).xz;
float gust = vnoise(root * 0.15 + vec2(uTime * 0.6, uTime * 0.25));
float sway = (sin(uTime * 1.8 + root.x * 0.7 + root.y * 0.4) * 0.35 + gust) * position.y * position.y;
transformed.x += sway * 0.22;
transformed.z += sway * 0.12;
vH = position.y;
vXZ = root;`);
      s.fragmentShader = s.fragmentShader
        .replace('#include <common>', `#include <common>
uniform sampler2D uFog;
uniform float uFogOn;
varying float vH;
varying vec2 vXZ;
${NOISE_GLSL}`)
        .replace('#include <map_fragment>', `
vec3 rootC = vec3(0.14, 0.22, 0.07), tipC = mix(vec3(0.38, 0.5, 0.17), vec3(0.5, 0.52, 0.22), vnoise(vXZ * 0.3));
diffuseColor.rgb = mix(rootC, tipC, vH);`)
        .replace('#include <opaque_fragment>', `
vec2 fuv = vec2((vXZ.x + 14.0) / 28.0, (14.0 - vXZ.y) / 28.0);
if (uFogOn > 0.5 && all(greaterThanEqual(fuv, vec2(0.0))) && all(lessThanEqual(fuv, vec2(1.0)))) {
  outgoingLight *= 1.0 - texture2D(uFog, fuv).a;
}
#include <opaque_fragment>`);
    };
    return m;
  }, [uniforms]);
  const mesh = useMemo(() => new THREE.InstancedMesh(geometry, material, CLUMPS), [geometry, material]);

  useEffect(() => {
    const ms = placements(obstacles, zone);
    ms.forEach((m, i) => mesh.setMatrixAt(i, m));
    mesh.count = ms.length;
    mesh.instanceMatrix.needsUpdate = true;
    mesh.computeBoundingSphere();
  }, [obstacles, zone, mesh]);

  useFrame((_, dt) => {
    uniforms.uTime.value += dt;
    uniforms.uFogOn.value = fogActive(useBattle.getState()) ? 1 : 0;
  });

  return <primitive object={mesh} receiveShadow />;
}
