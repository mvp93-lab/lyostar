<template>
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in"
    @click.self="$emit('close')"
  >
    <div 
      class="relative w-full max-w-md bg-[#11131b] border border-white/[0.1] rounded-3xl p-6 shadow-2xl shadow-black/80 flex flex-col max-h-[85vh]"
    >
      <!-- Close Button -->
      <button 
        @click="$emit('close')"
        class="absolute top-4 right-4 p-2 rounded-full text-slate-400 hover:text-white hover:bg-white/[0.06] transition-colors"
        title="Close"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Header: Book Info -->
      <div class="flex items-center gap-3 pb-4 border-b border-white/[0.08] pr-8">
        <div class="w-10 h-14 rounded-lg bg-[#090a0f] border border-white/[0.08] overflow-hidden flex-shrink-0 flex items-center justify-center">
          <img
            v-if="book.has_cover && book.cover_url"
            :src="book.cover_url"
            :alt="book.title"
            class="w-full h-full object-cover"
          />
          <Book v-else class="w-5 h-5 text-glacier-400/60" />
        </div>
        <div class="min-w-0 flex-1">
          <h3 class="text-sm font-bold text-white truncate">{{ book.title }}</h3>
          <p class="text-xs text-slate-400 truncate mt-0.5">
            {{ formatAuthors(book.authors) }}
          </p>
        </div>
      </div>

      <!-- Title / Description -->
      <div class="mt-4 mb-3 flex items-center justify-between">
        <h4 class="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center gap-1.5">
          <Bookmark class="w-3.5 h-3.5 text-glacier-400" />
          Add to Shelves
        </h4>
        <span v-if="loading" class="text-[11px] text-slate-500 flex items-center gap-1">
          <Loader2 class="w-3 h-3 animate-spin text-glacier-400" />
          Loading...
        </span>
      </div>

      <!-- Error message -->
      <div v-if="error" class="mb-3 p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300">
        {{ error }}
      </div>

      <!-- Shelves List -->
      <div class="flex-1 overflow-y-auto space-y-2 pr-1 min-h-[120px]">
        <div v-if="loading && shelves.length === 0" class="py-8 text-center text-slate-500 text-xs">
          <Loader2 class="w-6 h-6 animate-spin text-glacier-400 mx-auto mb-2" />
          Loading your shelves...
        </div>

        <div v-else-if="shelves.length === 0" class="py-8 text-center text-slate-500 text-xs">
          You don't have any custom shelves yet.<br />
          Create your first shelf below!
        </div>

        <div
          v-for="shelf in shelves"
          :key="shelf.id"
          @click="toggleShelf(shelf.id)"
          class="flex items-center justify-between p-3 rounded-2xl border transition-all cursor-pointer select-none"
          :class="isBookInShelf(shelf.id)
            ? 'bg-glacier-500/15 border-glacier-400/40 text-white'
            : 'bg-white/[0.03] hover:bg-white/[0.06] border-white/[0.06] text-slate-300'"
        >
          <div class="flex items-center gap-3 min-w-0 pr-2">
            <!-- Checkbox UI -->
            <div 
              class="w-5 h-5 rounded-lg flex items-center justify-center transition-colors flex-shrink-0 border"
              :class="isBookInShelf(shelf.id)
                ? 'bg-glacier-500 border-glacier-400 text-slate-950 shadow-sm shadow-glacier-500/30'
                : 'border-white/[0.2] bg-slate-900/60'"
            >
              <Check v-if="isBookInShelf(shelf.id)" class="w-3.5 h-3.5 stroke-[3]" />
            </div>

            <!-- Shelf Details -->
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <span class="text-xs font-semibold truncate">{{ shelf.name }}</span>
                <span 
                  v-if="shelf.is_public" 
                  class="text-[10px] px-1.5 py-0.2 rounded-md bg-white/[0.08] text-slate-400 border border-white/[0.06]"
                  title="Public Shelf"
                >
                  Public
                </span>
              </div>
              <p v-if="shelf.description" class="text-[11px] text-slate-400 truncate mt-0.5">
                {{ shelf.description }}
              </p>
            </div>
          </div>

          <!-- Book count badge -->
          <span class="text-[11px] px-2 py-0.5 rounded-full bg-white/[0.06] text-slate-400 font-mono flex-shrink-0">
            {{ shelf.book_count }} {{ shelf.book_count === 1 ? 'book' : 'books' }}
          </span>
        </div>
      </div>

      <!-- Quick Create New Shelf Accordion / Section -->
      <div class="mt-4 pt-3 border-t border-white/[0.08]">
        <div v-if="!showNewShelfForm">
          <button
            type="button"
            @click="showNewShelfForm = true"
            class="w-full flex items-center justify-center gap-2 py-2 px-3 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-xs font-medium text-slate-300 hover:text-white transition-all cursor-pointer"
          >
            <Plus class="w-3.5 h-3.5 text-glacier-400" />
            Create New Shelf
          </button>
        </div>

        <form v-else @submit.prevent="handleCreateShelf" class="space-y-2.5 text-xs animate-fade-in">
          <div class="flex items-center justify-between">
            <span class="font-semibold text-slate-300">New Shelf</span>
            <button
              type="button"
              @click="showNewShelfForm = false"
              class="text-slate-500 hover:text-slate-300 text-[11px]"
            >
              Cancel
            </button>
          </div>

          <input
            v-model="newShelfName"
            type="text"
            required
            placeholder="e.g. Must Read, Favorites, Tech..."
            class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 transition-colors"
          />

          <input
            v-model="newShelfDesc"
            type="text"
            placeholder="Optional short description..."
            class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 transition-colors"
          />

          <div class="flex items-center justify-between pt-1">
            <label class="flex items-center gap-2 text-slate-400 cursor-pointer select-none">
              <input 
                v-model="newShelfPublic" 
                type="checkbox" 
                class="rounded bg-slate-900 border-white/[0.2] text-glacier-500 focus:ring-glacier-400"
              />
              <span class="text-[11px]">Make shelf public to other users</span>
            </label>

            <button
              type="submit"
              :disabled="creatingShelf || !newShelfName.trim()"
              class="flex items-center gap-1.5 px-4 py-1.5 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs shadow-md shadow-glacier-500/20 transition-all cursor-pointer disabled:opacity-50"
            >
              <Loader2 v-if="creatingShelf" class="w-3 h-3 animate-spin" />
              <span>Create & Add</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { X, Book, Bookmark, Check, Plus, Loader2 } from 'lucide-vue-next'
