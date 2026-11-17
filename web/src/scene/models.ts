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

export type Character = { root: THREE.Object3D; mixer: THREE.AnimationMixer; actions: Record<string, THREE.AnimationAction>; clips: Record<string, THREE.AnimationClip> };

const ONCE = ['Sword_Attack', 'Spell_Simple_Shoot', 'Hit_Chest', 'Death01'];

// A fresh, independently animated copy of a look, holding its class weapon.
export async function makeCharacter(cls: number, look: string, team: number): Promise<Character> {
  const [scene, clips] = await Promise.all([loadScene(look, team), loadClips()]);
  const root = cloneSkinned(scene);
  const bones: Record<string, THREE.Bone> = {};
  root.traverse((o) => {
    if ((o as THREE.Bone).isBone) bones[o.name] = o as THREE.Bone;
    if ((o as THREE.Mesh).isMesh) { o.castShadow = true; o.frustumCulled = false; }
  });
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
  const mixer = new THREE.AnimationMixer(root);
  const actions: Record<string, THREE.AnimationAction> = {};
  for (const [name, clip] of Object.entries(fitClips(clips, root))) {
    const a = mixer.clipAction(clip);
    if (ONCE.includes(name)) { a.setLoop(THREE.LoopOnce, 1); a.clampWhenFinished = true; }
    actions[name] = a;
  }
  actions.Sword_Idle.play();
  mixer.setTime(Math.random() * 2); // idle out of step between units (presentation only)
  return { root, mixer, actions, clips };
}
