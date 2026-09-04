<template>
  <div class="fixed inset-0 z-50 bg-[#090a0f] text-slate-100 flex flex-col select-none overflow-hidden animate-fade-in">
    <!-- Top Bar Controls -->
    <header 
      v-show="showControls"
      class="h-14 bg-[#090a0f]/95 backdrop-blur border-b border-white/[0.08] px-4 flex items-center justify-between z-20 transition-all duration-200"
    >
      <div class="flex items-center gap-3 min-w-0">
        <button
          @click="handleClose"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
          title="Back to Library (Esc)"
        >
          <ArrowLeft class="w-5 h-5" />
        </button>
        <div class="min-w-0">
          <div class="flex items-center gap-2">
            <h2 class="text-sm font-semibold text-white truncate max-w-xs sm:max-w-md">
              {{ book.title }}
            </h2>
            <span class="px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-glacier-400/10 text-glacier-400 border border-glacier-400/20 flex-shrink-0">
              {{ book.format || 'EPUB' }}
            </span>
            <span 
              v-if="isFinished"
              class="hidden sm:inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
            >
              <CheckCircle2 class="w-3 h-3" />
              Finished
            </span>
          </div>
          <p v-if="currentSectionTitle && book.format !== 'pdf'" class="text-[11px] text-glacier-400 truncate">
            {{ currentSectionTitle }}
          </p>
        </div>
      </div>

      <!-- Controls for PDF -->
      <div v-if="book.format === 'pdf'" class="flex items-center gap-1.5 sm:gap-2">
        <!-- PDF Page tracker / quick bookmark -->
        <div class="flex items-center gap-1.5 px-2.5 py-1 rounded-xl bg-[#11131b] border border-white/[0.08] text-xs">
          <span class="text-slate-400">Page:</span>
          <input
            type="number"
            min="1"
            v-model.number="pdfCurrentPage"
            @keydown.enter="savePdfPage"
            class="w-12 bg-transparent text-white font-mono text-center outline-none border-b border-white/20 focus:border-glacier-400"
            title="Type page number and click Save"
          />
          <button
            @click="savePdfPage"
            class="text-[10px] text-glacier-400 hover:text-glacier-300 ml-1 cursor-pointer font-medium px-1.5 py-0.5 rounded bg-glacier-500/10 hover:bg-glacier-500/20 transition-colors"
            title="Jump to page & Save position"
          >
            {{ saveStatus || 'Save' }}
          </button>
        </div>

        <!-- Mark as Finished button -->
        <button
          @click="toggleFinished"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
          :class="{ 'text-emerald-400 bg-emerald-500/10': isFinished }"
          :title="isFinished ? 'Mark as Unfinished' : 'Mark as Finished'"
        >
          <CheckCircle2 class="w-4 h-4" />
        </button>

        <a
          :href="book.file_url"
          target="_blank"
          rel="noopener noreferrer"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors"
          title="Open in New Tab"
        >
          <ExternalLink class="w-4 h-4" />
        </a>
        <a
          :href="book.file_url"
          download
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors"
          title="Download PDF"
        >
          <Download class="w-4 h-4" />
        </a>
        <button
          @click="toggleFullscreen"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
          title="Toggle Fullscreen"
        >
          <Maximize2 v-if="!isFullscreen" class="w-4 h-4" />
          <Minimize2 v-else class="w-4 h-4" />
        </button>
      </div>

      <!-- Controls for EPUB -->
      <div v-else class="flex items-center gap-1.5 sm:gap-2">
        <!-- Mark as Finished button -->
        <button
          @click="toggleFinished"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
          :class="{ 'text-emerald-400 bg-emerald-500/10': isFinished }"
          :title="isFinished ? 'Mark as Unfinished' : 'Mark as Finished'"
        >
          <CheckCircle2 class="w-4 h-4" />
        </button>
        <!-- Font Size Decrease -->
        <button
          @click="adjustFontSize(-10)"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
          title="Decrease Font Size"
        >
          <span class="text-xs font-bold font-serif">A-</span>
        </button>
        <!-- Font Size Increase -->
        <button
          @click="adjustFontSize(10)"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
          title="Increase Font Size"
        >
          <span class="text-sm font-bold font-serif">A+</span>
        </button>
        <!-- Toggle Fullscreen -->
        <button
          @click="toggleFullscreen"
          class="p-2 rounded-xl text-slate-300 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
          title="Toggle Fullscreen"
        >
          <Maximize2 v-if="!isFullscreen" class="w-4 h-4" />
          <Minimize2 v-else class="w-4 h-4" />
        </button>
      </div>
    </header>

    <!-- Reader Viewport -->
    <main class="flex-1 relative w-full h-full overflow-hidden flex items-center justify-center">
      <!-- Loading State (both EPUB and PDF wait for saved progress to load) -->
      <div v-if="loading" class="flex flex-col items-center gap-3 text-slate-400">
        <Loader2 class="w-7 h-7 text-glacier-400 animate-spin" />
        <span class="text-xs font-medium tracking-wide">
          {{ resumeNotice || 'Loading book...' }}
        </span>
      </div>

      <!-- Error State -->
      <div v-else-if="error" class="p-6 max-w-md text-center bg-red-500/10 border border-red-500/20 rounded-2xl text-red-300 text-sm">
        <p class="font-semibold mb-2">Unable to open book</p>
        <p class="text-xs text-red-400/80 mb-4">{{ error }}</p>
        <button
          @click="handleClose"
          class="px-4 py-2 rounded-xl bg-red-500/20 hover:bg-red-500/30 text-red-200 text-xs font-medium cursor-pointer"
        >
          Return to Shelf
        </button>
      </div>

      <!-- PDF Reader Viewport: Mounted strictly AFTER progress is loaded -->
      <iframe
        v-else-if="book.format === 'pdf'"
        :key="iframeKey"
        :src="pdfSrc"
        class="w-full h-full border-0 bg-[#090a0f]"
        title="PDF Reader"
      />

      <!-- foliate-view container mounted here for EPUB -->
      <div 
        v-else
        ref="readerContainer" 
        class="w-full h-full select-text"
      ></div>

      <!-- Floating Prev / Next Click Zones (EPUB only) -->
      <template v-if="!loading && book.format !== 'pdf'">
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
      </template>
    </main>

    <!-- Bottom Progress Bar Controls (EPUB only) -->
    <footer 
      v-show="showControls && book.format !== 'pdf'"
      class="h-10 bg-[#090a0f]/95 backdrop-blur border-t border-white/[0.08] px-4 flex items-center justify-between text-xs text-slate-400 z-20"
    >
      <button 
        @click="prevPage" 
        class="flex items-center gap-1 hover:text-white transition-colors cursor-pointer"
      >
        <ChevronLeft class="w-4 h-4" />
        <span class="hidden sm:inline">Prev</span>
      </button>

      <div class="flex items-center gap-3 flex-1 max-w-xs mx-4">
        <div class="w-full bg-white/[0.08] h-1.5 rounded-full overflow-hidden">
          <div 
            class="h-full rounded-full transition-all duration-300"
            :class="isFinished ? 'bg-emerald-400' : 'bg-glacier-400'"
            :style="{ width: `${progressPercent}%` }"
          ></div>
        </div>
        <span class="font-mono text-[11px] min-w-8 text-right">{{ progressPercent }}%</span>
      </div>

      <button 
        @click="nextPage" 
        class="flex items-center gap-1 hover:text-white transition-colors cursor-pointer"
      >
        <span class="hidden sm:inline">Next</span>
        <ChevronRight class="w-4 h-4" />
      </button>
    </footer>
  </div>
