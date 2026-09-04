<template>
  <header class="sticky top-0 z-30 bg-[#090a0f]/90 backdrop-blur-md border-b border-white/[0.08] px-4 lg:px-8 py-3.5 transition-all">
    <div class="max-w-7xl mx-auto flex items-center justify-between gap-4">
      <!-- Brand Logo -->
      <div 
        @click="$emit('reset')" 
        class="flex items-center gap-2.5 cursor-pointer group select-none flex-shrink-0"
      >
        <div class="w-9 h-9 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400 group-hover:bg-glacier-400/20 group-hover:border-glacier-400/40 transition-colors">
          <BookOpen class="w-5 h-5" />
        </div>
        <div>
          <span class="text-lg font-bold tracking-tight text-white group-hover:text-glacier-400 transition-colors">
            Lyostar
          </span>
          <span class="hidden sm:inline-block ml-2 text-xs font-medium px-2 py-0.5 rounded-full bg-white/[0.06] text-slate-400 border border-white/[0.05]">
            Reader
          </span>
        </div>
      </div>

      <!-- Search Bar -->
      <div class="flex-1 max-w-xl relative">
        <div class="relative flex items-center">
          <Search class="w-4 h-4 text-slate-400 absolute left-3.5 pointer-events-none" />
          <input
            type="text"
            v-model="searchQuery"
            @input="onSearchInput"
            placeholder="Search titles, authors, series..."
            class="w-full bg-[#11131b] border border-white/[0.08] hover:border-white/[0.15] focus:border-glacier-400/50 focus:ring-1 focus:ring-glacier-400/50 rounded-xl pl-10 pr-9 py-2 text-sm text-slate-100 placeholder-slate-500 transition-all outline-none"
          />
          <button
            v-if="searchQuery"
            @click="clearSearch"
            class="absolute right-3 text-slate-400 hover:text-white transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Actions & Profile -->
      <div class="flex items-center gap-2 sm:gap-2.5 flex-shrink-0">
        <!-- Admin-only: Users management -->
        <button
          v-if="isAdmin"
          @click="$emit('open-users')"
          class="flex items-center gap-1.5 px-2.5 sm:px-3 py-2 rounded-xl text-xs font-medium bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] text-slate-300 hover:text-white transition-all cursor-pointer"
          title="Manage Users"
        >
          <Users class="w-3.5 h-3.5 text-glacier-400" />
          <span class="hidden md:inline">Users</span>
        </button>

        <!-- Admin-only: Rescan -->
        <button
          v-if="isAdmin"
          @click="$emit('scan')"
          :disabled="isScanning"
          class="flex items-center gap-1.5 px-2.5 sm:px-3 py-2 rounded-xl text-xs font-medium bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] text-slate-300 hover:text-white transition-all disabled:opacity-50 cursor-pointer"
          title="Rescan /books directory"
        >
          <RefreshCw class="w-3.5 h-3.5 text-glacier-400" :class="{ 'animate-spin': isScanning }" />
          <span class="hidden md:inline">{{ isScanning ? 'Scanning...' : 'Rescan' }}</span>
        </button>

        <!-- User Profile Pill -->
        <div v-if="user" class="flex items-center gap-2 pl-1 sm:pl-2 border-l border-white/[0.08]">
          <div class="flex items-center gap-2 bg-[#11131b] border border-white/[0.08] rounded-xl px-2.5 py-1.5">
            <div class="w-6 h-6 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-[11px] font-bold text-glacier-400">
              {{ (user.display_name || user.username).charAt(0).toUpperCase() }}
            </div>
            <div class="hidden lg:block text-left">
              <div class="text-xs font-medium text-slate-200 leading-tight">
                {{ user.display_name || user.username }}
              </div>
              <div class="text-[10px] text-slate-500 uppercase tracking-wider">
                {{ user.role }}
              </div>
            </div>
          </div>

          <!-- Logout Button -->
          <button
            @click="handleLogout"
            class="p-2 rounded-xl text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 border border-transparent hover:border-rose-500/20 transition-all cursor-pointer"
            title="Sign out"
          >
            <LogOut class="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref } from 'vue'
import { BookOpen, Search, X, RefreshCw, Users, LogOut } from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'

const props = defineProps({
  isScanning: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['search', 'scan', 'reset', 'open-users'])

const { user, isAdmin, logout } = useAuth()
const searchQuery = ref('')
let debounceTimer = null

function onSearchInput() {
  clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emit('search', searchQuery.value.trim())
  }, 250)
}

function clearSearch() {
  searchQuery.value = ''
  emit('search', '')
}

async function handleLogout() {
  await logout()
}
</script>