import { fetchShelves, fetchBookShelves, addBookToShelf, removeBookFromShelf, createShelf } from '../api/client'

const props = defineProps({
  book: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'updated'])

const shelves = ref([])
const assignedShelfIDs = ref(new Set())
const loading = ref(true)
const error = ref('')

const showNewShelfForm = ref(false)
const newShelfName = ref('')
const newShelfDesc = ref('')
const newShelfPublic = ref(false)
const creatingShelf = ref(false)

function formatAuthors(authors) {
  if (!authors || authors.length === 0) return 'Unknown Author'
  return authors.map(a => typeof a === 'string' ? a : a.name).filter(Boolean).join(', ')
}

function isBookInShelf(shelfId) {
  return assignedShelfIDs.value.has(shelfId)
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [allShelves, bookShelfIds] = await Promise.all([
      fetchShelves(),
      fetchBookShelves(props.book.id)
    ])
    shelves.value = allShelves || []
    assignedShelfIDs.value = new Set(bookShelfIds || [])
  } catch (err) {
    console.error('Failed to load shelves data:', err)
    error.value = err.message || 'Failed to load shelves'
  } finally {
    loading.value = false
  }
}

async function toggleShelf(shelfId) {
  const isCurrentlyIn = assignedShelfIDs.value.has(shelfId)
  try {
    if (isCurrentlyIn) {
      await removeBookFromShelf(shelfId, props.book.id)
      assignedShelfIDs.value.delete(shelfId)
      const target = shelves.value.find(s => s.id === shelfId)
      if (target && target.book_count > 0) target.book_count--
    } else {
      await addBookToShelf(shelfId, props.book.id)
      assignedShelfIDs.value.add(shelfId)
      const target = shelves.value.find(s => s.id === shelfId)
      if (target) target.book_count++
    }
    emit('updated', { bookId: props.book.id, shelfIds: Array.from(assignedShelfIDs.value) })
  } catch (err) {
    console.error('Failed to toggle shelf:', err)
    error.value = err.message || 'Failed to update shelf'
  }
}

async function handleCreateShelf() {
  if (!newShelfName.value.trim()) return
  creatingShelf.value = true
  error.value = ''

  try {
    const created = await createShelf({
      name: newShelfName.value.trim(),
      description: newShelfDesc.value.trim(),
      is_public: newShelfPublic.value
    })

    // Add book to this new shelf immediately
    await addBookToShelf(created.id, props.book.id)
    created.book_count = 1
    shelves.value.push(created)
    assignedShelfIDs.value.add(created.id)

    newShelfName.value = ''
    newShelfDesc.value = ''
    newShelfPublic.value = false
    showNewShelfForm.value = false

    emit('updated', { bookId: props.book.id, shelfIds: Array.from(assignedShelfIDs.value) })
  } catch (err) {
    console.error('Failed to create shelf:', err)
    error.value = err.message || 'Failed to create shelf'
  } finally {
    creatingShelf.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>
