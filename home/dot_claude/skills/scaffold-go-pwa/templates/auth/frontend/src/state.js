import { reactive } from 'vue'

// Global app state. Anything scoped to a single view belongs in that component.
export const state = reactive({
  user: null,
  build: null,
  // instance is served by GET /api/v1/instance before sign-in. dev_mode mirrors
  // the server's auth.dev_mode so the sign-in form can drop the password field.
  instance: { name: '__TITLE__', dev_mode: false },
  // banner is the app-wide notification shown at the top of the page. kind drives
  // the styling ('error' or 'notice'); an empty text hides it.
  banner: { kind: '', text: '' },
})

// setError shows a prominent error banner at the top of the page: errors must be
// visible and framed, not buried per-view.
export function setError(text) {
  state.banner = { kind: 'error', text }
}

// setNotice shows a neutral informational banner at the top of the page.
export function setNotice(text) {
  state.banner = { kind: 'notice', text }
}

export function clearBanner() {
  state.banner = { kind: '', text: '' }
}

// initials renders an avatar label from a display name.
export function initials(name) {
  if (!name) return '?'
  return name
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0].toUpperCase())
    .join('')
}
