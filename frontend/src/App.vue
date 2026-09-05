<template>
  <div class="min-h-screen bg-[#090a0f] text-slate-100 selection:bg-glacier-500/20 selection:text-glacier-400">
    <!-- Initial Auth Loading State -->
    <div v-if="authLoading" class="min-h-screen flex flex-col items-center justify-center text-slate-500">
      <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
      <p class="text-xs tracking-wider uppercase">Loading Lyostar...</p>
    </div>

    <!-- First-Run Admin Setup Wizard -->
    <SetupView 
      v-else-if="setupRequired || route.name === 'setup'" 
      @success="onAuthSuccess" 
    />

    <!-- Login Screen -->
    <LoginView 
      v-else-if="!isAuthenticated || route.name === 'login'" 
      @success="onAuthSuccess" 
    />

    <!-- Main Application (Authenticated) -->
    <template v-else>
      <div class="min-h-screen flex bg-[#090a0f]">
        <!-- Sidebar Navigation Drawer (Calibre-Web Style) -->
        <Sidebar
          :is-open="sidebarOpen"
          :active-nav="activeNav"
          :selected-shelf="selectedShelf"
          :shelves="shelvesList"
          :total-books="totalBooks"
          :continue-count="continueBooks.length"
          :tags-count="tags.length"
          :is-scanning="isScanning"
          @close="sidebarOpen = false"
          @select-nav="handleSelectNav"
          @select-shelf="handleSelectShelf"
          @create-shelf="showShelvesModal = true"
          @manage-shelves="showShelvesModal = true"
          @open-upload="showUploadModal = true"
          @open-users="showUsersModal = true"
          @scan="handleScan"
        />

        <!-- Right Side: Content Area with desktop left padding (md:pl-64) -->
        <div class="flex-1 flex flex-col min-w-0 md:pl-64">
          <!-- Topbar / Navbar (Streamlined & Clean) -->
          <Navbar
            :is-scanning="isScanning"
            @toggle-sidebar="sidebarOpen = !sidebarOpen"
            @search="handleSearch"
            @reset="resetView"
            @open-upload="showUploadModal = true"
          />

          <!-- Dedicated Book Detail Page (Calibre-Web Style) -->
          <BookDetailView
            v-if="route.name === 'book-detail'"
            :book-id="route.params.id"
            @read="openReader"
            @filter-tag="handleSelectTag"
            @shelf-updated="onShelfUpdated"
            @deleted="onBookDeleted"
          />

          <!-- Main Shelf Content -->
          <main v-else class="flex-1 max-w-7xl w-full mx-auto px-4 lg:px-8 py-6 sm:py-8">
            <!-- Section Header -->
            <div class="flex items-center justify-between mb-6">
              <div>
                <div class="flex items-center gap-3">
                  <h1 class="text-xl sm:text-2xl font-bold text-white tracking-tight">
                    {{ 
                      isSearching 
                        ? `Search Results for "${searchQuery}"` 
                        : (selectedShelf 
                            ? `Shelf: ${selectedShelf.name}` 
                            : (activeNav === 'continue' 
                                ? 'Continue Reading' 
                                : (activeNav === 'tags' 
                                    ? (selectedTag ? `Category: #${selectedTag}` : 'Categories & Tags') 
                                    : (selectedTag ? `Category: #${selectedTag}` : 'All Books'))))
                    }}
                  </h1>

                  <!-- Active Shelf Clear Badge -->
                  <button
                    v-if="selectedShelf"
                    @click="clearShelfFilter"
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-glacier-500/15 text-glacier-300 border border-glacier-400/30 hover:bg-glacier-500/25 transition-colors cursor-pointer"
                    title="Clear shelf filter"
                  >
                    <Bookmark class="w-3 h-3 text-glacier-400" />
                    <span>{{ selectedShelf.name }}</span>
                    <X class="w-3 h-3 text-glacier-400" />
                  </button>

                  <!-- Active Tag Clear Badge -->
                  <button
                    v-else-if="selectedTag"
                    @click="clearTagFilter"
                    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-glacier-500/15 text-glacier-300 border border-glacier-400/30 hover:bg-glacier-500/25 transition-colors cursor-pointer"
                    title="Clear tag filter"
                  >
                    <span>#{{ selectedTag }}</span>
                    <X class="w-3 h-3 text-glacier-400" />
                  </button>
                </div>

                <p class="text-xs text-slate-400 mt-1">
                  {{ selectedShelf ? `${totalBooks} books in this shelf` : (activeNav === 'continue' ? `${continueBooks.length} books in progress` : `${totalBooks} books available`) }}
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

            <!-- Tag Filter Pills Bar (When in Tags view or filtering) -->
            <div v-if="(activeNav === 'tags' || (!isSearching && !selectedShelf && activeNav !== 'continue')) && tags.length > 0" class="mb-6 flex items-center gap-2 overflow-x-auto pb-1 scrollbar-thin">
              <button
                @click="clearTagFilter"
                class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-medium transition-all whitespace-nowrap cursor-pointer border"
                :class="!selectedTag 
                  ? 'bg-glacier-500 text-slate-950 font-semibold border-glacier-500 shadow-sm shadow-glacier-500/20' 
                  : 'bg-[#11131b] hover:bg-[#161923] text-slate-400 hover:text-slate-200 border-white/[0.08]'"
              >
                All Genres
              </button>
              <button
                v-for="t in tags"
                :key="t.id"
                @click="handleSelectTag(t.name === selectedTag ? '' : t.name)"
                class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl text-xs font-medium transition-all whitespace-nowrap cursor-pointer border"
                :class="selectedTag === t.name 
                  ? 'bg-glacier-500/20 text-glacier-300 border-glacier-400/40 shadow-sm' 
                  : 'bg-[#11131b] hover:bg-[#161923] text-slate-400 hover:text-slate-200 border-white/[0.08] hover:border-white/[0.15]'"
              >
                <Tag class="w-3 h-3 text-glacier-400/70" />
                <span>#{{ t.name }}</span>
                <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-white/[0.06] text-slate-400 font-mono">{{ t.book_count }}</span>
              </button>
            </div>

            <!-- Dedicated View: Continue Reading (When activeNav === 'continue') -->
            <div v-if="activeNav === 'continue'">
              <!-- Empty Continue Reading -->
              <div v-if="continueBooks.length === 0" class="py-24 flex flex-col items-center justify-center text-center px-4">
                <div class="w-16 h-16 rounded-2xl bg-[#11131b] border border-white/[0.08] flex items-center justify-center text-slate-500 mb-4">
                  <BookOpen class="w-8 h-8 text-glacier-400/50" />
                </div>
                <h3 class="text-base font-semibold text-white mb-1">No books in progress</h3>
                <p class="text-xs text-slate-400 max-w-sm mb-6">
                  You haven't started reading any books yet. Open any book in the library and start reading to track your progress here.
                </p>
                <button
                  @click="handleSelectNav('books')"
                  class="px-4 py-2 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-semibold shadow-md shadow-glacier-500/20 transition-all cursor-pointer"
                >
                  Browse Library
                </button>
              </div>

              <!-- Full Grid of in-progress books -->
              <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                <div
                  v-for="item in continueBooks"
                  :key="item.book_id"
                  @click="openReader({ id: item.book_id, title: item.title, format: item.format, file_url: `/api/books/${item.book_id}/file`, cover_url: item.cover_url, has_cover: !!item.cover_url, progress: item.progress })"
                  class="group bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/30 rounded-2xl p-4 flex items-center gap-4 cursor-pointer transition-all hover:shadow-xl hover:shadow-glacier-500/5 hover:-translate-y-1"
                >
                  <!-- Mini cover -->
                  <div class="w-16 h-22 rounded-xl bg-[#090a0f] border border-white/[0.06] overflow-hidden flex-shrink-0 relative">
                    <img
                      v-if="item.cover_url"
                      :src="item.cover_url"
                      :alt="item.title"
                      class="w-full h-full object-cover"
                    />
                    <div v-else class="w-full h-full flex items-center justify-center text-glacier-400/50">
                      <BookOpen class="w-6 h-6" />
                    </div>
                  </div>

                  <!-- Info & Progress Bar -->
                  <div class="flex-1 min-w-0">
                    <h4 class="text-sm font-semibold text-white group-hover:text-glacier-400 transition-colors truncate">
                      {{ item.title }}
                    </h4>
                    <p class="text-xs text-slate-400 truncate mb-3">
                      {{ item.authors?.join(', ') || 'Unknown Author' }}
                    </p>

                    <div class="flex items-center gap-2.5">
                      <div class="flex-1 bg-white/[0.08] h-1.5 rounded-full overflow-hidden">
                        <div 
                          class="bg-glacier-400 h-full rounded-full transition-all duration-300"
                          :style="{ width: `${Math.round(item.progress * 100)}%` }"
                        ></div>
                      </div>
                      <span class="font-mono text-xs text-glacier-400 font-semibold">
                        {{ Math.round(item.progress * 100) }}%
                      </span>
                    </div>
                  </div>

                  <!-- Resume Action Button -->
                  <div class="w-9 h-9 rounded-xl bg-glacier-500/10 group-hover:bg-glacier-500 text-glacier-400 group-hover:text-slate-950 flex items-center justify-center transition-all flex-shrink-0 shadow-sm">
                    <BookOpen class="w-4 h-4" />
                  </div>
                </div>
              </div>
            </div>

            <!-- Standard View: Books Shelf + Preview Sections -->
            <div v-else>
              <!-- Continue Reading Preview Section (When browsing all books & has in-progress items) -->
              <div v-if="!isSearching && !selectedTag && !selectedShelf && continueBooks.length > 0" class="mb-10 animate-fade-in">
                <div class="flex items-center justify-between mb-3.5">
                  <div class="flex items-center gap-2">
                    <span class="w-2 h-2 rounded-full bg-glacier-400"></span>
                    <h2 class="text-sm sm:text-base font-bold text-white tracking-tight">
                      Continue Reading
                    </h2>
                  </div>
                  <button 
                    @click="handleSelectNav('continue')"
                    class="text-xs font-medium text-glacier-400 hover:text-glacier-300 transition-colors cursor-pointer"
                  >
                    View all ({{ continueBooks.length }})
                  </button>
                </div>

                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
                  <div
                    v-for="item in continueBooks.slice(0, 3)"
                    :key="item.book_id"
                    @click="openReader({ id: item.book_id, title: item.title, format: item.format, file_url: `/api/books/${item.book_id}/file`, cover_url: item.cover_url, has_cover: !!item.cover_url, progress: item.progress })"
                    class="group bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/30 rounded-xl p-3 flex items-center gap-3.5 cursor-pointer transition-all hover:shadow-lg hover:shadow-glacier-500/5 hover:-translate-y-0.5"
                  >
                    <!-- Mini cover -->
                    <div class="w-12 h-16 rounded-lg bg-[#090a0f] border border-white/[0.06] overflow-hidden flex-shrink-0 relative">
                      <img
                        v-if="item.cover_url"
                        :src="item.cover_url"
                        :alt="item.title"
                        class="w-full h-full object-cover"
                      />
                      <div v-else class="w-full h-full flex items-center justify-center text-glacier-400/50">
                        <BookOpen class="w-5 h-5" />
                      </div>
                    </div>

                    <!-- Info & Progress Bar -->
                    <div class="flex-1 min-w-0">
                      <h4 class="text-xs font-semibold text-white group-hover:text-glacier-400 transition-colors truncate">
                        {{ item.title }}
                      </h4>
                      <p class="text-[11px] text-slate-400 truncate mb-2">
                        {{ item.authors?.join(', ') || 'Unknown Author' }}
                      </p>

                      <div class="flex items-center gap-2">
                        <div class="flex-1 bg-white/[0.08] h-1.5 rounded-full overflow-hidden">
                          <div 
                            class="bg-glacier-400 h-full rounded-full transition-all duration-300"
                            :style="{ width: `${Math.round(item.progress * 100)}%` }"
                          ></div>
                        </div>
                        <span class="font-mono text-[10px] text-glacier-400 font-semibold">
                          {{ Math.round(item.progress * 100) }}%
                        </span>
                      </div>
                    </div>

                    <!-- Resume Action Button -->
                    <div class="w-8 h-8 rounded-lg bg-glacier-500/10 group-hover:bg-glacier-500 text-glacier-400 group-hover:text-slate-950 flex items-center justify-center transition-colors flex-shrink-0">
                      <BookOpen class="w-4 h-4" />
                    </div>
                  </div>
                </div>
              </div>

              <!-- Loading State -->
              <div v-if="loading && books.length === 0" class="py-24 flex flex-col items-center justify-center text-slate-500">
                <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
                <p class="text-sm">Loading collection...</p>
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
                  {{ isSearching ? 'No books found matching your search' : (selectedShelf ? `No books found in shelf "${selectedShelf.name}"` : (selectedTag ? `No books found with tag #${selectedTag}` : 'No books in library yet')) }}
                </h3>
                <p class="text-xs text-slate-400 max-w-sm mb-6">
                  {{ isSearching ? 'Try adjusting your search terms or keywords.' : (selectedShelf ? 'Add books to this shelf using the "Add to Shelf" button on book cards or detail view.' : (selectedTag ? 'Try selecting another genre or clear the active filter.' : 'Add .epub or .pdf files into your books directory and click Rescan.')) }}
                </p>
                <button
                  v-if="selectedShelf"
                  @click="clearShelfFilter"
                  class="px-4 py-2 rounded-xl bg-white/[0.06] hover:bg-white/[0.1] border border-white/[0.1] text-white text-xs font-medium transition-all cursor-pointer"
                >
                  Back to All Books
                </button>
                <button
                  v-else-if="selectedTag"
                  @click="clearTagFilter"
                  class="px-4 py-2 rounded-xl bg-white/[0.06] hover:bg-white/[0.1] border border-white/[0.1] text-white text-xs font-medium transition-all cursor-pointer"
                >
                  Clear Genre Filter
                </button>
                <button
                  v-else-if="!isSearching && isAdmin"
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
                  @select="openBookDetail(book)"
                  @read="openReader(book)"
                  @filter-tag="handleSelectTag"
                  @open-shelf="b => shelfSelectBook = b"
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
            </div>
          </main>
        </div>
      </div>


      <!-- Web Reader View -->
      <ReaderView
        v-if="readingBook"
        :book="readingBook"
        @close="closeReader"
        @progress-updated="onProgressUpdated"
      />

      <!-- Admin User Management Modal -->
      <UsersModal
        v-if="showUsersModal"
        @close="closeUsersModal"
      />

      <!-- Upload Books Modal -->
      <UploadModal
        v-if="showUploadModal"
        @close="closeUploadModal"
        @uploaded="onBooksUploaded"
      />

      <!-- Shelves Management Modal -->
      <ShelvesManageModal
        v-if="showShelvesModal"
        @close="showShelvesModal = false"
        @select-shelf="handleSelectShelf"
      />

      <!-- Quick Shelf Select Modal -->
      <ShelfSelectModal
        v-if="shelfSelectBook"
        :book="shelfSelectBook"
        @close="shelfSelectBook = null"
        @updated="onShelfUpdated"
      />
    </template>

    <!-- Global Floating Toast / Snackbar Notification -->
    <Transition
      enter-active-class="transition duration-300 ease-out"
      enter-from-class="transform translate-y-4 opacity-0 scale-95"
      enter-to-class="transform translate-y-0 opacity-100 scale-100"
      leave-active-class="transition duration-200 ease-in"
      leave-from-class="transform translate-y-0 opacity-100 scale-100"
      leave-to-class="transform translate-y-4 opacity-0 scale-95"
    >
      <div 
        v-if="toast.visible" 
        class="fixed bottom-6 right-6 z-[9999] flex items-center gap-3 px-4 py-3 rounded-2xl bg-[#11131b]/95 backdrop-blur-md border border-white/[0.12] shadow-2xl shadow-black/80 text-white text-xs font-medium"
      >
        <div 
          class="w-6 h-6 rounded-lg flex items-center justify-center flex-shrink-0"
          :class="toast.type === 'error' ? 'bg-rose-500/20 text-rose-400' : 'bg-emerald-500/20 text-emerald-400'"
        >
          <AlertCircle v-if="toast.type === 'error'" class="w-3.5 h-3.5" />
          <Check v-else class="w-3.5 h-3.5 stroke-[2.5]" />
        </div>
        <span class="max-w-xs truncate">{{ toast.message }}</span>
        <button 
          @click="hideToast" 
          class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/[0.05] transition-colors cursor-pointer"
        >
          <X class="w-3.5 h-3.5" />
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Loader2, BookX, BookOpen, Tag, X, Bookmark, Check, AlertCircle } from 'lucide-vue-next'
import Sidebar from './components/Sidebar.vue'
import Navbar from './components/Navbar.vue'
import BookCard from './components/BookCard.vue'
import BookDetailView from './components/BookDetailView.vue'
import ReaderView from './components/ReaderView.vue'
import SetupView from './components/SetupView.vue'
import LoginView from './components/LoginView.vue'
import UsersModal from './components/UsersModal.vue'
import UploadModal from './components/UploadModal.vue'
import ShelvesManageModal from './components/ShelvesManageModal.vue'
import ShelfSelectModal from './components/ShelfSelectModal.vue'
import { 
  fetchBooks, 
  searchBooks, 
  triggerScan, 
  fetchContinueReading, 
  fetchTags, 
  fetchShelfBooks, 
  fetchShelves, 
  fetchBookDetail 
} from './api/client.js'
import { useAuth } from './composables/useAuth'
import { useToast } from './composables/useToast'

