import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// Big libraries in their own chunks, so they cache across app updates.
const CHUNKS: [string, RegExp][] = [
  ['three', /node_modules\/(three|@react-three|three-stdlib|troika|camera-controls|maath|meshline|stats-gl)/],
  ['react', /node_modules\/(react|react-dom|scheduler|zustand)\//],
  ['rpc', /node_modules\/(@connectrpc|@bufbuild)\//],
];

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // Connect RPCs go to the Go server (pnpm server)
    proxy: { '/chill.v1.': 'http://localhost:8080' },
  },
  build: {
    // three.js (+ react-three) alone is ~975 kB (~260 kB gzipped) and can't be split usefully
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          return CHUNKS.find(([, re]) => re.test(id))?.[0];
        },
      },
    },
  },
});
