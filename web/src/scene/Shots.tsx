import { useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { CylinderGeometry, ConeGeometry, InstancedMesh, Object3D, Vector3 } from 'three';
import { mergeGeometries } from 'three/examples/jsm/utils/BufferGeometryUtils.js';
import { useBattle } from '../store.ts';
import { alpha } from './interp.ts';
import { trail, viewOf } from './effects.ts';
import { smokeShots } from '../units/looks.ts';

const MAX = 64, dummy = new Object3D(), ahead = new Vector3();

// Where a shot is at fraction f of its flight: straight for bolts, an arc for arrows.
function at(s: { sx: number; sz: number; ax: number; az: number; bolt: boolean; lob: boolean }, f: number, out: Vector3) {
  const d = Math.hypot(s.ax - s.sx, s.az - s.sz);
  const arc = s.bolt ? 0 : d * (s.lob ? 0.22 : 0.07);
  return out.set(s.sx + (s.ax - s.sx) * f, 1.3 + arc * 4 * f * (1 - f), s.sz + (s.az - s.sz) * f);
}

// Arrows (shaft + tip, nose first along the arc), thrown spears (burst shots: a
// longer, thicker shaft with an iron head), magic bolts (glowing orbs) and
// curses (bolts of a unit whose look draws them as smoke: a trail only).
export function Shots() {
  const arrows = useRef<InstancedMesh>(null), shafts = useRef<InstancedMesh>(null), heads = useRef<InstancedMesh>(null), boltsOuter = useRef<InstancedMesh>(null), boltsInner = useRef<InstancedMesh>(null);
  const arrowGeo = useMemo(() => {
    const shaft = new CylinderGeometry(0.014, 0.014, 0.7, 5).rotateX(Math.PI / 2);
    const tip = new ConeGeometry(0.035, 0.1, 5).rotateX(Math.PI / 2).translate(0, 0, 0.4);
    return mergeGeometries([shaft.toNonIndexed(), tip.toNonIndexed()]);
  }, []);
  const [shaftGeo, headGeo] = useMemo(() => [
    new CylinderGeometry(0.022, 0.022, 1.5, 6).rotateX(Math.PI / 2).translate(0, 0, -0.2),
    new ConeGeometry(0.045, 0.22, 4).rotateX(Math.PI / 2).translate(0, 0, 0.66),
  ], []);
  useFrame(() => {
    const st = useBattle.getState(), shots = st.curr?.shots ?? [], view = viewOf(st);
    const k = alpha();
    let na = 0, nb = 0, ns = 0;
    for (const s of shots) {
      const f = Math.min(1, (s.t + k) / s.dur);
      at(s, f, dummy.position);
      const smoke = s.bolt && smokeShots(st.curr?.units[s.shooter]?.cls ?? -1);
      if (f < 1) trail(dummy.position.x, dummy.position.y, dummy.position.z, smoke ? 'smoke' : s.bolt ? 'bolt' : 'arrow', view);
      if (smoke) continue;
      if (s.bolt) {
        dummy.rotation.set(0, 0, 0);
        dummy.updateMatrix();
        if (nb < MAX) { boltsOuter.current?.setMatrixAt(nb, dummy.matrix); boltsInner.current?.setMatrixAt(nb, dummy.matrix); nb++; }
      } else if (s.burst) {
        dummy.lookAt(at(s, Math.min(1, f + 0.02), ahead));
        dummy.updateMatrix();
        if (ns < MAX) { shafts.current?.setMatrixAt(ns, dummy.matrix); heads.current?.setMatrixAt(ns, dummy.matrix); ns++; }
      } else {
        dummy.lookAt(at(s, Math.min(1, f + 0.02), ahead)); // +Z (the tip) faces along the flight
        dummy.updateMatrix();
        if (na < MAX) arrows.current?.setMatrixAt(na++, dummy.matrix);
      }
    }
    for (const [m, c] of [[arrows.current, na], [shafts.current, ns], [heads.current, ns], [boltsOuter.current, nb], [boltsInner.current, nb]] as const) {
      if (!m) continue;
      m.count = c;
      m.instanceMatrix.needsUpdate = true;
    }
  });
  return (
    <>
      <instancedMesh ref={arrows} args={[arrowGeo, undefined, MAX]} castShadow frustumCulled={false}><meshStandardMaterial color="#d8c08a" /></instancedMesh>
      <instancedMesh ref={shafts} args={[shaftGeo, undefined, MAX]} castShadow frustumCulled={false}><meshStandardMaterial color="#7a5634" /></instancedMesh>
      <instancedMesh ref={heads} args={[headGeo, undefined, MAX]} castShadow frustumCulled={false}><meshStandardMaterial color="#b8bec6" metalness={0.8} roughness={0.35} /></instancedMesh>
      <instancedMesh ref={boltsOuter} args={[undefined, undefined, MAX]}>
        <sphereGeometry args={[0.14, 12, 8]} />
        <meshBasicMaterial color={[1.7, 1.15, 2.6]} transparent opacity={0.6} toneMapped={false} />{/* >1: glows under bloom */}
      </instancedMesh>
      <instancedMesh ref={boltsInner} args={[undefined, undefined, MAX]}>
        <sphereGeometry args={[0.07, 10, 6]} />
        <meshBasicMaterial color={[2.6, 2.4, 3]} toneMapped={false} />
      </instancedMesh>
    </>
  );
}
