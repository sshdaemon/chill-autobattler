// Bundle check (openspec change asset-loading), run after `vite build`:
// - three.js stays out of the startup graph: neither the entry nor the App
//   chunk imports the three chunk, directly or through other static imports;
// - the service worker's manifest lists every file in public/, with the same
//   hashes the app was built with;
// - no unused art (openspec change retire-classic): every file under
//   public/assets/parts is parts.json, a part or item model, a buffer or
//   image a model uses, or a material texture; every item is worn by some
//   look; every clip bundle belongs to a look file and holds only its clips.
//   Outfit parts are exempt from "worn": the bake keeps the whole outfit kit.
import { readdirSync, readFileSync } from 'node:fs';
import { buildManifest } from './asset-manifest.mjs';

const dist = new URL('./dist/', import.meta.url);
const vite = JSON.parse(readFileSync(new URL('.vite/manifest.json', dist), 'utf8'));
const fail = (msg) => { console.error('check-bundle:', msg); process.exitCode = 1; };

const staticGraph = (key, seen = new Set()) => {
  if (seen.has(key)) return seen;
  seen.add(key);
  for (const k of vite[key]?.imports ?? []) staticGraph(k, seen);
  return seen;
};
const startup = new Set([...staticGraph('index.html'), ...staticGraph('src/App.tsx')]);
const three = [...startup].filter((k) => /three/.test(vite[k]?.file ?? k));
if (three.length) fail(`three.js is in the startup bundle: ${three.join(', ')}`);

const sw = readFileSync(new URL('sw.js', dist), 'utf8');
const m = /const MANIFEST = (\{.*?\}); \/\/ public path/s.exec(sw);
if (!m) fail('dist/sw.js has no manifest');
else {
  const inSw = JSON.parse(m[1]);
  const want = buildManifest(new URL('./public/', import.meta.url).pathname);
  for (const [path, [hash]] of Object.entries(want)) {
    if (inSw[path]?.[0] !== hash) fail(`${path}: missing from the service worker's manifest or stale`);
  }
}
// unused art
const pub = new URL('./public/assets/', import.meta.url);
const looksDir = new URL('./src/units/looks/', import.meta.url);
const json = (url) => JSON.parse(readFileSync(url, 'utf8'));
const ALLOW = new Set([]); // public/assets paths loaded some other way, if ever needed
const looks = Object.fromEntries(readdirSync(looksDir).filter((f) => f.endsWith('.json')).map((f) => [f.slice(0, -5), json(new URL(f, looksDir))]));
const parts = json(new URL('parts/parts.json', pub));
const used = new Set(['parts.json']);
const models = [...Object.keys(parts.parts), ...Object.entries(parts.items).filter(([, it]) => it.kind !== 'cloth').map(([n]) => n)]; // cloth is simulated, not a model
for (const name of models) {
  const g = `${name}.gltf`;
  used.add(g);
  let model;
  try { model = json(new URL(`parts/${g}`, pub)); } catch { fail(`parts.json lists ${name}, but parts/${g} is missing`); continue; }
  for (const x of [...(model.buffers ?? []), ...(model.images ?? [])]) if (x.uri) used.add(x.uri);
}
for (const m of Object.values(parts.materials)) for (const f of [...m.variants, m.mask]) used.add(f);
for (const f of readdirSync(new URL('parts/', pub))) {
  if (!used.has(f) && !ALLOW.has(`parts/${f}`)) fail(`parts/${f} is used by no part, item or material: delete it (or re-bake)`);
}
const worn = new Set(Object.values(looks).flatMap((l) => (l.recipes ?? []).flatMap((r) => r.items ?? [])));
for (const it of Object.keys(parts.items)) {
  if (!worn.has(it)) fail(`item ${it} is worn by no look: drop it from tools/blender/items and parts.json`);
}
for (const f of readdirSync(new URL('clips/', pub))) {
  const name = f.replace(/\.json$/, '');
  if (!looks[name]) { fail(`clips/${f} belongs to no look file (src/units/looks/${name}.json)`); continue; }
  const own = new Set(Object.keys(looks[name].clips ?? {}));
  for (const c of json(new URL(`clips/${f}`, pub))) {
    if (!own.has(c.name)) fail(`clips/${f} packs ${c.name}, which looks/${name}.json doesn't list (re-run tools/assets/ual_clips.py)`);
  }
}

if (!process.exitCode) console.log(`check-bundle: ok (${startup.size} startup chunks, ${Object.keys(JSON.parse(m[1])).length} versioned assets)`);
