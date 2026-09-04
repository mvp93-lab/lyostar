<template>
  <div 
    v-if="book" 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto bg-black/75 backdrop-blur-sm animate-fade-in"
    @click.self="$emit('close')"
  >
    <div 
      class="relative w-full max-w-2xl bg-[#11131b] border border-white/[0.1] rounded-3xl p-6 sm:p-8 shadow-2xl shadow-black/80 flex flex-col md:flex-row gap-6 max-h-[90vh] overflow-y-auto"
    >
      <!-- Close Button -->
      <button 
        @click="$emit('close')"
        class="absolute top-4 right-4 p-2 rounded-full text-slate-400 hover:text-white hover:bg-white/[0.06] transition-colors"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Left Column: Cover & Actions -->
      <div class="w-full md:w-52 flex-shrink-0 flex flex-col items-center">
        <div class="w-44 md:w-full aspect-[2/3] rounded-2xl overflow-hidden bg-[#090a0f] border border-white/[0.08] shadow-lg">
          <img
            v-if="book.has_cover && book.cover_url"
            :src="book.cover_url"
            :alt="book.title"
            class="w-full h-full object-cover"
          />
          <div 
            v-else 
            class="w-full h-full flex flex-col justify-between p-4 bg-gradient-to-br from-slate-900 to-slate-950"
          >
            <Book class="w-6 h-6 text-glacier-400" />
            <div>
              <p class="text-xs font-semibold text-white line-clamp-3">{{ book.title }}</p>
              <p class="text-[11px] text-slate-400 mt-1">{{ formatAuthors(book.authors) }}</p>
            </div>
            <div class="w-6 h-0.5 bg-glacier-400/40"></div>
          </div>
        </div>

        <div class="mt-5 w-full flex flex-col gap-2.5">
          <button
            @click="$emit('read', book)"
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-sm transition-all shadow-md shadow-glacier-500/20"
          >
            <BookOpen class="w-4 h-4" />
            Read in Browser
          </button>

          <a
            :href="book.file_url"
            download
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-slate-300 hover:text-white text-xs font-medium transition-all"
          >
            <Download class="w-4 h-4" />
            Download EPUB
          </a>
        </div>
      </div>

      <!-- Right Column: Metadata & Description -->
      <div class="flex-1 min-w-0 flex flex-col">
        <!-- Series Badge -->
        <div v-if="book.series" class="mb-2">
          <span class="inline-flex items-center text-xs font-medium px-2.5 py-1 rounded-lg bg-glacier-400/10 text-glacier-400 border border-glacier-400/20">
            {{ book.series }} <span v-if="book.series_index" class="ml-1">#{{ book.series_index }}</span>
          </span>
        </div>

        <!-- Title -->
        <h2 class="text-xl sm:text-2xl font-bold text-white tracking-tight leading-snug">
          {{ book.title }}
        </h2>

        <!-- Authors -->
        <p class="text-sm font-medium text-slate-400 mt-1">
          By {{ formatAuthors(book.authors) }}
        </p>

        <!-- Description -->
        <div class="mt-4 flex-1">
          <h4 class="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">Description</h4>
          <div class="text-sm text-slate-300/90 leading-relaxed max-h-48 overflow-y-auto pr-2 space-y-2">
            <p v-if="book.description" class="whitespace-pre-line">{{ cleanDescription(book.description) }}</p>
            <p v-else class="italic text-slate-500">No description available for this title.</p>
          </div>
        </div>

        <!-- Details Grid -->
        <div class="mt-6 pt-4 border-t border-white/[0.08] grid grid-cols-2 gap-y-3 gap-x-4 text-xs">
          <div>
            <span class="text-slate-500 block">Publisher</span>
            <span class="text-slate-300 font-medium">{{ book.publisher || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-slate-500 block">Publication Date</span>
            <span class="text-slate-300 font-medium">{{ book.pub_date || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-slate-500 block">Language</span>
            <span class="text-slate-300 font-medium uppercase">{{ book.language || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-slate-500 block">File Size</span>
            <span class="text-slate-300 font-medium">{{ formatFileSize(book.file_size) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onBeforeUnmount } from 'vue'
import { X, Book, BookOpen, Download } from 'lucide-vue-next'

const props = defineProps({
  book: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['close', 'read'])

function handleKeyDown(e) {
  if (e.key === 'Escape') {
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
  return authors.join(', ')
}

function cleanDescription(desc) {
  if (!desc) return ''
  // Strip simple HTML tags if present in EPUB Dublin Core description
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
