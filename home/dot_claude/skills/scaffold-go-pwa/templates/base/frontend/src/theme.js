// Browser-chrome colour per palette, mirroring --bg in style.css so the mobile
// toolbar and status bar match the page instead of staying on one fixed colour.
const THEME_COLORS = { light: '__BG_COLOR__', dark: '__DARK_BG_COLOR__' }

// A single live listener for the OS preference, attached only while the pref is
// 'system'. Tracked at module scope so re-applying any other pref detaches it.
let osQuery = null
let osListener = null

function setMode(mode) {
  document.documentElement.dataset.theme = mode
  const meta = document.querySelector('meta[name="theme-color"]')
  if (meta) meta.setAttribute('content', THEME_COLORS[mode] || THEME_COLORS.light)
}

// applyTheme sets the active palette on <html data-theme>. 'system' follows the OS
// preference live (re-applying whenever the OS flips); an explicit 'light'/'dark'
// wins and detaches the OS listener.
export function applyTheme(pref) {
  if (osQuery && osListener) {
    osQuery.removeEventListener('change', osListener)
    osQuery = null
    osListener = null
  }
  if (pref === 'light' || pref === 'dark') {
    setMode(pref)
    return
  }
  osQuery = window.matchMedia('(prefers-color-scheme: dark)')
  osListener = (e) => setMode(e.matches ? 'dark' : 'light')
  osQuery.addEventListener('change', osListener)
  setMode(osQuery.matches ? 'dark' : 'light')
}
