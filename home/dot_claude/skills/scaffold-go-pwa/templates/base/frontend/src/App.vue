<script setup>
import { state, clearBanner } from './state.js'
</script>

<template>
  <header class="topbar">
    <span class="mark" aria-hidden="true">__EMOJI__</span>
    <h1>__TITLE__</h1>
  </header>

  <p v-if="state.banner.text" class="banner" :class="state.banner.kind" role="alert" @click="clearBanner">
    {{ state.banner.text }}
  </p>

  <main>
    <RouterView />
  </main>
</template>
