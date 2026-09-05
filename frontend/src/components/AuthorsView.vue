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
        <span class="text-white font-medium">Authors</span>
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
            <Users class="w-5 h-5" />
          </div>
          <div>
            <h1 class="text-xl sm:text-2xl font-bold text-white tracking-tight">
              Authors Catalog
            </h1>
            <p class="text-xs text-slate-400 mt-0.5">
              Browse all {{ authors.length }} authors and their works in your library
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
          placeholder="Filter authors by name..."
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

    <!-- Alphabet Quick-Jump Bar -->
    <div class="mb-8 p-2 rounded-2xl bg-[#11131b]/60 border border-white/[0.06] overflow-x-auto scrollbar-thin">
      <div class="flex items-center justify-between min-w-max gap-1 px-1">
        <button
          @click="selectedLetter = ''"
          class="px-2.5 py-1 rounded-lg text-xs font-semibold transition-all cursor-pointer"
          :class="!selectedLetter
            ? 'bg-glacier-500 text-slate-950 shadow-sm shadow-glacier-500/20'
            : 'text-slate-400 hover:text-white hover:bg-white/[0.06]'"
        >
          All ({{ authors.length }})
        </button>

        <button
          v-for="letter in alphabet"
          :key="letter"
          @click="onLetterClick(letter)"
          :disabled="!hasLetter(letter)"
          class="w-7 h-7 rounded-lg text-xs font-semibold flex items-center justify-center transition-all cursor-pointer"
          :class="[
            selectedLetter === letter
              ? 'bg-glacier-500 text-slate-950 shadow-sm shadow-glacier-500/20 font-bold'
              : hasLetter(letter)
                ? 'text-slate-300 hover:text-white hover:bg-white/[0.08]'
                : 'text-slate-600/40 cursor-not-allowed'
          ]"
        >
          {{ letter }}
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-24 flex flex-col items-center justify-center text-slate-500">
      <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
      <p class="text-xs uppercase tracking-wider">Loading authors...</p>
    </div>

    <!-- Empty State (No authors or filter mismatch) -->
    <div v-else-if="filteredAuthors.length === 0" class="py-20 flex flex-col items-center justify-center text-center px-4">
      <div class="w-14 h-14 rounded-2xl bg-white/[0.04] border border-white/[0.08] text-slate-500 flex items-center justify-center mb-3">
        <Users class="w-7 h-7 text-slate-500" />
      </div>
      <h3 class="text-sm font-semibold text-white mb-1">
        {{ searchQuery || selectedLetter ? 'No matching authors' : 'No authors found' }}
      </h3>
      <p class="text-xs text-slate-400 max-w-sm mb-4">
        {{ searchQuery || selectedLetter ? 'Try clearing your search or alphabet filter to see all authors.' : 'Upload books with author metadata or rescan your library to populate authors.' }}
      </p>
      <button
        v-if="searchQuery || selectedLetter"
        @click="resetFilters"
        class="px-4 py-2 rounded-xl bg-white/[0.06] hover:bg-white/[0.1] border border-white/[0.08] text-xs font-medium text-white transition-all cursor-pointer"
      >
        Clear Filters
      </button>
    </div>

    <!-- Authors Grid -->
    <div v-else class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3.5 sm:gap-4">
      <div
        v-for="author in filteredAuthors"
        :key="author.id"
        @click="selectAuthor(author.name)"
        class="group p-3.5 rounded-2xl bg-[#11131b] hover:bg-[#161923] border border-white/[0.06] hover:border-glacier-400/30 transition-all duration-200 cursor-pointer flex items-center gap-3.5 hover:shadow-lg hover:shadow-glacier-500/5 hover:-translate-y-0.5"
      >
        <!-- Author Avatar Initial -->
        <div class="w-11 h-11 rounded-xl bg-gradient-to-br from-glacier-500/15 to-glacier-600/5 border border-glacier-400/20 text-glacier-400 font-bold text-base flex items-center justify-center flex-shrink-0 group-hover:scale-105 group-hover:border-glacier-400/40 group-hover:text-glacier-300 transition-all shadow-sm">
          {{ getInitial(author.name) }}
        </div>

        <!-- Info -->
        <div class="flex-1 min-w-0">
          <h3 class="text-sm font-semibold text-white group-hover:text-glacier-400 transition-colors truncate">
            {{ author.name }}
          </h3>
          <p class="text-xs text-slate-400 flex items-center gap-1.5 mt-0.5 font-mono">
            <Book class="w-3 h-3 text-slate-500" />
            <span>{{ author.book_count }} {{ author.book_count === 1 ? 'book' : 'books' }}</span>
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
  Users, 
  Search, 
  X, 
  Book, 
  BookOpen, 
  ChevronRight, 
  ArrowLeft, 
  Loader2 
} from 'lucide-vue-next'
import { fetchAuthors } from '../api/client'

const props = defineProps({
  initialAuthors: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['select-author'])

const router = useRouter()
const authors = ref([])
const loading = ref(false)
const searchQuery = ref('')
const selectedLetter = ref('')

const alphabet = [
  '#', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
  'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'
]

async function loadAuthors() {
  if (props.initialAuthors && props.initialAuthors.length > 0) {
    authors.value = props.initialAuthors
    return
  }
  loading.value = true
  try {
    const data = await fetchAuthors()
    authors.value = data || []
  } catch (err) {
    console.error('Failed to load authors catalog:', err)
  } finally {
    loading.value = false
  }
}

function getInitial(name) {
  if (!name) return '?'
  const trimmed = name.trim()
  return trimmed.charAt(0).toUpperCase()
}

function hasLetter(letter) {
  if (letter === '#') {
    return authors.value.some(a => {
      const first = getInitial(a.name)
      return first < 'A' || first > 'Z'
    })
  }
  return authors.value.some(a => getInitial(a.name) === letter)
}

function onLetterClick(letter) {
  if (selectedLetter.value === letter) {
    selectedLetter.value = ''
  } else {
    selectedLetter.value = letter
  }
}

function resetFilters() {
  searchQuery.value = ''
  selectedLetter.value = ''
}

const filteredAuthors = computed(() => {
  let list = authors.value

  // 1. Search Query Filter
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim()
    list = list.filter(a => a.name.toLowerCase().includes(q))
  }

  // 2. Alphabet Filter
  if (selectedLetter.value) {
    if (selectedLetter.value === '#') {
      list = list.filter(a => {
        const first = getInitial(a.name)
        return first < 'A' || first > 'Z'
      })
    } else {
      list = list.filter(a => getInitial(a.name) === selectedLetter.value)
    }
  }

  return list
})

function selectAuthor(authorName) {
  emit('select-author', authorName)
  router.push(`/authors/${encodeURIComponent(authorName)}`)
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/books')
  }
}

onMounted(() => {
  loadAuthors()
})
</script>
