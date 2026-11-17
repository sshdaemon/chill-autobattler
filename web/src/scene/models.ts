// Character models baked from the prototype (tools/blender → tools/assets/pack_units.py),
// the shared animation clips, and the procedural weapons. Presentation only:
// nothing here can influence the battle (the sim runs in the worker).
import * as THREE from 'three';
import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader.js';
import { clone as cloneSkinned } from 'three/examples/jsm/utils/SkeletonUtils.js';

// Looks per sim class (0 warrior, 1 archer, 2 mage); the n-th unit of a class
// in a team gets look n (mod the list), so an army looks the same after a reload.
export const LOOKS = [['Mirko', 'Vesna', 'Radim'], ['Bogdan', 'Lada'], ['Zoran', 'Milena', 'Velimir']];
export const lookFor = (cls: number, nth: number) => LOOKS[cls][nth % LOOKS[cls].length];

const loader = new GLTFLoader();
const scenes = new Map<string, Promise<THREE.Object3D>>();

function loadScene(look: string, team: number) {
  const key = `${look}_${team ? 'red' : 'blue'}`;
  let p = scenes.get(key);
  if (!p) {
    p = loader.loadAsync(`/assets/units/${key}.gltf`).then((g) => g.scene);
    scenes.set(key, p);
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

function makeSword() {
  const g = new THREE.Group();
  const steel = new THREE.MeshStandardMaterial({ color: 0xc9ccd2, metalness: 0.8, roughness: 0.35 });
  const wood = new THREE.MeshStandardMaterial({ color: 0x5a3a1e, roughness: 0.8 });
  const blade = new THREE.Mesh(new THREE.BoxGeometry(0.05, 0.75, 0.012), steel); blade.position.y = 0.47;
  const guard = new THREE.Mesh(new THREE.BoxGeometry(0.2, 0.03, 0.04), steel); guard.position.y = 0.09;
  const grip = new THREE.Mesh(new THREE.CylinderGeometry(0.018, 0.018, 0.16, 6), wood);
  g.add(blade, guard, grip);
  return g;
}

function makeBow() {
  const g = new THREE.Group();
  const wood = new THREE.MeshStandardMaterial({ color: 0x7a4a22, roughness: 0.7 });
  const limb = new THREE.Mesh(new THREE.TorusGeometry(0.5, 0.02, 6, 20, Math.PI * 0.8), wood);
  limb.rotation.z = Math.PI / 2 - Math.PI * 0.4; // arc opens away from the archer
  limb.position.x = -0.42;
  const string = new THREE.Mesh(new THREE.CylinderGeometry(0.004, 0.004, 0.95, 4), new THREE.MeshBasicMaterial({ color: 0xeeeeee }));
  string.position.x = -0.08;
  g.add(limb, string);
  return g;
}

function makeStaff() {
  const g = new THREE.Group();
  const shaft = new THREE.Mesh(new THREE.CylinderGeometry(0.022, 0.03, 1.6, 6), new THREE.MeshStandardMaterial({ color: 0x5b3a1e, roughness: 0.8 }));
  shaft.position.y = 0.25;
  const orb = new THREE.Mesh(new THREE.SphereGeometry(0.075, 12, 8), new THREE.MeshBasicMaterial({ color: new THREE.Color(1.9, 1.5, 3), toneMapped: false })); // >1: glows under bloom
  orb.position.y = 1.1;
  g.add(shaft, orb);
  return g;
}

// The shared clips animate the full Quaternius rig, but the bake drops leaf
// "*_end" bones; tracks for nodes a model lacks only make three.js warn (once
// per unit and action). All looks share the rig, so the trimmed clips are
// computed once and reused.
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
export const TEAM_TRIM = [new THREE.Color('#2f6fe8'), new THREE.Color('#d63a2e')];
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

// A soft team-coloured edge glow (Fresnel), added to the emissive term.
function addRim(m: THREE.MeshStandardMaterial, rim: Character['rim']) {
  m.onBeforeCompile = (sh) => {
    sh.uniforms.uRim = rim.color;
    sh.uniforms.uRimK = rim.k;
    sh.fragmentShader = sh.fragmentShader
      .replace('#include <common>', '#include <common>\nuniform vec3 uRim;\nuniform float uRimK;')
      .replace('#include <emissivemap_fragment>', `#include <emissivemap_fragment>
float rimF = pow(1.0 - abs(dot(normalize(vNormal), normalize(vViewPosition))), 2.5);
totalEmissiveRadiance += uRim * uRimK * rimF;`);
  };
  m.customProgramCacheKey = () => 'unit-rim';
}

function trimMat(team: number, metal: boolean) {
  return new THREE.MeshStandardMaterial({ color: TEAM_TRIM[team], metalness: metal ? 0.6 : 0, roughness: metal ? 0.35 : 0.8 });
}

// Team trims, made in code and attached to bones like the weapons.
function addTrims(cls: number, team: number, root: THREE.Object3D, bones: Record<string, THREE.Bone>) {
  const at = (bone: THREE.Bone | undefined, mesh: THREE.Object3D, offset: THREE.Vector3) => {
    if (!bone) return;
    mesh.position.copy(bone.getWorldPosition(new THREE.Vector3())).add(offset);
    mesh.traverse((o) => { if ((o as THREE.Mesh).isMesh) o.castShadow = true; });
    root.add(mesh);
    bone.attach(mesh);
  };
  root.updateMatrixWorld(true);
  if (cls === 0) {
    for (const [side, sx] of [['l', 1], ['r', -1]] as const) {
      const p = new THREE.Mesh(new THREE.SphereGeometry(0.11, 10, 6, 0, Math.PI * 2, 0, Math.PI / 2), trimMat(team, true));
      p.scale.set(1, 0.7, 1.15);
      p.rotation.z = -sx * 0.5;
      at(bones[`upperarm_${side}`], p, new THREE.Vector3(sx * 0.03, 0.04, 0));
    }
  } else if (cls === 1) {
    const q = new THREE.Group();
    const tube = new THREE.Mesh(new THREE.CylinderGeometry(0.06, 0.05, 0.5, 8), new THREE.MeshStandardMaterial({ color: 0x5a3a1e, roughness: 0.8 }));
    const band = new THREE.Mesh(new THREE.CylinderGeometry(0.065, 0.065, 0.06, 8), trimMat(team, false));
    band.position.y = 0.12;
    const feathers = new THREE.Mesh(new THREE.ConeGeometry(0.05, 0.12, 6), new THREE.MeshStandardMaterial({ color: 0xe8e2d0, roughness: 0.9 }));
    feathers.position.y = 0.3;
    q.add(tube, band, feathers);
    q.rotation.set(-0.25, 0, 0.35);
    at(bones.spine_03, q, new THREE.Vector3(0.05, 0.02, -0.16));
    const scarf = new THREE.Mesh(new THREE.TorusGeometry(0.1, 0.035, 6, 14), trimMat(team, false));
    scarf.rotation.x = Math.PI / 2;
    scarf.scale.set(1, 1.1, 1);
    at(bones.neck_01, scarf, new THREE.Vector3(0, -0.03, 0.01));
  } else {
    const mantle = new THREE.Mesh(new THREE.ConeGeometry(0.3, 0.22, 12, 1, true), trimMat(team, false));
    (mantle.material as THREE.MeshStandardMaterial).side = THREE.DoubleSide;
    at(bones.neck_01, mantle, new THREE.Vector3(0, -0.12, 0));
  }
}


const ONCE = ['Sword_Attack', 'Spell_Simple_Shoot', 'Hit_Chest', 'Death01'];

// A fresh, independently animated copy of a look, holding its class weapon.
export async function makeCharacter(cls: number, look: string, team: number, nth = 0): Promise<Character> {
  const [scene, clips] = await Promise.all([loadScene(look, team), loadClips()]);
  const root = cloneSkinned(scene);
  const bones: Record<string, THREE.Bone> = {};
  const rnd = seeded(team, cls, nth);
  const rim: Character['rim'] = { color: { value: TEAM_RIM[team].clone() }, k: { value: 0.55 } };
  const hair = HAIR[Math.floor(rnd() * HAIR.length)], shade = 0.93 + rnd() * 0.14, hue = (rnd() - 0.5) * 0.08;
  const mats: THREE.MeshStandardMaterial[] = [];
  const own = new Map<THREE.Material, THREE.MeshStandardMaterial>();
  root.traverse((o) => {
    if ((o as THREE.Bone).isBone) bones[o.name] = o as THREE.Bone;
    const mesh = o as THREE.Mesh;
    if (!mesh.isMesh) return;
    mesh.castShadow = true;
    mesh.frustumCulled = false;
    // this unit's own materials: rim colour, variety and the death fade are per unit
    const swap = (m: THREE.Material) => {
      let c = own.get(m);
      if (!c) {
        c = (m as THREE.MeshStandardMaterial).clone();
        if (/^MI_Hair/.test(c.name)) c.color.copy(hair);
        else if (/^MI_(Peasant|Ranger)/.test(c.name)) c.color.offsetHSL(hue, 0, 0).multiplyScalar(shade);
        addRim(c, rim);
        own.set(m, c);
        mats.push(c);
      }
      return c;
    };
    mesh.material = Array.isArray(mesh.material) ? mesh.material.map(swap) : swap(mesh.material);
  });
  root.scale.setScalar(0.96 + rnd() * 0.08);
  root.userData.isCharacterRoot = true;
  // weapon in the right hand (sword, staff) or the left (bow), as in the prototype
  const weapon = cls === 1 ? makeBow() : cls === 2 ? makeStaff() : makeSword();
  const hand = cls === 1 ? bones.hand_l : bones.hand_r;
  if (hand) {
    root.updateMatrixWorld(true);
    weapon.position.copy(hand.getWorldPosition(new THREE.Vector3()));
    if (cls === 1) weapon.rotation.set(0, Math.PI / 2, 0);
    else if (cls === 0) weapon.rotation.set(Math.PI / 2, 0, 0); // the staff stays upright
    weapon.traverse((o) => { if ((o as THREE.Mesh).isMesh) o.castShadow = true; });
    root.add(weapon);
    hand.attach(weapon);
  }
  addTrims(cls, team, root, bones);
  const mixer = new THREE.AnimationMixer(root);
  const actions: Record<string, THREE.AnimationAction> = {};
  for (const [name, clip] of Object.entries(fitClips(clips, root))) {
    const a = mixer.clipAction(clip);
    if (ONCE.includes(name)) { a.setLoop(THREE.LoopOnce, 1); a.clampWhenFinished = true; }
    actions[name] = a;
  }
  actions.Sword_Idle.play();
  mixer.setTime(Math.random() * 2); // idle out of step between units (presentation only)
  return { root, mixer, actions, clips, mats, rim, spine1: bones.spine_01, spine2: bones.spine_02 };
}
