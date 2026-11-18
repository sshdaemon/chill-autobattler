import { Html } from '@react-three/drei';
import { useBattle } from '../store.ts';
import { UNIT_H } from './models.ts';

// Floating damage numbers and skill labels (from sim events).
export function Floats() {
  const floats = useBattle((s) => s.floats);
  return (
    <>
      {floats.map((f) => (
        <Html key={f.id} position={[f.x, UNIT_H.top + 0.9, f.z]} center style={{ pointerEvents: 'none' }}>
          <div className="float" style={{ color: f.color }}>{f.text}</div>
        </Html>
      ))}
    </>
  );
}