const router = useRouter()
const route = useRoute()
const { isAuthenticated, isAdmin, setupRequired, loading: authLoading, checkAuth, canRead } = useAuth()
const { toast, hideToast } = useToast()

const sidebarOpen = ref(false)
const activeNav = ref('books')
const shelvesList = ref([])

const books = ref([])
const continueBooks = ref([])
const tags = ref([])
const selectedTag = ref('')
const selectedShelf = ref(null)
const totalBooks = ref(0)
const currentPage = ref(1)
const pageSize = 24
const loading = ref(false)
const isScanning = ref(false)
const scanMessage = ref('')
const searchQuery = ref('')
const isSearching = computed(() => searchQuery.value.trim() !== '')

const readingBook = ref(null)
const showUsersModal = ref(false)
const showUploadModal = ref(false)
const showShelvesModal = ref(false)
const shelfSelectBook = ref(null)

const hasMore = computed(() => books.value.length < totalBooks.value)

async function loadShelves() {
  if (!isAuthenticated.value) return
  try {
    const res = await fetchShelves()
    shelvesList.value = res || []
  } catch (err) {
    console.warn('Failed to load shelves:', err)
  }
}

async function loadContinueReading() {
  if (!isAuthenticated.value) return
  try {
    const items = await fetchContinueReading(12)
    continueBooks.value = items || []
  } catch (err) {
    console.warn('Failed to load continue reading:', err)
  }
}

