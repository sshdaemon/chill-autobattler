// Asset URLs and loading stages (openspec change asset-loading). No three.js
// here: this module is part of the startup bundle.
import manifest from 'virtual:asset-manifest';
import { ARENA_FILES } from './scene/arenaFiles.ts';
import { CLIP_BUNDLES, lookOf } from './units/looks.ts';

// path → [content hash, bytes], from the build (empty in dev).
const M = manifest as Record<string, [string, number]>;

// assetUrl is the versioned URL of a public asset: `/x.json` → `/x.json?v=<hash>`.
// It takes a path or a same-origin URL (glTF loaders hand it resolved URLs);
// anything else, an already-versioned URL, or a file the manifest lacks (dev)
// passes through.
export function assetUrl(url: string): string {
  if (url.startsWith('data:') || url.startsWith('blob:') || url.includes('?')) return url;
  let origin = '', path = url;
  if (/^[a-z]+:\/\//i.test(url)) {
    const u = new URL(url);
    if (u.origin !== self.location.origin) return url;
    origin = u.origin;
    path = u.pathname;
  }
  if (!path.startsWith('/')) path = '/' + path;
  const e = M[decodeURIComponent(path)];
  return e ? `${origin}${path}?v=${e[0]}` : url;
}

// assetBytes is a public asset's size (0 when unknown, as in dev).
export const assetBytes = (path: string) => M[path]?.[1] ?? 0;

// ---- The battle stage: what a battle needs, prefetched during the draft ----


const PARTS = '/assets/parts/';
type PartsJson = {
  parts: Record<string, { materials: string[] }>;
  items: Record<string, { kind?: 'cloth' }>;
  materials: Record<string, { variants: string[]; mask: string }>;
};

// battleFiles lists every file a battle with these unit types needs: the
// arena, the parts (glTF and buffer) and outfit textures of each type's
// recipes, and the animation clips.
export function battleFiles(classes: number[], parts: PartsJson): string[] {
  const out = new Set<string>([...ARENA_FILES, `${PARTS}parts.json`, ...CLIP_BUNDLES.map((b) => `/assets/clips/${b}.json`)]);
  for (const cls of new Set(classes)) {
    for (const r of lookOf(cls).recipes) {
      const names = [...r.parts, r.head, ...r.hair, ...r.items.filter((it) => parts.items[it]?.kind !== 'cloth')];
      for (const n of names) {
        out.add(`${PARTS}${n}.gltf`);
        out.add(`${PARTS}${n}.bin`);
        for (const m of parts.parts[n]?.materials ?? []) {
          const spec = parts.materials[m];
          if (!spec) continue;
          out.add(PARTS + spec.variants[Math.min(r.variant, spec.variants.length - 1)]);
          out.add(PARTS + spec.mask);
        }
      }
    }
  }
  // only files the build has (in dev, every listed file is tried)
  return [...out].filter((f) => !Object.keys(M).length || M[f]);
}

export type Progress = { done: number; total: number }; // bytes

// prefetch downloads files into the browser's caches (the service worker's
// in production) a few at a time, reporting bytes as they arrive. The loaders
// later read them from there.
export async function prefetch(files: string[], onProgress: (p: Progress) => void, parallel = 4): Promise<void> {
  const sizes = new Map(files.map((f) => [f, assetBytes(f)]));
  const p: Progress = { done: 0, total: [...sizes.values()].reduce((a, b) => a + b, 0) };
  onProgress({ ...p });
  const queue = [...files];
  const worker = async () => {
    for (let f = queue.shift(); f; f = queue.shift()) {
      const res = await fetch(assetUrl(f), { priority: 'low' } as RequestInit);
      const body = await res.arrayBuffer(); // read it all so it lands in the cache
      const n = sizes.get(f) || body.byteLength;
      if (!sizes.get(f)) p.total += n; // dev: sizes learnt as files arrive
      p.done += n;
      onProgress({ ...p });
    }
  };
  await Promise.all(Array.from({ length: parallel }, worker));
}

// fetchParts reads the parts library's index (to list a battle's files).
export const fetchParts = () => fetch(assetUrl(`${PARTS}parts.json`)).then((r) => r.json() as Promise<PartsJson>);
