// Characters assembled at runtime from the parts library (tools/blender/bake_parts.py
// → web/public/assets/parts) by the recipes in units/looks.json, and the shared
// animation clips. Presentation only: nothing here can influence the battle
// (the sim runs in the worker).
import * as THREE from 'three';
import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader.js';
import { clone as cloneSkinned } from 'three/examples/jsm/utils/SkeletonUtils.js';
import { mergeGeometries } from 'three/examples/jsm/utils/BufferGeometryUtils.js';
import { BUILDS, CLIPS, lookOf, recipeNamed, type Recipe } from '../units/looks.ts';

// Which recipe a unit wears comes from its look: the n-th unit of a class in a
// team gets recipe n (mod the list), so an army looks the same after a reload.
export { modelFor as lookFor } from '../units/looks.ts';

const PARTS = '/assets/parts/';
const loader = new GLTFLoader();
const scenes = new Map<string, Promise<THREE.Object3D>>();
const load = (name: string) => {
  let p = scenes.get(name);
  if (!p) { p = loader.loadAsync(`${PARTS}${name}.gltf`).then((g) => g.scene); scenes.set(name, p); }
  return p;
};

type Manifest = {
  parts: Record<string, { slot: string; sex: string; materials: string[] }>;
  items: Record<string, { bone: string; pos?: number[]; rot?: number[]; scale?: number; team?: boolean }>;
  materials: Record<string, { variants: string[]; mask: string }>;
};
let manifestP: Promise<Manifest> | undefined;
const manifest = () => (manifestP ??= fetch(`${PARTS}parts.json`).then((r) => r.json() as Promise<Manifest>));

const textures = new Map<string, Promise<THREE.Texture>>();
function texture(file: string, color: boolean) {
  let p = textures.get(file);
  if (!p) {
    p = new THREE.TextureLoader().loadAsync(PARTS + file).then((t) => {
      t.flipY = false; // glTF UVs
      t.colorSpace = color ? THREE.SRGBColorSpace : THREE.NoColorSpace;
      return t;
    });
    textures.set(file, p);
  }
  return p;
}

let clipsPromise: Promise<Record<string, THREE.AnimationClip>> | undefined;
export function loadClips() {
  clipsPromise ??= fetch('/assets/clips.json')
    .then((r) => r.json())
    .then((list: Parameters<typeof THREE.AnimationClip.parse>[0][]) => Object.fromEntries(list.map((c) => {
      const clip = THREE.AnimationClip.parse(c);
      return [clip.name, clip];
    })));
  return clipsPromise;
}

// The shared clips animate the full rig, including leaf "*_end" bones the parts
// lack; tracks for missing nodes only make three.js warn. Every character has
// the same bones, so the trimmed clips are computed once and reused.
let fitted: Record<string, THREE.AnimationClip> | undefined;
function fitClips(clips: Record<string, THREE.AnimationClip>, root: THREE.Object3D) {
  if (fitted) return fitted;
  const names = new Set<string>();
  root.traverse((o) => names.add(o.name));
  fitted = Object.fromEntries(Object.entries(clips).map(([n, c]) => {
    const tracks = c.tracks.filter((t) => names.has(t.name.slice(0, t.name.lastIndexOf('.'))));
    return [n, new THREE.AnimationClip(c.name, c.duration, tracks)];
  }));
  return fitted;
}

export type Character = {
  root: THREE.Object3D; mixer: THREE.AnimationMixer; actions: Record<string, THREE.AnimationAction>; clips: Record<string, THREE.AnimationClip>;
  mats: THREE.MeshStandardMaterial[]; // this unit's own materials (rim, fade)
  rim: { color: { value: THREE.Color }; k: { value: number } };
  spine1?: THREE.Bone; spine2?: THREE.Bone;
};

