// Bakes procedural trees with EZ-Tree (MIT, a dev-only dependency) into one GLB
// for the game, so players never download the generator.
//
//   node scripts/bake-trees.mjs        (from web/)
//
// Output: public/assets/trees/
// - trees.glb: one root node per variant (bark + leaves meshes), no textures.
//   Material names say which textures to use (bark_<type>, leaves_<type>);
//   their extras hold the tint and the bark texture repeat.
// - bark_<type>_color.jpg / bark_<type>_normal.jpg (512 px), leaves_<type>.png (512 px):
//   EZ-Tree's textures (bark from Poly Haven and TextureCan, CC0).
//
// Geometry is generated in Node; the DOM stubs below only keep three's
// texture loader quiet (textures are applied in the game, not here).
import { execFileSync } from 'node:child_process';
import { mkdirSync, rmSync, writeFileSync, readdirSync, statSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

globalThis.document = { createElementNS: () => ({ addEventListener() {}, removeEventListener() {}, setAttribute() {}, style: {} }) };
globalThis.self = globalThis;
globalThis.FileReader = class {
  readAsArrayBuffer(blob) { blob.arrayBuffer().then((b) => { this.result = b; this.onloadend?.(); }); }
  readAsDataURL(blob) { blob.arrayBuffer().then((b) => { this.result = 'data:application/octet-stream;base64,' + Buffer.from(b).toString('base64'); this.onloadend?.(); }); }
};

const THREE = await import('three');
const { GLTFExporter } = await import('three/examples/jsm/exporters/GLTFExporter.js');
const { Tree } = await import('@dgreenheck/ez-tree');

const HERE = dirname(fileURLToPath(import.meta.url));
const WEB = join(HERE, '..');
const OUT = join(WEB, 'public', 'assets', 'trees');
const ASSETS = join(WEB, 'node_modules', '@dgreenheck', 'ez-tree', 'src', 'lib', 'assets');

// name, preset, seed, option tweaks
const VARIANTS = [
  ['Oak_1', 'Oak Medium', 11, {}],
  ['Oak_2', 'Oak Medium', 23, {}],
  ['Ash_1', 'Ash Medium', 5, {}],
  ['Aspen_1', 'Aspen Medium', 7, {}],
  ['Pine_1', 'Pine Medium', 3, {}],
  ['Bush_1', 'Bush 1', 9, {}],
];

const merge = (a, b) => { for (const [k, v] of Object.entries(b)) typeof v === 'object' ? merge(a[k], v) : (a[k] = v); };

const scene = new THREE.Scene();
const used = { bark: new Set(), leaves: new Set() };
let total = 0;
for (const [name, preset, seed, tweaks] of VARIANTS) {
  const t = new Tree();
  t.loadPreset(preset);
  t.options.seed = seed;
  merge(t.options, tweaks);
  t.generate();
  const root = new THREE.Group();
  root.name = name;
  const { bark, leaves } = t.options;
  t.children.forEach((mesh, i) => {
    const isBark = i === 0;
    const type = isBark ? bark.type : leaves.type;
    used[isBark ? 'bark' : 'leaves'].add(type);
    const m = new THREE.MeshStandardMaterial({ name: `${isBark ? 'bark' : 'leaves'}_${type}`, side: isBark ? THREE.FrontSide : THREE.DoubleSide });
    m.userData = isBark ? { tint: bark.tint, repeat: [bark.textureScale.x, bark.textureScale.y] } : { tint: leaves.tint };
    const out = new THREE.Mesh(mesh.geometry, m);
    out.name = isBark ? 'bark' : 'leaves';
    root.add(out);
    total += (mesh.geometry.index ? mesh.geometry.index.count : mesh.geometry.attributes.position.count) / 3;
  });
  scene.add(root);
  console.log(name.padEnd(8), preset.padEnd(13), 'tris', Math.round(root.children.reduce((n, c) => n + (c.geometry.index ? c.geometry.index.count : c.geometry.attributes.position.count) / 3, 0)));
}

rmSync(OUT, { recursive: true, force: true });
mkdirSync(OUT, { recursive: true });
const glb = await new GLTFExporter().parseAsync(scene, { binary: true });
writeFileSync(join(OUT, 'trees.glb'), Buffer.from(glb));

const sips = (args) => execFileSync('sips', args, { stdio: 'ignore' });
for (const type of used.bark) {
  sips(['-Z', '512', '-s', 'format', 'jpeg', '-s', 'formatOptions', '78', join(ASSETS, 'bark', `${type}_color_1k.jpg`), '--out', join(OUT, `bark_${type}_color.jpg`)]);
  sips(['-Z', '512', '-s', 'format', 'jpeg', '-s', 'formatOptions', '78', join(ASSETS, 'bark', `${type}_normal_1k.jpg`), '--out', join(OUT, `bark_${type}_normal.jpg`)]);
}
for (const type of used.leaves) sips(['-Z', '512', join(ASSETS, 'leaves', `${type}_color.png`), '--out', join(OUT, `leaves_${type}.png`)]);

let bytes = 0;
for (const f of readdirSync(OUT).sort()) { const s = statSync(join(OUT, f)).size; bytes += s; console.log(`${(s / 1e6).toFixed(2)} MB  ${f}`); }
console.log(`${(bytes / 1e6).toFixed(2)} MB  total, ${Math.round(total)} triangles over ${VARIANTS.length} variants`);
