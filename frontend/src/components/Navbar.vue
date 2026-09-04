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

      <!-- Actions -->
      <div class="flex items-center gap-2.5 flex-shrink-0">
        <button
          @click="$emit('scan')"
          :disabled="isScanning"
          class="flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-medium bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] text-slate-300 hover:text-white transition-all disabled:opacity-50"
          title="Rescan /books directory"
        >
          <RefreshCw class="w-3.5 h-3.5 text-glacier-400" :class="{ 'animate-spin': isScanning }" />
          <span class="hidden sm:inline">{{ isScanning ? 'Scanning...' : 'Rescan' }}</span>
        </button>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, watch } from 'vue'
import { BookOpen, Search, X, RefreshCw } from 'lucide-vue-next'

const props = defineProps({
  isScanning: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['search', 'scan', 'reset'])

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
</script>
