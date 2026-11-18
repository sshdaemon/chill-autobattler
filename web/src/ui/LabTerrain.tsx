// Dev only (unit lab): a strip of battlefield that scrolls under walking units,
// so speed reads naturally. A dirt path where the units walk, grass blades
// beside it, rocks, ferns, flowers, bushes and a few trees further out, all
// with the game's own textures and models. Laid out over one 24 m period and
// repeated, so it wraps seamlessly as it scrolls. No game store involved.
import { useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { useGLTF, useTexture } from '@react-three/drei';
import * as THREE from 'three';
import { NATURE_GLB, TREES_GLB, TREE_TEXTURES, dressTrees, instances, readModel, type Model, type Placement } from '../scene/nature.ts';
import { rng } from '../scene/natureShader.ts';

const P = 24; // period along the walk (m), a multiple of the 2 m texture tile
const W = 40; // width of the strip (m)
const PATH = 2.6; // half-width of the dirt path

// One clump of three tapered blades (as the battlefield's grass), base darker than the tip.
function clump() {
  const pos: number[] = [], col: number[] = [], idx: number[] = [];
  for (let b = 0; b < 3; b++) {
    const a = (b / 3) * Math.PI + 0.3, cx = Math.cos(a), cz = Math.sin(a), w = 0.035, base = pos.length / 3;
    pos.push(-cx * w, 0, -cz * w, cx * w, 0, cz * w, -cx * w * 0.6, 0.5, -cz * w * 0.6, cx * w * 0.6, 0.5, cz * w * 0.6, 0, 1, 0);
    for (const t of [0, 0, 0.5, 0.5, 1]) col.push(0.12 + 0.16 * t, 0.2 + 0.2 * t, 0.06 + 0.06 * t); // dark base, muted green tip (matches the ground)
    idx.push(base, base + 1, base + 2, base + 1, base + 3, base + 2, base + 2, base + 3, base + 4);
  }
  const g = new THREE.BufferGeometry();
  g.setAttribute('position', new THREE.Float32BufferAttribute(pos, 3));
  g.setAttribute('color', new THREE.Float32BufferAttribute(col, 3));
  g.setIndex(idx);
  g.computeVertexNormals();
  const n = g.getAttribute('normal');
  for (let i = 0; i < n.count; i++) n.setXYZ(i, 0, 1, 0); // lit like the ground below
  return g;
}

export function LabTerrain({ floor }: { floor: { current: { floor: number } } }) {
  const group = useRef<THREE.Group>(null);
  const tex = useTexture({ grass: '/assets/nature/Grass004_color.jpg', grassN: '/assets/nature/Grass004_normal.jpg', dirt: '/assets/nature/Ground037_color.jpg', dirtN: '/assets/nature/Ground037_normal.jpg' });
  const nature = useGLTF(NATURE_GLB), trees = useGLTF(TREES_GLB), treeTex = useTexture(TREE_TEXTURES);

  const models = useMemo(() => {
    const out: Record<string, Model> = {};
    for (const c of nature.scene.children) out[c.name] = readModel(c);
    dressTrees(trees.scene, treeTex);
    for (const c of trees.scene.children) out['ez' + c.name] = readModel(c);
    return out;
  }, [nature, trees, treeTex]);

  const scene = useMemo(() => {
    for (const [k, t] of Object.entries(tex)) {
      t.wrapS = t.wrapT = THREE.RepeatWrapping;
      t.anisotropy = 8;
      t.colorSpace = k.endsWith('N') ? THREE.NoColorSpace : THREE.SRGBColorSpace;
    }
    const g = new THREE.Group();
    const ground = (map: THREE.Texture, normal: THREE.Texture, width: number, y: number, tile: number, tint: string) => {
      const m = map.clone(), n = normal.clone();
      for (const t of [m, n]) { t.repeat.set(width / tile, (3 * P) / tile); t.needsUpdate = true; }
      const mesh = new THREE.Mesh(new THREE.PlaneGeometry(width, 3 * P), new THREE.MeshStandardMaterial({ map: m, normalMap: n, roughness: 0.95, color: tint }));
      mesh.rotation.x = -Math.PI / 2;
      mesh.position.y = y;
      mesh.receiveShadow = true;
      g.add(mesh);
    };
    ground(tex.grass, tex.grassN, W, 0, 2, '#b9c4a0');
    ground(tex.dirt, tex.dirtN, PATH * 2, 0.004, 2, '#c8b89c');

    // grass blades, rocks, plants and trees over one period, repeated three times
    const r = rng(4242), items: { model: string; p: Placement }[] = [];
    const blades: THREE.Matrix4[] = [];
    const m4 = new THREE.Matrix4(), q = new THREE.Quaternion(), up = new THREE.Vector3(0, 1, 0);
    for (let i = 0; i < 2600; i++) {
      const x = (r() * 2 - 1) * (W / 2 - 1), z = (r() - 0.5) * P;
      if (Math.abs(x) < PATH + 0.15 && r() < 0.92) continue; // a few tufts on the path's edge
      q.setFromAxisAngle(up, r() * 6.28);
      blades.push(m4.compose(new THREE.Vector3(x, 0, z), q, new THREE.Vector3(0.9 + r() * 0.4, 0.22 + r() * 0.3, 0.9 + r() * 0.4)).clone());
    }
    const scatter = (n: number, names: string[], minX: number, maxX: number, size: () => number, byHeight: boolean) => {
      for (let i = 0; i < n; i++) {
        const x = (r() < 0.5 ? -1 : 1) * (minX + r() * (maxX - minX)), z = (r() - 0.5) * P;
        const model = names[Math.floor(r() * names.length)], mm = models[model];
        if (!mm) continue;
        items.push({ model, p: { x, z, rot: r() * 6.28, scale: byHeight ? size() / mm.height : size() / mm.radius, squash: 0.7 + r() * 0.3, y: -0.04 } });
      }
    };
    scatter(10, ['Pebble_Round_2', 'Rock_Medium_1', 'Rock_Medium_2'], PATH - 0.3, PATH + 1.2, () => 0.15 + r() * 0.2, false); // stones along the path
    scatter(14, ['Fern_1', 'Flower_3_Group'], PATH + 0.6, 7, () => 0.35 + r() * 0.3, false);
    scatter(8, ['Rock_Medium_1', 'Rock_Medium_2', 'Rock_Medium_3'], PATH + 1.5, 9, () => 0.5 + r() * 0.8, false);
    scatter(6, ['Bush_Common', 'ezBush_1'], PATH + 2.5, 10, () => 0.8 + r() * 0.6, false);
    scatter(4, ['ezOak_1', 'ezAsh_1', 'ezAspen_1'], 6.5, 12, () => 5 + r() * 2.5, true);
    scatter(6, ['Pine_1', 'Pine_2', 'Pine_5'], 10, 18, () => 6 + r() * 3, true);
    const by = new Map<string, Placement[]>();
    for (const copy of [-P, 0, P]) {
      for (const it of items) by.set(it.model, [...(by.get(it.model) ?? []), { ...it.p, z: it.p.z + copy }]);
    }
    for (const [name, ps] of by) if (models[name]) g.add(instances(models[name], ps, true));
    const grass = new THREE.InstancedMesh(clump(), new THREE.MeshStandardMaterial({ vertexColors: true, roughness: 0.9, side: THREE.DoubleSide }), blades.length * 3);
    let k = 0;
    for (const copy of [-P, 0, P]) for (const b of blades) grass.setMatrixAt(k++, m4.makeTranslation(0, 0, copy).multiply(b));
    grass.receiveShadow = true;
    g.add(grass);
    return g;
  }, [tex, models]);

  // the ground slides back under the walking units, wrapping every period
  useFrame(() => { if (group.current) group.current.position.z = -(floor.current.floor % P); });
  return <group ref={group}><primitive object={scene} /></group>;
}
