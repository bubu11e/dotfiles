<script setup>
import { state } from '../state.js'
</script>

<template>
  <section class="card">
    <h2>Welcome, {{ state.user.display_name }}</h2>
    <p class="muted">TODO: replace this view with the real app.</p>
    <dl v-if="state.build" class="build">
      <dt>version</dt>
      <dd>{{ state.build.version }}</dd>
      <dt>commit</dt>
      <dd>{{ state.build.commit || 'local build' }}</dd>
      <dt>go</dt>
      <dd>{{ state.build.go_version }}</dd>
    </dl>
  </section>
</template>
