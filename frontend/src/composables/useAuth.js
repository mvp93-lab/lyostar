import { ref, computed } from 'vue'

const user = ref(null)
const setupRequired = ref(false)
const loading = ref(true)

export function useAuth() {
  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const canRead = computed(() => user.value?.permissions?.can_read ?? true)
  const canDownload = computed(() => user.value?.permissions?.can_download ?? true)
  const canUpload = computed(() => user.value?.permissions?.can_upload ?? false)
  const canEdit = computed(() => user.value?.permissions?.can_edit ?? false)
  const canDelete = computed(() => user.value?.permissions?.can_delete ?? false)

  async function checkAuth() {
    loading.value = true
    try {
      const res = await fetch('/api/auth/status')
      if (res.ok) {
        const data = await res.json()
        setupRequired.value = !!data.setup_required
        user.value = data.authenticated ? data.user : null
      }
    } catch (err) {
      console.error('[useAuth] Failed to check auth status:', err)
    } finally {
      loading.value = false
    }
  }

  async function setupAdmin({ username, password, displayName }) {
    const res = await fetch('/api/auth/setup', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        username,
        password,
        display_name: displayName
      })
    })

    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Failed to setup admin account')
    }

    user.value = data.user
    setupRequired.value = false
    return data.user
  }

  async function login({ username, password }) {
    const res = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password })
    })

    const data = await res.json()
    if (!res.ok) {
      throw new Error(data.error || 'Invalid username or password')
    }

    user.value = data.user
    setupRequired.value = false
    return data.user
  }

  async function logout() {
    try {
      await fetch('/api/auth/logout', { method: 'POST' })
    } catch (e) {
      console.warn('[useAuth] Logout request failed:', e)
    } finally {
      user.value = null
    }
  }

  return {
    user,
    isAuthenticated,
    isAdmin,
    canRead,
    canDownload,
    canUpload,
    canEdit,
    canDelete,
    setupRequired,
    loading,
    checkAuth,
    setupAdmin,
    login,
    logout
  }
}
