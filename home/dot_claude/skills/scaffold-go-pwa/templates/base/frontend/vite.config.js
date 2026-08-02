import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { VitePWA } from 'vite-plugin-pwa'

// The build is emitted into the Go module's embed directory so the binary serves
// the PWA as a single artifact (ADR-0001 / internal/spa).
export default defineConfig({
  plugins: [
    vue(),
    VitePWA({
      registerType: 'autoUpdate',
      strategies: 'injectManifest',
      srcDir: 'src',
      filename: 'sw.js',
      // The icons are rasterised from the app emoji into public/icons (ADR-0002),
      // so they are copied to the dist root and referenced from there.
      includeAssets: ['icons/apple-touch-icon.png'],
      manifest: {
        name: '__TITLE__',
        short_name: '__TITLE__',
        description: '__DESCRIPTION__',
        theme_color: '__BG_COLOR__',
        background_color: '__BG_COLOR__',
        display: 'standalone',
        start_url: '/',
        scope: '/',
        icons: [
          { src: '/icons/icon-192.png', sizes: '192x192', type: 'image/png', purpose: 'any' },
          { src: '/icons/icon-512.png', sizes: '512x512', type: 'image/png', purpose: 'any' },
          { src: '/icons/icon-maskable-512.png', sizes: '512x512', type: 'image/png', purpose: 'maskable' },
        ],
      },
    }),
  ],
  build: {
    outDir: '../internal/spa/dist',
    emptyOutDir: true,
  },
  server: {
    // For `npm run dev` only: proxy API calls to the Go backend. Matches the
    // backend's default port (config.example.yaml); if you run the backend on
    // another port (__ENV_PREFIX___PORT), point this at it too.
    proxy: {
      '/api': { target: 'http://localhost:__PORT__' },
    },
  },
})
