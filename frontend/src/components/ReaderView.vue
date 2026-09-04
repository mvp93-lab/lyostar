<template>
  <div class="fixed inset-0 z-50 bg-[#090a0f] text-slate-100 flex flex-col select-none overflow-hidden animate-fade-in">
    <!-- Top Bar Controls (Floating or Fixed) -->
    <header 
      v-show="showControls"
      class="h-14 bg-[#090a0f]/95 backdrop-blur border-b border-white/[0.08] px-4 flex items-center justify-between z-20 transition-all duration-200"
    >
      <div class="flex items-center gap-3 min-w-0">
        <button
          @click="$emit('close')"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors"
          title="Back to Library (Esc)"
        >
          <ArrowLeft class="w-5 h-5" />
        </button>
        <div class="min-w-0">
          <h2 class="text-sm font-semibold text-white truncate max-w-xs sm:max-w-md">
            {{ book.title }}
          </h2>
          <p v-if="currentSectionTitle" class="text-[11px] text-glacier-400 truncate">
            {{ currentSectionTitle }}
          </p>
        </div>
      </div>

      <div class="flex items-center gap-1.5 sm:gap-2">
        <!-- Font Size Decrease -->
        <button
          @click="adjustFontSize(-10)"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors"
          title="Decrease Font Size"
        >
          <span class="text-xs font-bold font-serif">A-</span>
        </button>
        <!-- Font Size Increase -->
        <button
          @click="adjustFontSize(10)"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors"
          title="Increase Font Size"
        >
          <span class="text-sm font-bold font-serif">A+</span>
        </button>
        <!-- Toggle Fullscreen -->
        <button
          @click="toggleFullscreen"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors"
          title="Toggle Fullscreen"
        >
          <Maximize2 v-if="!isFullscreen" class="w-4 h-4" />
          <Minimize2 v-else class="w-4 h-4" />
        </button>
      </div>
    </header>

    <!-- Reader Viewport -->
    <main class="flex-1 relative w-full h-full overflow-hidden flex items-center justify-center">
      <!-- Loading State -->
      <div v-if="loading" class="flex flex-col items-center gap-3 text-slate-400">
        <Loader2 class="w-7 h-7 text-glacier-400 animate-spin" />
        <span class="text-xs font-medium tracking-wide">Loading book content...</span>
      </div>

      <!-- Error State -->
      <div v-if="error" class="p-6 max-w-md text-center bg-red-500/10 border border-red-500/20 rounded-2xl text-red-300 text-sm">
        <p class="font-semibold mb-2">Unable to open EPUB</p>
        <p class="text-xs text-red-400/80 mb-4">{{ error }}</p>
        <button
          @click="$emit('close')"
          class="px-4 py-2 rounded-xl bg-red-500/20 hover:bg-red-500/30 text-red-200 text-xs font-medium"
        >
          Return to Shelf
        </button>
      </div>

      <!-- foliater-view container mounted here -->
      <div 
        ref="readerContainer" 
        class="w-full h-full select-text"
        :class="{ 'opacity-0': loading || error, 'opacity-100': !loading && !error }"
      ></div>

      <!-- Floating Prev / Next Click Zones -->
      <button 
        @click="prevPage" 
        class="absolute left-0 top-14 bottom-10 w-16 group flex items-center justify-start pl-3 z-10 cursor-pointer opacity-0 hover:opacity-100 transition-opacity"
        aria-label="Previous page"
      >
        <div class="p-2 rounded-full bg-black/60 backdrop-blur border border-white/10 text-white shadow-lg group-hover:scale-110 transition-transform">
          <ChevronLeft class="w-5 h-5" />
        </div>
      </button>

      <button 
        @click="nextPage" 
        class="absolute right-0 top-14 bottom-10 w-16 group flex items-center justify-end pr-3 z-10 cursor-pointer opacity-0 hover:opacity-100 transition-opacity"
        aria-label="Next page"
      >
        <div class="p-2 rounded-full bg-black/60 backdrop-blur border border-white/10 text-white shadow-lg group-hover:scale-110 transition-transform">
          <ChevronRight class="w-5 h-5" />
        </div>
      </button>
    </main>

    <!-- Bottom Progress Bar Controls -->
    <footer 
      v-show="showControls"
      class="h-10 bg-[#090a0f]/95 backdrop-blur border-t border-white/[0.08] px-4 flex items-center justify-between text-xs text-slate-400 z-20"
    >
      <button 
        @click="prevPage" 
        class="flex items-center gap-1 hover:text-white transition-colors"
      >
        <ChevronLeft class="w-4 h-4" />
        <span class="hidden sm:inline">Prev</span>
      </button>

      <div class="flex items-center gap-3 flex-1 max-w-xs mx-4">
        <div class="w-full bg-white/[0.08] h-1.5 rounded-full overflow-hidden">
          <div 
            class="bg-glacier-400 h-full rounded-full transition-all duration-300"
            :style="{ width: `${progressPercent}%` }"
          ></div>
        </div>
        <span class="font-mono text-[11px] min-w-8 text-right">{{ progressPercent }}%</span>
      </div>

      <button 
        @click="nextPage" 
        class="flex items-center gap-1 hover:text-white transition-colors"
      >
        <span class="hidden sm:inline">Next</span>
        <ChevronRight class="w-4 h-4" />
      </button>
    </footer>
  </div>
