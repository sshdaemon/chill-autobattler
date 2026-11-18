// Face close-ups, rendered once per unit look (look, team, nth: hair varies per unit) with a short-lived second
// renderer, cached as data URLs.
import * as THREE from 'three';
import { makeCharacter, UNIT_SCALE } from './models.ts';

const cache = new Map<string, Promise<string>>();
let queue: Promise<unknown> = Promise.resolve(); // one portrait at a time: one extra GL context, briefly

export function portrait(cls: number, look: string, team: number, nth = 0): Promise<string> {
  const key = `${look}_${team}_${nth}`;
  let p = cache.get(key);
  if (!p) {
    p = queue.then(() => render(cls, look, team, nth));
    queue = p.catch(() => undefined);
    cache.set(key, p);
  }
  return p;
}

async function render(cls: number, look: string, team: number, nth: number): Promise<string> {
  const ch = await makeCharacter(cls, look, team, nth);
  ch.rim.k.value = 0; // the frame shows the team; a rim would tint the face
  ch.mixer.setTime(0);
  const scene = new THREE.Scene();
  scene.add(new THREE.HemisphereLight(0xdfefff, 0x445533, 2.2));
  const sun = new THREE.DirectionalLight(0xfff2dd, 2.4);
  sun.position.set(1, 3, 4);
  scene.add(sun, ch.root);
  ch.root.updateMatrixWorld(true);
  let head: THREE.Object3D | undefined;
  ch.root.traverse((o) => { if ((o as THREE.Bone).isBone && o.name === 'Head') head = o; });
  const h = head ? head.getWorldPosition(new THREE.Vector3()) : new THREE.Vector3(0, 1.6, 0);
  h.y += 0.1 * UNIT_SCALE;
  const cam = new THREE.PerspectiveCamera(28, 1, 0.05, 60);
  cam.position.set(h.x, h.y + 0.04 * UNIT_SCALE, h.z + 0.8 * UNIT_SCALE); // models face +z
  cam.lookAt(h);
  const r = new THREE.WebGLRenderer({ antialias: true, alpha: true, preserveDrawingBuffer: true });
  r.setSize(108, 108);
  r.toneMapping = THREE.ACESFilmicToneMapping;
  r.outputColorSpace = THREE.SRGBColorSpace;
  r.render(scene, cam);
  const url = r.domElement.toDataURL();
  r.dispose();
  r.forceContextLoss();
  return url;
}
