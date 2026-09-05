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

      <!-- Controls for PDF (Official Mozilla PDF.js Engine) -->
      <div v-if="book.format === 'pdf'" class="flex items-center gap-1.5 sm:gap-2">
        <!-- Live Page Tracker -->
        <div class="flex items-center gap-1.5 px-3 py-1 rounded-xl bg-[#11131b] border border-white/[0.08] text-xs font-mono">
          <span class="text-slate-400">Page</span>
          <span class="text-white font-semibold">{{ pdfCurrentPage }}</span>
          <span class="text-slate-500">/</span>
          <span class="text-slate-400">{{ totalPdfPages || '...' }}</span>
          <span class="text-glacier-400 ml-1">({{ progressPercent }}%)</span>
        </div>

        <!-- 1-Click Bookmark Button -->
        <button
          @click="toggleBookmark"
          class="p-2 rounded-xl transition-all cursor-pointer"
          :class="isCurrentLocationBookmarked 
            ? 'text-amber-400 bg-amber-500/15 border border-amber-500/30 shadow-sm' 
            : 'text-slate-300 hover:text-white hover:bg-white/[0.06] border border-transparent'"
          :title="isCurrentLocationBookmarked ? 'Bookmarked (Click to remove)' : 'Bookmark Page'"
        >
          <Bookmark class="w-4 h-4" :class="{ 'fill-amber-400': isCurrentLocationBookmarked }" />
        </button>

        <!-- Notes & Marks Drawer Toggle -->
        <button
          @click="showSidePanel = !showSidePanel"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-xl border text-xs font-medium transition-all cursor-pointer"
          :class="showSidePanel 
            ? 'border-glacier-400/40 bg-glacier-500/15 text-glacier-300 shadow-sm' 
            : 'border-white/[0.08] hover:border-white/[0.15] bg-white/[0.04] hover:bg-white/[0.08] text-slate-300 hover:text-white'"
          title="Bookmarks & Notes Panel"
        >
          <BookMarked class="w-4 h-4 text-glacier-400" />
          <span class="hidden sm:inline">Notes</span>
          <span v-if="bookmarks.length + highlights.length > 0" class="text-[10px] font-mono px-1.5 py-0.2 rounded-full bg-white/[0.1] text-white">
            {{ bookmarks.length + highlights.length }}
          </span>
        </button>

        <!-- Mark as Finished button -->
        <button
          @click="toggleFinished"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-medium transition-colors cursor-pointer border"
          :class="isFinished 
            ? 'text-emerald-400 bg-emerald-500/10 border-emerald-500/30' 
            : 'text-slate-300 hover:text-white bg-white/[0.04] hover:bg-white/[0.08] border-white/[0.08]'"
          :title="isFinished ? 'Click to mark as unfinished' : 'Mark as finished'"
        >
          <CheckCircle2 class="w-3.5 h-3.5" />
          <span class="hidden md:inline">{{ isFinished ? 'Completed' : 'Mark Finished' }}</span>
        </button>

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
        <!-- 1-Click Bookmark Button -->
        <button
          @click="toggleBookmark"
          class="p-2 rounded-xl transition-all cursor-pointer"
          :class="isCurrentLocationBookmarked 
            ? 'text-amber-400 bg-amber-500/15 border border-amber-500/30 shadow-sm' 
            : 'text-slate-300 hover:text-white hover:bg-white/[0.06] border border-transparent'"
          :title="isCurrentLocationBookmarked ? 'Bookmarked (Click to remove)' : 'Bookmark Current Position'"
        >
          <Bookmark class="w-4 h-4" :class="{ 'fill-amber-400': isCurrentLocationBookmarked }" />
        </button>

        <!-- Notes & Marks Drawer Toggle -->
        <button
          @click="showSidePanel = !showSidePanel"
          class="flex items-center gap-1.5 px-2.5 py-1.5 rounded-xl border text-xs font-medium transition-all cursor-pointer"
          :class="showSidePanel 
            ? 'border-glacier-400/40 bg-glacier-500/15 text-glacier-300 shadow-sm' 
            : 'border-white/[0.08] hover:border-white/[0.15] bg-white/[0.04] hover:bg-white/[0.08] text-slate-300 hover:text-white'"
          title="Bookmarks, Notes & TOC"
        >
          <BookMarked class="w-4 h-4 text-glacier-400" />
          <span class="hidden sm:inline">Notes</span>
          <span v-if="bookmarks.length + highlights.length > 0" class="text-[10px] font-mono px-1.5 py-0.2 rounded-full bg-white/[0.1] text-white">
            {{ bookmarks.length + highlights.length }}
          </span>
        </button>

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

    <!-- Reader Viewport Container -->
    <div class="flex-1 relative w-full h-full overflow-hidden flex">
      <!-- Main Reader Content Area -->
      <main class="flex-1 relative w-full h-full overflow-hidden flex items-center justify-center">
        <!-- Loading State -->
        <div v-if="loading" class="flex flex-col items-center gap-3 text-slate-400">
          <Loader2 class="w-7 h-7 text-glacier-400 animate-spin" />
          <span class="text-xs font-medium tracking-wide">
            {{ resumeNotice || 'Loading book content...' }}
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

        <!-- PDF Reader Viewport: Official Mozilla PDF.js Viewer -->
        <iframe
          v-else-if="book.format === 'pdf'"
          ref="pdfFrame"
          :src="pdfViewerUrl"
          @load="onPdfFrameLoaded"
          class="w-full h-full border-0 bg-[#090a0f]"
          title="PDF Reader"
        />

        <!-- Foliate-view container mounted here for EPUB -->
        <div 
          v-else
          ref="readerContainer" 
          class="w-full h-full select-text"
        ></div>

        <!-- Floating Prev / Next Click Zones (EPUB only) -->
        <template v-if="!loading && !error && book.format !== 'pdf'">
          <button 
            @click="prevPage" 
            class="absolute left-0 top-0 bottom-0 w-16 group flex items-center justify-start pl-3 z-10 cursor-pointer opacity-0 hover:opacity-100 transition-opacity"
            aria-label="Previous page"
          >
            <div class="p-2 rounded-full bg-black/60 backdrop-blur border border-white/10 text-white shadow-lg group-hover:scale-110 transition-transform">
              <ChevronLeft class="w-5 h-5" />
            </div>
          </button>

          <button 
            @click="nextPage" 
            class="absolute right-0 top-0 bottom-0 w-16 group flex items-center justify-end pr-3 z-10 cursor-pointer opacity-0 hover:opacity-100 transition-opacity"
            aria-label="Next page"
          >
            <div class="p-2 rounded-full bg-black/60 backdrop-blur border border-white/10 text-white shadow-lg group-hover:scale-110 transition-transform">
              <ChevronRight class="w-5 h-5" />
            </div>
          </button>
        </template>
      </main>

      <!-- Slide-out Side Drawer Panel (Bookmarks, Notes, Table of Contents) -->
      <aside
        v-if="showSidePanel"
        class="w-80 sm:w-96 bg-[#0c0e15] border-l border-white/[0.08] shadow-2xl shadow-black/80 flex flex-col z-30 transition-all flex-shrink-0 animate-fade-in"
      >
        <!-- Drawer Header -->
        <div class="h-14 px-3.5 border-b border-white/[0.08] flex items-center justify-between flex-shrink-0">
          <!-- Navigation Tabs -->
          <div class="flex items-center gap-1 p-1 rounded-xl bg-white/[0.03] border border-white/[0.05]">
            <button
              @click="activeDrawerTab = 'bookmarks'"
              class="px-2.5 py-1 rounded-lg text-xs font-medium transition-all cursor-pointer"
              :class="activeDrawerTab === 'bookmarks' ? 'bg-glacier-500 text-slate-950 font-semibold shadow-sm' : 'text-slate-400 hover:text-white'"
            >
              Marks ({{ bookmarks.length }})
            </button>
            <button
              @click="activeDrawerTab = 'notes'"
              class="px-2.5 py-1 rounded-lg text-xs font-medium transition-all cursor-pointer"
              :class="activeDrawerTab === 'notes' ? 'bg-glacier-500 text-slate-950 font-semibold shadow-sm' : 'text-slate-400 hover:text-white'"
            >
              Notes ({{ highlights.length }})
            </button>
            <button
              v-if="book.format !== 'pdf' && tocItems.length > 0"
              @click="activeDrawerTab = 'toc'"
              class="px-2.5 py-1 rounded-lg text-xs font-medium transition-all cursor-pointer"
              :class="activeDrawerTab === 'toc' ? 'bg-glacier-500 text-slate-950 font-semibold shadow-sm' : 'text-slate-400 hover:text-white'"
            >
              TOC
            </button>
          </div>

          <!-- Close Drawer Button -->
          <button
            @click="showSidePanel = false"
            class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/[0.06] transition-colors cursor-pointer"
            title="Close panel"
          >
            <X class="w-4 h-4" />
          </button>
        </div>

        <!-- Drawer Body Content -->
        <div class="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-thin">
          <!-- TAB 1: BOOKMARKS -->
          <div v-if="activeDrawerTab === 'bookmarks'" class="space-y-3">
            <!-- Bookmark current location button -->
            <button
              @click="toggleBookmark"
              class="w-full flex items-center justify-center gap-2 py-2 px-3 rounded-xl border text-xs font-medium transition-all cursor-pointer"
              :class="isCurrentLocationBookmarked
                ? 'bg-amber-500/15 border-amber-500/40 text-amber-300'
                : 'bg-glacier-500/10 hover:bg-glacier-500/20 border-glacier-400/30 text-glacier-300 hover:text-white'"
            >
              <Bookmark class="w-3.5 h-3.5" :class="{ 'fill-amber-400 text-amber-400': isCurrentLocationBookmarked }" />
              <span>{{ isCurrentLocationBookmarked ? 'Remove Bookmark' : 'Bookmark Current Position' }}</span>
            </button>

            <!-- Bookmarks list -->
            <div v-if="bookmarks.length > 0" class="space-y-2">
              <div
                v-for="b in bookmarks"
                :key="b.id"
                @click="jumpToLocation(b.location)"
                class="group p-3 rounded-xl bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/40 flex items-center justify-between gap-3 cursor-pointer transition-all"
              >
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <Bookmark class="w-3.5 h-3.5 text-amber-400 flex-shrink-0 fill-amber-400" />
                    <span class="text-xs font-semibold text-white truncate group-hover:text-glacier-400 transition-colors">
                      {{ b.title }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2 mt-1 text-[10px] text-slate-400 font-mono">
                    <span class="text-glacier-400">{{ Math.round(b.progress * 100) }}%</span>
                    <span>•</span>
                    <span>{{ formatDate(b.created_at) }}</span>
                  </div>
                </div>

                <button
                  @click.stop="handleDeleteBookmark(b.id)"
                  class="p-1.5 rounded-lg text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 opacity-60 group-hover:opacity-100 transition-all cursor-pointer"
                  title="Delete bookmark"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>

            <!-- Empty Bookmarks State -->
            <div v-else class="py-12 text-center text-slate-500 text-xs leading-relaxed">
              No bookmarks saved yet.<br />Click "Bookmark Current Position" to save your spot!
            </div>
          </div>

          <!-- TAB 2: NOTES & HIGHLIGHTS -->
          <div v-if="activeDrawerTab === 'notes'" class="space-y-4">
            <!-- Add Note Card -->
            <div class="p-3.5 rounded-2xl bg-[#11131b] border border-white/[0.08] space-y-2.5">
              <div class="flex items-center justify-between">
                <span class="text-xs font-semibold text-white">Add Note / Highlight</span>
                <!-- Color palette -->
                <div class="flex items-center gap-1.5">
                  <button
                    v-for="c in ['yellow', 'green', 'blue', 'pink']"
                    :key="c"
                    @click="newNoteColor = c"
                    class="w-4 h-4 rounded-full transition-transform cursor-pointer"
                    :class="[
                      colorClasses[c].dot,
                      newNoteColor === c ? 'scale-125 ring-2 ring-white/60' : 'opacity-60 hover:opacity-100'
                    ]"
                    :title="c"
                  ></button>
                </div>
              </div>

              <textarea
                v-model="newNoteText"
                rows="2"
                placeholder="Write your note or thoughts here..."
                class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.08] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 transition-colors resize-none leading-relaxed"
              ></textarea>

              <input
                v-model="newSelectedText"
                type="text"
                placeholder="Optional quoted passage or phrase..."
                class="w-full px-3 py-1.5 rounded-xl bg-slate-900/80 border border-white/[0.08] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 transition-colors"
              />

              <div class="flex justify-end">
                <button
                  @click="handleAddHighlight"
                  :disabled="savingNote || (!newNoteText.trim() && !newSelectedText.trim())"
                  class="px-3.5 py-1.5 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs shadow-sm shadow-glacier-500/20 disabled:opacity-40 transition-all cursor-pointer"
                >
                  {{ savingNote ? 'Saving...' : 'Save Note' }}
                </button>
              </div>
            </div>

            <!-- Notes List -->
            <div v-if="highlights.length > 0" class="space-y-2.5">
              <div
                v-for="h in highlights"
                :key="h.id"
                @click="jumpToLocation(h.location)"
                class="group p-3 rounded-xl bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/40 space-y-2 cursor-pointer transition-all"
              >
                <!-- Quoted text -->
                <blockquote
                  v-if="h.selected_text"
                  class="text-xs italic pl-2.5 border-l-2 py-0.5 text-slate-300"
                  :class="colorClasses[h.color || 'yellow'].border"
                >
                  "{{ h.selected_text }}"
                </blockquote>

                <!-- Note comment -->
                <p v-if="h.note" class="text-xs text-white leading-relaxed">
                  {{ h.note }}
                </p>

                <!-- Footer row -->
                <div class="flex items-center justify-between text-[10px] text-slate-500 pt-1 border-t border-white/[0.04]">
                  <div class="flex items-center gap-1.5">
                    <span class="w-2 h-2 rounded-full" :class="colorClasses[h.color || 'yellow'].dot"></span>
                    <span>{{ formatDate(h.created_at) }}</span>
                  </div>
                  <button
                    @click.stop="handleDeleteHighlight(h.id)"
                    class="p-1 rounded text-slate-500 hover:text-rose-400 hover:bg-rose-500/10 opacity-60 group-hover:opacity-100 transition-all cursor-pointer"
                    title="Delete note"
                  >
                    <Trash2 class="w-3 h-3" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Empty Notes State -->
            <div v-else class="py-12 text-center text-slate-500 text-xs leading-relaxed">
              No notes or highlights saved yet.<br />Write down your insights above!
            </div>
          </div>

          <!-- TAB 3: TABLE OF CONTENTS (EPUB) -->
          <div v-if="activeDrawerTab === 'toc'" class="space-y-1">
            <button
              v-for="(item, idx) in tocItems"
              :key="idx"
              @click="jumpToToc(item)"
              class="w-full text-left px-3 py-2 rounded-xl text-xs font-medium text-slate-300 hover:text-white hover:bg-white/[0.05] transition-all truncate cursor-pointer"
              :class="{ 'text-glacier-400 bg-glacier-500/10 font-semibold': currentSectionTitle === item.label }"
            >
              {{ item.label }}
            </button>
          </div>
        </div>
      </aside>
    </div>

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
import { 
  ArrowLeft, 
  ChevronLeft, 
  ChevronRight, 
  Maximize2, 
  Minimize2, 
  Loader2, 
  Download, 
  CheckCircle2, 
  Bookmark, 
  BookMarked, 
  Trash2, 
  X 
} from 'lucide-vue-next'
import { 
  fetchBookProgress, 
  saveBookProgress, 
  fetchBookmarks, 
  createBookmark, 
  deleteBookmark, 
  fetchHighlights, 
  createHighlight, 
  deleteHighlight 
} from '../api/client'

