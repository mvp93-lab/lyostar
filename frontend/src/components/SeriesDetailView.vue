<template>
  <div class="flex-1 w-full max-w-7xl mx-auto px-4 lg:px-8 py-6 sm:py-8 animate-fade-in">
    <!-- Top Breadcrumbs & Navigation -->
    <div class="flex flex-wrap items-center justify-between gap-3 mb-6 pb-4 border-b border-white/[0.08]">
      <nav class="flex items-center gap-2 text-xs text-slate-400">
        <button
          @click="goHome"
          class="flex items-center gap-1.5 text-slate-400 hover:text-white transition-colors cursor-pointer group"
        >
          <ArrowLeft class="w-4 h-4 text-slate-500 group-hover:text-glacier-400 transition-colors" />
          <span>Library</span>
        </button>
        <ChevronRight class="w-3.5 h-3.5 text-slate-600 flex-shrink-0" />
        <button
          @click="goToSeriesCatalog"
          class="text-slate-400 hover:text-glacier-400 transition-colors cursor-pointer"
        >
          Series
        </button>
        <ChevronRight class="w-3.5 h-3.5 text-slate-600 flex-shrink-0" />
        <span class="text-white font-medium truncate max-w-[200px] sm:max-w-md">{{ effectiveSeriesName }}</span>
      </nav>

      <button
        @click="goToSeriesCatalog"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-xs font-medium text-slate-300 hover:text-white transition-all cursor-pointer"
      >
        <Layers class="w-3.5 h-3.5 text-glacier-400" />
        <span>All Series</span>
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-24 flex flex-col items-center justify-center text-slate-500">
      <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
      <p class="text-xs uppercase tracking-wider">Loading series collection...</p>
    </div>

    <!-- Error / Empty State -->
    <div v-else-if="books.length === 0" class="py-20 flex flex-col items-center justify-center text-center px-4">
      <div class="w-14 h-14 rounded-2xl bg-white/[0.04] border border-white/[0.08] text-slate-500 flex items-center justify-center mb-3">
        <Layers class="w-7 h-7 text-slate-500" />
      </div>
      <h3 class="text-sm font-semibold text-white mb-1">
        Series not found
      </h3>
      <p class="text-xs text-slate-400 max-w-sm mb-4">
        No books found for series "{{ effectiveSeriesName }}".
      </p>
      <button
        @click="goToSeriesCatalog"
        class="px-4 py-2 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-semibold transition-all cursor-pointer"
      >
        Back to Series Catalog
      </button>
    </div>

    <div v-else class="space-y-8">
      <!-- Series Hero Header -->
      <div class="relative overflow-hidden rounded-3xl bg-gradient-to-br from-[#131722] via-[#0e1017] to-[#090a0f] border border-white/[0.08] p-6 sm:p-8 shadow-xl shadow-black/40">
        <!-- Subtle Glow Effect -->
        <div class="absolute top-0 right-0 w-96 h-96 bg-glacier-500/10 rounded-full blur-3xl pointer-events-none -mr-20 -mt-20"></div>

        <div class="relative z-10 flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div class="space-y-3 max-w-2xl">
            <!-- Badges Row -->
            <div class="flex flex-wrap items-center gap-2">
              <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-glacier-400/10 border border-glacier-400/20 text-glacier-400 text-[11px] font-semibold tracking-wide uppercase">
                <Layers class="w-3.5 h-3.5" />
                <span>Book Series</span>
              </span>
              <span class="px-2.5 py-1 rounded-lg bg-white/[0.04] border border-white/[0.08] text-slate-300 text-[11px] font-mono">
                {{ books.length }} {{ books.length === 1 ? 'Volume' : 'Volumes' }}
              </span>
              <span 
                v-for="fmt in availableFormats" 
                :key="fmt" 
                class="px-2 py-0.5 rounded-md bg-white/[0.04] border border-white/[0.08] text-slate-400 text-[10px] font-bold uppercase tracking-wider"
              >
                {{ fmt }}
              </span>
            </div>

            <!-- Series Title -->
            <h1 class="text-2xl sm:text-3xl lg:text-4xl font-extrabold text-white tracking-tight">
              {{ effectiveSeriesName }}
            </h1>

            <!-- Authors Row -->
            <div v-if="uniqueAuthors.length > 0" class="flex items-center gap-2 text-xs sm:text-sm text-slate-400">
              <span class="text-slate-500">Written by</span>
              <div class="flex flex-wrap items-center gap-1.5">
                <button
                  v-for="author in uniqueAuthors"
                  :key="author"
                  @click="navigateToAuthor(author)"
                  class="font-medium text-glacier-400 hover:text-glacier-300 hover:underline transition-colors cursor-pointer"
                >
                  {{ author }}
                </button>
              </div>
            </div>
          </div>

          <!-- Series Reading Progress Box -->
          <div class="w-full lg:w-80 p-5 rounded-2xl bg-[#161923]/80 border border-white/[0.08] backdrop-blur-md flex flex-col gap-3.5 shadow-lg">
            <div class="flex items-center justify-between text-xs">
              <span class="font-semibold text-white flex items-center gap-1.5">
                <CheckCircle2 class="w-4 h-4 text-glacier-400" />
                <span>Reading Progress</span>
              </span>
              <span class="font-mono text-glacier-400 font-bold">
                {{ completedCount }} / {{ books.length }}
              </span>
            </div>

            <!-- Progress Bar -->
            <div class="w-full h-2 rounded-full bg-slate-800 overflow-hidden">
              <div
                class="h-full bg-gradient-to-r from-glacier-500 to-glacier-400 rounded-full transition-all duration-500"
                :style="{ width: `${progressPercent}%` }"
              ></div>
            </div>

            <div class="flex items-center justify-between text-[11px] text-slate-400">
              <span>{{ progressPercent }}% completed</span>
              <span v-if="progressPercent === 100" class="text-emerald-400 font-semibold flex items-center gap-1">
                Completed 🎉
              </span>
              <span v-else-if="nextBookToRead" class="text-slate-300 font-medium">
                Next: Vol. #{{ nextBookToRead.series_index || 1 }}
              </span>
            </div>

            <!-- Quick Continue Reading Button -->
            <button
              v-if="nextBookToRead && canRead"
              @click="handleReadBook(nextBookToRead)"
              class="w-full mt-1 px-3.5 py-2 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs flex items-center justify-center gap-2 transition-all cursor-pointer shadow-md shadow-glacier-500/10 hover:shadow-glacier-500/20"
            >
              <BookOpen class="w-4 h-4" />
              <span class="truncate">
                {{ nextBookToRead.progress > 0 ? 'Continue' : 'Start' }} Vol. #{{ nextBookToRead.series_index || 1 }}: {{ nextBookToRead.title }}
              </span>
            </button>
          </div>
        </div>
      </div>

      <!-- Controls & View Switcher Toolbar -->
      <div class="flex flex-wrap items-center justify-between gap-3 pb-2 border-b border-white/[0.06]">
        <div class="flex items-center gap-2">
          <span class="text-xs font-semibold text-slate-400 uppercase tracking-wider">
            Sequential Volumes ({{ filteredBooks.length }})
          </span>

          <!-- Format Filter (if mixed formats) -->
          <div v-if="availableFormats.length > 1" class="flex items-center gap-1 ml-3 bg-white/[0.03] p-0.5 rounded-lg border border-white/[0.06]">
            <button
              @click="selectedFormat = ''"
              class="px-2 py-0.5 rounded text-[10px] font-semibold transition-colors cursor-pointer"
              :class="!selectedFormat ? 'bg-glacier-500 text-slate-950' : 'text-slate-400 hover:text-white'"
            >
              All
            </button>
            <button
              v-for="fmt in availableFormats"
              :key="fmt"
              @click="selectedFormat = fmt.toLowerCase()"
              class="px-2 py-0.5 rounded text-[10px] font-semibold transition-colors cursor-pointer uppercase"
              :class="selectedFormat === fmt.toLowerCase() ? 'bg-glacier-500 text-slate-950' : 'text-slate-400 hover:text-white'"
            >
              {{ fmt }}
            </button>
          </div>
        </div>

        <!-- View Mode Buttons -->
        <div class="flex items-center gap-1 bg-white/[0.04] p-1 rounded-xl border border-white/[0.08]">
          <button
            @click="viewMode = 'grid'"
            class="p-1.5 rounded-lg text-xs font-medium flex items-center gap-1.5 transition-all cursor-pointer"
            :class="viewMode === 'grid' ? 'bg-glacier-500/20 text-glacier-300 shadow-sm' : 'text-slate-400 hover:text-white'"
            title="Grid View (Volume Cards)"
          >
            <LayoutGrid class="w-4 h-4" />
            <span class="hidden sm:inline text-xs">Grid</span>
          </button>
          <button
            @click="viewMode = 'timeline'"
            class="p-1.5 rounded-lg text-xs font-medium flex items-center gap-1.5 transition-all cursor-pointer"
            :class="viewMode === 'timeline' ? 'bg-glacier-500/20 text-glacier-300 shadow-sm' : 'text-slate-400 hover:text-white'"
            title="Timeline View (Sequential Order)"
          >
            <List class="w-4 h-4" />
            <span class="hidden sm:inline text-xs">Timeline</span>
          </button>
        </div>
      </div>

      <!-- 1. Numbered Grid View Mode -->
      <div v-if="viewMode === 'grid'" class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
        <div
          v-for="b in filteredBooks"
          :key="b.id"
          @click="openBookDetail(b.id)"
          class="group relative bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/40 rounded-2xl p-3 flex flex-col cursor-pointer transition-all duration-300 hover:shadow-xl hover:shadow-glacier-500/5 hover:-translate-y-1"
        >
          <!-- Prominent Volume Ribbon / Badge on Top-Left -->
          <div class="absolute top-2 left-2 z-20 px-2 py-0.5 rounded-lg bg-glacier-500 text-slate-950 text-[10px] font-black tracking-tight font-mono shadow-md shadow-glacier-500/20 flex items-center gap-1">
            <span>VOL #{{ b.series_index || 1 }}</span>
          </div>

          <!-- Cover Image Container (Aspect Ratio 2:3) -->
          <div class="relative w-full aspect-[2/3] rounded-xl overflow-hidden bg-[#090a0f] border border-white/[0.05] flex-shrink-0">
            <img
              v-if="b.has_cover && b.cover_url"
              :src="b.cover_url"
              :alt="b.title"
              loading="lazy"
              class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
            />
            <!-- Fallback Cover if no image -->
            <div 
              v-else 
              class="w-full h-full flex flex-col justify-between p-4 bg-gradient-to-br from-slate-900 via-[#131722] to-slate-950 border border-white/[0.06]"
            >
              <div class="flex items-center gap-1.5 text-glacier-400/60 text-xs font-semibold uppercase tracking-wider mt-6">
                <FileText v-if="b.format === 'pdf'" class="w-3.5 h-3.5" />
                <Book v-else class="w-3.5 h-3.5" />
                <span>{{ (b.format || 'epub').toUpperCase() }}</span>
              </div>
              <div>
                <h4 class="text-sm font-semibold text-white/90 line-clamp-3 leading-tight mb-2">
                  {{ b.title }}
                </h4>
                <p class="text-xs text-slate-400 line-clamp-1">
                  {{ formatAuthors(b.authors) }}
                </p>
              </div>
              <div class="w-8 h-1 rounded-full bg-glacier-400/30"></div>
            </div>

            <!-- Quick Shelf button on top-right (visible on hover) -->
            <button
              @click.stop="openShelfModal(b)"
              class="absolute top-2 right-2 p-1.5 rounded-lg bg-black/70 backdrop-blur-md border border-white/[0.1] text-slate-300 hover:text-glacier-400 hover:border-glacier-400/40 opacity-0 group-hover:opacity-100 transition-all z-10 cursor-pointer"
              title="Add to shelf"
            >
              <Bookmark class="w-3.5 h-3.5" />
            </button>

            <!-- Reading Status Pill on bottom -->
            <div v-if="b.is_finished || b.progress > 0" class="absolute bottom-0 inset-x-0 bg-black/80 backdrop-blur-sm px-2.5 py-1.5 flex flex-col gap-1">
              <div class="flex items-center justify-between text-[10px] font-medium leading-none">
                <span :class="b.is_finished ? 'text-emerald-400' : 'text-glacier-400'">
                  {{ b.is_finished ? 'Finished ✓' : `${Math.round(b.progress * 100)}% read` }}
                </span>
              </div>
              <div class="w-full h-1 bg-white/20 rounded-full overflow-hidden">
                <div 
                  class="h-full rounded-full transition-all duration-300"
                  :class="b.is_finished ? 'bg-emerald-400' : 'bg-glacier-400'"
                  :style="{ width: `${b.is_finished ? 100 : Math.round(b.progress * 100)}%` }"
                ></div>
              </div>
            </div>
          </div>

          <!-- Book Metadata -->
          <div class="mt-2.5 flex-1 flex flex-col justify-between">
            <div>
              <h3 class="text-xs sm:text-sm font-semibold text-white/90 group-hover:text-glacier-400 transition-colors line-clamp-2 leading-tight">
                {{ b.title }}
              </h3>
              <p class="text-[11px] text-slate-400 line-clamp-1 mt-0.5">
                {{ formatAuthors(b.authors) }}
              </p>
            </div>

            <div class="mt-2.5 pt-2 border-t border-white/[0.06] flex items-center justify-between text-[11px]">
              <span class="text-slate-500 uppercase font-mono text-[10px] font-semibold">
                {{ b.format }}
              </span>
              <button
                v-if="canRead"
                @click.stop="handleReadBook(b)"
                class="text-glacier-400 hover:text-glacier-300 font-medium flex items-center gap-1 transition-colors cursor-pointer"
              >
                <BookOpen class="w-3 h-3" />
                <span>{{ b.progress > 0 && !b.is_finished ? 'Resume' : 'Read' }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- 2. Sequential Timeline View Mode -->
      <div v-else class="space-y-4">
        <div 
          v-for="(b, idx) in filteredBooks" 
          :key="b.id"
          class="relative flex flex-col sm:flex-row gap-4 sm:gap-6 p-4 sm:p-5 rounded-2xl bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/30 transition-all duration-200 group shadow-md"
        >
          <!-- Left: Big Volume Index & Timeline Connector -->
          <div class="flex sm:flex-col items-center sm:items-center justify-between sm:justify-start gap-2 sm:w-16 flex-shrink-0">
            <div class="w-12 h-12 rounded-2xl bg-gradient-to-br from-glacier-500/20 to-glacier-600/5 border border-glacier-400/30 text-glacier-300 font-extrabold text-base font-mono flex items-center justify-center shadow-md shadow-glacier-500/10 group-hover:scale-105 group-hover:border-glacier-400/60 transition-all">
              #{{ b.series_index || idx + 1 }}
            </div>
            <span class="text-[10px] text-slate-500 font-bold uppercase tracking-wider font-mono">
              Vol. {{ b.series_index || idx + 1 }}
            </span>
          </div>

          <!-- Center Left: Book Cover Thumbnail -->
          <div 
            @click="openBookDetail(b.id)"
            class="relative w-24 sm:w-28 aspect-[2/3] rounded-xl overflow-hidden bg-[#090a0f] border border-white/[0.08] flex-shrink-0 cursor-pointer shadow-md group-hover:shadow-lg transition-transform group-hover:scale-102"
          >
            <img
              v-if="b.has_cover && b.cover_url"
              :src="b.cover_url"
              :alt="b.title"
              loading="lazy"
              class="w-full h-full object-cover"
            />
            <div v-else class="w-full h-full flex flex-col justify-center items-center p-2 bg-slate-900 text-center">
              <Book class="w-6 h-6 text-slate-600 mb-1" />
              <span class="text-[9px] text-slate-400 line-clamp-2">{{ b.title }}</span>
            </div>

            <!-- Reading Status Ribbon -->
            <div 
              v-if="b.is_finished"
              class="absolute top-1.5 left-1.5 px-1.5 py-0.5 rounded bg-emerald-500 text-slate-950 font-bold text-[9px]"
            >
              Done
            </div>
          </div>

          <!-- Center Right: Content & Description -->
          <div class="flex-1 min-w-0 flex flex-col justify-between">
            <div class="space-y-1.5">
              <div class="flex flex-wrap items-center gap-2">
                <span class="px-2 py-0.5 rounded bg-white/[0.05] border border-white/[0.08] text-[10px] font-bold font-mono text-slate-300 uppercase">
                  {{ b.format }}
                </span>
                <span v-if="b.is_finished" class="text-xs text-emerald-400 font-medium flex items-center gap-1">
                  <CheckCircle2 class="w-3.5 h-3.5" /> Finished
                </span>
                <span v-else-if="b.progress > 0" class="text-xs text-glacier-400 font-medium">
                  Reading {{ Math.round(b.progress * 100) }}%
                </span>
              </div>

              <h3 
                @click="openBookDetail(b.id)"
                class="text-base sm:text-lg font-bold text-white group-hover:text-glacier-400 transition-colors cursor-pointer"
              >
                {{ b.title }}
              </h3>

              <p class="text-xs text-slate-400">
                by <span class="text-slate-300 font-medium">{{ formatAuthors(b.authors) }}</span>
              </p>

              <!-- Description Snippet -->
              <p v-if="b.description" class="text-xs text-slate-400/90 line-clamp-2 sm:line-clamp-3 pt-1 leading-relaxed">
                {{ cleanDescription(b.description) }}
              </p>
            </div>

            <!-- Reading Progress Bar (if in progress) -->
            <div v-if="b.progress > 0 && !b.is_finished" class="mt-3 max-w-xs space-y-1">
              <div class="flex items-center justify-between text-[10px] text-slate-400">
                <span>Reading progress</span>
                <span class="font-mono text-glacier-400">{{ Math.round(b.progress * 100) }}%</span>
              </div>
              <div class="w-full h-1.5 rounded-full bg-slate-800 overflow-hidden">
                <div class="h-full bg-glacier-400 rounded-full" :style="{ width: `${Math.round(b.progress * 100)}%` }"></div>
              </div>
            </div>

            <!-- Action Buttons Row -->
            <div class="mt-4 pt-3 border-t border-white/[0.06] flex flex-wrap items-center justify-between gap-3">
              <div class="flex items-center gap-2">
                <button
                  v-if="canRead"
                  @click="handleReadBook(b)"
                  class="px-4 py-1.5 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-semibold flex items-center gap-1.5 transition-all cursor-pointer shadow-sm"
                >
                  <BookOpen class="w-3.5 h-3.5" />
                  <span>{{ b.progress > 0 && !b.is_finished ? 'Resume Book' : 'Read Book' }}</span>
                </button>

                <button
                  @click="openBookDetail(b.id)"
                  class="px-3 py-1.5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-xs font-medium text-slate-300 hover:text-white transition-all cursor-pointer"
                >
                  Details
                </button>
              </div>

              <!-- Shelf Button -->
              <button
                @click="openShelfModal(b)"
                class="p-2 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] text-slate-400 hover:text-glacier-400 transition-colors cursor-pointer"
                title="Add to shelf"
              >
                <Bookmark class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Shelf Assignment Modal -->
    <ShelfSelectModal
      :is-open="showShelfModal"
      :book="selectedBookForShelf"
      @close="showShelfModal = false"
      @shelf-updated="onShelfUpdated"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { 
  Layers, 
  Book, 
  BookOpen, 
  ChevronRight, 
  ArrowLeft, 
  Loader2, 
  FileText, 
  Bookmark, 
  LayoutGrid, 
  List, 
  CheckCircle2 
} from 'lucide-vue-next'
import { fetchBooks } from '../api/client'
import { useAuth } from '../composables/useAuth'
import ShelfSelectModal from './ShelfSelectModal.vue'

const props = defineProps({
  seriesName: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['read', 'filter-author', 'shelf-updated'])

const router = useRouter()
const route = useRoute()
const { canRead } = useAuth()

const books = ref([])
const loading = ref(true)
const viewMode = ref('grid') // 'grid' | 'timeline'
const selectedFormat = ref('')

const showShelfModal = ref(false)
const selectedBookForShelf = ref(null)

const effectiveSeriesName = computed(() => {
  if (props.seriesName) return props.seriesName
  if (route.params.series) {
    try {
      return decodeURIComponent(String(route.params.series))
    } catch {
      return String(route.params.series)
    }
  }
  return ''
})

async function loadSeriesBooks() {
  const name = effectiveSeriesName.value
  if (!name) {
    loading.value = false
    return
  }
  loading.value = true
  try {
    const res = await fetchBooks({ series: name, limit: 100 })
    books.value = res.items || res.books || []
  } catch (err) {
    console.error('Failed to load series books:', err)
  } finally {
    loading.value = false
  }
}

const availableFormats = computed(() => {
  const fmts = new Set()
  books.value.forEach(b => {
    if (b.format) fmts.add(b.format.toUpperCase())
  })
  return Array.from(fmts)
})

const filteredBooks = computed(() => {
  if (!selectedFormat.value) {
    return books.value
  }
  return books.value.filter(b => (b.format || '').toLowerCase() === selectedFormat.value)
})

const uniqueAuthors = computed(() => {
  const authorsSet = new Set()
  books.value.forEach(b => {
    if (b.authors && Array.isArray(b.authors)) {
      b.authors.forEach(a => {
        const name = typeof a === 'string' ? a : a.name
        if (name) authorsSet.add(name)
      })
    }
  })
  return Array.from(authorsSet)
})

const completedCount = computed(() => {
  return books.value.filter(b => b.is_finished || (b.progress && b.progress >= 0.95)).length
})

const progressPercent = computed(() => {
  if (books.value.length === 0) return 0
  return Math.round((completedCount.value / books.value.length) * 100)
})

const nextBookToRead = computed(() => {
  // Find first book that is in progress or unread
  const inProgress = books.value.find(b => b.progress > 0 && !b.is_finished)
  if (inProgress) return inProgress
  return books.value.find(b => !b.is_finished) || null
})

function formatAuthors(authors) {
  if (!authors || !authors.length) return 'Unknown Author'
  return authors.map(a => typeof a === 'string' ? a : a.name).filter(Boolean).join(', ')
}

function cleanDescription(desc) {
  if (!desc) return ''
  return desc.replace(/<[^>]*>?/gm, '').trim()
}

function handleReadBook(book) {
  emit('read', book)
  router.push(`/read/${book.id}`)
}

function openBookDetail(bookId) {
  router.push(`/books/${bookId}`)
}

function navigateToAuthor(authorName) {
  emit('filter-author', authorName)
  router.push(`/authors/${encodeURIComponent(authorName)}`)
}

function openShelfModal(book) {
  selectedBookForShelf.value = book
  showShelfModal.value = true
}

function onShelfUpdated() {
  emit('shelf-updated')
}

function goHome() {
  router.push('/books')
}

function goToSeriesCatalog() {
  router.push('/series')
}

watch(effectiveSeriesName, () => {
  loadSeriesBooks()
})

onMounted(() => {
  loadSeriesBooks()
})
</script>
