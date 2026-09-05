<template>
  <div 
    v-if="currentBook" 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto bg-black/75 backdrop-blur-sm animate-fade-in"
    @click.self="handleBackdropClick"
  >
    <div 
      class="relative w-full max-w-2xl bg-[#11131b] border border-white/[0.1] rounded-3xl p-6 sm:p-8 shadow-2xl shadow-black/80 flex flex-col md:flex-row gap-6 max-h-[90vh] overflow-y-auto"
    >
      <!-- Close Button -->
      <button 
        @click="handleClose"
        class="absolute top-4 right-4 p-2 rounded-full text-slate-400 hover:text-white hover:bg-white/[0.06] transition-colors z-10"
        title="Close"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Left Column: Cover & Actions -->
      <div class="w-full md:w-52 flex-shrink-0 flex flex-col items-center">
        <div class="w-44 md:w-full aspect-[2/3] rounded-2xl overflow-hidden bg-[#090a0f] border border-white/[0.08] shadow-lg relative">
          <img
            v-if="currentBook.has_cover && currentBook.cover_url"
            :src="currentBook.cover_url"
            :alt="currentBook.title"
            class="w-full h-full object-cover"
          />
          <div 
            v-else 
            class="w-full h-full flex flex-col justify-between p-4 bg-gradient-to-br from-slate-900 to-slate-950"
          >
            <component :is="currentBook.format === 'pdf' ? FileText : Book" class="w-6 h-6 text-glacier-400" />
            <div>
              <p class="text-xs font-semibold text-white line-clamp-3">{{ currentBook.title }}</p>
              <p class="text-[11px] text-slate-400 mt-1">{{ formatAuthors(currentBook.authors) }}</p>
            </div>
            <div class="w-6 h-0.5 bg-glacier-400/40"></div>
          </div>
        </div>

        <!-- Normal Actions (when not editing) -->
        <div v-if="!isEditing" class="mt-5 w-full flex flex-col gap-2">
          <button
            v-if="canRead"
            @click="$emit('read', currentBook)"
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-sm transition-all shadow-md shadow-glacier-500/20 cursor-pointer"
          >
            <BookOpen class="w-4 h-4" />
            Read in Browser
          </button>

          <a
            v-if="canDownload"
            :href="currentBook.download_url || currentBook.file_url"
            download
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-slate-300 hover:text-white text-xs font-medium transition-all"
          >
            <Download class="w-4 h-4" />
            Download {{ (currentBook.format || 'epub').toUpperCase() }}
          </a>

          <!-- Edit Metadata Button (can_edit) -->
          <button
            v-if="canEdit"
            @click="startEditing"
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-white/[0.03] hover:bg-white/[0.08] border border-white/[0.08] hover:border-amber-400/30 text-slate-300 hover:text-amber-300 text-xs font-medium transition-all cursor-pointer"
          >
            <Pencil class="w-4 h-4 text-amber-400" />
            Edit Metadata
          </button>

          <!-- Delete Book Button (can_delete) -->
          <button
            v-if="canDelete"
            @click="showDeleteConfirm = true"
            class="w-full flex items-center justify-center gap-2 py-2 px-4 rounded-xl bg-rose-500/5 hover:bg-rose-500/15 border border-rose-500/20 text-rose-400 hover:text-rose-300 text-xs font-medium transition-all cursor-pointer"
          >
            <Trash2 class="w-3.5 h-3.5" />
            Delete Book
          </button>
        </div>

        <!-- Editing Mode Hint -->
        <div v-else class="mt-5 w-full text-center">
          <span class="inline-flex items-center gap-1.5 text-xs font-medium text-amber-400/90 bg-amber-500/10 border border-amber-500/20 px-3 py-1.5 rounded-xl">
            <Pencil class="w-3.5 h-3.5" />
            Editing Metadata
          </span>
        </div>
      </div>

      <!-- Right Column: View Mode -->
      <div v-if="!isEditing" class="flex-1 min-w-0 flex flex-col">
        <!-- Series Badge -->
        <div v-if="currentBook.series" class="mb-2">
          <span class="inline-flex items-center text-xs font-medium px-2.5 py-1 rounded-lg bg-glacier-400/10 text-glacier-400 border border-glacier-400/20">
            {{ currentBook.series }} <span v-if="currentBook.series_index" class="ml-1">#{{ currentBook.series_index }}</span>
          </span>
        </div>

        <!-- Title -->
        <h2 class="text-xl sm:text-2xl font-bold text-white tracking-tight leading-snug">
          {{ currentBook.title }}
        </h2>

        <!-- Authors -->
        <p class="text-sm font-medium text-slate-400 mt-1">
          By {{ formatAuthors(currentBook.authors) }}
        </p>

        <!-- Tags / Categories -->
        <div v-if="currentBook.tags && currentBook.tags.length > 0" class="mt-3 flex flex-wrap gap-1.5">
          <span
            v-for="tag in currentBook.tags"
            :key="tag"
            @click="$emit('filter-tag', tag); $emit('close')"
            class="inline-flex items-center gap-1 text-[11px] font-medium px-2.5 py-1 rounded-lg bg-white/[0.04] hover:bg-glacier-500/15 text-slate-300 hover:text-glacier-300 border border-white/[0.08] hover:border-glacier-400/30 transition-all cursor-pointer"
            :title="`Filter library by #${tag}`"
          >
            <Tag class="w-3 h-3 text-glacier-400/80" />
            #{{ tag }}
          </span>
        </div>

        <!-- Description -->
        <div class="mt-4 flex-1">
          <h4 class="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">Description</h4>
          <div class="text-sm text-slate-300/90 leading-relaxed max-h-48 overflow-y-auto pr-2 space-y-2">
            <p v-if="currentBook.description" class="whitespace-pre-line">{{ cleanDescription(currentBook.description) }}</p>
            <p v-else class="italic text-slate-500">No description available for this title.</p>
          </div>
        </div>

        <!-- Details Grid -->
        <div class="mt-6 pt-4 border-t border-white/[0.08] grid grid-cols-2 gap-y-3 gap-x-4 text-xs">
          <div>
            <span class="text-slate-500 block">Publisher</span>
            <span class="text-slate-300 font-medium">{{ currentBook.publisher || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-slate-500 block">Publication Date</span>
            <span class="text-slate-300 font-medium">{{ currentBook.pub_date || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-slate-500 block">Language</span>
            <span class="text-slate-300 font-medium uppercase">{{ currentBook.language || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-slate-500 block">Format</span>
            <span class="text-slate-300 font-medium uppercase">{{ currentBook.format || 'EPUB' }}</span>
          </div>
          <div>
            <span class="text-slate-500 block">File Size</span>
            <span class="text-slate-300 font-medium">{{ formatFileSize(currentBook.file_size) }}</span>
          </div>
        </div>
      </div>

      <!-- Right Column: Edit Form Mode -->
      <div v-else class="flex-1 min-w-0 flex flex-col">
        <div class="flex items-center justify-between pb-3 border-b border-white/[0.08] mb-4">
          <h3 class="text-base font-bold text-white flex items-center gap-2">
            <Pencil class="w-4 h-4 text-amber-400" />
            Edit Book Information
          </h3>
          <span class="text-[11px] text-slate-400">Updates database metadata</span>
        </div>

        <!-- Error Alert -->
        <div v-if="editError" class="mb-4 p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300 flex items-start gap-2">
          <AlertCircle class="w-4 h-4 flex-shrink-0 mt-0.5" />
          <span>{{ editError }}</span>
        </div>

        <form @submit.prevent="saveChanges" class="flex-1 flex flex-col gap-3.5 text-xs">
          <!-- Title Input (Required) -->
          <div>
            <label class="block text-slate-400 font-medium mb-1">Book Title <span class="text-rose-400">*</span></label>
            <input
              v-model="editForm.title"
              type="text"
              required
              placeholder="e.g. Clean Architecture"
              class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
            />
          </div>

          <!-- Authors Input -->
          <div>
            <label class="block text-slate-400 font-medium mb-1">
              Authors <span class="text-slate-500 font-normal">(comma-separated)</span>
            </label>
            <input
              v-model="editForm.authorsStr"
              type="text"
              placeholder="e.g. Robert C. Martin, Martin Fowler"
              class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
            />
          </div>

          <!-- Tags / Categories Chips Input with Autocomplete Dropdown -->
          <div class="relative">
            <div class="flex items-center justify-between mb-1">
              <label class="block text-slate-400 font-medium">
                Categories / Tags <span class="text-slate-500 font-normal text-[11px]">(press Enter or comma to add)</span>
              </label>
              <span class="text-[10px] text-slate-500">
                {{ editForm.tags.length }} {{ editForm.tags.length === 1 ? 'tag' : 'tags' }}
              </span>
            </div>

            <!-- Chips Container Box -->
            <div
              @click="focusTagInput"
              class="w-full min-h-[44px] p-2 rounded-xl bg-slate-900/80 border border-white/[0.1] focus-within:border-glacier-400 focus-within:ring-1 focus-within:ring-glacier-400 transition-colors flex flex-wrap items-center gap-1.5 cursor-text"
            >
              <!-- Render existing tag chips -->
              <span
                v-for="(tag, index) in editForm.tags"
                :key="index"
                class="inline-flex items-center gap-1 px-2.5 py-1 rounded-lg bg-white/[0.08] hover:bg-white/[0.12] border border-white/[0.1] text-slate-200 text-xs transition-colors"
              >
                <Tag class="w-3 h-3 text-glacier-400/80" />
                <span class="font-medium">{{ tag }}</span>
                <button
                  type="button"
                  @click.stop="removeTag(index)"
                  class="ml-0.5 p-0.5 rounded hover:bg-white/20 text-slate-400 hover:text-rose-300 transition-colors cursor-pointer"
                  title="Remove tag"
                >
                  <X class="w-3 h-3" />
                </button>
              </span>

              <!-- Inline input for typing new tags -->
              <input
                ref="tagInputRef"
                v-model="tagInputText"
                type="text"
                placeholder="Enter tag and press Enter..."
                @focus="showTagDropdown = true"
                @input="showTagDropdown = true; selectedTagIndex = -1"
                @blur="handleTagBlur"
                @keydown="handleTagKeyDown"
                class="flex-1 min-w-[140px] bg-transparent text-white text-xs placeholder-slate-500 outline-none py-1"
              />
            </div>

            <!-- Autocomplete Suggestions Dropdown -->
            <div
              v-if="showTagDropdown && filteredTagSuggestions.length > 0"
              class="absolute left-0 right-0 top-full mt-1.5 z-30 bg-[#161923] border border-white/[0.12] rounded-xl shadow-2xl shadow-black/80 max-h-48 overflow-y-auto py-1 backdrop-blur-md"
            >
              <div class="px-3 py-1 text-[10px] font-semibold text-slate-400 uppercase tracking-wider border-b border-white/[0.06] flex items-center justify-between">
                <span>Existing Categories</span>
                <span>{{ filteredTagSuggestions.length }} matches</span>
              </div>
              <button
                v-for="(t, idx) in filteredTagSuggestions"
                :key="t.id"
                type="button"
                @mousedown.prevent="addTag(t.name)"
                class="w-full px-3 py-2 flex items-center justify-between text-left text-xs transition-colors cursor-pointer"
                :class="selectedTagIndex === idx ? 'bg-glacier-500/20 text-glacier-300' : 'text-slate-300 hover:bg-white/[0.06] hover:text-white'"
              >
                <div class="flex items-center gap-2 min-w-0">
                  <Tag class="w-3.5 h-3.5 text-glacier-400/80 flex-shrink-0" />
                  <span class="font-medium truncate">{{ t.name }}</span>
                </div>
                <span class="text-[10px] px-1.5 py-0.5 rounded-full bg-white/[0.06] text-slate-400 font-mono flex-shrink-0 ml-2">
                  {{ t.book_count }} {{ t.book_count === 1 ? 'book' : 'books' }}
                </span>
              </button>
            </div>
          </div>

          <!-- Series and Index Grid -->
          <div class="grid grid-cols-3 gap-2.5">
            <div class="col-span-2">
              <label class="block text-slate-400 font-medium mb-1">Series</label>
              <input
                v-model="editForm.series"
                type="text"
                placeholder="e.g. Robert C. Martin Series"
                class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
              />
            </div>
            <div>
              <label class="block text-slate-400 font-medium mb-1">Series #</label>
              <input
                v-model.number="editForm.series_index"
                type="number"
                step="0.1"
                min="0"
                placeholder="1"
                class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
              />
            </div>
          </div>

          <!-- Publisher, Date & Language Grid -->
          <div class="grid grid-cols-3 gap-2.5">
            <div>
              <label class="block text-slate-400 font-medium mb-1">Publisher</label>
              <input
                v-model="editForm.publisher"
                type="text"
                placeholder="e.g. O'Reilly"
                class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
              />
            </div>
            <div>
              <label class="block text-slate-400 font-medium mb-1">Pub Date</label>
              <input
                v-model="editForm.pub_date"
                type="text"
                placeholder="YYYY-MM-DD"
                class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
              />
            </div>
            <div>
              <label class="block text-slate-400 font-medium mb-1">Language</label>
              <input
                v-model="editForm.language"
                type="text"
                placeholder="en"
                class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors uppercase"
              />
            </div>
          </div>

          <!-- Description Textarea -->
          <div class="flex-1 flex flex-col min-h-[100px]">
            <label class="block text-slate-400 font-medium mb-1">Description</label>
            <textarea
              v-model="editForm.description"
              rows="4"
              placeholder="Book summary or synopsis..."
              class="w-full flex-1 px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors resize-none leading-relaxed"
            ></textarea>
          </div>

          <!-- Form Buttons -->
          <div class="mt-2 pt-3 border-t border-white/[0.08] flex items-center justify-end gap-2.5">
            <button
              type="button"
              @click="cancelEditing"
              :disabled="saving"
              class="px-4 py-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/[0.06] text-xs font-medium transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="saving"
              class="flex items-center gap-2 px-5 py-2 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-semibold shadow-md shadow-glacier-500/20 transition-all cursor-pointer disabled:opacity-50"
            >
              <Loader2 v-if="saving" class="w-3.5 h-3.5 animate-spin" />
              <Save v-else class="w-3.5 h-3.5" />
              {{ saving ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </form>
      </div>

      <!-- Delete Confirmation Overlay -->
      <div 
        v-if="showDeleteConfirm" 
        class="absolute inset-0 z-20 bg-black/85 backdrop-blur-sm rounded-3xl p-6 flex flex-col items-center justify-center text-center animate-fade-in"
      >
        <div class="w-12 h-12 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-400 flex items-center justify-center mb-3">
          <Trash2 class="w-6 h-6" />
        </div>
        <h3 class="text-base font-bold text-white mb-1">Delete Book from Library?</h3>
        <p class="text-xs text-slate-400 max-w-sm mb-6 leading-relaxed">
          Are you sure you want to delete <span class="text-white font-medium">"{{ currentBook.title }}"</span>? 
          This will remove it from the catalog and delete any reading progress.
        </p>

        <div v-if="deleteError" class="mb-4 p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300 max-w-sm">
          {{ deleteError }}
        </div>

        <div class="flex items-center gap-3">
          <button
            @click="showDeleteConfirm = false"
            :disabled="deleting"
            class="px-4 py-2 rounded-xl text-slate-300 hover:text-white bg-white/[0.05] hover:bg-white/[0.1] border border-white/[0.08] text-xs font-medium transition-colors cursor-pointer"
          >
            Cancel
          </button>
          <button
            @click="handleDelete"
            :disabled="deleting"
            class="flex items-center gap-2 px-5 py-2 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold shadow-md shadow-rose-600/30 transition-all cursor-pointer disabled:opacity-50"
          >
            <Loader2 v-if="deleting" class="w-3.5 h-3.5 animate-spin" />
            <Trash2 v-else class="w-3.5 h-3.5" />
            {{ deleting ? 'Deleting...' : 'Yes, Delete Book' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { X, Book, BookOpen, Download, FileText, Pencil, Trash2, Save, Loader2, AlertCircle, Tag } from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'
import { updateBookMetadata, deleteBook, fetchTags } from '../api/client'

const props = defineProps({
  book: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'read', 'update', 'delete', 'filter-tag'])
const { canRead, canDownload, canEdit, canDelete } = useAuth()

const currentBook = ref(null)
const isEditing = ref(false)
const saving = ref(false)
const editError = ref('')
const showDeleteConfirm = ref(false)
const deleting = ref(false)
const deleteError = ref('')

const availableTags = ref([])
const showTagDropdown = ref(false)
const selectedTagIndex = ref(-1)
const tagInputRef = ref(null)

const editForm = reactive({
  title: '',
  authorsStr: '',
  tags: [],
  series: '',
  series_index: 0,
  publisher: '',
  pub_date: '',
  language: '',
  description: ''
})

const tagInputText = ref('')

function focusTagInput() {
  tagInputRef.value?.focus()
}

// Current set of tags already added as chips
const currentEnteredTags = computed(() => {
  return new Set(editForm.tags.map(t => t.toLowerCase()))
})

// Current text query typed in input
const currentTagQuery = computed(() => {
  return tagInputText.value.trim().toLowerCase()
})

// Dynamic suggestions: filtered by query, excluding already added tags
const filteredTagSuggestions = computed(() => {
  const entered = currentEnteredTags.value
  const q = currentTagQuery.value

  let list = availableTags.value.filter(t => !entered.has(t.name.toLowerCase()))
  if (q) {
    list = list.filter(t => t.name.toLowerCase().includes(q))
  }
  return list.slice(0, 8)
})

function addTag(name) {
  const val = (typeof name === 'string' ? name : tagInputText.value).trim()
  if (!val) return

  // Support comma-separated input or pasting multiple tags
  const newItems = val.split(',').map(s => s.trim()).filter(Boolean)
  for (const item of newItems) {
    const exists = editForm.tags.some(t => t.toLowerCase() === item.toLowerCase())
    if (!exists) {
      editForm.tags.push(item)
    }
  }

  tagInputText.value = ''
  selectedTagIndex.value = -1
  showTagDropdown.value = false
  tagInputRef.value?.focus()
}

function removeTag(index) {
  editForm.tags.splice(index, 1)
  tagInputRef.value?.focus()
}

function handleTagKeyDown(e) {
  // Backspace on empty input removes the last chip
  if (e.key === 'Backspace' && tagInputText.value === '' && editForm.tags.length > 0) {
    editForm.tags.pop()
    return
  }

  // Enter or Comma creates a new chip
  if (e.key === 'Enter' || e.key === ',') {
    e.preventDefault()
    if (showTagDropdown.value && selectedTagIndex.value >= 0 && selectedTagIndex.value < filteredTagSuggestions.value.length) {
      addTag(filteredTagSuggestions.value[selectedTagIndex.value].name)
    } else {
      addTag()
    }
    return
  }

  if (!showTagDropdown.value || filteredTagSuggestions.value.length === 0) {
    if (e.key === 'ArrowDown') {
      showTagDropdown.value = true
    }
    return
  }

  if (e.key === 'ArrowDown') {
    e.preventDefault()
    selectedTagIndex.value = (selectedTagIndex.value + 1) % filteredTagSuggestions.value.length
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    selectedTagIndex.value = (selectedTagIndex.value - 1 + filteredTagSuggestions.value.length) % filteredTagSuggestions.value.length
  } else if (e.key === 'Escape') {
    showTagDropdown.value = false
  }
}

function handleTagBlur() {
  setTimeout(() => {
    showTagDropdown.value = false
  }, 180)
}

watch(() => props.book, (newBook) => {
  currentBook.value = newBook ? { ...newBook } : null
  isEditing.value = false
  showDeleteConfirm.value = false
  showTagDropdown.value = false
  tagInputText.value = ''
}, { immediate: true })

async function startEditing() {
  if (!currentBook.value) return
  editError.value = ''
  showTagDropdown.value = false
  selectedTagIndex.value = -1
  tagInputText.value = ''

  // Load existing tags from server to power autocomplete suggestions
  try {
    const list = await fetchTags()
    availableTags.value = list || []
  } catch (err) {
    console.warn('Failed to load tags for autocomplete:', err)
  }

  // Format authors to comma-separated string
  let authorsText = ''
  if (Array.isArray(currentBook.value.authors)) {
    authorsText = currentBook.value.authors.map(a => typeof a === 'string' ? a : a.name).filter(Boolean).join(', ')
  }

  editForm.title = currentBook.value.title || ''
  editForm.authorsStr = authorsText
  editForm.tags = Array.isArray(currentBook.value.tags) ? [...currentBook.value.tags] : []
  editForm.series = currentBook.value.series || ''
  editForm.series_index = currentBook.value.series_index || 0
  editForm.publisher = currentBook.value.publisher || ''
  editForm.pub_date = currentBook.value.pub_date || ''
  editForm.language = currentBook.value.language || ''
  editForm.description = currentBook.value.description || ''
  
  isEditing.value = true
}

function cancelEditing() {
  isEditing.value = false
  editError.value = ''
  showTagDropdown.value = false
  tagInputText.value = ''
}

async function saveChanges() {
  if (!currentBook.value) return
  if (!editForm.title.trim()) {
    editError.value = 'Book title cannot be empty.'
    return
  }

  // Commit any unsubmitted text remaining in input before saving
  if (tagInputText.value.trim()) {
    addTag()
  }

  saving.value = true
  editError.value = ''

  try {
    const authors = editForm.authorsStr
      .split(',')
      .map(s => s.trim())
      .filter(Boolean)

    const payload = {
      title: editForm.title.trim(),
      authors,
      tags: editForm.tags,
      series: editForm.series.trim(),
      series_index: Number(editForm.series_index) || 0,
      publisher: editForm.publisher.trim(),
      pub_date: editForm.pub_date.trim(),
      language: editForm.language.trim().toLowerCase(),
      description: editForm.description.trim()
    }

    const updated = await updateBookMetadata(currentBook.value.id, payload)
    currentBook.value = updated
    isEditing.value = false
    emit('update', updated)
  } catch (err) {
    console.error('Failed to update book:', err)
    editError.value = err.message || 'Failed to update book metadata'
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  if (!currentBook.value) return
  deleting.value = true
  deleteError.value = ''

  try {
    const bookId = currentBook.value.id
    await deleteBook(bookId)
    showDeleteConfirm.value = false
    emit('delete', bookId)
    emit('close')
  } catch (err) {
    console.error('Failed to delete book:', err)
    deleteError.value = err.message || 'Failed to delete book'
  } finally {
    deleting.value = false
  }
}

function handleBackdropClick() {
  if (showDeleteConfirm.value) {
    showDeleteConfirm.value = false
    return
  }
  if (isEditing.value) {
    if (confirm('Discard unsaved metadata changes?')) {
      isEditing.value = false
      emit('close')
    }
    return
  }
  emit('close')
}

function handleClose() {
  if (isEditing.value) {
    if (confirm('Discard unsaved metadata changes?')) {
      isEditing.value = false
      emit('close')
    }
    return
  }
  emit('close')
}

function handleKeyDown(e) {
  if (e.key === 'Escape') {
    if (showDeleteConfirm.value) {
      showDeleteConfirm.value = false
      return
    }
    if (isEditing.value) {
      cancelEditing()
      return
    }
    emit('close')
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeyDown)
})

function formatAuthors(authors) {
  if (!authors || authors.length === 0) return 'Unknown Author'
  return authors.map(a => typeof a === 'string' ? a : a.name).filter(Boolean).join(', ')
}

function cleanDescription(desc) {
  if (!desc) return ''
  return desc.replace(/<[^>]*>?/gm, '').trim()
}

function formatFileSize(bytes) {
  if (!bytes) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}
</script>
