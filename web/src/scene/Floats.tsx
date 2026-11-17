import { Html } from '@react-three/drei';
import { useBattle } from '../store.ts';

// Floating damage numbers and skill labels (from sim events).
export function Floats() {
  const floats = useBattle((s) => s.floats);
  return (
    <>
      {floats.map((f) => (
        <Html key={f.id} position={[f.x, 2.5, f.z]} center style={{ pointerEvents: 'none' }}>
          <div className="float" style={{ color: f.color }}>{f.text}</div>
        </Html>
      ))}
    </>
  );
}
