import { Suspense } from 'react';
import { useBattle } from '../store.ts';
import { Ground } from './Ground.tsx';
import { Water } from './Water.tsx';
import { Nature } from './Nature.tsx';
import { Grass } from './Grass.tsx';

// The battlefield: textured ground, the ponds, rocks and trees where the sim
// put them, grass (High only), the wall and the scenery around it.
export function Arena() {
  const high = useBattle((s) => s.quality === 'high');
  const deploying = useBattle((s) => ['starting', 'deploying', 'locking'].includes(s.phase));
  return (
    <>
      <Suspense fallback={<FlatGround />}><Ground /></Suspense>
      <Water />
      <Suspense fallback={null}><Nature /></Suspense>
      {high && <Grass />}
      {/* Blue's deployment zone, only while placing units */}
      {deploying && (
        <mesh position={[0, 0.05, 10]} rotation-x={-Math.PI / 2} renderOrder={3}>
          <planeGeometry args={[20, 6]} />
          <meshBasicMaterial color="#3a8cff" transparent opacity={0.12} depthWrite={false} />
        </mesh>
      )}
    </>
  );
}

function FlatGround() {
  return (
    <mesh rotation-x={-Math.PI / 2} receiveShadow>
      <planeGeometry args={[120, 120]} />
      <meshStandardMaterial color="#6f8a4a" />
    </mesh>
  );
}
