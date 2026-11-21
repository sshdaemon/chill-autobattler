import type { Plugin } from 'vite';
export function buildManifest(publicDir: string): Record<string, [string, number]>;
export function assetManifest(opts?: { sw?: boolean }): Plugin;
