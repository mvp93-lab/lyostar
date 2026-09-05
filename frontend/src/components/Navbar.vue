<template>
  <header class="sticky top-0 z-30 bg-[#090a0f]/90 backdrop-blur-md border-b border-white/[0.08] px-4 lg:px-8 py-3 transition-all">
    <div class="max-w-7xl mx-auto flex items-center justify-between gap-3 sm:gap-4">
      <!-- Left: Mobile Sidebar Toggle Button & Brand -->
      <div class="flex items-center gap-2">
        <button
          @click="$emit('toggle-sidebar')"
          class="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/[0.06] border border-transparent hover:border-white/[0.08] md:hidden transition-all cursor-pointer"
          title="Toggle Navigation Menu"
        >
          <Menu class="w-5 h-5" />
        </button>

        <!-- Brand Name (Mobile view when sidebar closed) -->
        <div 
          @click="onBrandClick" 
          class="flex items-center gap-2 cursor-pointer md:hidden select-none"
        >
          <div class="w-7 h-7 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400">
            <BookOpen class="w-4 h-4" />
          </div>
          <span class="text-sm font-bold tracking-tight text-white">Lyostar</span>
        </div>
      </div>

      <!-- Center: Expanded & Focused Search Bar -->
      <div class="flex-1 max-w-2xl relative">
        <div class="relative flex items-center">
          <Search class="w-4 h-4 text-slate-400 absolute left-3.5 pointer-events-none" />
          <input
            type="text"
            v-model="searchQuery"
            @input="onSearchInput"
            placeholder="Search books, authors, series..."
            class="w-full bg-[#11131b] border border-white/[0.08] hover:border-white/[0.15] focus:border-glacier-400/50 focus:ring-1 focus:ring-glacier-400/50 rounded-xl pl-10 pr-9 py-2 text-xs sm:text-sm text-slate-100 placeholder-slate-500 transition-all outline-none"
          />
          <button
            v-if="searchQuery"
            @click="clearSearch"
            class="absolute right-3 text-slate-400 hover:text-white transition-colors cursor-pointer"
            title="Clear search"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Right: Primary Upload Action & User Menu with 3-Dots -->
      <div class="flex items-center gap-2 sm:gap-2.5 flex-shrink-0">
        <!-- Scanning indicator badge -->
        <div 
          v-if="isScanning" 
          class="hidden sm:flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-glacier-500/10 border border-glacier-400/30 text-[11px] font-medium text-glacier-300 animate-pulse"
        >
          <RefreshCw class="w-3 h-3 animate-spin text-glacier-400" />
          <span>Scanning...</span>
        </div>

        <!-- Quick Upload Book Button (can_upload) -->
        <button
          v-if="canUpload"
          @click="onUploadClick"
          class="flex items-center gap-1.5 px-3 py-1.5 sm:py-2 rounded-xl text-xs font-semibold bg-glacier-500 hover:bg-glacier-400 text-slate-950 shadow-md shadow-glacier-500/20 hover:shadow-glacier-500/30 transition-all cursor-pointer"
          title="Upload Ebook (.epub, .pdf)"
        >
          <Upload class="w-3.5 h-3.5" />
          <span class="hidden sm:inline">Upload Book</span>
        </button>

        <!-- User Avatar Pill with 3-Dots Options Menu -->
        <div v-if="user" class="relative" ref="userMenuRef">
          <div 
            @click="toggleUserMenu"
            class="flex items-center gap-1.5 bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-white/[0.15] rounded-xl pl-2 pr-1.5 py-1.5 cursor-pointer transition-all select-none"
            :class="{ 'border-glacier-400/40 bg-[#161923]': showUserMenu }"
          >
            <!-- Avatar -->
            <div class="w-6 h-6 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-[11px] font-bold text-glacier-400">
              {{ (user.display_name || user.username).charAt(0).toUpperCase() }}
            </div>

            <!-- Display Name -->
            <span class="hidden sm:inline-block text-xs font-medium text-slate-300 max-w-[100px] truncate">
              {{ user.display_name || user.username }}
            </span>

            <!-- 3 Dots Options Button -->
            <div class="p-0.5 text-slate-400 hover:text-white transition-colors">
              <MoreVertical class="w-4 h-4" />
            </div>
          </div>

          <!-- Dropdown Options Popup -->
          <div
            v-if="showUserMenu"
            class="absolute right-0 mt-2 w-56 rounded-2xl bg-[#11131b] border border-white/[0.1] shadow-2xl shadow-black/80 py-2 z-50 animate-fade-in divide-y divide-white/[0.06]"
          >
            <!-- Profile Info Header -->
            <div class="px-3.5 py-2.5">
              <div class="flex items-center gap-2.5 mb-1.5">
                <div class="w-8 h-8 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-xs font-bold text-glacier-400 flex-shrink-0">
                  {{ (user.display_name || user.username).charAt(0).toUpperCase() }}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="text-xs font-semibold text-white truncate">
                    {{ user.display_name || user.username }}
                  </div>
                  <div class="text-[10px] text-slate-400 truncate">
                    @{{ user.username }}
                  </div>
                </div>
              </div>

              <div class="mt-2 pt-2 border-t border-white/[0.04] flex items-center justify-between text-[11px]">
                <span class="text-slate-400">Role:</span>
                <span class="font-semibold px-2 py-0.5 rounded-md bg-glacier-500/10 text-glacier-300 border border-glacier-500/20 uppercase tracking-wider text-[10px]">
                  {{ user.role }}
                </span>
              </div>
            </div>

            <!-- Sign Out Action -->
            <div class="p-1.5">
              <button
                @click="handleLogout"
                class="w-full flex items-center gap-2.5 px-3 py-2 rounded-xl text-xs font-medium text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 transition-colors cursor-pointer"
              >
                <LogOut class="w-3.5 h-3.5" />
                <span>Sign Out</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, watch, onMounted, onBeforeUnmount } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { BookOpen, Search, X, Menu, RefreshCw, Upload, MoreVertical, LogOut } from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'

const props = defineProps({
  isScanning: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['search', 'reset', 'toggle-sidebar', 'open-upload'])

const router = useRouter()
const route = useRoute()
const { user, canUpload, logout } = useAuth()
const searchQuery = ref(route.query.q || '')
let debounceTimer = null

watch(() => route.query.q, (newQ) => {
  if (newQ !== undefined && newQ !== searchQuery.value) {
    searchQuery.value = newQ
  } else if (route.name !== 'search' && !route.query.q && searchQuery.value) {
    searchQuery.value = ''
  }
})

const showUserMenu = ref(false)
const userMenuRef = ref(null)

function toggleUserMenu() {
  showUserMenu.value = !showUserMenu.value
}

function handleClickOutside(e) {
  if (userMenuRef.value && !userMenuRef.value.contains(e.target)) {
    showUserMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
})

async function handleLogout() {
  showUserMenu.value = false
  await logout()
  router.push('/login')
}

function onUploadClick() {
  emit('open-upload')
  router.push('/upload')
}

function onBrandClick() {
  emit('reset')
  router.push('/books')
}

function onSearchInput() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    const q = searchQuery.value.trim()
    emit('search', q)
    if (q) {
      router.push({ path: '/search', query: { q } })
    } else if (route.name === 'search') {
      router.push('/books')
    }
  }, 300)
}

function clearSearch() {
  searchQuery.value = ''
  emit('search', '')
  if (route.name === 'search' || route.query.q) {
    router.push('/books')
  }
}
</script>