const props = defineProps({
  book: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'progress-updated'])

// Containers
const readerContainer = ref(null)
const pdfFrame = ref(null)

// AGENTS.md rule: ALWAYS wrap reader instance strictly inside shallowRef() (never ref() or reactive())
const readerInstance = shallowRef(null)

// UI State
const loading = ref(true)
const error = ref(null)
const progressPercent = ref(0)
const isFinished = ref(false)
const currentSectionTitle = ref('')
const showControls = ref(true)
const isFullscreen = ref(false)
const fontSize = ref(100)
const resumeNotice = ref('')

// Side Drawer State (Bookmarks, Notes, TOC)
const showSidePanel = ref(false)
const activeDrawerTab = ref('bookmarks') // 'bookmarks' | 'notes' | 'toc'
const bookmarks = ref([])
const highlights = ref([])
const tocItems = ref([])

// Form state for creating a note
const newNoteText = ref('')
const newSelectedText = ref('')
const newNoteColor = ref('yellow')
const savingNote = ref(false)

// EPUB state
const lastCfi = ref('')

// PDF state
const totalPdfPages = ref(0)
const pdfCurrentPage = ref(1)
const initialPdfPage = ref(1)

let saveTimeout = null
let pollAppTimer = null

const colorClasses = {
  yellow: { dot: 'bg-amber-400', border: 'border-amber-400/60' },
  green: { dot: 'bg-emerald-400', border: 'border-emerald-400/60' },
  blue: { dot: 'bg-sky-400', border: 'border-sky-400/60' },
  pink: { dot: 'bg-rose-400', border: 'border-rose-400/60' }
}

