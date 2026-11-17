import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // Connect RPCs go to the Go server (pnpm server)
    proxy: { '/chill.v1.': 'http://localhost:8080' },
  },
});
