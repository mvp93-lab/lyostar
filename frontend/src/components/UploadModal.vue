<template>
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 overflow-y-auto bg-black/75 backdrop-blur-sm animate-fade-in"
    @click.self="$emit('close')"
  >
    <div 
      class="relative w-full max-w-xl bg-[#11131b] border border-white/[0.1] rounded-3xl p-6 sm:p-8 shadow-2xl shadow-black/80 flex flex-col gap-5 max-h-[90vh] overflow-y-auto"
    >
      <!-- Close Button -->
      <button 
        @click="$emit('close')"
        class="absolute top-4 right-4 p-2 rounded-full text-slate-400 hover:text-white hover:bg-white/[0.06] transition-colors"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Modal Header -->
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-glacier-400/10 border border-glacier-400/20 flex items-center justify-center text-glacier-400">
          <Upload class="w-5 h-5" />
        </div>
        <div>
          <h2 class="text-lg sm:text-xl font-bold text-white tracking-tight">Upload Books</h2>
          <p class="text-xs text-slate-400">Add EPUB or PDF ebooks to your Lyostar library</p>
        </div>
      </div>

      <!-- Drag & Drop Zone -->
      <div 
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="onFileDrop"
        @click="triggerFileInput"
        :class="[
          'relative border-2 border-dashed rounded-2xl p-6 sm:p-8 flex flex-col items-center justify-center text-center cursor-pointer transition-all duration-200 select-none',
          isDragging 
            ? 'border-glacier-400 bg-glacier-500/10 scale-[1.01]' 
            : 'border-white/[0.12] hover:border-glacier-400/40 bg-[#090a0f]/60 hover:bg-[#090a0f]'
        ]"
      >
        <input 
          ref="fileInputRef"
          type="file" 
          multiple 
          accept=".epub,.pdf,application/epub+zip,application/pdf" 
          class="hidden" 
          @change="onFileInputChange"
        />

        <div class="w-12 h-12 rounded-2xl bg-white/[0.04] border border-white/[0.08] flex items-center justify-center text-slate-400 mb-3 group-hover:text-glacier-400 transition-colors">
          <FileUp class="w-6 h-6 text-glacier-400" />
        </div>

        <p class="text-sm font-semibold text-white mb-1">
          Drag & drop your files here, or <span class="text-glacier-400 underline underline-offset-2">browse</span>
        </p>
        <p class="text-xs text-slate-400">
          Supports <span class="text-slate-200 font-medium">EPUB</span> and <span class="text-slate-200 font-medium">PDF</span> (up to 100MB per file)
        </p>
      </div>

      <!-- Selected Files Queue -->
      <div v-if="filesQueue.length > 0" class="flex flex-col gap-2.5">
        <div class="flex items-center justify-between text-xs font-semibold text-slate-400 px-1">
          <span>Queue ({{ filesQueue.length }})</span>
          <button 
            v-if="!isUploading && hasPendingFiles" 
            @click="clearCompletedOrAll"
            class="text-slate-500 hover:text-slate-300 transition-colors"
          >
            Clear
          </button>
        </div>

        <div class="flex flex-col gap-2 max-h-60 overflow-y-auto pr-1">
          <div 
            v-for="(item, index) in filesQueue" 
            :key="item.id"
            class="flex items-center justify-between gap-3 p-3 rounded-xl bg-[#0d0f17] border border-white/[0.06] text-xs transition-colors"
          >
            <div class="flex items-center gap-3 min-w-0 flex-1">
              <span 
                :class="[
                  'px-2 py-0.5 rounded font-mono text-[10px] font-bold uppercase tracking-wider',
                  item.format === 'epub' 
                    ? 'bg-purple-500/10 border border-purple-500/20 text-purple-400' 
                    : 'bg-amber-500/10 border border-amber-500/20 text-amber-400'
                ]"
              >
                {{ item.format }}
              </span>
              <div class="min-w-0 flex-1">
                <p class="text-slate-200 font-medium truncate" :title="item.file.name">
                  {{ item.file.name }}
                </p>
                <p class="text-[11px] text-slate-500">
                  {{ formatBytes(item.file.size) }}
                  <span v-if="item.error" class="text-rose-400 ml-1.5">• {{ item.error }}</span>
                  <span v-else-if="item.status === 'success'" class="text-emerald-400 ml-1.5">• Indexed successfully</span>
                </p>
              </div>
            </div>

            <!-- Status / Action icon -->
            <div class="flex items-center gap-2 flex-shrink-0">
              <div v-if="item.status === 'uploading'" class="flex items-center gap-1.5 text-glacier-400">
                <Loader2 class="w-4 h-4 animate-spin" />
                <span class="text-[11px] font-medium hidden sm:inline">Uploading...</span>
              </div>

              <div v-else-if="item.status === 'success'" class="text-emerald-400 flex items-center gap-1">
                <CheckCircle2 class="w-4 h-4" />
                <span class="text-[11px] font-medium hidden sm:inline">Done</span>
              </div>

              <div v-else-if="item.status === 'error'" class="text-rose-400 flex items-center gap-1">
                <AlertCircle class="w-4 h-4" />
                <span class="text-[11px] font-medium hidden sm:inline">Failed</span>
              </div>

              <button 
                v-if="item.status === 'pending'"
                @click="removeFile(index)"
                class="p-1 rounded-lg text-slate-500 hover:text-slate-300 hover:bg-white/[0.06] transition-colors"
                title="Remove"
              >
                <Trash2 class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex items-center justify-end gap-3 pt-2 border-t border-white/[0.08]">
        <button 
          @click="$emit('close')"
          class="px-4 py-2 rounded-xl text-xs font-medium text-slate-400 hover:text-white bg-white/[0.04] hover:bg-white/[0.08] transition-colors"
        >
          {{ anySuccess ? 'Done' : 'Cancel' }}
        </button>

        <button 
          v-if="hasPendingFiles"
          @click="startUpload"
          :disabled="isUploading"
          class="flex items-center gap-2 px-5 py-2 rounded-xl text-xs font-semibold bg-glacier-500 hover:bg-glacier-400 text-slate-950 transition-all shadow-md shadow-glacier-500/20 disabled:opacity-50 cursor-pointer"
        >
          <Loader2 v-if="isUploading" class="w-3.5 h-3.5 animate-spin" />
          <Upload v-else class="w-3.5 h-3.5" />
          <span>{{ isUploading ? 'Uploading...' : `Upload ${pendingCount} ${pendingCount === 1 ? 'Book' : 'Books'}` }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { Upload, FileUp, X, CheckCircle2, AlertCircle, Loader2, Trash2 } from 'lucide-vue-next'
import { uploadBook } from '../api/client.js'

const emit = defineEmits(['close', 'uploaded'])

const fileInputRef = ref(null)
const isDragging = ref(false)
const isUploading = ref(false)
const filesQueue = ref([])
const anySuccess = ref(false)

const hasPendingFiles = computed(() => filesQueue.value.some(f => f.status === 'pending'))
const pendingCount = computed(() => filesQueue.value.filter(f => f.status === 'pending').length)

function triggerFileInput() {
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
    fileInputRef.value.click()
  }
}

