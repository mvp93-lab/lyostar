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
          <FileText v-if="book.format === 'pdf'" class="w-3.5 h-3.5" />
          <Book v-else class="w-3.5 h-3.5" />
          <span>{{ (book.format || 'epub').toUpperCase() }}</span>
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

      <!-- Format badge on top-left -->
      <div class="absolute top-2 left-2 px-1.5 py-0.5 rounded-md bg-black/60 backdrop-blur-md border border-white/[0.1] text-[10px] font-bold uppercase tracking-wider text-slate-300">
        {{ book.format || 'epub' }}
      </div>

      <!-- Quick Shelf button on top-right (visible on hover) -->
      <button
        @click.stop="$emit('open-shelf', book)"
        class="absolute top-2 right-2 p-1.5 rounded-lg bg-black/70 backdrop-blur-md border border-white/[0.1] text-slate-300 hover:text-glacier-400 hover:border-glacier-400/40 opacity-0 group-hover:opacity-100 transition-all z-10 cursor-pointer"
        title="Add to shelf"
      >
        <Bookmark class="w-3.5 h-3.5" />
      </button>

      <!-- Reading Progress Bar on bottom of cover -->
      <div v-if="book.progress > 0" class="absolute bottom-0 inset-x-0 bg-black/70 backdrop-blur-sm px-2 py-1.5 flex flex-col gap-1">
        <div class="flex items-center justify-between text-[10px] font-medium leading-none">
          <span :class="book.is_finished ? 'text-emerald-400' : 'text-glacier-400'">
            {{ book.is_finished ? 'Finished' : `${Math.round(book.progress * 100)}%` }}
          </span>
          <span v-if="!book.is_finished" class="text-[9px] text-slate-400">reading</span>
        </div>
        <div class="w-full h-1 bg-white/20 rounded-full overflow-hidden">
          <div 
            class="h-full rounded-full transition-all duration-300"
            :class="book.is_finished ? 'bg-emerald-400' : 'bg-glacier-400'"
            :style="{ width: `${Math.min(100, Math.round(book.progress * 100))}%` }"
          ></div>
        </div>
      </div>

      <!-- Quick Read Floating Overlay -->
      <div v-if="canRead" class="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center p-4">
        <button
          @click.stop="$emit('read', book)"
          class="flex items-center gap-2 px-4 py-2 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs shadow-lg shadow-glacier-500/20 transform scale-95 group-hover:scale-100 transition-all cursor-pointer"
        >
          <BookOpen class="w-3.5 h-3.5" />
          {{ book.progress > 0 && !book.is_finished ? 'Continue' : 'Read' }}
        </button>
      </div>
    </div>

    <!-- Metadata Section -->
    <div class="mt-3 flex-1 flex flex-col justify-between">
      <div>
        <!-- Series Tag -->
        <div v-if="book.series" class="mb-1">
          <span 
            @click.stop="$emit('filter-series', book.series)"
            class="inline-flex items-center text-[11px] font-medium px-2 py-0.5 rounded-md bg-glacier-400/10 hover:bg-glacier-400/20 text-glacier-400 border border-glacier-400/20 hover:border-glacier-400/40 max-w-full truncate cursor-pointer transition-all"
            :title="`View series: ${book.series}`"
          >
            {{ book.series }} <span v-if="book.series_index" class="font-mono ml-0.5">#{{ book.series_index }}</span>
          </span>
        </div>

        <!-- Title -->
        <h3 class="text-sm font-semibold text-white group-hover:text-glacier-400 transition-colors line-clamp-2 leading-snug">
          {{ book.title }}
        </h3>
      </div>

      <!-- Authors -->
      <p class="text-xs text-slate-400 mt-1 line-clamp-1">
        <template v-if="book.authors && book.authors.length > 0">
          <span
            v-for="(author, idx) in book.authors"
            :key="author"
            @click.stop="$emit('filter-author', author)"
            class="hover:text-glacier-400 hover:underline cursor-pointer transition-colors"
          >
            {{ author }}<span v-if="idx < book.authors.length - 1">, </span>
          </span>
        </template>
        <span v-else>Unknown Author</span>
      </p>

      <!-- Tags Pills -->
      <div v-if="book.tags && book.tags.length > 0" class="mt-2 flex flex-wrap items-center gap-1 overflow-hidden">
        <span
          v-for="tag in book.tags.slice(0, 2)"
          :key="tag"
          @click.stop="$emit('filter-tag', tag)"
          class="inline-block text-[10px] px-1.5 py-0.5 rounded-md bg-white/[0.03] hover:bg-glacier-500/15 text-slate-400 hover:text-glacier-300 border border-white/[0.06] hover:border-glacier-400/30 transition-all truncate max-w-[100px] cursor-pointer"
          :title="`Filter by ${tag}`"
        >
          #{{ tag }}
        </span>
        <span v-if="book.tags.length > 2" class="text-[9px] text-slate-500 font-mono">
          +{{ book.tags.length - 2 }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Book, BookOpen, FileText, Bookmark } from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'

const props = defineProps({
  book: {
    type: Object,
    required: true
  }
})

defineEmits(['select', 'read', 'filter-tag', 'filter-author', 'filter-series', 'open-shelf'])
const { canRead } = useAuth()

function formatAuthors(authors) {
  if (!authors || !authors.length) return 'Unknown Author'
  return authors.map(a => typeof a === 'string' ? a : (a.name || '')).filter(Boolean).join(', ') || 'Unknown Author'
}
</script>