const pdfViewerUrl = computed(() => {
  const page = initialPdfPage.value > 1 ? initialPdfPage.value : 1
  return `/pdfjs/viewer.html?file=${encodeURIComponent(props.book.file_url)}#page=${page}`
})

const isCurrentLocationBookmarked = computed(() => {
  if (props.book.format === 'pdf') {
    const loc = `page-${pdfCurrentPage.value}`
    return bookmarks.value.some(b => b.location === loc)
  }
  if (!lastCfi.value) return false
  return bookmarks.value.some(b => b.location === lastCfi.value)
})

function formatDate(dateStr) {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
  } catch {
    return dateStr
  }
}

async function loadBookmarksAndHighlights() {
  try {
    const [bms, hls] = await Promise.all([
      fetchBookmarks(props.book.id).catch(() => []),
      fetchHighlights(props.book.id).catch(() => [])
    ])
    bookmarks.value = bms || []
    highlights.value = hls || []
  } catch (err) {
    console.warn('[Lyostar Reader] Could not load marks:', err)
  }
}

async function toggleBookmark() {
  const loc = props.book.format === 'pdf' ? `page-${pdfCurrentPage.value}` : lastCfi.value
  if (!loc) return

  const existing = bookmarks.value.find(b => b.location === loc)
  if (existing) {
    await handleDeleteBookmark(existing.id)
  } else {
    try {
      const title = props.book.format === 'pdf'
        ? `Page ${pdfCurrentPage.value}`
        : (currentSectionTitle.value || `Section at ${progressPercent.value}%`)
      const created = await createBookmark(props.book.id, {
        location: loc,
        title: title,
        progress: progressPercent.value / 100
      })
      bookmarks.value = [created, ...bookmarks.value]
    } catch (err) {
      console.warn('[Lyostar Reader] Failed to create bookmark:', err)
    }
  }
}

