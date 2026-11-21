// Versioned assets (openspec change asset-loading): at build time every file
// under public/ is hashed, and the hashes become `virtual:asset-manifest`
// (path → [hash, bytes]), bundled into the app and the sim worker. The client
// asks for each asset as `path?v=<hash>` (src/assets.ts), so a file's URL
// changes exactly when its content does. The same manifest is written into the
// service worker (sw.template.js → dist/sw.js), which keeps those URLs cached
// across visits. In dev the manifest is empty and paths pass through.
import { createHash } from 'node:crypto';
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative, sep } from 'node:path';

const ID = 'virtual:asset-manifest';
const RESOLVED = '\0' + ID;

function walk(dir, out = []) {
  for (const name of readdirSync(dir).sort()) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) walk(p, out);
    else out.push(p);
  }
  return out;
}

export function buildManifest(publicDir) {
  const manifest = {};
  for (const file of walk(publicDir)) {
    const rel = '/' + relative(publicDir, file).split(sep).join('/');
    const data = readFileSync(file);
    manifest[rel] = [createHash('sha256').update(data).digest('hex').slice(0, 10), data.length];
  }
  return manifest;
}

// sw: write dist/sw.js (the main build only, not the worker's)
export function assetManifest({ sw = true } = {}) {
  let publicDir = '', build = false, manifest;
  const get = () => (manifest ??= build ? buildManifest(publicDir) : {});
  return {
    name: 'asset-manifest',
    configResolved(c) {
      publicDir = c.publicDir;
      build = c.command === 'build';
    },
    resolveId(id) {
      return id === ID ? RESOLVED : undefined;
    },
    load(id) {
      return id === RESOLVED ? `export default ${JSON.stringify(get())};` : undefined;
    },
    generateBundle(_, bundle) {
      if (!build || !sw) return;
      // the release's hashed chunks (Vite names them <name>-<hash>.js/.css)
      const chunks = Object.keys(bundle).filter((f) => /\.(js|css)$/.test(f) && f.startsWith('assets/')).map((f) => '/' + f);
      const template = readFileSync(new URL('./sw.template.js', import.meta.url), 'utf8');
      this.emitFile({
        type: 'asset', fileName: 'sw.js',
        source: template.replace('self.__MANIFEST__', JSON.stringify(get())).replace('self.__CHUNKS__', JSON.stringify(chunks)),
      });
    },
  };
}
