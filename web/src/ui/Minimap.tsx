import { useEffect, useRef } from 'react';
import { useBattle } from '../store.ts';
import { OBSTACLE, UF } from '../sim/snapshot.ts';
import { camView } from '../scene/RtsCamera.tsx';
import { FOGGED_FIGHT } from '../scene/Fog.tsx';

const SIZE = 160, ARENA = 28, S = SIZE / ARENA;
const px = (v: number) => (v + ARENA / 2) * S;
const SHOWN = ['deploying', 'locking', 'countdown', 'fighting', 'done'];

// Top-down overview: obstacles, units (fog of war applies), the camera's
// view; click to move the camera there.
export function Minimap() {
  const ref = useRef<HTMLCanvasElement>(null);
  const phase = useBattle((s) => s.phase);
  const visible = SHOWN.includes(phase);
  useEffect(() => {
    if (!visible) return;
    const draw = () => {
      const c = ref.current?.getContext('2d');
      if (!c) return;
      const { obstacles, curr, fogOn, phase } = useBattle.getState();
      c.fillStyle = '#5f7f45'; c.fillRect(0, 0, SIZE, SIZE);
      c.fillStyle = 'rgba(58,140,255,.18)'; c.fillRect(px(-10), px(7), 20 * S, 6 * S);
      c.fillStyle = 'rgba(255,64,64,.14)'; c.fillRect(px(-10), px(-13), 20 * S, 6 * S);
      for (const o of obstacles) {
        c.fillStyle = o.kind === OBSTACLE.water ? '#2f6f9f' : o.kind === OBSTACLE.rock ? '#9a9a95' : '#2f5d2a';
        c.beginPath(); c.arc(px(o.x), px(o.z), Math.max(1.5, o.r * S), 0, Math.PI * 2); c.fill();
      }
      const fogged = fogOn && FOGGED_FIGHT.includes(phase), deploying = ['deploying', 'locking'].includes(phase);
      for (const u of curr?.units ?? []) {
        if (u.flags & UF.dead) continue;
        if (u.team === 1 && (deploying || (fogged && !(u.flags & UF.seen)))) continue;
        c.fillStyle = u.team ? '#ff4040' : '#3a8cff';
        c.strokeStyle = '#fff'; c.lineWidth = 1;
        c.beginPath(); c.arc(px(u.x), px(u.z), 3.2, 0, Math.PI * 2); c.fill(); c.stroke();
      }
      // the camera's view, roughly: fov 45°, tilted 56°
      const h = camView.dist * 0.62, w = h * camView.aspect;
      c.strokeStyle = 'rgba(255,255,255,.85)'; c.lineWidth = 1.5;
      c.strokeRect(px(camView.x - w / 2), px(camView.z - h * 0.55), w * S, h * S);
      c.strokeStyle = '#1d2330'; c.lineWidth = 2; c.strokeRect(1, 1, SIZE - 2, SIZE - 2);
    };
    draw();
    const t = setInterval(draw, 100);
    return () => clearInterval(t);
  }, [visible]);
  if (!visible) return null;
  const jump = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const r = e.currentTarget.getBoundingClientRect();
    useBattle.getState().centerAt(((e.clientX - r.left) / r.width) * ARENA - ARENA / 2, ((e.clientY - r.top) / r.height) * ARENA - ARENA / 2);
  };
  return <canvas ref={ref} className="minimap" width={SIZE} height={SIZE} onClick={jump} title="Click to move the camera" data-testid="minimap" />;
}