</template>

<script setup>
import { shallowRef, ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { ArrowLeft, ChevronLeft, ChevronRight, Maximize2, Minimize2, Loader2, Download, ExternalLink, CheckCircle2 } from 'lucide-vue-next'
import { fetchBookProgress, saveBookProgress } from '../api/client'

const props = defineProps({
  book: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'progress-updated'])

const readerContainer = ref(null)
// AGENTS.md rule: ALWAYS wrap reader instance strictly inside shallowRef() (never ref() or reactive())
const readerInstance = shallowRef(null)

const loading = ref(true)
const error = ref(null)
const progressPercent = ref(0)
const isFinished = ref(false)
const currentSectionTitle = ref('')
const showControls = ref(true)
const isFullscreen = ref(false)
const fontSize = ref(100)
const resumeNotice = ref('')

const lastCfi = ref('')
const pdfCurrentPage = ref(1)
const iframeKey = ref(0)
const saveStatus = ref('')

let saveTimeout = null

const pdfSrc = computed(() => {
  const page = pdfCurrentPage.value > 1 ? pdfCurrentPage.value : 1
  return `${props.book.file_url}#page=${page}`
})

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

async function persistProgress({ location, progress, currentPage, isFinishedVal }) {
  try {
    await saveBookProgress(props.book.id, {
      location: location ?? lastCfi.value,
      progress: progress ?? (progressPercent.value / 100),
      currentPage: currentPage ?? pdfCurrentPage.value,
      isFinished: isFinishedVal ?? isFinished.value
    })
    emit('progress-updated', {
      bookId: props.book.id,
      progress: progress ?? (progressPercent.value / 100),
      isFinished: isFinishedVal ?? isFinished.value
    })
  } catch (err) {
    console.warn('[Lyostar Reader] Failed to save progress:', err)
  }
}

function queueSaveProgress(cfi, fraction) {
  clearTimeout(saveTimeout)
  saveTimeout = setTimeout(() => {
    persistProgress({
      location: cfi,
      progress: fraction,
      isFinishedVal: fraction >= 0.999 ? true : isFinished.value
    })
  }, 1200)
}

async function savePdfPage() {
  if (pdfCurrentPage.value < 1) pdfCurrentPage.value = 1
  iframeKey.value++
  saveStatus.value = 'Saved!'
  setTimeout(() => {
    saveStatus.value = ''
  }, 1800)

  await persistProgress({
    location: `page-${pdfCurrentPage.value}`,
    currentPage: pdfCurrentPage.value,
    progress: isFinished.value ? 1.0 : (pdfCurrentPage.value > 1 ? 0.5 : 0.05)
  })
}

async function toggleFinished() {
  isFinished.value = !isFinished.value
  if (isFinished.value) {
    progressPercent.value = 100
  }
  await persistProgress({
    isFinishedVal: isFinished.value,
    progress: isFinished.value ? 1.0 : (progressPercent.value / 100)
  })
}

function handleKeyDown(e) {
  if (e.key === 'Escape') {
    handleClose()
  } else if (props.book.format !== 'pdf') {
    if (e.key === 'ArrowRight' || e.key === ' ' || e.key === 'PageDown') {
      nextPage()
    } else if (e.key === 'ArrowLeft' || e.key === 'PageUp') {
      prevPage()
    }
  }
}

function handleClose() {
  // Flush pending save
  if (saveTimeout) {
    clearTimeout(saveTimeout)
    persistProgress({})
  }
  emit('close')
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeyDown)

  // 1. Fetch initial progress from server
  let savedProgress = null
  try {
    savedProgress = await fetchBookProgress(props.book.id)
    if (savedProgress) {
      if (savedProgress.current_page > 0) {
        pdfCurrentPage.value = savedProgress.current_page
      }
      if (savedProgress.progress > 0) {
        progressPercent.value = Math.round(savedProgress.progress * 100)
      }
      if (savedProgress.is_finished) {
        isFinished.value = true
      }
      if (savedProgress.location) {
        lastCfi.value = savedProgress.location
      }
    }
  } catch (err) {
    console.warn('[Lyostar Reader] Could not fetch saved progress:', err)
  }

  // 2. Handle PDF: pdfCurrentPage is now populated with saved page, reveal iframe
  if (props.book.format === 'pdf') {
    if (pdfCurrentPage.value > 1) {
      resumeNotice.value = `Opening page ${pdfCurrentPage.value}...`
    }
    loading.value = false
    return
  }

  // 3. Handle EPUB
  try {
    await import('foliate-js/view.js')

    if (!readerContainer.value) return

    const view = document.createElement('foliate-view')
    view.style.width = '100%'
    view.style.height = '100%'
    view.style.display = 'block'
    readerContainer.value.appendChild(view)

    readerInstance.value = view

    // Listen to relocate
    view.addEventListener('relocate', (e) => {
      if (e.detail?.fraction != null) {
        progressPercent.value = Math.round(e.detail.fraction * 100)
      }
      if (e.detail?.cfi) {
        lastCfi.value = e.detail.cfi
        queueSaveProgress(e.detail.cfi, e.detail.fraction ?? 0)
      }
      if (e.detail?.tocItem?.label) {
        currentSectionTitle.value = e.detail.tocItem.label
      }
    })

    const res = await fetch(props.book.file_url)
    if (!res.ok) {
      throw new Error(`Failed to load book file: ${res.statusText}`)
    }
    const blob = await res.blob()
    const file = new File([blob], `${props.book.title}.epub`, { type: 'application/epub+zip' })

    await view.open(file)

    // Resume reading from saved CFI if available
    if (savedProgress?.location && savedProgress.location.startsWith('epubcfi')) {
      resumeNotice.value = `Resuming from ${Math.round(savedProgress.progress * 100)}%...`
      try {
        await view.goTo(savedProgress.location)
      } catch (resumeErr) {
        console.warn('[Lyostar Reader] Could not restore location CFI:', resumeErr)
      }
    }

    loading.value = false
  } catch (err) {
    console.error('[Lyostar Reader] Error initializing reader:', err)
    error.value = err.message || 'Failed to initialize reader'
    loading.value = false
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeyDown)

  if (saveTimeout) {
    clearTimeout(saveTimeout)
    persistProgress({})
  }

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
:deep(foliate-view) {
  width: 100%;
  height: 100%;
  display: block;
}
</style>
