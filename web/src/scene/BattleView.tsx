// The 3D battle view and the battle-only UI around it (portraits, minimap,
// unit info). Loaded on demand (App.tsx lazy import, store's battle stage), so
// three.js and the scene stay out of the startup bundle.
import './loaderUrls.ts'; // first: before any scene module starts a load
import { Canvas } from '@react-three/fiber';
import { useGLTF, useTexture } from '@react-three/drei';
import { useBattle } from '../store.ts';
import { sceneHooks } from './hooks.ts';
import { lastKill, onSnapshot, viewOf } from './effects.ts';
import { warmUnits } from './models.ts';
import { GROUND_TEXTURES, NATURE_GLB, TREE_TEXTURES, TREES_GLB } from './arenaFiles.ts';
import { Arena } from './Arena.tsx';
import { Units } from './Units.tsx';
import { Shots } from './Shots.tsx';
import { Floats } from './Floats.tsx';
import { Particles } from './Particles.tsx';
import { Fog } from './Fog.tsx';
import { RtsCamera } from './RtsCamera.tsx';
import { Lighting } from './Lighting.tsx';
import { Effects } from './Effects.tsx';
import { Perf } from './Perf.tsx';
import { Portraits } from '../ui/Portraits.tsx';
import { InfoPanel } from '../ui/InfoPanel.tsx';
import { Minimap } from '../ui/Minimap.tsx';

sceneHooks.onSnapshot = (s, events) => onSnapshot(s, events, viewOf(useBattle.getState()));
sceneHooks.lastKill = lastKill;

// warmBattle fills the in-memory caches for a battle: the arena's models and
// textures (drei's cache) and the unit types' characters (models.ts).
export async function warmBattle(classes: number[]): Promise<void> {
  useGLTF.preload(NATURE_GLB);
  useGLTF.preload(TREES_GLB);
  useTexture.preload(Object.values(TREE_TEXTURES));
  useTexture.preload(Object.values(GROUND_TEXTURES));
  await warmUnits(classes);
}

export default function BattleView() {
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
      <Portraits />
      <Minimap />
      <InfoPanel />
    </>
  );
}
