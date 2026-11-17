import { useBattle } from '../store.ts';
import { OBSTACLE } from '../sim/snapshot.ts';

// Ground, border and the seeded obstacles. Debug look, not the final art.
export function Arena() {
  const obstacles = useBattle((s) => s.obstacles);
  return (
    <>
      <mesh rotation-x={-Math.PI / 2} receiveShadow>
        <planeGeometry args={[28, 28]} />
        <meshStandardMaterial color="#7a9a5a" />
      </mesh>
      {[[0, -14.4, 29.6, 0.8], [0, 14.4, 29.6, 0.8], [-14.4, 0, 0.8, 28], [14.4, 0, 0.8, 28]].map(([x, z, w, d], i) => (
        <mesh key={i} position={[x, 0.35, z]} castShadow receiveShadow>
          <boxGeometry args={[w, 0.7, d]} />
          <meshStandardMaterial color="#9c9284" />
        </mesh>
      ))}
      {/* deployment zones */}
      {[[0, 10, '#3a8cff'], [0, -10, '#ff4040']].map(([x, z, c]) => (
        <mesh key={String(c)} position={[x as number, 0.02, z as number]} rotation-x={-Math.PI / 2}>
          <planeGeometry args={[20, 6]} />
          <meshBasicMaterial color={c as string} transparent opacity={0.12} depthWrite={false} />
        </mesh>
      ))}
      {obstacles.map((o, i) =>
        o.kind === OBSTACLE.water ? (
          <mesh key={i} position={[o.x, 0.03, o.z]} rotation-x={-Math.PI / 2}>
            <circleGeometry args={[o.r, 32]} />
            <meshStandardMaterial color="#2f6f9f" transparent opacity={0.85} />
          </mesh>
        ) : o.kind === OBSTACLE.rock ? (
          <mesh key={i} position={[o.x, o.r * 0.45, o.z]} scale={[1, 0.75, 1]} rotation-y={o.x} castShadow receiveShadow>
            <dodecahedronGeometry args={[o.r, 0]} />
            <meshStandardMaterial color="#8a8a85" flatShading />
          </mesh>
        ) : (
          <group key={i} position={[o.x, 0, o.z]}>
            <mesh position-y={0.9} castShadow><cylinderGeometry args={[0.25, 0.35, 1.8, 7]} /><meshStandardMaterial color="#6b4a2b" /></mesh>
            <mesh position-y={3.3} castShadow><coneGeometry args={[1, 3, 7]} /><meshStandardMaterial color="#2f5d2a" flatShading /></mesh>
          </group>
        ),
      )}
    </>
  );
}
