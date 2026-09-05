<template>
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-black/75 backdrop-blur-sm animate-fade-in"
    @click.self="$emit('close')"
  >
    <div 
      class="relative w-full max-w-xl bg-[#11131b] border border-white/[0.1] rounded-3xl p-6 sm:p-8 shadow-2xl shadow-black/80 flex flex-col max-h-[85vh]"
    >
      <!-- Close Button -->
      <button 
        @click="$emit('close')"
        class="absolute top-4 right-4 p-2 rounded-full text-slate-400 hover:text-white hover:bg-white/[0.06] transition-colors"
        title="Close"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Header -->
      <div class="flex items-center justify-between pb-4 border-b border-white/[0.08] pr-8">
        <div class="flex items-center gap-2.5">
          <div class="w-9 h-9 rounded-xl bg-glacier-500/10 border border-glacier-500/20 text-glacier-400 flex items-center justify-center">
            <Bookmark class="w-5 h-5" />
          </div>
          <div>
            <h3 class="text-base font-bold text-white tracking-tight">Custom Shelves & Collections</h3>
            <p class="text-xs text-slate-400 mt-0.5">Organize and group your personal reading lists</p>
          </div>
        </div>

        <button
          v-if="!showForm"
          @click="openCreateForm"
          class="flex items-center gap-1.5 px-3.5 py-1.5 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs shadow-md shadow-glacier-500/20 transition-all cursor-pointer"
        >
          <Plus class="w-3.5 h-3.5" />
          <span>New Shelf</span>
        </button>
      </div>

      <!-- Error / Success Alert -->
      <div v-if="error" class="mt-4 p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300">
        {{ error }}
      </div>

      <!-- Create / Edit Shelf Form -->
      <div v-if="showForm" class="mt-4 p-4 rounded-2xl bg-white/[0.03] border border-white/[0.08] space-y-3 animate-fade-in">
        <div class="flex items-center justify-between">
          <h4 class="text-xs font-semibold text-white">
            {{ editingShelfId ? 'Edit Shelf' : 'Create New Shelf' }}
          </h4>
          <button @click="closeForm" class="text-slate-500 hover:text-slate-300 text-xs">
            Cancel
          </button>
        </div>

        <div>
          <label class="block text-[11px] text-slate-400 mb-1">Shelf Name <span class="text-rose-400">*</span></label>
          <input
            v-model="formName"
            type="text"
            required
            placeholder="e.g. Science Fiction Best, Recommended 2024..."
            class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 transition-colors"
          />
        </div>

        <div>
          <label class="block text-[11px] text-slate-400 mb-1">Description <span class="text-slate-500">(optional)</span></label>
          <input
            v-model="formDesc"
            type="text"
            placeholder="Short notes about this collection..."
            class="w-full px-3 py-2 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 transition-colors"
          />
        </div>

        <div class="flex items-center justify-between pt-1">
          <label class="flex items-center gap-2 text-slate-400 cursor-pointer select-none">
            <input 
              v-model="formPublic" 
              type="checkbox" 
              class="rounded bg-slate-900 border-white/[0.2] text-glacier-500 focus:ring-glacier-400"
            />
            <span class="text-xs">Public shelf (visible to other users)</span>
          </label>

          <button
            type="button"
            @click="saveShelf"
            :disabled="saving || !formName.trim()"
            class="flex items-center gap-1.5 px-4 py-1.5 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs shadow-md shadow-glacier-500/20 transition-all cursor-pointer disabled:opacity-50"
          >
            <Loader2 v-if="saving" class="w-3 h-3 animate-spin" />
            <span>{{ editingShelfId ? 'Save Changes' : 'Create Shelf' }}</span>
          </button>
        </div>
      </div>

      <!-- Shelves List -->
      <div class="mt-4 flex-1 overflow-y-auto space-y-2 pr-1 min-h-[160px]">
        <div v-if="loading && shelves.length === 0" class="py-12 text-center text-slate-500 text-xs">
          <Loader2 class="w-6 h-6 animate-spin text-glacier-400 mx-auto mb-2" />
          Loading shelves...
        </div>

        <div v-else-if="shelves.length === 0" class="py-12 text-center text-slate-500 text-xs">
          No shelves created yet.<br />
          Click <strong class="text-glacier-400">New Shelf</strong> to start organizing your books!
        </div>

        <div
          v-for="shelf in shelves"
          :key="shelf.id"
          class="flex items-center justify-between p-3.5 rounded-2xl bg-white/[0.03] hover:bg-white/[0.06] border border-white/[0.06] transition-all group"
        >
          <!-- Left: Shelf Info & Click to view -->
          <div 
            @click="selectShelf(shelf)" 
            class="flex-1 min-w-0 pr-3 cursor-pointer select-none"
            title="Click to view books in this shelf"
          >
            <div class="flex items-center gap-2">
              <span class="text-xs font-semibold text-white group-hover:text-glacier-300 transition-colors">
                {{ shelf.name }}
              </span>
              <span 
                v-if="shelf.is_public" 
                class="text-[10px] px-1.5 py-0.2 rounded-md bg-white/[0.06] text-slate-400 border border-white/[0.06]"
              >
                Public
              </span>
              <span 
                v-if="!shelf.is_owner" 
                class="text-[10px] px-1.5 py-0.2 rounded-md bg-white/[0.04] text-slate-500"
              >
                by @{{ shelf.owner_username }}
              </span>
            </div>
            <p v-if="shelf.description" class="text-[11px] text-slate-400 truncate mt-0.5">
              {{ shelf.description }}
            </p>
          </div>

          <!-- Right: Book Count & Actions -->
          <div class="flex items-center gap-2 flex-shrink-0">
            <span class="text-[11px] px-2.5 py-1 rounded-full bg-white/[0.06] text-slate-300 font-mono">
              {{ shelf.book_count }} {{ shelf.book_count === 1 ? 'book' : 'books' }}
            </span>

            <!-- Owner Actions -->
            <template v-if="shelf.is_owner || isAdmin">
              <button
                type="button"
                @click.stop="openEditForm(shelf)"
                class="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/[0.08] transition-colors cursor-pointer"
                title="Edit Shelf"
              >
                <Pencil class="w-3.5 h-3.5" />
              </button>

              <button
                type="button"
                @click.stop="handleDeleteShelf(shelf)"
                class="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors cursor-pointer"
                title="Delete Shelf"
              >
                <Trash2 class="w-3.5 h-3.5" />
              </button>
            </template>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { X, Bookmark, Plus, Pencil, Trash2, Loader2 } from 'lucide-vue-next'
