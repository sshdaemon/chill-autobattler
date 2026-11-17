import { Canvas } from '@react-three/fiber';
import { OrbitControls } from '@react-three/drei';
import { useBattle } from './store.ts';
import { Arena } from './scene/Arena.tsx';
import { Units } from './scene/Units.tsx';
import { Shots } from './scene/Shots.tsx';
import { Floats } from './scene/Floats.tsx';
import { Picker } from './ui/Picker.tsx';
import { DeployBar } from './ui/DeployBar.tsx';
import { Banner, DevPanel, Result, RunBar, Status } from './ui/Overlays.tsx';

export function App() {
  const phase = useBattle((s) => s.phase);
  // The camera is fixed while deploying: OrbitControls sees pointerdown before
  // R3F, so a unit drag would also orbit. It frees up once the fight starts.
  const noOrbit = useBattle((s) => s.phase === 'deploying' || s.phase === 'locking' || !!s.dragging);
  return (
    <>
      <Canvas shadows camera={{ position: [0, 25, 24], fov: 45 }}>
        <color attach="background" args={['#8fb6d8']} />
        <hemisphereLight args={['#dfefff', '#445533', 1.5]} />
        <directionalLight position={[10, 25, 8]} intensity={2.4} castShadow shadow-mapSize={[2048, 2048]}
          shadow-camera-left={-18} shadow-camera-right={18} shadow-camera-top={18} shadow-camera-bottom={-18} />
        <Arena />
        <Units />
        <Shots />
        <Floats />
        <OrbitControls enabled={!noOrbit} maxPolarAngle={1.3} minDistance={10} maxDistance={45} target={[0, 0, 5]} />
      </Canvas>
      <RunBar />
      <Status />
      <Banner />
      {phase === 'picker' && <Picker />}
      {(phase === 'deploying' || phase === 'locking') && <DeployBar />}
      {phase === 'done' && <Result />}
      <DevPanel />
    </>
  );
}
