<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
    <div class="bg-[#11131b] border border-white/[0.1] rounded-2xl w-full max-w-2xl max-h-[92vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-white/[0.08] bg-[#090a0f]/80">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400">
            <Users class="w-4 h-4" />
          </div>
          <div>
            <h2 class="text-base font-semibold text-white">User & Permission Management</h2>
            <p class="text-xs text-slate-400">Configure roles and granular permissions (Calibre-Web model)</p>
          </div>
        </div>
        <button
          @click="$emit('close')"
          class="text-slate-400 hover:text-white p-1.5 rounded-lg hover:bg-white/[0.05] transition-colors cursor-pointer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Content area (scrollable) -->
      <div class="flex-1 overflow-y-auto p-6 space-y-6">
        <!-- New User Form -->
        <div class="bg-[#090a0f] border border-white/[0.06] rounded-xl p-4">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-xs font-semibold text-glacier-400 uppercase tracking-wider flex items-center gap-1.5">
              <UserPlus class="w-3.5 h-3.5" />
              <span>Add New User</span>
            </h3>
            <span class="text-[11px] text-slate-500">Fine-tune individual permissions</span>
          </div>

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
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white placeholder-slate-500 outline-none transition-colors"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-400 mb-1">Display Name</label>
                <input
                  type="text"
                  v-model="newUser.displayName"
                  placeholder="Display name"
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white placeholder-slate-500 outline-none transition-colors"
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
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white placeholder-slate-500 outline-none transition-colors"
                />
              </div>
              <div>
                <label class="block text-xs font-medium text-slate-400 mb-1">Role *</label>
                <select
                  v-model="newUser.role"
                  @change="onNewUserRoleChange"
                  class="w-full bg-[#161923] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white outline-none cursor-pointer"
                >
                  <option value="reader">Reader</option>
                  <option value="admin">Administrator</option>
                </select>
              </div>
            </div>

            <!-- Granular Permissions Grid -->
            <div class="pt-2">
              <label class="block text-[11px] font-medium text-slate-400 uppercase tracking-wider mb-2">
                Permissions & Capabilities
              </label>
              <div class="grid grid-cols-2 sm:grid-cols-3 gap-2 bg-[#121520] p-3 rounded-lg border border-white/[0.04]">
                <label class="flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    v-model="newUser.permissions.can_read"
                    class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                  />
                  <span class="text-xs text-slate-300 flex items-center gap-1.5">
                    <BookOpen class="w-3.5 h-3.5 text-glacier-400" /> Read
                  </span>
                </label>

                <label class="flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    v-model="newUser.permissions.can_download"
                    class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                  />
                  <span class="text-xs text-slate-300 flex items-center gap-1.5">
                    <Download class="w-3.5 h-3.5 text-emerald-400" /> Download
                  </span>
                </label>

                <label class="flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    v-model="newUser.permissions.can_upload"
                    class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                  />
                  <span class="text-xs text-slate-300 flex items-center gap-1.5">
                    <Upload class="w-3.5 h-3.5 text-blue-400" /> Upload
                  </span>
                </label>

                <label class="flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    v-model="newUser.permissions.can_edit"
                    class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                  />
                  <span class="text-xs text-slate-300 flex items-center gap-1.5">
                    <Pencil class="w-3.5 h-3.5 text-amber-400" /> Edit Info
                  </span>
                </label>

                <label class="flex items-center gap-2 cursor-pointer select-none">
                  <input
                    type="checkbox"
                    v-model="newUser.permissions.can_delete"
                    class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                  />
                  <span class="text-xs text-slate-300 flex items-center gap-1.5">
                    <Trash2 class="w-3.5 h-3.5 text-rose-400" /> Delete Book
                  </span>
                </label>
              </div>
            </div>

            <div class="flex justify-end pt-2">
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

          <div v-else class="space-y-2.5">
            <div
              v-for="u in users"
              :key="u.id"
              class="p-3.5 rounded-xl bg-[#090a0f] border border-white/[0.06] hover:border-white/[0.12] transition-colors"
            >
              <div class="flex items-center justify-between gap-3">
                <div class="flex items-center gap-3 min-w-0">
                  <div class="w-9 h-9 rounded-full bg-slate-800 border border-white/[0.1] flex items-center justify-center text-xs font-bold text-slate-300 flex-shrink-0">
                    {{ (u.display_name || u.username).charAt(0).toUpperCase() }}
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center gap-2 flex-wrap">
                      <span class="text-xs font-semibold text-white truncate">
                        {{ u.display_name || u.username }}
                      </span>
                      <span
                        class="text-[10px] font-semibold px-2 py-0.5 rounded-full uppercase tracking-wider"
                        :class="u.role === 'admin' ? 'bg-amber-500/15 text-amber-300 border border-amber-500/30' : 'bg-slate-700/50 text-slate-300 border border-slate-600/30'"
                      >
                        {{ u.role }}
                      </span>
                      <span v-if="currentUser?.id === u.id" class="text-[10px] text-glacier-400/80 bg-glacier-500/10 px-1.5 py-0.5 rounded border border-glacier-400/20 font-medium">
                        You
                      </span>
                    </div>
                    <span class="text-[11px] text-slate-500 block truncate">@{{ u.username }}</span>
                  </div>
                </div>

                <!-- Action buttons -->
                <div class="flex items-center gap-1.5 flex-shrink-0">
                  <button
                    @click="openEditUser(u)"
                    class="p-1.5 rounded-lg text-slate-400 hover:text-glacier-400 hover:bg-glacier-500/10 transition-colors cursor-pointer"
                    title="Edit user & permissions"
                  >
                    <Pencil class="w-4 h-4" />
                  </button>
                  <button
                    v-if="currentUser?.id !== u.id"
                    @click="handleDeleteUser(u)"
                    class="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
                    title="Delete user"
                  >
                    <Trash2 class="w-4 h-4" />
                  </button>
                </div>
              </div>

              <!-- Permission Badges Display -->
              <div class="mt-2.5 pt-2.5 border-t border-white/[0.04] flex items-center gap-1.5 flex-wrap">
                <span class="text-[10px] text-slate-500 mr-1">Permissions:</span>
                <span
                  class="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-md border"
                  :class="u.permissions?.can_read ? 'bg-glacier-500/10 text-glacier-300 border-glacier-500/20' : 'bg-slate-800/40 text-slate-500 border-white/[0.04] line-through'"
                >
                  <BookOpen class="w-3 h-3" /> Read
                </span>
                <span
                  class="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-md border"
                  :class="u.permissions?.can_download ? 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20' : 'bg-slate-800/40 text-slate-500 border-white/[0.04] line-through'"
                >
                  <Download class="w-3 h-3" /> Download
                </span>
                <span
                  class="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-md border"
                  :class="u.permissions?.can_upload ? 'bg-blue-500/10 text-blue-300 border-blue-500/20' : 'bg-slate-800/40 text-slate-500 border-white/[0.04] line-through'"
                >
                  <Upload class="w-3 h-3" /> Upload
                </span>
                <span
                  class="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-md border"
                  :class="u.permissions?.can_edit ? 'bg-amber-500/10 text-amber-300 border-amber-500/20' : 'bg-slate-800/40 text-slate-500 border-white/[0.04] line-through'"
                >
                  <Pencil class="w-3 h-3" /> Edit
                </span>
                <span
                  class="inline-flex items-center gap-1 text-[10px] px-2 py-0.5 rounded-md border"
                  :class="u.permissions?.can_delete ? 'bg-rose-500/10 text-rose-300 border-rose-500/20' : 'bg-slate-800/40 text-slate-500 border-white/[0.04] line-through'"
                >
                  <Trash2 class="w-3 h-3" /> Delete
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit User Modal -->
    <div v-if="editingUser" class="fixed inset-0 z-60 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-fade-in">
      <div class="bg-[#141724] border border-white/[0.12] rounded-2xl w-full max-w-lg p-6 shadow-2xl space-y-4">
        <div class="flex items-center justify-between pb-3 border-b border-white/[0.08]">
          <div class="flex items-center gap-2">
            <Pencil class="w-4 h-4 text-glacier-400" />
            <h3 class="text-sm font-semibold text-white">Edit User: @{{ editingUser.username }}</h3>
          </div>
          <button @click="editingUser = null" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/[0.05]">
            <X class="w-4 h-4" />
          </button>
        </div>

        <div v-if="editError" class="p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs flex items-center gap-2">
          <AlertCircle class="w-4 h-4 flex-shrink-0" />
          <span>{{ editError }}</span>
        </div>

        <form @submit.prevent="handleSaveEdit" class="space-y-3.5">
          <div>
            <label class="block text-xs font-medium text-slate-400 mb-1">Display Name</label>
            <input
              type="text"
              v-model="editForm.displayName"
              placeholder="Display name"
              class="w-full bg-[#1b1f30] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white outline-none"
            />
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">Role</label>
              <select
                v-model="editForm.role"
                class="w-full bg-[#1b1f30] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white outline-none cursor-pointer"
              >
                <option value="reader">Reader</option>
                <option value="admin">Administrator</option>
              </select>
            </div>
            <div>
              <label class="block text-xs font-medium text-slate-400 mb-1">New Password (optional)</label>
              <input
                type="password"
                v-model="editForm.password"
                placeholder="Leave blank to keep current"
                class="w-full bg-[#1b1f30] border border-white/[0.08] focus:border-glacier-400/50 rounded-lg px-3 py-2 text-xs text-white placeholder-slate-500 outline-none"
              />
            </div>
          </div>

          <div>
            <label class="block text-[11px] font-medium text-slate-400 uppercase tracking-wider mb-2">
              Permissions Configuration
            </label>
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-2 bg-[#0c0e15] p-3 rounded-lg border border-white/[0.06]">
              <label class="flex items-center gap-2 cursor-pointer select-none">
                <input
                  type="checkbox"
                  v-model="editForm.permissions.can_read"
                  class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                />
                <span class="text-xs text-slate-300 flex items-center gap-1.5">
                  <BookOpen class="w-3.5 h-3.5 text-glacier-400" /> Read
                </span>
              </label>

              <label class="flex items-center gap-2 cursor-pointer select-none">
                <input
                  type="checkbox"
                  v-model="editForm.permissions.can_download"
                  class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                />
                <span class="text-xs text-slate-300 flex items-center gap-1.5">
                  <Download class="w-3.5 h-3.5 text-emerald-400" /> Download
                </span>
              </label>

              <label class="flex items-center gap-2 cursor-pointer select-none">
                <input
                  type="checkbox"
                  v-model="editForm.permissions.can_upload"
                  class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                />
                <span class="text-xs text-slate-300 flex items-center gap-1.5">
                  <Upload class="w-3.5 h-3.5 text-blue-400" /> Upload
                </span>
              </label>

              <label class="flex items-center gap-2 cursor-pointer select-none">
                <input
                  type="checkbox"
                  v-model="editForm.permissions.can_edit"
                  class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                />
                <span class="text-xs text-slate-300 flex items-center gap-1.5">
                  <Pencil class="w-3.5 h-3.5 text-amber-400" /> Edit
                </span>
              </label>

              <label class="flex items-center gap-2 cursor-pointer select-none">
                <input
                  type="checkbox"
                  v-model="editForm.permissions.can_delete"
                  class="rounded bg-[#1a1d29] border-white/20 text-glacier-500 focus:ring-glacier-500/30"
                />
                <span class="text-xs text-slate-300 flex items-center gap-1.5">
                  <Trash2 class="w-3.5 h-3.5 text-rose-400" /> Delete
                </span>
              </label>
            </div>
          </div>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-white/[0.08]">
            <button
              type="button"
              @click="editingUser = null"
              class="px-3.5 py-2 rounded-lg bg-white/[0.05] hover:bg-white/[0.1] text-slate-300 text-xs font-medium cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="savingEdit"
              class="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-semibold shadow-md shadow-glacier-500/20 transition-all disabled:opacity-50 cursor-pointer"
            >
              <Loader2 v-if="savingEdit" class="w-3.5 h-3.5 animate-spin" />
              <span>{{ savingEdit ? 'Saving...' : 'Save Changes' }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Users, UserPlus, Trash2, X, AlertCircle, Loader2, Pencil, BookOpen, Download, Upload } from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'

const emit = defineEmits(['close'])
const { user: currentUser, checkAuth } = useAuth()

const users = ref([])
const loadingUsers = ref(false)
const creating = ref(false)
const formError = ref('')

const newUser = reactive({
  username: '',
  displayName: '',
  password: '',
  role: 'reader',
  permissions: {
    can_read: true,
    can_download: true,
    can_upload: false,
    can_edit: false,
    can_delete: false
  }
})

function onNewUserRoleChange() {
  if (newUser.role === 'admin') {
    newUser.permissions.can_read = true
    newUser.permissions.can_download = true
    newUser.permissions.can_upload = true
    newUser.permissions.can_edit = true
    newUser.permissions.can_delete = true
  } else {
    newUser.permissions.can_read = true
    newUser.permissions.can_download = true
    newUser.permissions.can_upload = false
    newUser.permissions.can_edit = false
    newUser.permissions.can_delete = false
  }
}

// Edit User State
const editingUser = ref(null)
const editError = ref('')
const savingEdit = ref(false)
const editForm = reactive({
  displayName: '',
  role: 'reader',
  password: '',
  permissions: {
    can_read: true,
    can_download: true,
    can_upload: false,
    can_edit: false,
    can_delete: false
  }
})

function openEditUser(userToEdit) {
  editingUser.value = userToEdit
  editError.value = ''
  editForm.displayName = userToEdit.display_name || userToEdit.username
  editForm.role = userToEdit.role
  editForm.password = ''
  editForm.permissions = {
    can_read: userToEdit.permissions?.can_read ?? true,
    can_download: userToEdit.permissions?.can_download ?? true,
    can_upload: userToEdit.permissions?.can_upload ?? false,
    can_edit: userToEdit.permissions?.can_edit ?? false,
    can_delete: userToEdit.permissions?.can_delete ?? false
  }
}

async function handleSaveEdit() {
  if (!editingUser.value) return
  editError.value = ''

  if (editForm.password && editForm.password.length < 6) {
    editError.value = 'Password must be at least 6 characters'
    return
  }

  savingEdit.value = true
  try {
    const payload = {
      display_name: editForm.displayName.trim(),
      role: editForm.role,
      permissions: editForm.permissions
    }
    if (editForm.password) {
      payload.password = editForm.password
    }

    const res = await fetch(`/api/users/${editingUser.value.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Failed to update user')
    }

    // Refresh current user info if admin edited themselves
    if (currentUser.value?.id === editingUser.value.id) {
      await checkAuth()
    }

    editingUser.value = null
    await fetchUsers()
  } catch (err) {
    editError.value = err.message
  } finally {
    savingEdit.value = false
  }
}

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
        role: newUser.role,
        permissions: newUser.permissions
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
    newUser.permissions = {
      can_read: true,
      can_download: true,
      can_upload: false,
      can_edit: false,
      can_delete: false
    }

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
