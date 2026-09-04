<template>
  <div 
    @click="$emit('select', book)"
    class="group relative bg-[#11131b] hover:bg-[#161923] border border-white/[0.08] hover:border-glacier-400/30 rounded-2xl p-3 flex flex-col cursor-pointer transition-all duration-300 hover:shadow-xl hover:shadow-glacier-500/5 hover:-translate-y-1"
  >
    <!-- Cover Image Container (Aspect Ratio 2:3) -->
    <div class="relative w-full aspect-[2/3] rounded-xl overflow-hidden bg-[#090a0f] border border-white/[0.05] flex-shrink-0">
      <img
        v-if="book.has_cover && book.cover_url"
        :src="book.cover_url"
        :alt="book.title"
        loading="lazy"
        class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
      />
      <!-- Fallback Cover if no image -->
      <div 
        v-else 
        class="w-full h-full flex flex-col justify-between p-4 bg-gradient-to-br from-slate-900 via-[#131722] to-slate-950 border border-white/[0.06]"
      >
        <div class="flex items-center gap-1.5 text-glacier-400/60 text-xs font-semibold uppercase tracking-wider">
          <Book class="w-3.5 h-3.5" />
          <span>EPUB</span>
        </div>
        <div>
          <h4 class="text-sm font-semibold text-white/90 line-clamp-3 leading-tight mb-2">
            {{ book.title }}
          </h4>
          <p class="text-xs text-slate-400 line-clamp-1">
            {{ formatAuthors(book.authors) }}
          </p>
        </div>
        <div class="w-8 h-1 rounded-full bg-glacier-400/30"></div>
      </div>

      <!-- Quick Read Floating Overlay -->
      <div class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center p-4">
        <button
          @click.stop="$emit('read', book)"
          class="flex items-center gap-2 px-4 py-2 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs shadow-lg shadow-glacier-500/20 transform scale-95 group-hover:scale-100 transition-all"
        >
          <BookOpen class="w-3.5 h-3.5" />
          Read Now
        </button>
      </div>
    </div>

    <!-- Metadata Section -->
    <div class="mt-3 flex-1 flex flex-col justify-between">
      <div>
        <!-- Series Tag -->
        <div v-if="book.series" class="mb-1">
          <span class="inline-flex items-center text-[11px] font-medium px-2 py-0.5 rounded-md bg-glacier-400/10 text-glacier-400 border border-glacier-400/20 max-w-full truncate">
            {{ book.series }} <span v-if="book.series_index">#{{ book.series_index }}</span>
          </span>
        </div>

        <!-- Title -->
        <h3 class="text-sm font-semibold text-white group-hover:text-glacier-400 transition-colors line-clamp-2 leading-snug">
          {{ book.title }}
        </h3>
      </div>

      <!-- Authors -->
      <p class="text-xs text-slate-400 mt-1 line-clamp-1">
        {{ formatAuthors(book.authors) }}
      </p>
    </div>
  </div>
</template>

<script setup>
import { Book, BookOpen } from 'lucide-vue-next'

const props = defineProps({
  book: {
    type: Object,
    required: true
  }
})

defineEmits(['select', 'read'])

function formatAuthors(authors) {
  if (!authors || authors.length === 0) return 'Unknown Author'
  return authors.join(', ')
}
</script>
