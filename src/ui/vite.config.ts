import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  // Relative base path — required for Docker Desktop Extensions, which serve
  // UI assets from the container filesystem rather than a web-server root.
  base: './',
  build: {
    outDir: 'dist',
  },
});
