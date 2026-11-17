// The CC0 nature models (Quaternius Stylized Nature MegaKit, packed by
// tools/assets/pack_nature.py) and a helper that draws many copies of one
// model as instanced meshes. Presentation only.
import * as THREE from 'three';

export const NATURE_GLB = '/assets/nature/nature.glb';
export const TREES_GLB = '/assets/trees/trees.glb';

const TR = '/assets/trees/';
export const TREE_TEXTURES = {
  bark_oak: TR + 'bark_oak_color.jpg', bark_oak_n: TR + 'bark_oak_normal.jpg',
  bark_birch: TR + 'bark_birch_color.jpg', bark_birch_n: TR + 'bark_birch_normal.jpg',
  bark_pine: TR + 'bark_pine_color.jpg', bark_pine_n: TR + 'bark_pine_normal.jpg',
  leaves_oak: TR + 'leaves_oak.png', leaves_ash: TR + 'leaves_ash.png',
  leaves_aspen: TR + 'leaves_aspen.png', leaves_pine: TR + 'leaves_pine.png',
};

// The baked trees carry no textures: their material names (bark_<type>,
// leaves_<type>) pick them, and the bake's extras give tint and bark repeat.
export function dressTrees(root: THREE.Object3D, tex: Record<string, THREE.Texture>) {
  root.traverse((o) => {
    const mesh = o as THREE.Mesh;
    if (!mesh.isMesh) return;
    const m = mesh.material as THREE.MeshStandardMaterial;
    if (m.userData.dressed) return;
    m.userData.dressed = true;
    const tint = m.userData.tint as number | undefined;
    if (tint !== undefined) m.color.setHex(tint);
    if (m.name.startsWith('bark_')) {
      const map = tex[m.name].clone(), normal = tex[m.name + '_n'].clone();
      const [rx, ry] = (m.userData.repeat as number[] | undefined) ?? [1, 1];
      for (const t of [map, normal]) { t.wrapS = t.wrapT = THREE.RepeatWrapping; t.repeat.set(rx, ry); t.needsUpdate = true; }
      map.colorSpace = THREE.SRGBColorSpace;
      Object.assign(m, { map, normalMap: normal, roughness: 0.9, metalness: 0 });
    } else if (m.name.startsWith('leaves_')) {
      const map = tex[m.name];
      map.colorSpace = THREE.SRGBColorSpace;
      Object.assign(m, { map, alphaTest: 0.5, side: THREE.DoubleSide, roughness: 0.8, metalness: 0 });
      // leaf cards lit like a canopy: normals bent upward, so they catch the sun from above
      const n = mesh.geometry.getAttribute('normal');
      const v = new THREE.Vector3();
      for (let i = 0; i < n.count; i++) {
        v.fromBufferAttribute(n, i).multiplyScalar(0.35).add(new THREE.Vector3(0, 0.65, 0)).normalize();
        n.setXYZ(i, v.x, v.y, v.z);
      }
      m.userData.baked = true; // keep its own colours (see readModel)
    }
    m.needsUpdate = true;
  });
}

// Shared clock for leaves swaying in the wind (advanced by Nature's frame loop).
export const natureTime = { value: 0 };

function swayLeaves(m: THREE.MeshStandardMaterial) {
  m.onBeforeCompile = (s) => {
    s.uniforms.uTime = natureTime;
    s.vertexShader = s.vertexShader
      .replace('#include <common>', '#include <common>\nuniform float uTime;')
      .replace('#include <begin_vertex>', `#include <begin_vertex>
#ifdef USE_INSTANCING
vec2 root = instanceMatrix[3].xz;
#else
vec2 root = vec2(0.0);
#endif
float sway = sin(uTime * 1.4 + root.x * 0.5 + root.y * 0.3 + position.y * 0.6) + 0.5 * sin(uTime * 2.3 + position.x * 2.0);
transformed.xz += vec2(0.6, 0.35) * sway * 0.02 * max(position.y, 0.0);`);
  };
}

export type Model = {
  parts: { geometry: THREE.BufferGeometry; material: THREE.Material; local: THREE.Matrix4 }[];
  radius: number; // horizontal extent from the origin, in model units
  height: number;
};

// Flattens a model's meshes (with their transforms inside the model) and
// measures it; leaf materials get alpha testing so they cast leaf shadows.
export function readModel(root: THREE.Object3D): Model {
  root.updateMatrixWorld(true);
  const inv = new THREE.Matrix4().copy(root.matrixWorld).invert();
  const parts: Model['parts'] = [];
  const box = new THREE.Box3();
  root.traverse((o) => {
    const m = o as THREE.Mesh;
    if (!m.isMesh) return;
    const local = new THREE.Matrix4().multiplyMatrices(inv, m.matrixWorld);
    for (const mat of Array.isArray(m.material) ? m.material : [m.material]) {
      const std = mat as THREE.MeshStandardMaterial;
      if (std.map && std.transparent) { std.transparent = false; std.alphaTest = 0.5; }
      if (/leaf|leaves|bush/i.test(std.name) && !std.userData.sway) { std.userData.sway = true; swayLeaves(std); }
      // match the darker battlefield: the pack's rocks and needles are bright
      if (!std.userData.toned) { std.userData.toned = true; std.color?.multiplyScalar(/rock/i.test(std.name) ? 0.72 : /leaf|leaves/i.test(std.name) && !std.userData.baked ? 0.82 : 1); }
    }
    parts.push({ geometry: m.geometry, material: Array.isArray(m.material) ? m.material[0] : m.material, local });
    m.geometry.computeBoundingBox();
    box.union(m.geometry.boundingBox!.clone().applyMatrix4(local));
  });
  const radius = Math.max(-box.min.x, box.max.x, -box.min.z, box.max.z);
  return { parts, radius, height: box.max.y };
}

export type Placement = { x: number; z: number; y?: number; rot: number; scale: number; squash?: number };

const tmp = new THREE.Matrix4(), q = new THREE.Quaternion(), up = new THREE.Vector3(0, 1, 0);
const pos = new THREE.Vector3(), scl = new THREE.Vector3();

// One InstancedMesh per part of the model, all copies in one draw call each.
export function instances(model: Model, places: Placement[], shadows = true) {
  const group = new THREE.Group();
  if (!places.length) return group;
  for (const part of model.parts) {
    const mesh = new THREE.InstancedMesh(part.geometry, part.material, places.length);
    places.forEach((p, i) => {
      q.setFromAxisAngle(up, p.rot);
      scl.set(p.scale, p.scale * (p.squash ?? 1), p.scale);
      tmp.compose(pos.set(p.x, p.y ?? 0, p.z), q, scl).multiply(part.local);
      mesh.setMatrixAt(i, tmp);
    });
    mesh.castShadow = shadows;
    mesh.receiveShadow = true;
    mesh.computeBoundingSphere();
    group.add(mesh);
  }
  return group;
}