export const TEAM_RIM = [new THREE.Color('#4f8dff'), new THREE.Color('#ff5a4a')];
export const TEAM_CLOTH = [new THREE.Color('#2f6fe8'), new THREE.Color('#d63a2e')];
const HAIR = ['#2a1c12', '#4a3020', '#6b4325', '#8a5a2b', '#b08850', '#c8a165', '#7a3b1e', '#9a9690'].map((c) => new THREE.Color(c));

// Stable per-unit randomness (team, class, nth in the army): same army, same looks.
function seeded(team: number, cls: number, nth: number) {
  let h = (team * 7919 + cls * 104729 + nth * 1299709 + 0x9e3779b9) >>> 0;
  return () => {
    h = Math.imul(h ^ (h >>> 15), 0x2c1b3c6d) >>> 0;
    h = (h ^ (h >>> 12)) >>> 0;
    return h / 4294967296;
  };
}

type Hooks = {
  rim: Character['rim'];
  girth: { value: number[] }; // metres to push the surface out per bone (body build), before skinning
  team?: { color: { value: THREE.Color }; k: { value: number }; mask: { value: THREE.Texture } };
};

// One shader hook per unit material: a soft team-coloured edge glow (Fresnel),
// the build's girth (vertices pushed along their bind-pose normal, weighted by
// bone) and, on outfit cloth, the team colour through the outfit's cloth mask.
function addHooks(m: THREE.MeshStandardMaterial, h: Hooks) {
  m.onBeforeCompile = (sh) => {
    sh.uniforms.uRim = h.rim.color;
    sh.uniforms.uRimK = h.rim.k;
    sh.uniforms.uGirth = h.girth;
    sh.vertexShader = sh.vertexShader
      .replace('#include <common>', `#include <common>\nuniform float uGirth[${h.girth.value.length}];`)
      .replace('#include <begin_vertex>', `#include <begin_vertex>
#ifdef USE_SKINNING
transformed += normal * (uGirth[int(skinIndex.x)] * skinWeight.x + uGirth[int(skinIndex.y)] * skinWeight.y
  + uGirth[int(skinIndex.z)] * skinWeight.z + uGirth[int(skinIndex.w)] * skinWeight.w);
#endif`);
    let frag = sh.fragmentShader
      .replace('#include <common>', '#include <common>\nuniform vec3 uRim;\nuniform float uRimK;')
      .replace('#include <emissivemap_fragment>', `#include <emissivemap_fragment>
float rimF = pow(1.0 - abs(dot(normalize(vNormal), normalize(vViewPosition))), 2.5);
totalEmissiveRadiance += uRim * uRimK * rimF;`);
    if (h.team) {
      sh.uniforms.uTeam = h.team.color;
      sh.uniforms.uTeamK = h.team.k;
      sh.uniforms.uMask = h.team.mask;
      frag = frag
        .replace('#include <common>', '#include <common>\nuniform vec3 uTeam;\nuniform float uTeamK;\nuniform sampler2D uMask;')
        .replace('#include <map_fragment>', `#include <map_fragment>
float teamM = texture2D(uMask, vMapUv).r * uTeamK;
float lum = dot(diffuseColor.rgb, vec3(0.299, 0.587, 0.114));
diffuseColor.rgb = mix(diffuseColor.rgb, uTeam * lum * 2.2, teamM);`);
    }
    sh.fragmentShader = frag;
  };
  m.customProgramCacheKey = () => (h.team ? 'unit-team' : 'unit');
}

