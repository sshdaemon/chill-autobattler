# Tasks

## 1. Versioned URLs

- [x] 1.1 Vite plugin `asset-manifest.mjs`: hash `public/` files, `virtual:asset-manifest` (path → hash, size)
- [x] 1.2 `src/assets.ts`: `assetUrl`, manifest types; dev passes paths through
- [x] 1.3 Route every loader through it: `models.ts` fetches and loaders, drei loaders (URL modifier), the worker's WASM and `wasm_exec.js`

## 2. Stages

- [x] 2.1 Store no longer imports scene code (scene hooks registered by the scene)
- [x] 2.2 `BattleView` (canvas and scene) lazy-loaded, mounted from deployment
- [x] 2.3 2D backdrop for faction and draft (CSS; the WebP was left out, see design)
- [x] 2.4 `battleAssets(classes)` and `prefetch` with progress; start it when the draft starts
- [x] 2.5 Warm the in-memory caches (models, clips, arena) once the scene chunk is in
- [x] 2.6 Loading screen component; at start (WASM progress from the worker) and before deployment; the deploy countdown starts after it

## 3. Service worker

- [x] 3.1 `sw.js` built with the manifest: cache-first for `?v=` URLs, prune on activate, `skipWaiting`/`clients.claim`
- [x] 3.2 Register in production builds only

## 4. Server

- [x] 4.1 `server/web`: serve `CHILL_WEB` (default `web/dist`) with cache headers and gzip; wire into `cmd/server`
- [x] 4.2 Go tests for headers, gzip and routing
- [x] 4.3 `pnpm build` at the root: WASM, then the web build; `pnpm serve` runs the server on it

## 5. Checks

- [x] 5.1 Bundle check: three.js not in the startup graph; manifest covers `public/`
- [x] 5.2 Browser checks on a production build (cold start, prefetch, throttled loading screen, reload from cache, one changed file)
- [x] 5.3 `pnpm test`; `CLAUDE.md` notes for loading and caching
