import { useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { useGLTF } from '@react-three/drei';
import * as THREE from 'three';
import { useBattle } from '../store.ts';
import { OBSTACLE, type Obstacle } from '../sim/snapshot.ts';
import { NATURE_GLB, instances, readModel, type Model, type Placement } from './nature.ts';
import { hash2, rng } from './natureShader.ts';

const ROCKS = ['Rock_Medium_1', 'Rock_Medium_2', 'Rock_Medium_3'];
const TREES = ['CommonTree_1', 'CommonTree_2', 'CommonTree_3', 'Pine_1', 'Pine_2', 'Pine_5'];
const HALF = 14, WALL = 14.4;

function useModels() {
  const gltf = useGLTF(NATURE_GLB);
  return useMemo(() => {
    const out: Record<string, Model> = {};
    for (const child of gltf.scene.children) out[child.name] = readModel(child);
    return out;
  }, [gltf]);
}

// Groups placements by model and builds the instanced meshes.
function build(models: Record<string, Model>, items: { model: string; p: Placement }[], shadows = true) {
  const g = new THREE.Group();
  const by = new Map<string, Placement[]>();
  for (const it of items) by.set(it.model, [...(by.get(it.model) ?? []), it.p]);
  for (const [name, ps] of by) if (models[name]) g.add(instances(models[name], ps, shadows));
  return g;
}

// Rocks and trees where the sim put them. Variant, rotation and size come
// from the obstacle's position, so a map always looks the same. Rocks span
// the sim's circle (what you see is what blocks); a tree's circle is its trunk.
function obstaclePlacements(models: Record<string, Model>, obstacles: Obstacle[]) {
  const items: { model: string; p: Placement }[] = [];
  for (const o of obstacles) {
    if (o.kind === OBSTACLE.rock) {
      const model = ROCKS[Math.floor(hash2(o.x, o.z) * ROCKS.length)], m = models[model];
      if (!m) continue;
      const scale = (o.r * 1.05) / m.radius;
      items.push({ model, p: { x: o.x, z: o.z, rot: hash2(o.x, o.z, 1) * Math.PI * 2, scale, squash: 0.75 + 0.25 * hash2(o.x, o.z, 2), y: -0.05 } });
    } else if (o.kind === OBSTACLE.tree) {
      const model = TREES[Math.floor(hash2(o.x, o.z) * TREES.length)], m = models[model];
      if (!m) continue;
      const scale = (4.5 + 1.5 * hash2(o.x, o.z, 3)) / m.height;
      items.push({ model, p: { x: o.x, z: o.z, rot: hash2(o.x, o.z, 1) * Math.PI * 2, scale } });
    }
  }
  return items;
}

// A low wall of stones around the arena, and scenery outside it (fixed seed).
function sceneryPlacements(models: Record<string, Model>) {
  const r = rng(1234), items: { model: string; p: Placement }[] = [];
  const stone = (x: number, z: number, size: number) => {
    const model = ROCKS[Math.floor(r() * ROCKS.length)];
    items.push({ model, p: { x, z, y: -0.05, rot: r() * 6.28, scale: size / models[model].radius, squash: 0.55 + r() * 0.25 } });
  };
  for (let t = -WALL; t <= WALL; t += 0.85) {
    for (const [x, z] of [[t, -WALL], [t, WALL], [-WALL, t], [WALL, t]]) {
      stone(x + (r() - 0.5) * 0.15, z + (r() - 0.5) * 0.15, 0.5 + r() * 0.15);
    }
  }
  const scatter = (count: number, names: string[], minR: number, maxR: number, size: () => number, byHeight: boolean, nearSide = minR) => {
    for (let i = 0; i < count; i++) {
      const a = r() * Math.PI * 2, d = minR + r() * (maxR - minR);
      const x = Math.cos(a) * d * 1.15, z = Math.sin(a) * d;
      // keep tall things off the camera's side, so they never hide the arena
      if (Math.max(Math.abs(x), Math.abs(z)) < HALF + 2 || (z > HALF && d < nearSide)) { i--; continue; }
      const model = names[Math.floor(r() * names.length)], m = models[model];
      items.push({ model, p: { x, z, rot: r() * 6.28, scale: byHeight ? size() / m.height : size() / m.radius } });
    }
  };
  scatter(46, TREES, 17, 42, () => 5 + r() * 3, true, 30);
  scatter(24, ['Bush_Common'], 16, 30, () => 0.9 + r() * 0.6, false);
  scatter(18, ROCKS, 16, 34, () => 0.8 + r() * 1.2, false);
  scatter(40, ['Fern_1', 'Flower_3_Group', 'Pebble_Round_2'], 15.2, 22, () => 0.35 + r() * 0.3, false);
  return items;
}

export function Nature() {
  const models = useModels();
  const obstacles = useBattle((s) => s.obstacles);
  const rocksTrees = useMemo(() => build(models, obstaclePlacements(models, obstacles)), [models, obstacles]);
  const scenery = useMemo(() => build(models, sceneryPlacements(models)), [models]);
  return (
    <>
      <primitive object={rocksTrees} />
      <primitive object={scenery} />
      <Banner x={-10.5} z={WALL + 0.2} color="#3a6cff" />
      <Banner x={10.5} z={WALL + 0.2} color="#3a6cff" />
      <Banner x={-10.5} z={-WALL - 0.2} color="#e04040" />
      <Banner x={10.5} z={-WALL - 0.2} color="#e04040" />
      {[[-1, -1], [1, -1], [-1, 1], [1, 1]].map(([sx, sz]) => <Torch key={`${sx}${sz}`} x={sx * (WALL + 0.1)} z={sz * (WALL + 0.1)} />)}
    </>
  );
}

const wood = new THREE.MeshStandardMaterial({ color: '#5a3d22', roughness: 0.85 });

// A team banner: a pole and a cloth that waves (vertex sway, pinned at the pole).
function Banner({ x, z, color }: { x: number; z: number; color: string }) {
  const mat = useMemo(() => {
    const m = new THREE.MeshStandardMaterial({ color, side: THREE.DoubleSide, roughness: 0.8 });
    const time = { value: 0 };
    m.userData.time = time;
    m.onBeforeCompile = (s) => {
      s.uniforms.uTime = time;
      s.vertexShader = s.vertexShader
        .replace('#include <common>', '#include <common>\nuniform float uTime;')
        .replace('#include <begin_vertex>', `#include <begin_vertex>
float k = (position.x + 0.6) / 1.2;   // 0 at the pole, 1 at the free edge
transformed.z += sin(uTime * 2.6 + position.x * 3.0 + position.y * 1.5) * 0.12 * k;`);
    };
    return m;
  }, [color]);
  useFrame((_, dt) => { mat.userData.time.value += dt; });
  return (
    <group position={[x, 0, z]}>
      <mesh position-y={1.6} material={wood} castShadow><cylinderGeometry args={[0.06, 0.07, 3.2, 6]} /></mesh>
      <mesh position={[0.62, 2.45, 0]} material={mat} castShadow>
        <planeGeometry args={[1.2, 1.3, 8, 4]} />
      </mesh>
    </group>
  );
}

// A corner torch: the flame is an HDR colour so it glows on High.
function Torch({ x, z }: { x: number; z: number }) {
  const flame = useRef<THREE.Mesh>(null);
  const seed = useMemo(() => hash2(x, z) * 10, [x, z]);
  useFrame(({ clock }) => {
    const t = clock.elapsedTime + seed;
    flame.current?.scale.set(1, 1 + Math.sin(t * 13) * 0.12 + Math.sin(t * 7.3) * 0.08, 1);
  });
  return (
    <group position={[x, 0, z]}>
      <mesh position-y={0.9} material={wood} castShadow><cylinderGeometry args={[0.07, 0.09, 1.8, 6]} /></mesh>
      <mesh position-y={1.85}><cylinderGeometry args={[0.2, 0.1, 0.18, 8]} /><meshStandardMaterial color="#3a3a3a" metalness={0.6} roughness={0.5} /></mesh>
      <mesh ref={flame} position-y={2.1}>
        <coneGeometry args={[0.14, 0.42, 8]} />
        <meshBasicMaterial color={[4, 1.9, 0.5]} toneMapped={false} />
      </mesh>
    </group>
  );
}

useGLTF.preload(NATURE_GLB);
