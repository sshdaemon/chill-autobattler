import { Bloom, BrightnessContrast, EffectComposer, HueSaturation, SMAA, TiltShift2, ToneMapping, Vignette } from '@react-three/postprocessing';
import { ToneMappingMode } from 'postprocessing';
import { useBattle } from '../store.ts';

// High quality only: glow on magic, a mild grade, cheap anti-aliasing.
// Budget (Apple M3, 1536x1152): N8AO cost ~8-14 ms and MSAA ~5 ms per frame,
// so ambient occlusion is left out and SMAA smooths edges instead; bloom
// runs at half resolution with 4 mip levels (~0.1 ms).
export function Effects() {
  const { quality, tiltShift } = useBattle();
  if (quality !== 'high') return null;
  return (
    <EffectComposer multisampling={0}>
      <Bloom luminanceThreshold={1} luminanceSmoothing={0.2} intensity={0.8} mipmapBlur levels={4} resolutionScale={0.5} />
      <HueSaturation saturation={0.08} />
      <BrightnessContrast contrast={0.06} />
      {tiltShift ? <TiltShift2 blur={0.08} /> : <></>}
      <ToneMapping mode={ToneMappingMode.ACES_FILMIC} />
      <Vignette offset={0.3} darkness={0.45} />
      <SMAA />
    </EffectComposer>
  );
}
