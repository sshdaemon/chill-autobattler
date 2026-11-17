import { Bloom, BrightnessContrast, EffectComposer, HueSaturation, N8AO, TiltShift2, ToneMapping, Vignette } from '@react-three/postprocessing';
import { ToneMappingMode } from 'postprocessing';
import { useBattle } from '../store.ts';

// High quality only: ambient occlusion, glow on magic, a mild grade.
export function Effects() {
  const { quality, tiltShift } = useBattle();
  if (quality !== 'high') return null;
  return (
    <EffectComposer multisampling={2}>
      <N8AO aoRadius={1.2} intensity={2.2} distanceFalloff={0.6} quality="performance" halfRes />
      <Bloom luminanceThreshold={1} luminanceSmoothing={0.2} intensity={0.8} mipmapBlur />
      <HueSaturation saturation={0.08} />
      <BrightnessContrast contrast={0.06} />
      {tiltShift ? <TiltShift2 blur={0.08} /> : <></>}
      <ToneMapping mode={ToneMappingMode.ACES_FILMIC} />
      <Vignette offset={0.3} darkness={0.45} />
    </EffectComposer>
  );
}
