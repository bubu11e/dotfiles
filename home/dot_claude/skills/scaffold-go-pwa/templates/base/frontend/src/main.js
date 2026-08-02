import { createApp } from 'vue'
import App from './App.vue'
import router from './router.js'
import { state } from './state.js'
import { api } from './api.js'
import { applyTheme } from './theme.js'
import './style.css'

// Build info is fetched before mount so the shell can render it without a
// second paint. A failure here is not fatal: the app still works offline.
async function boot() {
  try {
    state.build = await api.version()
  } catch {
    state.build = null
  }
  applyTheme('system')
  createApp(App).use(router).mount('#app')
}

boot()
