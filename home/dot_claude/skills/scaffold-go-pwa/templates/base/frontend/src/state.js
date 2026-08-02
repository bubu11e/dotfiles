import { reactive } from 'vue'

// Global app state. Anything scoped to a single view belongs in that component.
export const state = reactive({
  build: null,
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
