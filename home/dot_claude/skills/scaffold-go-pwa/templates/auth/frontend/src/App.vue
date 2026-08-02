<script setup>
import { useRouter } from 'vue-router'
import { state, clearBanner, initials } from './state.js'
import { api } from './api.js'

const router = useRouter()

async function logout() {
  await api.logout()
  state.user = null
  clearBanner()
  router.push('/login')
}
</script>

<template>
  <header class="topbar">
    <span class="mark" aria-hidden="true">__EMOJI__</span>
    <h1>{{ state.instance.name }}</h1>
    <div v-if="state.user" class="account">
      <span class="avatar" :title="state.user.display_name">{{ initials(state.user.display_name) }}</span>
      <button class="ghost" @click="logout">Sign out</button>
    </div>
  </header>

  <p v-if="state.banner.text" class="banner" :class="state.banner.kind" role="alert" @click="clearBanner">
    {{ state.banner.text }}
  </p>

  <main>
    <RouterView />
  </main>
</template>
