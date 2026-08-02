// Custom service worker (vite-plugin-pwa injectManifest): it precaches the app
// shell so the PWA opens offline.
import { precacheAndRoute, cleanupOutdatedCaches } from 'workbox-precaching'

// Take over as soon as a new build is installed instead of sitting in "waiting"
// until every tab closes. Without this an injectManifest worker keeps serving the
// previous bundle, so users never see updates on a plain reload.
self.skipWaiting()
self.addEventListener('activate', (event) => event.waitUntil(self.clients.claim()))
cleanupOutdatedCaches()

precacheAndRoute(self.__WB_MANIFEST)