async function handleDeleteBookmark(bookmarkId) {
  try {
    await deleteBookmark(props.book.id, bookmarkId)
    bookmarks.value = bookmarks.value.filter(b => b.id !== bookmarkId)
  } catch (err) {
    console.warn('[Lyostar Reader] Failed to delete bookmark:', err)
  }
}

async function handleAddHighlight() {
  const loc = props.book.format === 'pdf' ? `page-${pdfCurrentPage.value}` : (lastCfi.value || 'cfi-start')
  if (!newNoteText.trim() && !newSelectedText.trim()) return

  savingNote.value = true
  try {
    const created = await createHighlight(props.book.id, {
      location: loc,
      selected_text: newSelectedText.value.trim(),
      note: newNoteText.value.trim(),
      color: newNoteColor.value
    })
    highlights.value = [created, ...highlights.value]
    newNoteText.value = ''
    newSelectedText.value = ''
  } catch (err) {
    console.warn('[Lyostar Reader] Failed to create highlight:', err)
  } finally {
    savingNote.value = false
  }
}

async function handleDeleteHighlight(highlightId) {
  try {
    await deleteHighlight(props.book.id, highlightId)
    highlights.value = highlights.value.filter(h => h.id !== highlightId)
  } catch (err) {
    console.warn('[Lyostar Reader] Failed to delete highlight:', err)
  }
}

