// The service worker (openspec change asset-loading), written to dist/sw.js by
// the asset-manifest Vite plugin with this release's manifest. It keeps
// versioned assets (`?v=<hash>` URLs) and the release's hashed JS chunks in
// Cache Storage across visits, answers them from there first, and on each new
// release drops whatever the release no longer lists. Everything else (the
// page, RPCs) goes to the network untouched.
const MANIFEST = self.__MANIFEST__; // public path → [hash, bytes]
const CHUNKS = new Set(self.__CHUNKS__); // this release's hashed JS/CSS chunk paths
const CACHE = 'chill-assets';

const cacheable = (u) => u.origin === self.location.origin && (u.searchParams.has('v') || CHUNKS.has(u.pathname));
const current = (u) => {
  if (CHUNKS.has(u.pathname)) return true;
  const e = MANIFEST[decodeURIComponent(u.pathname)];
  return !!e && u.searchParams.get('v') === e[0];
};

self.addEventListener('install', () => self.skipWaiting());

self.addEventListener('activate', (event) => event.waitUntil((async () => {
  const cache = await caches.open(CACHE);
  for (const req of await cache.keys()) {
    if (!current(new URL(req.url))) await cache.delete(req);
  }
  await self.clients.claim();
})()));

self.addEventListener('fetch', (event) => {
  const req = event.request;
  if (req.method !== 'GET') return;
  const u = new URL(req.url);
  if (!cacheable(u)) return;
  event.respondWith((async () => {
    const cache = await caches.open(CACHE);
    const hit = await cache.match(req);
    if (hit) return hit;
    const res = await fetch(req);
    if (res.ok && current(u)) await cache.put(req, res.clone());
    return res;
  })());
});