// Merged geometry per recipe and material (same for every unit wearing the recipe).
const merged = new Map<string, THREE.BufferGeometry>();
function mergeGroup(key: string, geos: THREE.BufferGeometry[]) {
  let g = merged.get(key);
  if (!g) {
    const fixed = geos.map((src) => {
      const geo = src.clone();
      for (const n of Object.keys(geo.attributes)) if (!['position', 'normal', 'uv', 'skinIndex', 'skinWeight'].includes(n)) geo.deleteAttribute(n);
      const si = geo.getAttribute('skinIndex'); // ubyte or ushort per part: one type for all
      geo.setAttribute('skinIndex', new THREE.BufferAttribute(new Uint16Array(si.array as ArrayLike<number>), 4));
      // parts store weights, normals and UVs quantised (bake_parts.py), not always alike: floats for all
      for (const n of ['skinWeight', 'normal', 'uv']) {
        const at = geo.getAttribute(n);
        if (!at || at.array instanceof Float32Array) continue;
        const f = new Float32Array(at.count * at.itemSize);
        for (let i = 0; i < at.count; i++) for (let k = 0; k < at.itemSize; k++) f[i * at.itemSize + k] = at.getComponent(i, k);
        geo.setAttribute(n, new THREE.BufferAttribute(f, at.itemSize));
      }
      return geo;
    });
    g = fixed.length === 1 ? fixed[0] : mergeGeometries(fixed, false)!;
    merged.set(key, g);
  }
  return g;
}

