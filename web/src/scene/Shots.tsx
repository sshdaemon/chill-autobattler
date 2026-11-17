import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { InstancedMesh, Object3D } from 'three';
import { useBattle } from '../store.ts';
import { alpha } from './interp.ts';

const MAX = 64, dummy = new Object3D();

// Arrows (light) and bolts (purple), flying along an arc; drawn from the latest snapshot.
export function Shots() {
  const arrows = useRef<InstancedMesh>(null), bolts = useRef<InstancedMesh>(null);
  useFrame(() => {
    const shots = useBattle.getState().curr?.shots ?? [];
    const k = alpha();
    const counts = [0, 0];
    for (const s of shots) {
      const f = Math.min(1, (s.t + k) / s.dur);
      const d = Math.hypot(s.ax - s.sx, s.az - s.sz);
      const arc = s.bolt ? 0 : d * (s.lob ? 0.22 : 0.07);
      dummy.position.set(s.sx + (s.ax - s.sx) * f, 1.3 + arc * 4 * f * (1 - f), s.sz + (s.az - s.sz) * f);
      dummy.scale.setScalar(s.bolt ? 1.6 : 1);
      dummy.updateMatrix();
      const m = s.bolt ? bolts.current : arrows.current, i = counts[s.bolt ? 1 : 0]++;
      if (m && i < MAX) m.setMatrixAt(i, dummy.matrix);
    }
    for (const [m, c] of [[arrows.current, counts[0]], [bolts.current, counts[1]]] as const) {
      if (!m) continue;
      m.count = Math.min(MAX, c);
      m.instanceMatrix.needsUpdate = true;
    }
  });
  return (
    <>
      <instancedMesh ref={arrows} args={[undefined, undefined, MAX]}><sphereGeometry args={[0.08, 8, 6]} /><meshBasicMaterial color="#f5e3b0" /></instancedMesh>
      <instancedMesh ref={bolts} args={[undefined, undefined, MAX]}><sphereGeometry args={[0.1, 10, 8]} /><meshBasicMaterial color="#c9a8ff" /></instancedMesh>
    </>
  );
}