function jumpToLocation(loc) {
  if (!loc) return
  if (props.book.format === 'pdf') {
    if (loc.startsWith('page-')) {
      const pageNum = parseInt(loc.replace('page-', ''), 10)
      if (pageNum > 0) {
        const iframeWin = pdfFrame.value?.contentWindow
        if (iframeWin?.PDFViewerApplication) {
          iframeWin.PDFViewerApplication.page = pageNum
        }
      }
    }
  } else {
    if (readerInstance.value && typeof readerInstance.value.goTo === 'function') {
      readerInstance.value.goTo(loc).catch(err => {
        console.warn('[Lyostar Reader] Could not jump to location:', err)
      })
    }
  }
}

function jumpToToc(item) {
  if (!item?.href) return
  if (readerInstance.value && typeof readerInstance.value.goTo === 'function') {
    readerInstance.value.goTo(item.href)
  }
}

function onPdfFrameLoaded() {
  try {
    const iframeWin = pdfFrame.value?.contentWindow
    if (!iframeWin) return

    // Poll until Mozilla PDFViewerApplication has initialized
    let attempts = 0
    pollAppTimer = setInterval(() => {
      attempts++
      const app = iframeWin.PDFViewerApplication
      if (app && app.eventBus) {
        clearInterval(pollAppTimer)
        pollAppTimer = null

        // Initial setup when pages are ready
        app.eventBus.on('pagesinit', () => {
          if (initialPdfPage.value > 1) {
            app.page = initialPdfPage.value
          }
          if (app.pagesCount) {
            totalPdfPages.value = app.pagesCount
          }
        })

        // Listen to live page changes during continuous vertical scrolling
        app.eventBus.on('pagechanging', (evt) => {
          const page = evt.pageNumber
          const total = app.pagesCount || totalPdfPages.value || 1
          pdfCurrentPage.value = page
          totalPdfPages.value = total
          const fraction = Math.min(1.0, page / total)
          progressPercent.value = Math.round(fraction * 100)

          // Debounced background auto-save to SQLite
          queueSaveProgress(`page-${page}`, fraction, page)
        })
      } else if (attempts > 100) {
        clearInterval(pollAppTimer)
        pollAppTimer = null
      }
    }, 100)
  } catch (err) {
    console.warn('[Lyostar Reader] Could not bind to PDF frame:', err)
  }
}

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
    const curPage = currentPage ?? (props.book.format === 'pdf' ? pdfCurrentPage.value : 0)
    const prog = progress ?? (progressPercent.value / 100)
    const loc = location ?? (props.book.format === 'pdf' ? `page-${curPage}` : lastCfi.value)
    const finished = isFinishedVal ?? isFinished.value

    await saveBookProgress(props.book.id, {
      location: loc,
      progress: prog,
      currentPage: curPage,
      isFinished: finished
    })
    emit('progress-updated', {
      bookId: props.book.id,
      progress: prog,
      isFinished: finished
    })
  } catch (err) {
    console.warn('[Lyostar Reader] Failed to save progress:', err)
  }
}

