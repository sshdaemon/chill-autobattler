// Bundle check (openspec change asset-loading), run after `vite build`:
// - three.js stays out of the startup graph: neither the entry nor the App
//   chunk imports the three chunk, directly or through other static imports;
// - the service worker's manifest lists every file in public/, with the same
//   hashes the app was built with.
import { readFileSync } from 'node:fs';
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
if (!process.exitCode) console.log(`check-bundle: ok (${startup.size} startup chunks, ${Object.keys(JSON.parse(m[1])).length} versioned assets)`);
