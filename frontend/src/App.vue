<template>
  <div class="min-h-screen bg-[#090a0f] text-slate-100 flex flex-col selection:bg-glacier-500/20 selection:text-glacier-400">
    <!-- Initial Auth Loading State -->
    <div v-if="authLoading" class="min-h-screen flex flex-col items-center justify-center text-slate-500">
      <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
      <p class="text-xs tracking-wider uppercase">Loading Lyostar...</p>
    </div>

    <!-- First-Run Admin Setup Wizard -->
    <SetupView 
      v-else-if="setupRequired" 
      @success="onAuthSuccess" 
    />

    <!-- Login Screen -->
    <LoginView 
      v-else-if="!isAuthenticated" 
      @success="onAuthSuccess" 
    />

    <!-- Main Application (Authenticated) -->
    <template v-else>
      <!-- Navbar -->
      <Navbar
        :is-scanning="isScanning"
        @search="handleSearch"
        @scan="handleScan"
        @reset="resetView"
        @open-users="showUsersModal = true"
      />

      <!-- Main Shelf Content -->
      <main class="flex-1 max-w-7xl w-full mx-auto px-4 lg:px-8 py-8">
        <!-- Section Header -->
        <div class="flex items-center justify-between mb-6">
          <div>
            <h1 class="text-xl sm:text-2xl font-bold text-white tracking-tight">
              {{ isSearching ? `Search Results for "${searchQuery}"` : 'Library Shelf' }}
            </h1>
            <p class="text-xs text-slate-400 mt-1">
              {{ totalBooks }} {{ totalBooks === 1 ? 'book' : 'books' }} available
            </p>
          </div>

          <!-- Scan message banner if active -->
          <div 
            v-if="scanMessage" 
            class="text-xs px-3 py-1.5 rounded-xl bg-glacier-500/10 border border-glacier-500/20 text-glacier-400 flex items-center gap-2 animate-fade-in"
          >
            <span class="w-2 h-2 rounded-full bg-glacier-400 animate-pulse"></span>
            <span>{{ scanMessage }}</span>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="loading && books.length === 0" class="py-24 flex flex-col items-center justify-center text-slate-500">
          <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
          <p class="text-sm">Loading your collection...</p>
        </div>

        <!-- Empty State -->
        <div 
          v-else-if="!loading && books.length === 0" 
          class="py-24 flex flex-col items-center justify-center text-center px-4"
        >
          <div class="w-16 h-16 rounded-2xl bg-[#11131b] border border-white/[0.08] flex items-center justify-center text-slate-500 mb-4">
            <BookX class="w-8 h-8" />
          </div>
          <h3 class="text-base font-semibold text-white mb-1">
            {{ isSearching ? 'No books found matching your search' : 'No books in library yet' }}
          </h3>
          <p class="text-xs text-slate-400 max-w-sm mb-6">
            {{ isSearching ? 'Try adjusting your search terms or keywords.' : 'Add .epub or .pdf files into your books directory and click Rescan.' }}
          </p>
          <button
            v-if="!isSearching && isAdmin"
            @click="handleScan"
            :disabled="isScanning"
            class="px-4 py-2 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-semibold shadow-md shadow-glacier-500/20 transition-all cursor-pointer"
          >
            Scan Books Now
          </button>
        </div>

        <!-- Books Grid -->
        <div 
          v-else 
          class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4 sm:gap-6"
        >
          <BookCard
            v-for="book in books"
            :key="book.id"
            :book="book"
            @select="selectedBook = book"
            @read="openReader(book)"
          />
        </div>

        <!-- Load More / Pagination -->
        <div v-if="hasMore" class="mt-12 flex justify-center">
          <button
            @click="loadMore"
            :disabled="loading"
            class="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/30 text-xs font-medium text-slate-300 hover:text-white transition-all disabled:opacity-50 cursor-pointer"
          >
            <Loader2 v-if="loading" class="w-4 h-4 animate-spin text-glacier-400" />
            <span>{{ loading ? 'Loading...' : 'Load More Books' }}</span>
          </button>
        </div>
      </main>

      <!-- Book Details Modal -->
      <BookDetailModal
        v-if="selectedBook"
        :book="selectedBook"
        @close="selectedBook = null"
        @read="openReader"
      />

      <!-- Foliate Web Reader View -->
      <ReaderView
        v-if="readingBook"
        :book="readingBook"
        @close="readingBook = null"
      />

      <!-- Admin User Management Modal -->
      <UsersModal
        v-if="showUsersModal"
        @close="showUsersModal = false"
      />
    </template>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { Loader2, BookX } from 'lucide-vue-next'
import Navbar from './components/Navbar.vue'
import BookCard from './components/BookCard.vue'
import BookDetailModal from './components/BookDetailModal.vue'
import ReaderView from './components/ReaderView.vue'
import SetupView from './components/SetupView.vue'
import LoginView from './components/LoginView.vue'
import UsersModal from './components/UsersModal.vue'
import { fetchBooks, searchBooks, triggerScan } from './api/client.js'
import { useAuth } from './composables/useAuth'

const { isAuthenticated, isAdmin, setupRequired, loading: authLoading, checkAuth } = useAuth()

const books = ref([])
const totalBooks = ref(0)
const currentPage = ref(1)
const pageSize = 24
const loading = ref(false)
const isScanning = ref(false)
const scanMessage = ref('')
const searchQuery = ref('')
const isSearching = computed(() => searchQuery.value.trim() !== '')

const selectedBook = ref(null)
const readingBook = ref(null)
const showUsersModal = ref(false)

const hasMore = computed(() => books.value.length < totalBooks.value)

async function loadData(page = 1, append = false) {
  if (!isAuthenticated.value) return
  loading.value = true
  try {
    let res
    if (isSearching.value) {
      res = await searchBooks({ q: searchQuery.value, page, limit: pageSize })
    } else {
      res = await fetchBooks({ page, limit: pageSize })
    }

    if (append) {
      books.value = [...books.value, ...(res.items || [])]
    } else {
      books.value = res.items || []
    }
    totalBooks.value = res.total || 0
    currentPage.value = page
  } catch (err) {
    console.error('Failed to load books:', err)
  } finally {
    loading.value = false
  }
}

function handleSearch(query) {
  searchQuery.value = query
  currentPage.value = 1
  loadData(1, false)
}

function resetView() {
  searchQuery.value = ''
  selectedBook.value = null
  readingBook.value = null
  loadData(1, false)
}

function loadMore() {
  if (hasMore.value && !loading.value) {
    loadData(currentPage.value + 1, true)
  }
}

async function handleScan() {
  if (isScanning.value) return
  isScanning.value = true
  scanMessage.value = 'Scanning books in background...'

  try {
    await triggerScan()
    // Poll for changes after 3 seconds
    setTimeout(async () => {
      await loadData(1, false)
      isScanning.value = false
      scanMessage.value = 'Library updated!'
      setTimeout(() => {
        scanMessage.value = ''
      }, 3000)
    }, 3000)
  } catch (err) {
    console.error('Scan error:', err)
    isScanning.value = false
    scanMessage.value = 'Scan failed'
  }
}

function openReader(book) {
  selectedBook.value = null
  readingBook.value = book
}

function onAuthSuccess() {
  loadData(1, false)
}

watch(isAuthenticated, (newVal) => {
  if (newVal) {
    loadData(1, false)
  } else {
    books.value = []
    totalBooks.value = 0
    selectedBook.value = null
    readingBook.value = null
    showUsersModal.value = false
  }
})

onMounted(async () => {
  await checkAuth()
  if (isAuthenticated.value) {
    loadData(1, false)
  }
})
</script>
