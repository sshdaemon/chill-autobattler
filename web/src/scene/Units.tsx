import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { Billboard } from '@react-three/drei';
import type { Group, Mesh } from 'three';
import { useBattle } from '../store.ts';
import { UF } from '../sim/snapshot.ts';
import { alpha, lerp } from './interp.ts';

const TEAM = ['#3a8cff', '#ff4040'];
const CLASS_TOP = ['#d9d9d9', '#3f7a3a', '#8f74ff']; // warrior steel, archer green, mage purple

function Unit({ index }: { index: number }) {
  const g = useRef<Group>(null), bar = useRef<Mesh>(null), body = useRef<Group>(null);
  const info = useBattle((s) => s.curr?.units[index]);
  useFrame(() => {
    const { prev, curr } = useBattle.getState();
    const a = prev?.units[index], b = curr?.units[index];
    if (!a || !b || !g.current) return;
    const k = alpha();
    g.current.position.set(lerp(a.x, b.x, k), b.flags & UF.wet ? -0.12 : 0, lerp(a.z, b.z, k));
    g.current.rotation.y = Math.atan2(lerp(a.fx, b.fx, k), lerp(a.fz, b.fz, k));
    if (body.current) {
      body.current.rotation.x = b.flags & UF.dead ? -Math.PI / 2 : b.flags & UF.swinging ? 0.25 : 0;
      body.current.position.y = b.flags & UF.dead ? 0.35 : 0;
    }
    if (bar.current) bar.current.scale.x = Math.max(0.001, b.hp / b.maxHp);
  });
  if (!info) return null;
  const dead = !!(info.flags & UF.dead);
  return (
    <group ref={g}>
      <group ref={body}>
        <mesh position-y={0.8} castShadow>
          <capsuleGeometry args={[0.32, 0.9]} />
          <meshStandardMaterial color={dead ? '#555' : TEAM[info.team]} />
        </mesh>
        <mesh position={[0, 1.55, 0]} castShadow>
          <coneGeometry args={[0.22, info.cls === 2 ? 0.6 : 0.3, 8]} />
          <meshStandardMaterial color={dead ? '#444' : CLASS_TOP[info.cls]} />
        </mesh>
        <mesh position={[0, 1.0, 0.36]}>{/* nose: facing */}
          <boxGeometry args={[0.12, 0.12, 0.2]} />
          <meshStandardMaterial color="#222" />
        </mesh>
      </group>
      {!dead && (
        <Billboard position-y={2.1}>
          <mesh><planeGeometry args={[0.9, 0.12]} /><meshBasicMaterial color="#222" /></mesh>
          <mesh ref={bar} position-z={0.001}><planeGeometry args={[0.9, 0.12]} /><meshBasicMaterial color={TEAM[info.team]} /></mesh>
        </Billboard>
      )}
      {!!(info.flags & UF.slowed) && !dead && (
        <mesh position-y={0.05} rotation-x={-Math.PI / 2}><ringGeometry args={[0.45, 0.6, 24]} /><meshBasicMaterial color="#9fd8ff" /></mesh>
      )}
    </group>
  );
}

export function Units() {
  const n = useBattle((s) => s.curr?.units.length ?? 0);
  return <>{Array.from({ length: n }, (_, i) => <Unit key={i} index={i} />)}</>;
}
