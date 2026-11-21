# Design

## Context

From the code map (2026-11-21):

- `App.tsx` always mounts `<Canvas>` with `Arena`, `Units` and the rest. The
  arena's GLBs and textures load through drei (`useGLTF`, `useTexture`); unit
  parts and clips through `scene/models.ts` (`fetch`, `GLTFLoader`,
  `TextureLoader`, memoised promise maps); the sim through the worker
  (`fetch('/sim.wasm')`, `import('/wasm_exec.js')`).
- `store.ts` imports `scene/effects.ts` (snapshot effects), which pulls three.js
  into the startup graph.
- Faction pick and the draft are HTML only. Looks, factions and fx JSON are
  bundled into the App chunk.
- `public/` is served verbatim with no hashes; Vite hashes only the JS. No
  service worker, no cache headers; the Go server serves RPCs only. Dev uses
  Vite's server with a proxy to `:8080`.

## Goals / Non-Goals

**Goals:** a fast first screen; battle assets ready by deployment on a normal
connection; a loading screen with progress when they aren't; no re-downloads of
unchanged files; one process serving the game in production.

**Non-Goals:** making the assets smaller; offline RPCs; the dev pages.

## Decisions

**Two stages, by screen.**

| Stage | What | When |
|---|---|---|
| Shell | HTML, app JS without three.js, the sim WASM, the 2D backdrop | at start; loading screen until the sim is ready |
| Battle | three.js and the scene chunk; arena (`nature.glb`, `trees.glb`, ground and tree textures); `parts.json`, the parts and textures of the battle's unit types, `_shared` clips and each type's clips | prefetch from the draft; loading screen at deployment if not ready |

The battle's unit types are known when the draft starts: the player's faction
and Red's drafted army come from the sim's draft plan. The prefetch takes the
union of the player's faction's units and Red's units.

**Keep three.js out of the shell.** The 3D view moves behind
`React.lazy(() => import('./scene/BattleView.tsx'))` (the canvas and
everything in it), mounted from `deploying` on. The store stops importing scene
code: snapshot effects go through a hook the scene registers when it mounts
(`setSceneHooks({ onSnapshot, lastKill })`), and the store skips them while no
scene is mounted (nothing to show). The check is the build's chunk graph: the
entry and App chunks must not import the three chunk statically (a test over
`dist/.vite/manifest.json`).

**Asset groups and prefetch** (`web/src/assets.ts`): `battleAssets(classes)`
lists the URLs; `prefetch(list, onProgress)` downloads them in order with
`fetch` at low priority (`priority: 'low'`) into the service worker's cache (or
the HTTP cache in dev) and resolves with bytes done / total (sizes from the
manifest). The same loaders then read them from cache. The loaders also warm
the in-memory caches (`models.ts` maps, drei's `useGLTF.preload` for the arena)
once the scene chunk has loaded, so deployment renders units without capsules.

**Versioned URLs.** A Vite plugin (`web/asset-manifest.mjs`) at build time:
hashes every file under `public/` (sha256, first 10 hex), and emits the
manifest as a virtual module (`virtual:asset-manifest`: path → `{hash, size}`)
bundled into the App chunk, so a new manifest gives the chunk a new name. Files
keep their paths; URLs become `path?v=<hash>`. `assetUrl(path)` resolves them;
in dev the manifest is empty and paths pass through. Every loader goes through
it:
- `fetch` calls (`parts.json`, clips, the WASM) use `assetUrl`;
- three.js loaders: `THREE.DefaultLoadingManager.setURLModifier(assetUrl)`
  plus the same modifier on drei's loader instances (`useGLTF`'s
  `extendLoader`), which also catches a glTF's relative `.bin` and texture
  URIs;
- the worker gets the WASM and `wasm_exec.js` URLs in its first message.
`?v=` rather than renaming files keeps glTF's relative references working with
one modifier.

**Service worker** (`web/public/sw.js` template, filled at build with the
manifest): on install it does nothing heavy (no precache; stages fetch what
they need); on fetch, a GET whose URL has `?v=` is answered cache-first from
the `assets` cache, and put there when fetched; anything else goes to the
network. On activate it deletes cached entries whose path's hash is not the
manifest's current one. `sw.js` changes whenever the manifest does, so the
browser installs the new worker on the next visit; `skipWaiting` +
`clients.claim` so it takes over at once (a page mid-battle keeps its
in-memory assets). Registered only in production builds.

**Server** (`server/web`): serves `CHILL_WEB` (default `web/dist`) under `/`
next to the RPC routes. Headers:
- `?v=` URLs and Vite's hashed `/assets/*-<hash>.*`: `public, max-age=31536000, immutable`;
- `index.html`, `sw.js`: `no-cache`;
- other files: `max-age=300`.
Gzip (on the fly, cached per file in memory) for `wasm`, `js`, `json`,
`gltf`, `bin`, `glb`, `svg` when the client accepts it; images are left as
they are. The RPC routes keep their auth middleware; static files don't need
it.

**Loading screens.** One component (`ui/Loading.tsx`): title, a progress bar
with MB done / total, and what is loading ("the battlefield", "your army"). At
start it tracks the WASM download (the worker reports progress from a
streaming read); at deployment it tracks the battle prefetch. The deploy
countdown starts when the battle stage is ready (`startDeployTimer` after the
gate), within the server's 3-minute window.

**Backdrop.** The faction and draft screens get a CSS backdrop (sky gradient,
a ground band and a soft vignette). The planned static image of the arena was
left out: the CSS alone reads well behind the panels and costs nothing.

## Changes while building

- **The service worker also keeps Vite's hashed JS chunks** (the build hands
  it the release's chunk list), so a returning player downloads no JS either.
- **The loaded 3D view is shared** (`battleView.ts`): `React.lazy` suspends on
  its first render even when the module is already loaded, which flashed the
  loading screen at every deployment; App renders the loaded module directly.
- **Loader URLs are set in a module imported first** (`scene/loaderUrls.ts`):
  `Nature.tsx` preloads its model when evaluated, before `BattleView`'s body
  ran, and that one request went out unversioned.
- **The server mounts at `/`, not `GET /`**: Go's router rejects `GET /` next
  to Connect's method-less RPC patterns; the handler refuses other methods.
- **The server checks `?v=`**: a hash that doesn't match the file on disk gets
  a short cache, so an old client can't pin new content under an old URL.

## Tests and checks

- The web package has no unit-test runner; the typecheck stays, and the
  build check and browser checks below cover `assetUrl` and `battleAssets`.
- A build check (`pnpm --filter web check:bundle`): after `vite build`, the
  entry and App chunks don't statically import the three chunk, and the
  manifest lists every `public/` file.
- Go tests for the static handler: cache headers per kind of path, gzip when
  accepted and not otherwise, `index.html` at `/`, RPC routes untouched.
- In the browser, on a production build served by the Go server:
  - cold start: the faction screen requests no arena, part, clip or three.js
    file;
  - a draft prefetches, deployment then starts with no loading screen;
  - with throttling, the deployment loading screen shows and the countdown
    waits;
  - a reload downloads nothing (all from the service worker);
  - after changing one clip file and rebuilding, only that file downloads.

## Risks / Trade-offs

- **Service worker pitfalls** (a stuck old worker): no-cache on `sw.js`, and
  it serves only `?v=` URLs from cache, so the HTML and RPCs never come from it.
- **The first battle waits longer on a slow connection** (assets that used to
  load during the faction screen now load from the draft): the draft lasts up
  to six 15-second turns, which covers about 18 MB on a 10 Mbit/s line.
- **More moving parts in production than in dev:** the browser checks run on a
  production build.
