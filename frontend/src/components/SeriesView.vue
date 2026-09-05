<template>
  <div class="flex-1 w-full max-w-7xl mx-auto px-4 lg:px-8 py-6 sm:py-8 animate-fade-in">
    <!-- Top Breadcrumbs & Back -->
    <div class="flex flex-wrap items-center justify-between gap-3 mb-6 pb-4 border-b border-white/[0.08]">
      <nav class="flex items-center gap-2 text-xs text-slate-400">
        <button
          @click="goBack"
          class="flex items-center gap-1.5 text-slate-400 hover:text-white transition-colors cursor-pointer group"
        >
          <ArrowLeft class="w-4 h-4 text-slate-500 group-hover:text-glacier-400 transition-colors" />
          <span>Library</span>
        </button>
        <ChevronRight class="w-3.5 h-3.5 text-slate-600 flex-shrink-0" />
        <span class="text-white font-medium">Series</span>
      </nav>

      <button
        @click="goBack"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-xs font-medium text-slate-300 hover:text-white transition-all cursor-pointer"
      >
        <BookOpen class="w-3.5 h-3.5 text-glacier-400" />
        <span>All Books</span>
      </button>
    </div>

    <!-- Header & Search Controls -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
      <div>
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-2xl bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400 shadow-sm shadow-glacier-500/10">
            <Layers class="w-5 h-5" />
          </div>
          <div>
            <h1 class="text-xl sm:text-2xl font-bold text-white tracking-tight">
              Book Series
            </h1>
            <p class="text-xs text-slate-400 mt-0.5">
              Browse all {{ seriesList.length }} book series and multi-volume sagas in your library
            </p>
          </div>
        </div>
      </div>

      <!-- Quick Search Bar -->
      <div class="relative w-full sm:w-72">
        <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter series by name..."
          class="w-full pl-9 pr-8 py-2 rounded-xl bg-[#11131b] border border-white/[0.08] focus:border-glacier-400/40 text-xs text-white placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-glacier-400/30 transition-all"
        />
        <button
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-500 hover:text-white p-0.5 rounded-md transition-colors"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-24 flex flex-col items-center justify-center text-slate-500">
      <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
      <p class="text-xs uppercase tracking-wider">Loading series collection...</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredSeries.length === 0" class="py-20 flex flex-col items-center justify-center text-center px-4">
      <div class="w-14 h-14 rounded-2xl bg-white/[0.04] border border-white/[0.08] text-slate-500 flex items-center justify-center mb-3">
        <Layers class="w-7 h-7 text-slate-500" />
      </div>
      <h3 class="text-sm font-semibold text-white mb-1">
        {{ searchQuery ? 'No matching series' : 'No series found' }}
      </h3>
      <p class="text-xs text-slate-400 max-w-sm mb-4">
        {{ searchQuery ? 'Try adjusting your search terms.' : 'Books with series information in their metadata will automatically appear here.' }}
      </p>
      <button
        v-if="searchQuery"
        @click="searchQuery = ''"
        class="px-4 py-2 rounded-xl bg-white/[0.06] hover:bg-white/[0.1] border border-white/[0.08] text-xs font-medium text-white transition-all cursor-pointer"
      >
        Clear Filter
      </button>
    </div>

    <!-- Series Grid -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3.5 sm:gap-4">
      <div
        v-for="s in filteredSeries"
        :key="s.name"
        @click="selectSeries(s.name)"
        class="group p-4 rounded-2xl bg-[#11131b] hover:bg-[#161923] border border-white/[0.06] hover:border-glacier-400/30 transition-all duration-200 cursor-pointer flex items-center gap-3.5 hover:shadow-lg hover:shadow-glacier-500/5 hover:-translate-y-0.5"
      >
        <!-- Icon Avatar -->
        <div class="w-11 h-11 rounded-xl bg-gradient-to-br from-glacier-500/15 to-glacier-600/5 border border-glacier-400/20 text-glacier-400 flex items-center justify-center flex-shrink-0 group-hover:scale-105 group-hover:border-glacier-400/40 group-hover:text-glacier-300 transition-all shadow-sm">
          <Layers class="w-5 h-5" />
        </div>

        <!-- Info -->
        <div class="flex-1 min-w-0">
          <h3 class="text-sm font-semibold text-white group-hover:text-glacier-400 transition-colors truncate">
            {{ s.name }}
          </h3>
          <p class="text-xs text-slate-400 flex items-center gap-1.5 mt-0.5 font-mono">
            <Book class="w-3 h-3 text-slate-500" />
            <span>{{ s.book_count }} {{ s.book_count === 1 ? 'volume' : 'volumes' }}</span>
          </p>
        </div>

        <!-- Arrow Action Icon -->
        <div class="w-7 h-7 rounded-lg bg-white/[0.03] group-hover:bg-glacier-400/10 text-slate-500 group-hover:text-glacier-400 flex items-center justify-center transition-all flex-shrink-0">
          <ChevronRight class="w-4 h-4" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Layers, 
  Search, 
  X, 
  Book, 
  BookOpen, 
  ChevronRight, 
  ArrowLeft, 
  Loader2 
} from 'lucide-vue-next'
import { fetchSeries } from '../api/client'

const props = defineProps({
  initialSeries: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['select-series'])

const router = useRouter()
const seriesList = ref([])
const loading = ref(false)
const searchQuery = ref('')

async function loadSeries() {
  if (props.initialSeries && props.initialSeries.length > 0) {
    seriesList.value = props.initialSeries
    return
  }
  loading.value = true
  try {
    const data = await fetchSeries()
    seriesList.value = data || []
  } catch (err) {
    console.error('Failed to load series catalog:', err)
  } finally {
    loading.value = false
  }
}

const filteredSeries = computed(() => {
  if (!searchQuery.value.trim()) {
    return seriesList.value
  }
  const q = searchQuery.value.toLowerCase().trim()
  return seriesList.value.filter(s => s.name.toLowerCase().includes(q))
})

function selectSeries(seriesName) {
  emit('select-series', seriesName)
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/books')
  }
}

onMounted(() => {
  loadSeries()
})
</script>
