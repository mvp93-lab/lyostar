<template>
  <div>
    <!-- Mobile Backdrop Overlay -->
    <div
      v-if="isOpen"
      @click="$emit('close')"
      class="fixed inset-0 bg-black/80 backdrop-blur-sm z-40 md:hidden transition-opacity animate-fade-in"
    ></div>

    <!-- Sidebar Container (Fixed desktop + Drawer on mobile) -->
    <aside
      class="fixed top-0 bottom-0 left-0 z-50 w-64 bg-[#0b0d14] border-r border-white/[0.07] flex flex-col transition-transform duration-300 ease-in-out md:translate-x-0"
      :class="isOpen ? 'translate-x-0 shadow-2xl shadow-black/80' : '-translate-x-full md:translate-x-0'"
    >
      <!-- Top Brand Header -->
      <div class="h-16 px-5 flex items-center justify-between border-b border-white/[0.07] flex-shrink-0">
        <div 
          @click="onNavClick('books')" 
          class="flex items-center gap-2.5 cursor-pointer group select-none"
        >
          <div class="w-9 h-9 rounded-xl bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400 group-hover:bg-glacier-400/20 group-hover:border-glacier-400/40 group-hover:scale-105 transition-all shadow-sm shadow-glacier-500/10">
            <BookOpen class="w-5 h-5" />
          </div>
          <div>
            <span class="text-base font-bold tracking-tight text-white group-hover:text-glacier-400 transition-colors">
              Lyostar
            </span>
            <span class="ml-1.5 text-[10px] font-semibold px-1.5 py-0.5 rounded-full bg-white/[0.06] text-slate-400 border border-white/[0.05]">
              Reader
            </span>
          </div>
        </div>

        <!-- Close button for mobile -->
        <button
          @click="$emit('close')"
          class="p-1.5 rounded-xl text-slate-400 hover:text-white hover:bg-white/[0.06] md:hidden transition-colors cursor-pointer"
          title="Close menu"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Scrollable Navigation Area -->
      <div class="flex-1 overflow-y-auto py-4 px-3 space-y-6 scrollbar-thin">
        <!-- Section: BROWSE -->
        <div>
          <div class="px-3 mb-2 flex items-center justify-between text-[11px] font-bold uppercase tracking-wider text-slate-400/80">
            <span>Browse</span>
          </div>

          <nav class="space-y-1">
            <!-- All Books -->
            <button
              @click="onNavClick('books')"
              class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium transition-all group cursor-pointer"
              :class="activeNav === 'books' && !selectedShelf
                ? 'bg-glacier-500/15 text-glacier-300 font-semibold border border-glacier-400/30 shadow-sm'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04]'"
            >
              <div class="flex items-center gap-2.5">
                <Book class="w-4 h-4 transition-colors" :class="activeNav === 'books' && !selectedShelf ? 'text-glacier-400' : 'text-slate-500 group-hover:text-slate-300'" />
                <span>All Books</span>
              </div>
              <span v-if="totalBooks > 0" class="text-[10px] font-mono px-1.5 py-0.5 rounded-md bg-white/[0.05] text-slate-500">
                {{ totalBooks }}
              </span>
            </button>

            <!-- Continue Reading -->
            <button
              @click="onNavClick('continue')"
              class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium transition-all group cursor-pointer"
              :class="activeNav === 'continue' && !selectedShelf
                ? 'bg-glacier-500/15 text-glacier-300 font-semibold border border-glacier-400/30 shadow-sm'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04]'"
            >
              <div class="flex items-center gap-2.5">
                <Clock class="w-4 h-4 transition-colors" :class="activeNav === 'continue' && !selectedShelf ? 'text-glacier-400' : 'text-slate-500 group-hover:text-slate-300'" />
                <span>Continue Reading</span>
              </div>
              <span v-if="continueCount > 0" class="text-[10px] font-mono px-1.5 py-0.5 rounded-md bg-glacier-500/20 text-glacier-400 font-semibold">
                {{ continueCount }}
              </span>
            </button>

            <!-- Categories / Tags -->
            <button
              @click="onNavClick('tags')"
              class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium transition-all group cursor-pointer"
              :class="activeNav === 'tags' && !selectedShelf
                ? 'bg-glacier-500/15 text-glacier-300 font-semibold border border-glacier-400/30 shadow-sm'
                : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04]'"
            >
              <div class="flex items-center gap-2.5">
                <Tag class="w-4 h-4 transition-colors" :class="activeNav === 'tags' && !selectedShelf ? 'text-glacier-400' : 'text-slate-500 group-hover:text-slate-300'" />
                <span>Categories & Tags</span>
              </div>
              <span v-if="tagsCount > 0" class="text-[10px] font-mono px-1.5 py-0.5 rounded-md bg-white/[0.05] text-slate-500">
                {{ tagsCount }}
              </span>
            </button>
          </nav>
        </div>

        <!-- Section: SHELVES (Collections) -->
        <div>
          <div class="px-3 mb-2 flex items-center justify-between text-[11px] font-bold uppercase tracking-wider text-slate-400/80">
            <span>Shelves</span>
            <button
              @click="$emit('manage-shelves')"
              class="p-1 rounded-md text-slate-500 hover:text-glacier-400 hover:bg-white/[0.05] transition-colors cursor-pointer"
              title="Manage all shelves"
            >
              <Settings class="w-3.5 h-3.5" />
            </button>
          </div>

          <div class="space-y-1">
            <!-- Shelf List Items -->
            <template v-if="shelves && shelves.length > 0">
              <button
                v-for="shelf in shelves"
                :key="shelf.id"
                @click="onShelfClick(shelf)"
                class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium transition-all group cursor-pointer text-left"
                :class="selectedShelf && selectedShelf.id === shelf.id
                  ? 'bg-glacier-500/15 text-glacier-300 font-semibold border border-glacier-400/30 shadow-sm'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04]'"
              >
                <div class="flex items-center gap-2.5 min-w-0 pr-2">
                  <Bookmark class="w-3.5 h-3.5 flex-shrink-0 transition-colors" :class="selectedShelf && selectedShelf.id === shelf.id ? 'text-glacier-400' : 'text-slate-500 group-hover:text-slate-300'" />
                  <span class="truncate">{{ shelf.name }}</span>
                  <span v-if="shelf.is_public" class="text-[9px] px-1 py-0.2 rounded bg-glacier-500/10 text-glacier-400 border border-glacier-500/20 flex-shrink-0 font-medium">Public</span>
                </div>
                <span class="text-[10px] font-mono px-1.5 py-0.5 rounded-md bg-white/[0.05] text-slate-500 flex-shrink-0">
                  {{ shelf.book_count || 0 }}
                </span>
              </button>
            </template>
            <div v-else class="px-3 py-2 text-[11px] text-slate-500 italic">
              No custom shelves yet
            </div>

            <!-- Create A Shelf Button (Calibre-Web Style) -->
            <button
              @click="$emit('create-shelf')"
              class="w-full mt-2 flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-medium text-glacier-400 hover:text-glacier-300 bg-glacier-400/5 hover:bg-glacier-400/10 border border-dashed border-glacier-400/30 hover:border-glacier-400/50 transition-all cursor-pointer group"
            >
              <Plus class="w-3.5 h-3.5 transition-transform group-hover:rotate-90 duration-300" />
              <span>Create a Shelf</span>
            </button>
          </div>
        </div>

        <!-- Section: MANAGEMENT (System & Admin) -->
        <div>
          <div class="px-3 mb-2 text-[11px] font-bold uppercase tracking-wider text-slate-400/80">
            <span>Management</span>
          </div>

          <nav class="space-y-1">
            <!-- Upload Book (can_upload) -->
            <button
              v-if="canUpload"
              @click="onUploadClick"
              class="w-full flex items-center gap-2.5 px-3 py-2 rounded-xl text-xs font-medium text-slate-300 hover:text-white hover:bg-white/[0.04] transition-all cursor-pointer group"
            >
              <Upload class="w-4 h-4 text-glacier-400/80 group-hover:text-glacier-400" />
              <span>Upload Book</span>
            </button>

            <!-- Users Management (Admin) -->
            <button
              v-if="isAdmin"
              @click="onUsersClick"
              class="w-full flex items-center gap-2.5 px-3 py-2 rounded-xl text-xs font-medium text-slate-300 hover:text-white hover:bg-white/[0.04] transition-all cursor-pointer group"
            >
              <Users class="w-4 h-4 text-glacier-400/80 group-hover:text-glacier-400" />
              <span>Manage Users</span>
            </button>

            <!-- Rescan Library (Admin) -->
            <button
              v-if="isAdmin"
              @click="$emit('scan')"
              :disabled="isScanning"
              class="w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium text-slate-300 hover:text-white hover:bg-white/[0.04] transition-all disabled:opacity-50 cursor-pointer group"
            >
              <div class="flex items-center gap-2.5">
                <RefreshCw class="w-4 h-4 text-glacier-400/80 group-hover:text-glacier-400" :class="{ 'animate-spin': isScanning }" />
                <span>{{ isScanning ? 'Scanning...' : 'Rescan Library' }}</span>
              </div>
            </button>
          </nav>
        </div>
      </div>

      <!-- User Profile & Sign Out Footer -->
      <div v-if="user" class="p-3 border-t border-white/[0.07] bg-[#090a0f]/60 flex-shrink-0">
        <div class="flex items-center justify-between gap-2 p-2 rounded-xl bg-white/[0.03] border border-white/[0.05]">
          <div class="flex items-center gap-2.5 min-w-0">
            <div class="w-8 h-8 rounded-lg bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-xs font-bold text-glacier-400 flex-shrink-0">
              {{ (user.display_name || user.username).charAt(0).toUpperCase() }}
            </div>
            <div class="min-w-0">
              <div class="text-xs font-semibold text-white truncate">
                {{ user.display_name || user.username }}
              </div>
              <div class="text-[10px] text-slate-500 font-medium uppercase tracking-wider">
                {{ user.role }}
              </div>
            </div>
          </div>

          <!-- Sign Out Button -->
          <button
            @click="handleLogout"
            class="p-2 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors flex-shrink-0 cursor-pointer"
            title="Sign out"
          >
            <LogOut class="w-4 h-4" />
          </button>
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup>
import { 
  BookOpen, 
  Book, 
  Clock, 
  Tag, 
  Bookmark, 
  Plus, 
  Settings, 
  Upload, 
  Users, 
  RefreshCw, 
  LogOut, 
  X 
} from 'lucide-vue-next'
import { useRouter } from 'vue-router'
import { useAuth } from '../composables/useAuth'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  activeNav: {
    type: String,
    default: 'books'
  },
  selectedShelf: {
    type: Object,
    default: null
  },
  shelves: {
    type: Array,
    default: () => []
  },
  totalBooks: {
    type: Number,
    default: 0
  },
  continueCount: {
    type: Number,
    default: 0
  },
  tagsCount: {
    type: Number,
    default: 0
  },
  isScanning: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits([
  'close', 
  'select-nav', 
  'select-shelf', 
  'create-shelf', 
  'manage-shelves', 
  'open-upload', 
  'open-users', 
  'scan'
])

const router = useRouter()
const { user, isAdmin, canUpload, logout } = useAuth()

function onNavClick(nav) {
  emit('select-nav', nav)
  emit('close')
  if (nav === 'books') {
    router.push('/books')
  } else if (nav === 'continue') {
    router.push('/continue-reading')
  } else if (nav === 'tags') {
    router.push('/tags')
  }
}

function onShelfClick(shelf) {
  emit('select-shelf', shelf)
  emit('close')
  router.push(`/shelves/${shelf.id}`)
}

function onUploadClick() {
  emit('open-upload')
  emit('close')
  router.push('/upload')
}

function onUsersClick() {
  emit('open-users')
  emit('close')
  router.push('/users')
}

async function handleLogout() {
  emit('close')
  await logout()
  router.push('/login')
}
</script>
