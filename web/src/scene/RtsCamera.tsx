import { shakeOffset } from './shake.ts';
import { useEffect, useRef } from 'react';
import { useFrame, useThree } from '@react-three/fiber';
import { useBattle } from '../store.ts';

// Warcraft III-style camera (ported from the prototype): fixed tilt, looking at
// a ground point; pan with WASD / arrows / screen edge, zoom with the wheel,
// F follows the selected unit, Space jumps to it.
const PITCH = (56 * Math.PI) / 180, MIN = 10, MAX = 45, PAN = 28, EDGE = 14, LIM = 16;

// Where the camera looks (read by the minimap without re-rendering React).
export const camView = { x: 0, z: 4, dist: 26, aspect: 1.5 };

export function RtsCamera() {
  const { camera, gl } = useThree();
  const st = useRef({ tx: 0, tz: 4, dist: 26, keys: new Set<string>(), mouse: null as null | { x: number; y: number }, lastCenter: 0 });

  useEffect(() => {
    const s = st.current, el = gl.domElement;
    const typing = (e: KeyboardEvent) => e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement;
    const down = (e: KeyboardEvent) => {
      if (typing(e)) return;
      s.keys.add(e.code);
      const b = useBattle.getState();
      if (e.code === 'KeyF') b.toggleFollow();
      if (e.code === 'KeyV') b.toggleFog();
      if (e.code === 'Space' && b.selected !== undefined) { e.preventDefault(); b.center(b.selected); }
    };
    const up = (e: KeyboardEvent) => s.keys.delete(e.code);
    const blur = () => s.keys.clear();
    const move = (e: PointerEvent) => { s.mouse = { x: e.clientX, y: e.clientY }; };
    const leave = () => { s.mouse = null; };
    const wheel = (e: WheelEvent) => { e.preventDefault(); s.dist = Math.max(MIN, Math.min(MAX, s.dist * Math.exp(e.deltaY * 0.001))); };
    addEventListener('keydown', down);
    addEventListener('keyup', up);
    addEventListener('blur', blur);
    el.addEventListener('pointermove', move);
    el.addEventListener('pointerleave', leave);
    el.addEventListener('wheel', wheel, { passive: false });
    return () => {
      removeEventListener('keydown', down); removeEventListener('keyup', up); removeEventListener('blur', blur);
      el.removeEventListener('pointermove', move); el.removeEventListener('pointerleave', leave); el.removeEventListener('wheel', wheel);
    };
  }, [gl]);

  useFrame((_, dt) => {
    const s = st.current, b = useBattle.getState();
    let px = 0, pz = 0;
    const k = s.keys;
    if (k.has('KeyA') || k.has('ArrowLeft')) px -= 1;
    if (k.has('KeyD') || k.has('ArrowRight')) px += 1;
    if (k.has('KeyW') || k.has('ArrowUp')) pz -= 1;
    if (k.has('KeyS') || k.has('ArrowDown')) pz += 1;
    if (s.mouse && !b.dragging) {
      const r = gl.domElement.getBoundingClientRect();
      if (s.mouse.x - r.left < EDGE) px -= 1;
      if (r.right - s.mouse.x < EDGE) px += 1;
      if (s.mouse.y - r.top < EDGE) pz -= 1;
      if (r.bottom - s.mouse.y < EDGE) pz += 1;
    }
    if ((px || pz) && b.follow) useBattle.setState({ follow: false }); // panning stops following
    const speed = PAN * dt * (s.dist / 26);
    s.tx += Math.sign(px) * speed;
    s.tz += Math.sign(pz) * speed;
    const unit = b.selected !== undefined ? b.curr?.units[b.selected] : undefined;
    if (b.centerOn && b.centerOn.n !== s.lastCenter) {
      s.lastCenter = b.centerOn.n;
      const u = b.centerOn.index !== undefined ? b.curr?.units[b.centerOn.index] : undefined;
      if (u) { s.tx = u.x; s.tz = u.z; }
      else if (b.centerOn.x !== undefined && b.centerOn.z !== undefined) { s.tx = b.centerOn.x; s.tz = b.centerOn.z; }
    }
    if (b.follow && unit) {
      const t = 1 - Math.exp(-6 * dt);
      s.tx += (unit.x - s.tx) * t;
      s.tz += (unit.z - s.tz) * t;
    }
    s.tx = Math.max(-LIM, Math.min(LIM, s.tx));
    s.tz = Math.max(-LIM, Math.min(LIM, s.tz));
    // the camera sits south (+z) of the target, raised by the pitch angle
    const [ox, oy, oz] = shakeOffset(dt, performance.now() / 1000);
    camera.position.set(s.tx + ox, Math.sin(PITCH) * s.dist + oy, s.tz + Math.cos(PITCH) * s.dist + oz);
    camera.lookAt(s.tx + ox, oy, s.tz + oz);
    camView.x = s.tx; camView.z = s.tz; camView.dist = s.dist;
    camView.aspect = gl.domElement.clientWidth / Math.max(1, gl.domElement.clientHeight);
  });
  return null;
}
