// The CC0 nature models (Quaternius Stylized Nature MegaKit, packed by
// tools/assets/pack_nature.py) and a helper that draws many copies of one
// model as instanced meshes. Presentation only.
import * as THREE from 'three';

export const NATURE_GLB = '/assets/nature/nature.glb';

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