function onFileInputChange(e) {
  const files = Array.from(e.target.files || [])
  addFilesToQueue(files)
}

function onFileDrop(e) {
  isDragging.value = false
  const files = Array.from(e.dataTransfer.files || [])
  addFilesToQueue(files)
}

function addFilesToQueue(files) {
  for (const file of files) {
    const ext = file.name.split('.').pop().toLowerCase()
    if (ext !== 'epub' && ext !== 'pdf') {
      continue
    }
    // Prevent duplicate entries of the same filename in queue
    if (filesQueue.value.some(q => q.file.name === file.name && q.file.size === file.size)) {
      continue
    }

    filesQueue.value.push({
      id: `${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
      file,
      format: ext,
      status: 'pending',
      error: null
    })
  }
}

function removeFile(index) {
  filesQueue.value.splice(index, 1)
}

function clearCompletedOrAll() {
  filesQueue.value = []
}

async function startUpload() {
  if (isUploading.value) return
  isUploading.value = true

  let successOccurred = false

  for (const item of filesQueue.value) {
    if (item.status !== 'pending') continue

    item.status = 'uploading'
    item.error = null

    try {
      await uploadBook(item.file)
      item.status = 'success'
      successOccurred = true
      anySuccess.value = true
    } catch (err) {
      item.status = 'error'
      item.error = err.message || 'Upload failed'
    }
  }

  isUploading.value = false

  if (successOccurred) {
    emit('uploaded')
  }
}

function formatBytes(bytes) {
  if (!bytes || bytes <= 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`
}
</script>