// A fresh, independently animated character for a unit, built from its recipe.
export async function makeCharacter(cls: number, look: string, team: number, nth = 0): Promise<Character> {
  const recipe: Recipe = recipeNamed(cls, look);
  const outfit = recipe.parts, torsoName = outfit[0]; // the first part's skeleton is the unit's
  const [man, clips, torso, ...rest] = await Promise.all([manifest(), loadClips(), load(torsoName),
    ...[...outfit.slice(1), recipe.head, ...recipe.hair].map(load)]);
  const itemScenes = await Promise.all(recipe.items.map(load));
  const root = cloneSkinned(torso);
  let skeleton: THREE.Skeleton | undefined, bindMatrix = new THREE.Matrix4();
  root.traverse((o) => { if ((o as THREE.SkinnedMesh).isSkinnedMesh && !skeleton) { skeleton = (o as THREE.SkinnedMesh).skeleton; bindMatrix = (o as THREE.SkinnedMesh).bindMatrix; } });
  const skel = skeleton!;
  const boneIndex = new Map(skel.bones.map((b, i) => [b.name, i]));
  const bones: Record<string, THREE.Bone> = Object.fromEntries(skel.bones.map((b) => [b.name, b]));

  // every skinned mesh of every part, grouped by material (outfit, skin, hair...)
  const groups = new Map<string, { mat: THREE.MeshStandardMaterial; geos: THREE.BufferGeometry[] }>();
  const collect = (scene: THREE.Object3D) => scene.traverse((o) => {
    const m = o as THREE.SkinnedMesh;
    if (!m.isSkinnedMesh) return;
    // parts share the canonical joint list (bake_parts.py); remap defensively if not
    const order = m.skeleton.bones.map((b) => boneIndex.get(b.name) ?? 0);
    let geo = m.geometry;
    if (order.some((v, i) => v !== i)) {
      geo = geo.clone();
      const si = geo.getAttribute('skinIndex');
      for (let i = 0; i < si.count; i++) for (let k = 0; k < 4; k++) si.setComponent(i, k, order[si.getComponent(i, k)]);
    }
    const mat = m.material as THREE.MeshStandardMaterial;
    const g = groups.get(mat.name) ?? { mat, geos: [] };
    g.geos.push(geo);
    groups.set(mat.name, g);
  });
  collect(torso);
  rest.forEach(collect);
  const originals: THREE.Object3D[] = []; // replaced by the merged meshes below
  root.traverse((o) => { if ((o as THREE.SkinnedMesh).isSkinnedMesh) originals.push(o); });
  originals.forEach((o) => o.removeFromParent());
  const armature = skel.bones[0].parent ?? root;

  const rnd = seeded(team, cls, nth);
  const rim: Character['rim'] = { color: { value: TEAM_RIM[team].clone() }, k: { value: 0.55 } };
  const build = BUILDS[recipe.build] ?? BUILDS.normal;
  const girth = { value: skel.bones.map((b) => build.girth[b.name.replace(/_[lr]$/, '')] ?? 0) };
  const hair = HAIR[Math.floor(rnd() * HAIR.length)], shade = 0.93 + rnd() * 0.14, hue = (rnd() - 0.5) * 0.08;
  const mats: THREE.MeshStandardMaterial[] = [];
  const recipeKey = [recipe.name, ...outfit, recipe.head, ...recipe.hair].join('|');
  for (const [name, { mat, geos }] of groups) {
    const m = mat.clone();
    const spec = man.materials[name];
    let team_: Hooks['team'];
    if (spec) { // outfit cloth: the recipe's colour variant, team colour through the mask
      m.map = await texture(spec.variants[Math.min(recipe.variant, spec.variants.length - 1)], true);
      team_ = { color: { value: TEAM_CLOTH[team].clone() }, k: { value: recipe.team }, mask: { value: await texture(spec.mask, false) } };
      m.color.offsetHSL(hue, 0, 0).multiplyScalar(shade);
      if (recipe.tint) m.color.multiply(new THREE.Color(recipe.tint));
    } else if (/Hair|Eyebrow/i.test(name)) {
      m.color.copy(hair);
    }
    addHooks(m, { rim, girth, team: team_ });
    mats.push(m);
    const mesh = new THREE.SkinnedMesh(mergeGroup(`${recipeKey}#${name}`, geos), m);
    mesh.castShadow = true;
    mesh.frustumCulled = false;
    armature.add(mesh);
    mesh.bind(skel, bindMatrix);
  }

  // items (weapons, shields, quiver, hats): placed in the character's space at
  // rest, then attached to their bones keeping that placement
  root.updateMatrixWorld(true);
  recipe.items.forEach((name, i) => {
    const spec = man.items[name];
    const bone = spec && bones[spec.bone];
    if (!bone) return;
    const item = itemScenes[i].clone();
    item.position.copy(root.worldToLocal(bone.getWorldPosition(new THREE.Vector3())));
    if (spec.pos) item.position.add(new THREE.Vector3().fromArray(spec.pos));
    if (spec.rot) item.rotation.fromArray(spec.rot as [number, number, number]);
    if (spec.scale) item.scale.setScalar(spec.scale);
    root.add(item);
    item.updateMatrixWorld(true);
    bone.attach(item);
    item.traverse((o) => {
      const mesh = o as THREE.Mesh;
      if (!mesh.isMesh) return;
      mesh.castShadow = true;
      const m = (mesh.material as THREE.MeshStandardMaterial).clone();
      if (spec.team) m.color.copy(TEAM_CLOTH[team]); // e.g. a cape in the team colour
      addHooks(m, { rim, girth: { value: girth.value.map(() => 0) } });
      mats.push(m);
      mesh.material = m;
    });
  });

  // the build: wider shoulders, bigger or smaller head, height (girth is in the shader)
  for (const s of ['l', 'r']) bones[`upperarm_${s}`]?.position.multiplyScalar(build.shoulders);
  bones.Head?.scale.setScalar(build.head);
  root.scale.setScalar(build.height * (0.97 + rnd() * 0.06));
  root.userData.isCharacterRoot = true;

  const mixer = new THREE.AnimationMixer(root);
  const actions: Record<string, THREE.AnimationAction> = {};
  for (const [name, clip] of Object.entries(fitClips(clips, root))) {
    const a = mixer.clipAction(clip);
    if (!CLIPS[name]?.loop) { a.setLoop(THREE.LoopOnce, 1); a.clampWhenFinished = true; }
    actions[name] = a;
  }
  actions[lookOf(cls).anims.idle]?.play();
  mixer.setTime(Math.random() * 2); // idle out of step between units (presentation only)
  return { root, mixer, actions, clips, mats, rim, spine1: bones.spine_01, spine2: bones.spine_02 };
}