function queueSaveProgress(loc, fraction, pageNum) {
  clearTimeout(saveTimeout)
  saveTimeout = setTimeout(() => {
    persistProgress({
      location: loc,
      progress: fraction,
      currentPage: pageNum,
      isFinishedVal: fraction >= 0.99 ? true : isFinished.value
    })
  }, 1200)
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
    if (showSidePanel.value) {
      showSidePanel.value = false
    } else {
      handleClose()
    }
  } else if (props.book.format !== 'pdf') {
    if (e.key === 'ArrowRight' || e.key === ' ' || e.key === 'PageDown') {
      nextPage()
    } else if (e.key === 'ArrowLeft' || e.key === 'PageUp') {
      prevPage()
    }
  }
}

function handleClose() {
  if (saveTimeout) {
    clearTimeout(saveTimeout)
    persistProgress({})
  }
  emit('close')
}

onMounted(async () => {
  window.addEventListener('keydown', handleKeyDown)

  // 1. Fetch saved reading progress & marks in parallel
  loadBookmarksAndHighlights()

  let savedProgress = null
  try {
    savedProgress = await fetchBookProgress(props.book.id)
    if (savedProgress) {
      if (savedProgress.current_page > 0) {
        initialPdfPage.value = savedProgress.current_page
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

  // 2. Handle PDF: Mount Mozilla PDF.js viewer iframe
  if (props.book.format === 'pdf') {
    if (initialPdfPage.value > 1) {
      resumeNotice.value = `Resuming at page ${initialPdfPage.value}...`
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
        queueSaveProgress(e.detail.cfi, e.detail.fraction ?? 0, 0)
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

    // Load Table of Contents
    if (view.toc && Array.isArray(view.toc)) {
      tocItems.value = view.toc
    }

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

  if (pollAppTimer) {
    clearInterval(pollAppTimer)
    pollAppTimer = null
  }

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
