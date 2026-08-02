import { createRouter, createWebHistory } from 'vue-router'
import { state } from './state.js'

import AuthView from './views/AuthView.vue'
import HomeView from './views/HomeView.vue'

const routes = [
  { path: '/login', component: AuthView },
  { path: '/', component: HomeView },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

const router = createRouter({ history: createWebHistory(), routes })

// The session is restored before mount (see main.js), so this guard sees the truth.
router.beforeEach((to) => {
  if (to.path !== '/login' && !state.user) return '/login'
  if (to.path === '/login' && state.user) return '/'
  return true
})

export default router
