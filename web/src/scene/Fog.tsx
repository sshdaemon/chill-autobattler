import { useEffect } from 'react';
import { DataTexture, LinearFilter } from 'three';
import { useBattle } from '../store.ts';

const N = 56, ARENA = 28, DARK = 0.62 * 255;
export const FOGGED_FIGHT = ['countdown', 'fighting'];
const DEPLOYING = ['starting', 'deploying', 'locking'];

// The fog grid as a texture (alpha = darkness), shared with the grass,
// which stands above the overlay and darkens itself.
export const fogTex = new DataTexture(new Uint8Array(N * N * 4), N, N);
fogTex.magFilter = fogTex.minFilter = LinearFilter; // soft edges

export const fogActive = (s: { fogOn: boolean; phase: string }) =>
  s.fogOn && (DEPLOYING.includes(s.phase) || FOGGED_FIGHT.includes(s.phase));

// Darkens what Blue can't see. The grid comes from the sim (same sight rules
// as the battle); while deploying only Blue's own zone is lit.
export function Fog() {
  const { fog, fogOn, phase, data } = useBattle();
  const tex = fogTex;
  const deploying = DEPLOYING.includes(phase);
  const active = fogActive({ fogOn, phase });
  useEffect(() => {
    const d = tex.image.data as Uint8Array, z = data?.deployZone;
    for (let j = 0; j < N; j++) for (let i = 0; i < N; i++) {
      const x = -ARENA / 2 + (i + 0.5) * 0.5, zz = -ARENA / 2 + (j + 0.5) * 0.5;
      const lit = deploying ? !!z && x >= z[0] && x <= z[1] && zz >= z[2] && zz <= z[3] : !!fog?.[j * N + i];
      d[((N - 1 - j) * N + i) * 4 + 3] = lit ? 0 : DARK; // texture row 0 is the plane's +z edge
    }
    tex.needsUpdate = true;
  }, [fog, deploying, data, tex]);
  if (!active) return null;
  return (
    <mesh rotation-x={-Math.PI / 2} position-y={0.06} renderOrder={5}>
      <planeGeometry args={[ARENA, ARENA]} />
      <meshBasicMaterial map={tex} transparent depthWrite={false} color="#000" />
    </mesh>
  );
}
