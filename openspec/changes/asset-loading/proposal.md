# Proposal

## Why

The client downloads far more than it needs before the player can do
anything, and downloads it again whenever the browser drops its cache:

- The 3D canvas, and with it the arena, is mounted on every screen, so the
  faction screen pulls the sim (4.8 MB of WASM) plus about 11 MB of battlefield
  models and textures (`nature.glb`, `trees.glb`, ground and tree textures) and
  the three.js chunk (1.3 MB), though faction pick and the draft are plain HTML.
- Unit parts and every unit's animation clips (up to about 7 MB) load at
  deployment while units stand as capsules; nothing loads them ahead.
- Nothing in `public/` carries a version, so a returning player gets whatever
  the HTTP cache kept, with no way to tell a changed file from an unchanged one.
- No production server serves the client at all.

## What Changes

- **Load in stages, by need.** At start only the app shell and the sim; the
  faction and draft screens get a light 2D backdrop (no 3D). Once a battle's
  draft is known, the battle's assets (three.js and the scene code, the arena,
  the parts of both armies' unit types, their animation clips) prefetch in the
  background. Deployment waits on them only if they're not there yet, behind a
  loading screen with progress.
- **Loading screens with progress** at start (until the sim is ready) and
  before deployment (until the battle's assets are ready).
- **Versioned assets.** A build step hashes every runtime asset and writes a
  manifest; the client asks for each asset by its hash, so a changed file gets
  a new URL and an unchanged one keeps its old one.
- **A service worker** keeps hashed assets in Cache Storage across visits,
  serves them cache-first, and drops entries whose hash is no longer in the
  manifest. A new release re-downloads only the files that changed.
- **The Go server serves the built client** (`web/dist`) with long immutable
  cache headers for hashed files, no-cache for the HTML, manifest and service
  worker, and gzip for compressible files (the WASM shrinks to about a third).

Out of scope:

- Shrinking the assets themselves (mesh compression, smaller textures).
- Offline play beyond replaying cached assets (the RPCs still need the server).
- The dev pages (`?lab`, `?preview`): they keep loading what they need, unhashed.

## Capabilities

### New Capabilities

- `asset-delivery`: when the client loads which assets, its loading screens,
  versioned caching, and how the server serves the client.

### Modified Capabilities

None.

## Impact

- `web/`: lazy 3D scene (three.js out of the startup bundle; the store no
  longer imports scene code), 2D backdrop, asset groups and prefetch, loading
  screens, URL resolution through the manifest (fetches, GLTF and texture
  loaders, the WASM), a Vite plugin for hashing and the manifest, the service
  worker.
- `server/`: static serving of `web/dist` with cache headers and gzip.
- `package.json`: a production build that includes the WASM; a size check.
