<template>
  <div class="min-h-screen bg-[#090a0f] flex items-center justify-center p-4 selection:bg-glacier-500/20 selection:text-glacier-400">
    <div class="w-full max-w-sm bg-[#11131b] border border-white/[0.08] rounded-2xl p-6 sm:p-8 shadow-2xl relative overflow-hidden">
      <!-- Glow decoration -->
      <div class="absolute -top-24 -left-24 w-48 h-48 bg-glacier-500/10 rounded-full blur-3xl pointer-events-none"></div>

      <!-- Header -->
      <div class="text-center mb-8">
        <div class="w-12 h-12 rounded-xl bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400 mx-auto mb-3">
          <BookOpen class="w-6 h-6" />
        </div>
        <h1 class="text-2xl font-bold text-white tracking-tight">Lyostar</h1>
        <p class="text-xs text-slate-400 mt-1">Sign in to your library</p>
      </div>

      <!-- Error Message -->
      <div v-if="error" class="mb-5 p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs flex items-center gap-2">
        <AlertCircle class="w-4 h-4 flex-shrink-0" />
        <span>{{ error }}</span>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-xs font-medium text-slate-300 mb-1.5">Username</label>
          <input
            type="text"
            v-model="form.username"
            required
            autocomplete="username"
            placeholder="Username"
            class="w-full bg-[#090a0f] border border-white/[0.08] focus:border-glacier-400/50 focus:ring-1 focus:ring-glacier-400/50 rounded-xl px-3.5 py-2.5 text-sm text-slate-100 placeholder-slate-600 transition-all outline-none"
          />
        </div>

        <div>
          <label class="block text-xs font-medium text-slate-300 mb-1.5">Password</label>
          <input
            type="password"
            v-model="form.password"
            required
            autocomplete="current-password"
            placeholder="Password"
            class="w-full bg-[#090a0f] border border-white/[0.08] focus:border-glacier-400/50 focus:ring-1 focus:ring-glacier-400/50 rounded-xl px-3.5 py-2.5 text-sm text-slate-100 placeholder-slate-600 transition-all outline-none"
          />
        </div>

        <div class="pt-2">
          <button
            type="submit"
            :disabled="submitting"
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-sm font-semibold shadow-lg shadow-glacier-500/20 transition-all disabled:opacity-50 cursor-pointer"
          >
            <Loader2 v-if="submitting" class="w-4 h-4 animate-spin" />
            <span>{{ submitting ? 'Signing in...' : 'Sign In' }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { BookOpen, AlertCircle, Loader2 } from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'

const emit = defineEmits(['success'])
const { login } = useAuth()

const form = reactive({
  username: '',
  password: ''
})

const error = ref('')
const submitting = ref(false)

async function handleSubmit() {
  error.value = ''
  if (!form.username.trim() || !form.password) {
    error.value = 'Please enter both username and password'
    return
  }

  submitting.value = true
  try {
    await login({
      username: form.username.trim(),
      password: form.password
    })
    emit('success')
  } catch (err) {
    error.value = err.message || 'Invalid username or password'
  } finally {
    submitting.value = false
  }
}
</script>
