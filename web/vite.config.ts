import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    // The output is embedded into the Go binary (web/embed.go): one file to
    // deploy, and no window where the API is new and the assets are old.
    outDir: 'dist',
    emptyOutDir: true,
    sourcemap: false,
    // three.js is ~600 kB and only the home hero needs it; it is split out
    // and loaded after first paint (README → Hero ball → Performance).
    chunkSizeWarningLimit: 700,
  },
  server: {
    port: 5173,
    proxy: {
      // Development only. In production the Go binary serves everything.
      '/api': { target: 'http://127.0.0.1:8098', changeOrigin: true },
      '/media': { target: 'http://127.0.0.1:8098', changeOrigin: true },
    },
  },
})
