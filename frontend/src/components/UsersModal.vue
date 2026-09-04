<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in">
    <div class="bg-[#11131b] border border-white/[0.08] rounded-2xl w-full max-w-2xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-white/[0.08] bg-[#090a0f]/60">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400">
            <Users class="w-4 h-4" />
          </div>
          <div>
            <h2 class="text-base font-semibold text-white">User Management</h2>
            <p class="text-xs text-slate-400">Manage reader and administrator accounts</p>
          </div>
        </div>
        <button
          @click="$emit('close')"
          class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/[0.05] transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Content area (scrollable) -->
      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <!-- New User Form -->
        <div class="bg-[#090a0f] border border-white/[0.06] rounded-xl p-4">
          <h3 class="text-xs font-semibold text-glacier-400 uppercase tracking-wider mb-3 flex items-center gap-1.5">
            <UserPlus class="w-3.5 h-3.5" />
            <span>Add New User</span>
          </h3>

          <div v-if="formError" class="mb-3 p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs flex items-center gap-2">
            <AlertCircle class="w-4 h-4 flex-shrink-0" />
            <span>{{ formError }}</span>
          </div>

          <form @submit.prevent="handleCreateUser" class="space-y-3">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-medium text-slate-400 mb-1">Username *</label>
                <input
                  type="text"
                  v-model="newUser.username"
                  required
                  placeholder="Username"
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white placeholder-slate-500 outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-400 mb-1">Display Name</label>
                <input
                  type="text"
                  v-model="newUser.displayName"
                  placeholder="Display name"
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white placeholder-slate-500 outline-none"
                />
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-xs font-medium text-slate-400 mb-1">Password *</label>
                <input
                  type="password"
                  v-model="newUser.password"
                  required
                  placeholder="Min 6 characters"
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white placeholder-slate-500 outline-none"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-400 mb-1">Role *</label>
                <select
                  v-model="newUser.role"
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white outline-none cursor-pointer"
                >
                  <option value="reader">Reader</option>
                  <option value="admin">Administrator</option>
                </select>
              </div>
            </div>

            <div class="flex justify-end pt-1">
              <button
                type="submit"
                :disabled="creating"
                class="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-semibold shadow-md shadow-glacier-500/20 transition-all disabled:opacity-50 cursor-pointer"
              >
                <Loader2 v-if="creating" class="w-3.5 h-3.5 animate-spin" />
                <span>{{ creating ? 'Creating...' : 'Create Account' }}</span>
              </button>
            </div>
          </form>
        </div>

        <!-- Users List -->
        <div>
          <h3 class="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
            Existing Accounts ({{ users.length }})
          </h3>

          <div v-if="loadingUsers" class="py-8 flex justify-center text-slate-500">
            <Loader2 class="w-6 h-6 text-glacier-400 animate-spin" />
          </div>

          <div v-else class="space-y-2">
            <div
              v-for="u in users"
              :key="u.id"
              class="flex items-center justify-between p-3 rounded-xl bg-[#090a0f] border border-white/[0.06] hover:border-white/[0.12] transition-colors"
            >
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-full bg-slate-800 border border-white/[0.1] flex items-center justify-center text-xs font-bold text-slate-300">
                  {{ (u.display_name || u.username).charAt(0).toUpperCase() }}
                </div>
                <div>
                  <div class="flex items-center gap-2">
                    <span class="text-xs font-semibold text-white">
                      {{ u.display_name || u.username }}
                    </span>
                    <span
                      class="text-[10px] font-semibold px-2 py-0.5 rounded-full"
                      :class="u.role === 'admin' ? 'bg-amber-500/15 text-amber-300 border border-amber-500/30' : 'bg-slate-700/50 text-slate-300 border border-slate-600/30'"
                    >
                      {{ u.role }}
                    </span>
                  </div>
                  <span class="text-[11px] text-slate-500">@{{ u.username }}</span>
                </div>
              </div>

              <div class="flex items-center gap-2">
                <button
                  v-if="currentUser?.id !== u.id"
                  @click="handleDeleteUser(u)"
                  class="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors"
                  title="Delete user"
                >
                  <Trash2 class="w-4 h-4" />
                </button>
                <span v-else class="text-[10px] text-slate-500 px-2 py-1 italic">
                  (You)
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Users, UserPlus, Trash2, X, AlertCircle, Loader2 } from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'

const emit = defineEmits(['close'])
const { user: currentUser } = useAuth()

const users = ref([])
const loadingUsers = ref(false)
const creating = ref(false)
const formError = ref('')

const newUser = reactive({
  username: '',
  displayName: '',
  password: '',
  role: 'reader'
})

async function fetchUsers() {
  loadingUsers.value = true
  try {
    const res = await fetch('/api/users')
    if (res.ok) {
      users.value = await res.json()
    }
  } catch (err) {
    console.error('Failed to load users:', err)
  } finally {
    loadingUsers.value = false
  }
}

async function handleCreateUser() {
  formError.value = ''
  if (newUser.username.trim().length < 3) {
    formError.value = 'Username must be at least 3 characters'
    return
  }
  if (newUser.password.length < 6) {
    formError.value = 'Password must be at least 6 characters'
    return
  }

  creating.value = true
  try {
    const res = await fetch('/api/users', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username: newUser.username.trim(),
        display_name: newUser.displayName.trim(),
        password: newUser.password,
        role: newUser.role
      })
    })
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Failed to create user')
    }

    // Reset form
    newUser.username = ''
    newUser.displayName = ''
    newUser.password = ''
    newUser.role = 'reader'

    await fetchUsers()
  } catch (err) {
    formError.value = err.message
  } finally {
    creating.value = false
  }
}

async function handleDeleteUser(userToDelete) {
  if (!confirm(`Are you sure you want to delete user "${userToDelete.username}"?`)) {
    return
  }

  try {
    const res = await fetch(`/api/users/${userToDelete.id}`, { method: 'DELETE' })
    if (res.ok) {
      await fetchUsers()
    } else {
      const data = await res.json()
      alert(data.error || 'Failed to delete user')
    }
  } catch (err) {
    console.error('Failed to delete user:', err)
  }
}

onMounted(() => {
  fetchUsers()
})
</script>
