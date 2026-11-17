import { Canvas } from '@react-three/fiber';
import { useBattle } from './store.ts';
import { Arena } from './scene/Arena.tsx';
import { Units } from './scene/Units.tsx';
import { Shots } from './scene/Shots.tsx';
import { Floats } from './scene/Floats.tsx';
import { Fog } from './scene/Fog.tsx';
import { RtsCamera } from './scene/RtsCamera.tsx';
import { Portraits } from './ui/Portraits.tsx';
import { InfoPanel } from './ui/InfoPanel.tsx';
import { Picker } from './ui/Picker.tsx';
import { DeployBar } from './ui/DeployBar.tsx';
import { Banner, DevPanel, Result, RunBar, Status } from './ui/Overlays.tsx';

export function App() {
  const phase = useBattle((s) => s.phase);
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
        <Fog />
        <RtsCamera />
      </Canvas>
      <RunBar />
      <Status />
      <Banner />
      {phase === 'picker' && <Picker />}
      {(phase === 'deploying' || phase === 'locking') && <DeployBar />}
      {phase === 'done' && <Result />}
      <Portraits />
      <InfoPanel />
      <DevPanel />
    </>
  );
}
