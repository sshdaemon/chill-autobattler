import { useEffect, useMemo } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import { useBattle } from '../store.ts';
import { UF } from '../sim/snapshot.ts';
import { alpha, lerp } from './interp.ts';
import { NOISE_GLSL, POND_GLSL, pondUniforms } from './natureShader.ts';

const ARENA = 28, MAX_WET = 10;
const SUN = new THREE.Vector3(-0.45, 0.62, 0.64).normalize();

// All ponds as one surface: the shader draws the union of the sim's water
// circles, so overlapping blobs read as one pond. A soft shore, moving
// ripples, a sky reflection and a ring around every wading unit.
export function Water() {
  const obstacles = useBattle((s) => s.obstacles);
  const material = useMemo(() => new THREE.ShaderMaterial({
    transparent: true,
    depthWrite: false,
    uniforms: {
      ...pondUniforms([]),
      uTime: { value: 0 },
      uWet: { value: Array.from({ length: MAX_WET }, () => new THREE.Vector2()) },
      uWetCount: { value: 0 },
      uSun: { value: SUN },
      uSky: { value: new THREE.Color('#bcd6ee') },
      uDeep: { value: new THREE.Color('#1d4a63') },
      uShallow: { value: new THREE.Color('#4f8f8c') },
    },
    vertexShader: /* glsl */ `
      varying vec3 vWPos;
      void main() {
        vec4 w = modelMatrix * vec4(position, 1.0);
        vWPos = w.xyz;
        gl_Position = projectionMatrix * viewMatrix * w;
      }`,
    fragmentShader: /* glsl */ `
      varying vec3 vWPos;
      uniform float uTime;
      uniform vec2 uWet[${MAX_WET}];
      uniform int uWetCount;
      uniform vec3 uSun, uSky, uDeep, uShallow;
      ${NOISE_GLSL}
      ${POND_GLSL}
      float height(vec2 p) {
        float h = vnoise(p * 1.6 + vec2(uTime * 0.35, uTime * 0.2)) * 0.6 + vnoise(p * 3.7 - vec2(uTime * 0.25, -uTime * 0.4)) * 0.4;
        for (int i = 0; i < ${MAX_WET}; i++) {
          if (i >= uWetCount) break;
          float r = length(p - uWet[i]);
          h += sin(r * 14.0 - uTime * 7.0) * exp(-r * 1.6) * 0.9 * smoothstep(2.2, 0.15, r);
        }
        return h;
      }
      void main() {
        vec2 p = vWPos.xz;
        float d = pondDist(p);
        if (d > 0.05) discard;
        float depth = clamp(-d / 1.6, 0.0, 1.0);
        vec2 e = vec2(0.04, 0.0);
        vec3 n = normalize(vec3(height(p - e.xy) - height(p + e.xy), 0.6, height(p - e.yx) - height(p + e.yx)));
        vec3 view = normalize(cameraPosition - vWPos);
        float fres = 0.08 + 0.6 * pow(1.0 - max(dot(view, n), 0.0), 3.0);
        vec3 col = mix(uShallow, uDeep, smoothstep(0.0, 1.0, depth));
        col = mix(col, uSky, fres);
        col += vec3(1.0, 0.95, 0.85) * pow(max(dot(reflect(-uSun, n), view), 0.0), 120.0) * 1.6;
        float foam = smoothstep(0.05, -0.05, d) * smoothstep(-0.35, -0.05, d) * (0.6 + 0.4 * vnoise(p * 6.0 + uTime));
        col = mix(col, vec3(0.92, 0.96, 0.95), foam * 0.55);
        float a = mix(0.55, 0.92, smoothstep(0.0, 0.35, depth)) * smoothstep(0.05, -0.12, d);
        gl_FragColor = vec4(col, a);
        #include <tonemapping_fragment>
        #include <colorspace_fragment>
      }`,
  }), []);

  useEffect(() => {
    const p = pondUniforms(obstacles);
    material.uniforms.uPonds.value = p.uPonds.value;
    material.uniforms.uPondCount.value = p.uPondCount.value;
  }, [obstacles, material]);

  useFrame((_, dt) => {
    const u = material.uniforms;
    u.uTime.value += dt;
    const { prev, curr, fogOn } = useBattle.getState();
    let n = 0;
    if (curr && prev) {
      const k = alpha();
      for (let i = 0; i < curr.units.length && n < MAX_WET; i++) {
        const b = curr.units[i], a = prev.units[i] ?? b;
        const visible = b.team === 0 || !fogOn || (b.flags & UF.seen) !== 0; // no ripples that would give away hidden enemies
        if ((b.flags & UF.wet) !== 0 && (b.flags & UF.dead) === 0 && visible) {
          u.uWet.value[n++].set(lerp(a.x, b.x, k), lerp(a.z, b.z, k));
        }
      }
    }
    u.uWetCount.value = n;
  });

  return (
    <mesh rotation-x={-Math.PI / 2} position-y={0.04} material={material} renderOrder={2}>
      <planeGeometry args={[ARENA, ARENA]} />
    </mesh>
  );
}