import { fetchShelves, createShelf, updateShelf, deleteShelf } from '../api/client'
import { useAuth } from '../composables/useAuth'

const emit = defineEmits(['close', 'select-shelf'])
const { isAdmin } = useAuth()

const shelves = ref([])
const loading = ref(true)
const error = ref('')

const showForm = ref(false)
const editingShelfId = ref(null)
const formName = ref('')
const formDesc = ref('')
const formPublic = ref(false)
const saving = ref(false)

async function loadShelves() {
  loading.value = true
  error.value = ''
  try {
    const list = await fetchShelves()
    shelves.value = list || []
  } catch (err) {
    console.error('Failed to load shelves:', err)
    error.value = err.message || 'Failed to load shelves'
  } finally {
    loading.value = false
  }
}

function openCreateForm() {
  editingShelfId.value = null
  formName.value = ''
  formDesc.value = ''
  formPublic.value = false
  showForm.value = true
}

function openEditForm(shelf) {
  editingShelfId.value = shelf.id
  formName.value = shelf.name
  formDesc.value = shelf.description || ''
  formPublic.value = shelf.is_public
  showForm.value = true
}

function closeForm() {
  showForm.value = false
  editingShelfId.value = null
}

async function saveShelf() {
  if (!formName.value.trim()) return
  saving.value = true
  error.value = ''

  try {
    if (editingShelfId.value) {
      const updated = await updateShelf(editingShelfId.value, {
        name: formName.value.trim(),
        description: formDesc.value.trim(),
        is_public: formPublic.value
      })
      const idx = shelves.value.findIndex(s => s.id === editingShelfId.value)
      if (idx !== -1) {
        shelves.value[idx] = { ...shelves.value[idx], ...updated }
      }
    } else {
      const created = await createShelf({
        name: formName.value.trim(),
        description: formDesc.value.trim(),
        is_public: formPublic.value
      })
      shelves.value.unshift(created)
    }
    closeForm()
  } catch (err) {
    console.error('Failed to save shelf:', err)
    error.value = err.message || 'Failed to save shelf'
  } finally {
    saving.value = false
  }
}

async function handleDeleteShelf(shelf) {
  if (!confirm(`Delete shelf "${shelf.name}"? Books in this shelf will remain in your library.`)) {
    return
  }

  try {
    await deleteShelf(shelf.id)
    shelves.value = shelves.value.filter(s => s.id !== shelf.id)
  } catch (err) {
    console.error('Failed to delete shelf:', err)
    error.value = err.message || 'Failed to delete shelf'
  }
}

function selectShelf(shelf) {
  emit('select-shelf', shelf)
  emit('close')
}

onMounted(() => {
  loadShelves()
})
</script>
