<script setup>
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { state, setError, clearBanner } from '../state.js'
import { api } from '../api.js'

const router = useRouter()

const mode = ref('login')
const form = ref({ email: '', display_name: '', password: '' })
const pending = ref(false)
const message = ref('')

// In development mode the identity is a plain pseudo and the password is
// optional, so the form asks for less and labels it differently (ADR-0004).
const dev = computed(() => state.instance.dev_mode)
const identityLabel = computed(() => (dev.value ? 'Username' : 'Email'))

async function submit() {
  pending.value = true
  message.value = ''
  clearBanner()
  try {
    const body = { email: form.value.email, password: form.value.password }
    if (mode.value === 'register') {
      const user = await api.register({ ...body, display_name: form.value.display_name })
      // Production registration returns no session: the account waits on the
      // verification link, so stay on this page and say so.
      if (!user || !user.id) {
        message.value = user.message || 'Check your email to verify your address.'
        mode.value = 'login'
        return
      }
      state.user = user
    } else {
      state.user = await api.login(body)
    }
    router.push('/')
  } catch (err) {
    setError(err.message)
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <section class="card">
    <h2>{{ mode === 'login' ? 'Sign in' : 'Create an account' }}</h2>

    <p v-if="dev" class="muted">
      Development mode: sign in with a username, no password needed.
    </p>
    <p v-if="message" class="muted">{{ message }}</p>

    <form @submit.prevent="submit">
      <label>
        <span>{{ identityLabel }}</span>
        <input
          v-model="form.email"
          :type="dev ? 'text' : 'email'"
          :autocomplete="dev ? 'username' : 'email'"
          required
        />
      </label>

      <label v-if="mode === 'register' && !dev">
        <span>Display name</span>
        <input v-model="form.display_name" type="text" autocomplete="name" required />
      </label>

      <label>
        <span>Password<template v-if="dev"> (optional)</template></span>
        <input
          v-model="form.password"
          type="password"
          :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
          :required="!dev"
        />
      </label>

      <button type="submit" :disabled="pending">
        {{ mode === 'login' ? 'Sign in' : 'Create account' }}
      </button>
    </form>

    <p class="muted switch">
      <button class="ghost" type="button" @click="mode = mode === 'login' ? 'register' : 'login'">
        {{ mode === 'login' ? 'Create an account' : 'I already have an account' }}
      </button>
    </p>
  </section>
</template>
