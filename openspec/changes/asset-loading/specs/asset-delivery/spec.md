# Spec Delta

## Purpose

Decides what the client downloads, when, and how it is cached, so the player
waits only for what the screen in front of them needs and never downloads an
unchanged asset twice.

## ADDED Requirements

### Requirement: Start with the shell and the sim only
Until a battle is being prepared, the client SHALL download only the app shell (HTML, its scripts and styles), the sim (WASM) and a light 2D backdrop. It SHALL NOT download three.js, the arena (models and textures), unit parts or animation clips on the faction screen.

#### Scenario: Faction screen on a cold start
- **WHEN** a new player opens the game and reaches the faction screen
- **THEN** no three.js chunk, arena model or texture, unit part or animation clip has been requested

### Requirement: Loading screen at start
Until the sim is ready, the client SHALL show a loading screen with its progress.

#### Scenario: Slow connection
- **WHEN** the sim takes several seconds to download
- **THEN** the loading screen shows how much has arrived

### Requirement: Battle assets prefetch during the draft
Once a battle's draft is known, the client SHALL start downloading that battle's assets in the background: three.js and the scene code, the arena, and the parts and animation clips of every unit type that can appear in the battle (the player's faction and Red's units). Prefetching SHALL NOT block the draft.

#### Scenario: Draft
- **WHEN** the player is drafting
- **THEN** the battle's assets are downloading in the background and the draft stays responsive

### Requirement: Loading screen before deployment
Deployment SHALL start only when the battle's assets are ready. If they aren't when the draft ends, the client SHALL show a loading screen with progress until they are; the deployment countdown SHALL start after it.

#### Scenario: Fast connection
- **WHEN** the battle's assets finished downloading during the draft
- **THEN** deployment starts straight after the draft, with no loading screen

#### Scenario: Slow connection
- **WHEN** the draft ends before the battle's assets are ready
- **THEN** a loading screen shows progress, and the deployment countdown starts only once they are ready

### Requirement: Versioned asset URLs
Every runtime asset in a production build SHALL be requested under a URL that changes exactly when its content changes. The client SHALL resolve asset paths through a manifest built with the client, including files a model loads by relative path (a glTF's buffers and textures).

#### Scenario: One file changes
- **WHEN** a release changes one unit's animation clips and nothing else
- **THEN** only that file's URL changes; every other asset keeps its URL

### Requirement: Assets cached across visits
In a production build, a service worker SHALL keep versioned assets across visits, serve them from its cache before the network, and remove cached files that the current manifest no longer lists. The HTML, the manifest and the service worker itself SHALL always be checked with the server.

#### Scenario: Returning player
- **WHEN** a player who has played before opens the game again with no new release
- **THEN** no asset is downloaded again

#### Scenario: New release
- **WHEN** a player returns after a release that changed the sim
- **THEN** only the new sim and other changed files are downloaded, and the old sim is removed from the cache

### Requirement: The server serves the client
The game server SHALL serve the built client with cache headers: versioned files cached for a year as immutable; the HTML, the manifest and the service worker revalidated on every load. It SHALL compress compressible files (WASM, JSON, glTF, JavaScript) for clients that accept it.

#### Scenario: Hashed file
- **WHEN** a browser requests a versioned asset
- **THEN** the response says it may be cached for a year without revalidation

#### Scenario: Compressed sim
- **WHEN** a browser that accepts gzip requests the sim
- **THEN** it receives it gzip-compressed
