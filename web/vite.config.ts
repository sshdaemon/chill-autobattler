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
  // Pre-bundle the 3D stack together at startup: if Vite discovers one of these
  // mid-session it re-optimises, and the page ends up with two copies of React.
  optimizeDeps: {
    include: ['three', '@react-three/fiber', '@react-three/drei', '@react-three/postprocessing', 'postprocessing', 'react', 'react-dom', 'zustand'],
  },
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
