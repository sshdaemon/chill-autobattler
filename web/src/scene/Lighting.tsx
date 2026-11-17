import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { Environment, Lightformer, Sky } from '@react-three/drei';
import type { DirectionalLight } from 'three';
import { useBattle } from '../store.ts';
import { camView } from './RtsCamera.tsx';

// Mid-afternoon sun from the south-west, high enough that trees don't throw
// long shadows across the field; the sky and the image-based
// lighting are generated in code (no HDRI download).
const SUN = { x: -0.42, y: 0.95, z: 0.55 };
const SHADOW_HALF = 14, SHADOW_MAP = 2048, TEXEL = (2 * SHADOW_HALF) / SHADOW_MAP;

export function Lighting() {
  const high = useBattle((s) => s.quality === 'high');
  const sun = useRef<DirectionalLight>(null);

  // The shadow box follows where the camera looks, snapped to whole shadow
  // texels so edges don't shimmer while panning.
  useFrame(() => {
    const l = sun.current;
    if (!l) return;
    const tx = Math.round(camView.x / TEXEL) * TEXEL, tz = Math.round(camView.z / TEXEL) * TEXEL;
    l.target.position.set(tx, 0, tz);
    l.position.set(tx + SUN.x * 40, SUN.y * 40, tz + SUN.z * 40);
    l.target.updateMatrixWorld();
  });

  return (
    <>
      <color attach="background" args={['#8fb2cf']} />
      <fog attach="fog" args={['#93adc2', 60, 150]} />
      <Sky distance={450} sunPosition={[SUN.x, SUN.y, SUN.z]} turbidity={6} rayleigh={1.4} mieCoefficient={0.004} mieDirectionalG={0.85} />
      {high ? (
        <Environment frames={1} resolution={256} environmentIntensity={0.7}>
          <Sky sunPosition={[SUN.x, SUN.y, SUN.z]} turbidity={6} rayleigh={1.4} />
          <Lightformer intensity={2.2} color="#fff1d6" position={[SUN.x * 8, SUN.y * 8, SUN.z * 8]} scale={[6, 6, 1]} form="circle" />
          <Lightformer intensity={0.6} color="#b9d4ff" position={[6, 4, -6]} scale={[10, 6, 1]} />
          <Lightformer intensity={0.5} color="#6f8a4a" position={[0, -5, 0]} rotation-x={Math.PI / 2} scale={[30, 30, 1]} />
        </Environment>
      ) : (
        <hemisphereLight args={['#d4e4f5', '#3a4a2c', 1.15]} />
      )}
      {high && <hemisphereLight args={['#d4e4f5', '#3a4a2c', 0.25]} />}
      <directionalLight ref={sun} color="#fff2dd" intensity={high ? 2.3 : 2.0} castShadow
        shadow-intensity={0.75} shadow-mapSize={[high ? SHADOW_MAP : 1024, high ? SHADOW_MAP : 1024]} shadow-bias={-0.0004} shadow-normalBias={0.03} shadow-radius={high ? 4 : 1} shadow-blurSamples={high ? 12 : 4}
        shadow-camera-left={-SHADOW_HALF} shadow-camera-right={SHADOW_HALF} shadow-camera-top={SHADOW_HALF} shadow-camera-bottom={-SHADOW_HALF}
        shadow-camera-near={1} shadow-camera-far={90} />
    </>
  );
}
