import { createApp } from 'vue'
import App from './App.vue'
import router from './router.js'
import { state } from './state.js'
import { api } from './api.js'
import { applyTheme } from './theme.js'
import './style.css'

// The session is restored before mount so the router guard sees the real auth
// state instead of briefly redirecting a signed-in user to the sign-in page.
async function boot() {
  try {
    state.user = await api.me()
  } catch {
    state.user = null
  }
  try {
    state.instance = await api.instance()
  } catch {
    // Keep the built-in defaults; the app still renders offline.
  }
  try {
    state.build = await api.version()
  } catch {
    state.build = null
  }
  document.title = state.instance.name || '__TITLE__'
  applyTheme('system')
  createApp(App).use(router).mount('#app')
}

boot()