</template>

<script setup>
import { shallowRef, ref, onMounted, onBeforeUnmount } from 'vue'
import { ArrowLeft, ChevronLeft, ChevronRight, Maximize2, Minimize2, Loader2 } from 'lucide-vue-next'

const props = defineProps({
  book: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close'])

const readerContainer = ref(null)
// AGENTS.md rule: ALWAYS wrap reader instance strictly inside shallowRef() (never ref() or reactive())
const readerInstance = shallowRef(null)

const loading = ref(true)
const error = ref(null)
const progressPercent = ref(0)
const currentSectionTitle = ref('')
const showControls = ref(true)
const isFullscreen = ref(false)
const fontSize = ref(100)

function prevPage() {
  if (readerInstance.value && typeof readerInstance.value.prev === 'function') {
    readerInstance.value.prev()
  }
}

function nextPage() {
  if (readerInstance.value && typeof readerInstance.value.next === 'function') {
    readerInstance.value.next()
  }
}

function adjustFontSize(delta) {
  const newSize = Math.max(70, Math.min(200, fontSize.value + delta))
  fontSize.value = newSize
  if (readerInstance.value?.renderer?.setAttribute) {
    readerInstance.value.renderer.setAttribute('style', `font-size: ${newSize}%;`)
  }
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().catch(() => {})
    isFullscreen.value = true
  } else {
    document.exitFullscreen().catch(() => {})
    isFullscreen.value = false
  }
}

function handleKeyDown(e) {
  if (e.key === 'Escape') {
    emit('close')
  } else if (e.key === 'ArrowRight' || e.key === ' ' || e.key === 'PageDown') {
    nextPage()
  } else if (e.key === 'ArrowLeft' || e.key === 'PageUp') {
    prevPage()
  }
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeyDown)

  try {
    // Dynamically import foliate-js view component
    await import('foliate-js/view.js')

    if (!readerContainer.value) return

    // Create foliate-view element
    const view = document.createElement('foliate-view')
    view.style.width = '100%'
    view.style.height = '100%'
    view.style.display = 'block'
    readerContainer.value.appendChild(view)

    // Store strictly in shallowRef
    readerInstance.value = view

    // Listen to relocation/progress events
    view.addEventListener('relocate', (e) => {
      if (e.detail?.fraction != null) {
        progressPercent.value = Math.round(e.detail.fraction * 100)
      }
      if (e.detail?.tocItem?.label) {
        currentSectionTitle.value = e.detail.tocItem.label
      }
    })

    // Fetch EPUB file as blob from Lyostar backend
    const res = await fetch(props.book.file_url)
    if (!res.ok) {
      throw new Error(`Failed to load book file: ${res.statusText}`)
    }
    const blob = await res.blob()
    const file = new File([blob], `${props.book.title}.epub`, { type: 'application/epub+zip' })

    await view.open(file)
    loading.value = false
  } catch (err) {
    console.error('[Lyostar Reader] Error initializing reader:', err)
    error.value = err.message || 'Failed to initialize reader'
    loading.value = false
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeyDown)

  // AGENTS.md rule: Clean up instances explicitly in onBeforeUnmount
  if (readerInstance.value) {
    try {
      if (typeof readerInstance.value.close === 'function') {
        readerInstance.value.close()
      }
    } catch (err) {
      console.warn('[Lyostar Reader] Error closing reader:', err)
    }
    try {
      readerInstance.value.remove()
    } catch (e) {}
    readerInstance.value = null
  }
})
</script>

<style scoped>
/* Ensure the custom foliate-view occupies entire space */
:deep(foliate-view) {
  width: 100%;
  height: 100%;
  display: block;
}
</style>
