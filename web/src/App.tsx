import { Canvas } from '@react-three/fiber';
import { useBattle } from './store.ts';
import { Arena } from './scene/Arena.tsx';
import { Units } from './scene/Units.tsx';
import { Shots } from './scene/Shots.tsx';
import { Floats } from './scene/Floats.tsx';
import { Particles } from './scene/Particles.tsx';
import { Fog } from './scene/Fog.tsx';
import { RtsCamera } from './scene/RtsCamera.tsx';
import { Lighting } from './scene/Lighting.tsx';
import { Effects } from './scene/Effects.tsx';
import { Perf } from './scene/Perf.tsx';
import { Portraits } from './ui/Portraits.tsx';
import { InfoPanel } from './ui/InfoPanel.tsx';
import { Minimap } from './ui/Minimap.tsx';
import { Picker } from './ui/Picker.tsx';
import { DeployBar } from './ui/DeployBar.tsx';
import { Banner, DevPanel, Result, RunBar, Status } from './ui/Overlays.tsx';

export function App() {
  const phase = useBattle((s) => s.phase);
  const quality = useBattle((s) => s.quality);
  const lowDpr = useBattle((s) => s.lowDpr);
  return (
    <>
      {/* cap pixel density: full Retina (2x) with post-processing costs 4x the pixels */}
      <Canvas shadows dpr={quality === 'high' && !lowDpr ? [1, 1.5] : 1} camera={{ position: [0, 25, 24], fov: 45, far: 600 }}>
        <Lighting />
        <Arena />
        <Units />
        <Shots />
        <Floats />
        <Particles />
        <Fog />
        <RtsCamera />
        <Effects />
        <Perf />
      </Canvas>
      <RunBar />
      <Status />
      <Banner />
      {phase === 'picker' && <Picker />}
      {(phase === 'deploying' || phase === 'locking') && <DeployBar />}
      {phase === 'done' && <Result />}
      <Portraits />
      <Minimap />
      <InfoPanel />
      <DevPanel />
    </>
  );
}
