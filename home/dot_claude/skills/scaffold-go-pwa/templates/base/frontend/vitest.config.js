import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

// Test config kept separate from vite.config.js so the PWA/service-worker plugin
// is not loaded during unit tests (it has no place in jsdom).
export default defineConfig({
  plugins: [vue()],
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.{test,spec}.{js,mjs}'],
    coverage: {
      provider: 'v8',
      include: ['src/**/*.{js,vue}'],
      // Service worker, bootstrap, and the specs themselves are not production
      // code under test.
      exclude: ['src/sw.js', 'src/main.js', 'src/**/*.{test,spec}.{js,mjs}'],
    },
  },
})