async function loadTags() {
  if (!isAuthenticated.value) return
  try {
    const items = await fetchTags()
    tags.value = items || []
  } catch (err) {
    console.warn('Failed to load tags:', err)
  }
}

async function loadData(page = 1, append = false) {
  if (!isAuthenticated.value) return
  loading.value = true
  try {
    let res
    if (isSearching.value) {
      res = await searchBooks({ q: searchQuery.value, page, limit: pageSize })
    } else if (selectedShelf.value) {
      res = await fetchShelfBooks(selectedShelf.value.id, { page, limit: pageSize })
    } else {
      res = await fetchBooks({ page, limit: pageSize, tag: selectedTag.value })
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

async function syncFromRoute(r = route) {
  if (!isAuthenticated.value) return

  const name = r.name
  const params = r.params
  const query = r.query

  // 1. Modals
  showUsersModal.value = (name === 'users')
  showUploadModal.value = (name === 'upload')

  // 2. Reader view (/read/:id)
  if (name === 'reader') {
    const bookId = Number(params.id)
    if (!readingBook.value || readingBook.value.id !== bookId) {
      try {
        const b = await fetchBookDetail(bookId)
        readingBook.value = b
      } catch (err) {
        console.error('Failed to load book for reader:', err)
        router.replace('/books')
        return
      }
    }
    return
  } else {
    readingBook.value = null
  }

  // 3. Book detail page (/books/:id) handled by BookDetailView component
  if (name === 'book-detail') {
    return
  }

  // 4. Content navigation
  if (name === 'continue-reading') {
    activeNav.value = 'continue'
    selectedShelf.value = null
    selectedTag.value = ''
    searchQuery.value = ''
    await loadContinueReading()
  } else if (name === 'tags') {
    activeNav.value = 'tags'
    selectedShelf.value = null
    selectedTag.value = ''
    searchQuery.value = ''
    await loadTags()
    await loadData(1, false)
  } else if (name === 'tag-filter') {
    activeNav.value = 'tags'
    selectedShelf.value = null
    selectedTag.value = decodeURIComponent(params.tag || '')
    searchQuery.value = ''
    await loadData(1, false)
  } else if (name === 'shelf-books') {
    activeNav.value = 'shelf'
    const shelfId = Number(params.id)
    if (shelvesList.value.length === 0) {
      await loadShelves()
    }
    selectedShelf.value = shelvesList.value.find(s => s.id === shelfId) || { id: shelfId, name: 'Custom Shelf' }
    selectedTag.value = ''
    searchQuery.value = ''
    await loadData(1, false)
  } else if (name === 'search') {
    activeNav.value = 'books'
    selectedShelf.value = null
    selectedTag.value = ''
    searchQuery.value = query.q || ''
    await loadData(1, false)
  } else {
    // 'books', 'users', 'upload'
    activeNav.value = 'books'
    selectedShelf.value = null
    selectedTag.value = query.tag || ''
    searchQuery.value = query.q || ''
    const page = Number(query.page) || 1
    await loadData(page, false)
  }
}

watch(() => route.fullPath, () => {
  syncFromRoute(route)
})

function handleSelectNav(nav) {
  sidebarOpen.value = false
  if (nav === 'books') {
    router.push('/books')
  } else if (nav === 'continue') {
    router.push('/continue-reading')
  } else if (nav === 'tags') {
    router.push('/tags')
  }
}

function handleSelectShelf(shelf) {
  sidebarOpen.value = false
  showShelvesModal.value = false
  router.push(`/shelves/${shelf.id}`)
}

function clearShelfFilter() {
  router.push('/books')
}

function onShelfUpdated() {
  loadShelves()
  if (selectedShelf.value) {
    loadData(1, false)
  }
}

function handleSelectTag(tag) {
  if (tag) {
    router.push(`/tags/${encodeURIComponent(tag)}`)
  } else {
    router.push('/books')
  }
}

function clearTagFilter() {
  router.push('/books')
}

function handleSearch(query) {
  if (query) {
    router.push({ path: '/search', query: { q: query } })
  } else if (route.name === 'search') {
    router.push('/books')
  }
}

function resetView() {
  router.push('/books')
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
      await loadTags()
      await loadShelves()
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

function openBookDetail(book) {
  router.push(`/books/${book.id}`)
}

function openReader(book) {
  if (!canRead.value) return
  readingBook.value = book
  router.push(`/read/${book.id}`)
}

function closeReader() {
  readingBook.value = null
  loadContinueReading()
  if (route.name === 'reader') {
    if (window.history.length > 1) {
      router.back()
    } else {
      router.push('/books')
    }
  }
}

function closeUsersModal() {
  showUsersModal.value = false
  if (route.name === 'users') {
    if (window.history.length > 1) {
      router.back()
    } else {
      router.push('/books')
    }
  }
}

function closeUploadModal() {
  showUploadModal.value = false
  if (route.name === 'upload') {
    if (window.history.length > 1) {
      router.back()
    } else {
      router.push('/books')
    }
  }
}

function onProgressUpdated({ bookId, progress, isFinished }) {
  const b = books.value.find(x => x.id === bookId)
  if (b) {
    b.progress = progress
    b.is_finished = isFinished
  }
  loadContinueReading()
}

function onAuthSuccess() {
  loadShelves()
  loadTags()
  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/')) {
    router.push(redirect)
  } else {
    router.push('/books')
  }
}

function onBooksUploaded() {
  loadData(1, false)
  loadTags()
  loadShelves()
}

function onBookUpdated(updatedBook) {
  const idx = books.value.findIndex(b => b.id === updatedBook.id)
  if (idx !== -1) {
    const prev = books.value[idx]
    books.value[idx] = {
      ...prev,
      ...updatedBook,
      authors: updatedBook.authors?.map(a => typeof a === 'string' ? a : a.name) || []
    }
  }

  const cIdx = continueBooks.value.findIndex(b => b.book_id === updatedBook.id)
  if (cIdx !== -1) {
    continueBooks.value[cIdx].title = updatedBook.title
    continueBooks.value[cIdx].authors = updatedBook.authors?.map(a => typeof a === 'string' ? a : a.name) || []
  }

  loadTags()
  loadShelves()
}

function onBookDeleted(bookId) {
  books.value = books.value.filter(b => b.id !== bookId)
  continueBooks.value = continueBooks.value.filter(b => b.book_id !== bookId)
  if (totalBooks.value > 0) {
    totalBooks.value--
  }
  if (route.name === 'book-detail') {
    router.replace('/books')
  }
  loadTags()
  loadShelves()
}

watch(isAuthenticated, async (newVal) => {
  if (newVal) {
    await loadShelves()
    await loadTags()
    await loadContinueReading()
    await syncFromRoute(route)
  } else {
    sidebarOpen.value = false
    activeNav.value = 'books'
    shelvesList.value = []
    books.value = []
    continueBooks.value = []
    tags.value = []
    selectedTag.value = ''
    selectedShelf.value = null
    totalBooks.value = 0
    readingBook.value = null
    showUsersModal.value = false
    showUploadModal.value = false
    showShelvesModal.value = false
    shelfSelectBook.value = null
  }
})

onMounted(async () => {
  await checkAuth()
  if (isAuthenticated.value) {
    await loadShelves()
    await loadTags()
    await loadContinueReading()
    await syncFromRoute(route)
  }
})
</script>
