import { useEffect, useMemo } from 'react';
import { useTexture } from '@react-three/drei';
import * as THREE from 'three';
import { useBattle } from '../store.ts';
import { NOISE_GLSL, POND_GLSL, pondUniforms } from './natureShader.ts';

const T = '/assets/nature/';
const SIZE = 120;

// The ground: CC0 grass, dirt and mud (ambientCG) blended in world space,
// with two texture scales mixed by noise so the tiling doesn't show, and mud
// around the ponds. It runs past the arena so the field sits in a landscape.
export function Ground() {
  const obstacles = useBattle((s) => s.obstacles);
  const tex = useTexture({
    grass: T + 'Grass004_color.jpg', dirt: T + 'Ground037_color.jpg', mud: T + 'Ground054_color.jpg',
    grassN: T + 'Grass004_normal.jpg', dirtN: T + 'Ground037_normal.jpg', mudN: T + 'Ground054_normal.jpg',
  });
  const uniforms = useMemo(() => ({ ...pondUniforms([]), ...Object.fromEntries(Object.entries(tex).map(([k, t]) => [k, { value: t }])) }), [tex]);

  useMemo(() => {
    for (const [k, t] of Object.entries(tex)) {
      t.wrapS = t.wrapT = THREE.RepeatWrapping;
      t.anisotropy = 8;
      t.colorSpace = k.endsWith('N') ? THREE.NoColorSpace : THREE.SRGBColorSpace;
      t.needsUpdate = true;
    }
  }, [tex]);

  useEffect(() => {
    const p = pondUniforms(obstacles);
    uniforms.uPonds.value = p.uPonds.value;
    uniforms.uPondCount.value = p.uPondCount.value;
  }, [obstacles, uniforms]);

  const material = useMemo(() => {
    const m = new THREE.MeshStandardMaterial({ roughness: 0.95, metalness: 0, normalMap: tex.grassN, normalScale: new THREE.Vector2(0.8, 0.8) });
    m.onBeforeCompile = (s) => {
      Object.assign(s.uniforms, uniforms);
      s.vertexShader = s.vertexShader
        .replace('#include <common>', '#include <common>\nvarying vec3 vWPos;')
        .replace('#include <begin_vertex>', '#include <begin_vertex>\nvWPos = (modelMatrix * vec4(transformed, 1.0)).xyz;');
      s.fragmentShader = s.fragmentShader
        .replace('#include <common>', `#include <common>
varying vec3 vWPos;
uniform sampler2D grass, dirt, mud, grassN, dirtN, mudN;
${NOISE_GLSL}
${POND_GLSL}
// two scales, the second rotated, mixed by noise: hides the tile repeat
vec4 tri(sampler2D t, vec2 p, float k) {
  vec4 a = texture2D(t, p * 0.25);
  vec4 b = texture2D(t, mat2(0.8, -0.6, 0.6, 0.8) * p * 0.11 + 0.37);
  return mix(a, b, k);
}
vec3 groundNormal(vec2 p, float k, vec3 w) {
  vec3 n = tri(grassN, p, k).xyz * w.x;
  if (w.y > 0.002) n += tri(dirtN, p, k).xyz * w.y;
  if (w.z > 0.002) n += tri(mudN, p, k).xyz * w.z;
  return n;
}
vec3 gWeights() {
  vec2 p = vWPos.xz;
  float edge = max(abs(p.x), abs(p.y)) - 14.0;              // > 0 outside the arena
  float dirtW = smoothstep(0.58, 0.78, fbm(p * 0.09) + (edge > 0.0 ? 0.06 : 0.0));
  float mudW = 1.0 - smoothstep(-0.2, 1.3 + vnoise(p * 0.8) * 0.6, pondDist(p));
  return vec3(1.0 - max(dirtW, mudW), dirtW * (1.0 - mudW), mudW);
}`)
        .replace('#include <map_fragment>', `
vec2 gp = vec2(vWPos.x, -vWPos.z);
float gk = vnoise(gp * 0.05);
vec3 gw = gWeights();
// sample dirt and mud only where they show (most of the field is grass)
vec3 gcol = tri(grass, gp, gk).rgb * vec3(1.02, 1.08, 0.86) * gw.x;
if (gw.y > 0.002) gcol += tri(dirt, gp, gk).rgb * gw.y;
if (gw.z > 0.002) gcol += tri(mud, gp, gk).rgb * 0.8 * gw.z;
gcol *= 0.92 + 0.16 * vnoise(gp * 0.35);                  // soft large-scale variation
diffuseColor.rgb *= gcol;`)
        .replace('#include <normal_fragment_maps>', THREE.ShaderChunk.normal_fragment_maps.replace(
          'texture2D( normalMap, vNormalMapUv ).xyz',
          'groundNormal(gp, gk, gw)'));
    };
    return m;
  }, [tex, uniforms]);

  return (
    <mesh rotation-x={-Math.PI / 2} receiveShadow material={material}>
      <planeGeometry args={[SIZE, SIZE]} />
    </mesh>
  );
}
