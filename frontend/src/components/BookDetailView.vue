<template>
  <div class="flex-1 w-full max-w-7xl mx-auto px-4 lg:px-8 py-6 sm:py-8 animate-fade-in">
    <!-- Top Breadcrumbs Bar -->
    <div class="flex flex-wrap items-center justify-between gap-3 mb-6 pb-4 border-b border-white/[0.08]">
      <nav class="flex items-center gap-2 text-xs text-slate-400">
        <button
          @click="goBack"
          class="flex items-center gap-1.5 text-slate-400 hover:text-white transition-colors cursor-pointer group"
        >
          <ArrowLeft class="w-4 h-4 text-slate-500 group-hover:text-glacier-400 transition-colors" />
          <span>Library</span>
        </button>
        <ChevronRight class="w-3.5 h-3.5 text-slate-600 flex-shrink-0" />
        
        <button
          v-if="book && book.tags && book.tags.length > 0"
          @click="navigateToTag(book.tags[0])"
          class="text-slate-400 hover:text-glacier-400 transition-colors cursor-pointer truncate max-w-[150px]"
        >
          {{ book.tags[0] }}
        </button>
        <span v-else class="text-slate-500">Books</span>

        <ChevronRight class="w-3.5 h-3.5 text-slate-600 flex-shrink-0" />
        <span class="text-white font-medium truncate max-w-[200px] sm:max-w-[320px]">
          {{ book ? book.title : 'Loading...' }}
        </span>
      </nav>

      <!-- Quick Back Button on right -->
      <button
        @click="goBack"
        class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-xs font-medium text-slate-300 hover:text-white transition-all cursor-pointer"
      >
        <BookOpen class="w-3.5 h-3.5 text-glacier-400" />
        <span>Back to Books</span>
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-32 flex flex-col items-center justify-center text-slate-500">
      <Loader2 class="w-8 h-8 text-glacier-400 animate-spin mb-3" />
      <p class="text-xs uppercase tracking-wider">Loading book details...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error || !book" class="py-24 flex flex-col items-center justify-center text-center px-4">
      <div class="w-16 h-16 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-400 flex items-center justify-center mb-4">
        <AlertCircle class="w-8 h-8" />
      </div>
      <h3 class="text-base font-semibold text-white mb-1">Book Not Found</h3>
      <p class="text-xs text-slate-400 max-w-sm mb-6">{{ error || 'The requested book could not be found or has been removed.' }}</p>
      <button
        @click="goBack"
        class="px-5 py-2.5 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-semibold text-xs transition-all shadow-md shadow-glacier-500/20 cursor-pointer"
      >
        Return to Library
      </button>
    </div>

    <!-- Main Content: 2-Column Responsive Layout -->
    <div v-else class="grid grid-cols-1 md:grid-cols-12 gap-8 items-start">
      <!-- Left Column: Book Cover & Quick Action Buttons (md: 4 cols, lg: 3 cols) -->
      <div class="md:col-span-5 lg:col-span-4 xl:col-span-3 space-y-5">
        <!-- Cover Card with Glow / Shadow -->
        <div class="relative w-full aspect-[2/3] rounded-2xl overflow-hidden bg-[#0c0e15] border border-white/[0.1] shadow-2xl shadow-black/80 flex items-center justify-center group">
          <img
            v-if="book.has_cover && book.cover_url"
            :src="book.cover_url"
            :alt="book.title"
            class="w-full h-full object-cover transition-transform duration-500 group-hover:scale-105"
          />
          <div 
            v-else 
            class="w-full h-full flex flex-col justify-between p-6 bg-gradient-to-br from-slate-900 via-[#131722] to-slate-950"
          >
            <div class="flex items-center gap-1.5 text-glacier-400 font-semibold text-xs uppercase tracking-wider">
              <FileText v-if="book.format === 'pdf'" class="w-4 h-4" />
              <Book v-else class="w-4 h-4" />
              <span>{{ (book.format || 'epub').toUpperCase() }}</span>
            </div>
            <div>
              <p class="text-base font-bold text-white line-clamp-3 leading-snug">{{ book.title }}</p>
              <p class="text-xs text-slate-400 mt-2">{{ formatAuthors(book.authors) }}</p>
            </div>
            <div class="w-10 h-1 rounded-full bg-glacier-400/40"></div>
          </div>

          <!-- Format Badge on Top Left -->
          <div class="absolute top-3 left-3 px-2 py-0.5 rounded-md bg-black/70 backdrop-blur-md border border-white/[0.15] text-[10px] font-bold uppercase tracking-wider text-slate-200">
            {{ (book.format || 'epub').toUpperCase() }}
          </div>

          <!-- Reading Progress Badge on Bottom -->
          <div v-if="book.progress > 0" class="absolute bottom-0 inset-x-0 bg-black/80 backdrop-blur-md px-3 py-2 border-t border-white/[0.1] flex flex-col gap-1.5">
            <div class="flex items-center justify-between text-xs font-semibold leading-none">
              <span :class="book.is_finished ? 'text-emerald-400' : 'text-glacier-400'">
                {{ book.is_finished ? 'Completed' : `${Math.round(book.progress * 100)}% Read` }}
              </span>
              <span class="text-[10px] text-slate-400 font-normal">
                {{ book.is_finished ? 'Finished' : 'In Progress' }}
              </span>
            </div>
            <div class="w-full h-1.5 bg-white/20 rounded-full overflow-hidden">
              <div 
                class="h-full rounded-full transition-all duration-300"
                :class="book.is_finished ? 'bg-emerald-400' : 'bg-glacier-400'"
                :style="{ width: `${Math.round(book.progress * 100)}%` }"
              ></div>
            </div>
          </div>
        </div>

        <!-- Action Buttons Stack -->
        <div class="space-y-2.5">
          <!-- Primary Action: Read in Browser -->
          <button
            v-if="canRead"
            @click="handleRead"
            class="w-full flex items-center justify-center gap-2.5 py-3 px-5 rounded-2xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 font-bold text-sm transition-all shadow-lg shadow-glacier-500/25 hover:shadow-glacier-500/35 cursor-pointer group"
          >
            <BookOpen class="w-4 h-4 transition-transform group-hover:scale-110" />
            <span>{{ book.progress > 0 && !book.is_finished ? 'Resume Reading' : 'Read in Browser' }}</span>
          </button>

          <!-- Download Action -->
          <a
            v-if="canDownload"
            :href="book.download_url || book.file_url"
            download
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-slate-300 hover:text-white text-xs font-medium transition-all"
          >
            <Download class="w-4 h-4 text-glacier-400" />
            <span>Download {{ (book.format || 'epub').toUpperCase() }} ({{ formatFileSize(book.file_size) }})</span>
          </a>

          <!-- Add to Shelf Button -->
          <button
            @click="showShelfModal = true"
            class="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] text-slate-300 hover:text-white text-xs font-medium transition-all cursor-pointer"
          >
            <Bookmark class="w-4 h-4 text-glacier-400" />
            <span>Add to Shelf</span>
          </button>
        </div>

        <!-- Quick Technical Info Box -->
        <div class="bg-[#11131b] border border-white/[0.08] rounded-2xl p-4 space-y-2.5 text-xs text-slate-400">
          <div class="flex items-center justify-between">
            <span class="text-slate-500">File Format</span>
            <span class="font-mono uppercase text-slate-300 font-semibold">{{ book.format }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-500">File Size</span>
            <span class="font-mono text-slate-300">{{ formatFileSize(book.file_size) }}</span>
          </div>
          <div v-if="book.created_at" class="flex items-center justify-between">
            <span class="text-slate-500">Added to Library</span>
            <span class="text-slate-300">{{ formatDate(book.created_at) }}</span>
          </div>
          <div v-if="book.updated_at" class="flex items-center justify-between">
            <span class="text-slate-500">Last Updated</span>
            <span class="text-slate-300">{{ formatDate(book.updated_at) }}</span>
          </div>
        </div>
      </div>

      <!-- Right Column: Metadata, Description & In-Page Editor (md: 7 cols, lg: 8 cols) -->
      <div class="md:col-span-7 lg:col-span-8 xl:col-span-9 bg-[#11131b] border border-white/[0.08] rounded-3xl p-6 sm:p-8">
        <!-- NORMAL VIEW MODE -->
        <div v-if="!isEditing" class="space-y-6">
          <!-- Header: Title, Authors, Series, Edit/Delete Action Toolbar -->
          <div class="flex flex-col sm:flex-row sm:items-start justify-between gap-4 pb-6 border-b border-white/[0.08]">
            <div class="space-y-2 min-w-0 flex-1">
              <!-- Series Badge -->
              <div 
                v-if="book.series" 
                @click="navigateToSeries(book.series)"
                class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-glacier-500/15 hover:bg-glacier-500/25 text-glacier-300 border border-glacier-400/30 cursor-pointer transition-all"
                :title="`View series: ${book.series}`"
              >
                <Bookmark class="w-3.5 h-3.5 text-glacier-400" />
                <span>{{ book.series }}</span>
                <span v-if="book.series_index" class="text-glacier-400 font-mono">#{{ book.series_index }}</span>
              </div>

              <!-- Book Title -->
              <h1 class="text-2xl sm:text-3xl font-extrabold text-white tracking-tight leading-tight">
                {{ book.title }}
              </h1>

              <!-- Authors List -->
              <div class="flex items-center gap-2 text-sm text-slate-300">
                <span class="text-slate-500 font-medium">By</span>
                <div class="flex flex-wrap gap-1.5">
                  <span 
                    v-for="(author, idx) in formattedAuthorsList" 
                    :key="idx"
                    @click="navigateToAuthor(author)"
                    class="font-semibold text-glacier-300 hover:text-glacier-200 hover:underline cursor-pointer transition-colors"
                    :title="`View books by ${author}`"
                  >
                    {{ author }}<span v-if="idx < formattedAuthorsList.length - 1" class="text-slate-500 no-underline">,</span>
                  </span>
                </div>
              </div>

              <!-- Star Rating Widget -->
              <div class="flex items-center gap-3 pt-1">
                <div class="flex items-center gap-0.5">
                  <button
                    v-for="star in 5"
                    :key="star"
                    type="button"
                    @click="handleSetRating(star)"
                    @mouseenter="hoverRating = star"
                    @mouseleave="hoverRating = 0"
                    class="p-0.5 text-slate-600 hover:scale-110 transition-transform cursor-pointer"
                    :title="`Rate ${star} star${star > 1 ? 's' : ''}`"
                  >
                    <Star
                      class="w-4 h-4 transition-colors"
                      :class="[
                        (hoverRating || userRating) >= star
                          ? 'text-amber-400 fill-amber-400'
                          : 'text-slate-600'
                      ]"
                    />
                  </button>
                </div>

                <div class="flex items-center gap-2 text-xs text-slate-400">
                  <span v-if="userRating > 0" class="font-semibold text-amber-400">
                    Your rating: {{ userRating }}★
                  </span>
                  <button
                    v-if="userRating > 0"
                    type="button"
                    @click="handleClearRating"
                    class="text-[10px] text-slate-500 hover:text-rose-400 transition-colors cursor-pointer"
                  >
                    (Clear)
                  </button>
                  <span v-if="avgRating > 0" class="text-slate-500 font-mono">
                    • Avg: {{ avgRating.toFixed(1) }}★
                  </span>
                </div>
              </div>
            </div>

            <!-- Action Toolbar on Top Right -->
            <div class="flex items-center gap-2 flex-shrink-0">
              <button
                v-if="canEdit"
                @click="startEditing"
                class="flex items-center gap-1.5 px-3.5 py-2 rounded-xl bg-white/[0.04] hover:bg-white/[0.08] border border-white/[0.08] hover:border-glacier-400/30 text-xs font-semibold text-slate-300 hover:text-white transition-all cursor-pointer"
                title="Edit book metadata"
              >
                <Pencil class="w-3.5 h-3.5 text-glacier-400" />
                <span>Edit Metadata</span>
              </button>

              <button
                v-if="canDelete"
                @click="showDeleteConfirm = true"
                class="p-2 rounded-xl text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 border border-transparent hover:border-rose-500/20 transition-all cursor-pointer"
                title="Delete book"
              >
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>

          <!-- Tags / Categories Pills -->
          <div v-if="book.tags && book.tags.length > 0" class="space-y-2">
            <h3 class="text-xs font-bold uppercase tracking-wider text-slate-500">Categories & Genres</h3>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="tag in book.tags"
                :key="tag"
                @click="navigateToTag(tag)"
                class="inline-flex items-center gap-1.5 px-3 py-1 rounded-xl text-xs font-medium bg-white/[0.04] hover:bg-glacier-500/15 border border-white/[0.08] hover:border-glacier-400/30 text-slate-300 hover:text-glacier-300 transition-all cursor-pointer group"
              >
                <Tag class="w-3 h-3 text-glacier-400/70 group-hover:text-glacier-400" />
                <span>#{{ tag }}</span>
              </button>
            </div>
          </div>

          <!-- Description / Synopsis Section -->
          <div class="space-y-2.5 pt-2">
            <h3 class="text-xs font-bold uppercase tracking-wider text-slate-500">Synopsis</h3>
            <div 
              v-if="book.description" 
              class="text-sm text-slate-300 leading-relaxed space-y-3 whitespace-pre-line prose prose-invert max-w-none"
            >
              {{ cleanDescription(book.description) }}
            </div>
            <div v-else class="text-xs text-slate-500 italic py-2">
              No description available for this book.
            </div>
          </div>

          <!-- Publication & Details Table -->
          <div class="pt-6 border-t border-white/[0.08] space-y-3">
            <h3 class="text-xs font-bold uppercase tracking-wider text-slate-500">Book Details</h3>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
              <div class="p-3.5 rounded-2xl bg-white/[0.02] border border-white/[0.05]">
                <span class="text-slate-500 block mb-1">Publisher</span>
                <span class="font-medium text-slate-200">{{ book.publisher || 'Not specified' }}</span>
              </div>
              <div class="p-3.5 rounded-2xl bg-white/[0.02] border border-white/[0.05]">
                <span class="text-slate-500 block mb-1">Publication Date</span>
                <span class="font-medium text-slate-200">{{ book.pub_date || 'Unknown' }}</span>
              </div>
              <div class="p-3.5 rounded-2xl bg-white/[0.02] border border-white/[0.05]">
                <span class="text-slate-500 block mb-1">Language</span>
                <span class="font-medium uppercase text-slate-200">{{ book.language || 'en' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- EDIT METADATA MODE -->
        <div v-else class="space-y-6">
          <div class="flex items-center justify-between pb-4 border-b border-white/[0.08]">
            <div class="flex items-center gap-2">
              <Pencil class="w-5 h-5 text-glacier-400" />
              <h2 class="text-lg font-bold text-white">Edit Book Metadata</h2>
            </div>
            <button
              @click="cancelEditing"
              class="text-xs text-slate-400 hover:text-white transition-colors cursor-pointer"
            >
              Cancel
            </button>
          </div>

          <!-- Edit Error -->
          <div v-if="editError" class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300">
            {{ editError }}
          </div>

          <form @submit.prevent="handleSaveMetadata" class="space-y-4 text-xs">
            <!-- Title -->
            <div>
              <label class="block text-slate-400 font-semibold mb-1">Title *</label>
              <input
                v-model="editForm.title"
                type="text"
                required
                class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
              />
            </div>

            <!-- Authors -->
            <div>
              <label class="block text-slate-400 font-semibold mb-1">
                Authors <span class="text-slate-500 font-normal">(comma-separated)</span>
              </label>
              <input
                v-model="editForm.authorsStr"
                type="text"
                class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
              />
            </div>

            <!-- Tags / Categories Chips Input with Autocomplete Dropdown -->
            <div class="relative">
              <div class="flex items-center justify-between mb-1">
                <label class="block text-slate-400 font-semibold">
                  Categories / Tags <span class="text-slate-500 font-normal text-[11px]">(press Enter or comma to add)</span>
                </label>
                <span class="text-[10px] text-slate-500">
                  {{ editForm.tags.length }} tags
                </span>
              </div>

              <!-- Chips Container Box -->
              <div
                @click="focusTagInput"
                class="w-full min-h-[48px] p-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] focus-within:border-glacier-400 focus-within:ring-1 focus-within:ring-glacier-400 transition-colors flex flex-wrap items-center gap-1.5 cursor-text"
              >
                <!-- Render existing tag chips -->
                <span
                  v-for="(tag, index) in editForm.tags"
                  :key="index"
                  class="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg bg-white/[0.08] hover:bg-white/[0.12] border border-white/[0.1] text-slate-200 text-xs transition-colors"
                >
                  <Tag class="w-3 h-3 text-glacier-400/80" />
                  <span class="font-medium">{{ tag }}</span>
                  <button
                    type="button"
                    @click.stop="removeTag(index)"
                    class="p-0.5 rounded hover:bg-white/20 text-slate-400 hover:text-rose-300 transition-colors cursor-pointer"
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
                  placeholder="Type tag name and press Enter..."
                  @focus="showTagDropdown = true"
                  @input="showTagDropdown = true; selectedTagIndex = -1"
                  @blur="handleTagBlur"
                  @keydown="handleTagKeyDown"
                  class="flex-1 min-w-[150px] bg-transparent text-white text-xs placeholder-slate-500 outline-none py-1"
                />
              </div>

              <!-- Autocomplete Suggestions Dropdown -->
              <div
                v-if="showTagDropdown && filteredTagSuggestions.length > 0"
                class="absolute left-0 right-0 top-full mt-1.5 z-30 bg-[#161923] border border-white/[0.12] rounded-xl shadow-2xl shadow-black/80 max-h-48 overflow-y-auto py-1 backdrop-blur-md"
              >
                <div class="px-3 py-1 text-[10px] font-semibold text-slate-400 uppercase tracking-wider border-b border-white/[0.06] flex items-center justify-between">
                  <span>Available Categories</span>
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
                    {{ t.book_count }} books
                  </span>
                </button>
              </div>
            </div>

            <!-- Series and Index Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div class="sm:col-span-2">
                <label class="block text-slate-400 font-semibold mb-1">Series Name</label>
                <input
                  v-model="editForm.series"
                  type="text"
                  placeholder="e.g. The Dune Chronicles"
                  class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
                />
              </div>
              <div>
                <label class="block text-slate-400 font-semibold mb-1">Series #</label>
                <input
                  v-model.number="editForm.series_index"
                  type="number"
                  step="0.1"
                  min="0"
                  placeholder="1"
                  class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
                />
              </div>
            </div>

            <!-- Publisher, Date & Language Grid -->
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div>
                <label class="block text-slate-400 font-semibold mb-1">Publisher</label>
                <input
                  v-model="editForm.publisher"
                  type="text"
                  placeholder="e.g. O'Reilly"
                  class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
                />
              </div>
              <div>
                <label class="block text-slate-400 font-semibold mb-1">Publication Date</label>
                <input
                  v-model="editForm.pub_date"
                  type="text"
                  placeholder="YYYY-MM-DD"
                  class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors"
                />
              </div>
              <div>
                <label class="block text-slate-400 font-semibold mb-1">Language</label>
                <input
                  v-model="editForm.language"
                  type="text"
                  placeholder="en"
                  class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors uppercase"
                />
              </div>
            </div>

            <!-- Description Textarea -->
            <div>
              <label class="block text-slate-400 font-semibold mb-1">Description / Synopsis</label>
              <textarea
                v-model="editForm.description"
                rows="6"
                placeholder="Write a synopsis or overview of the book..."
                class="w-full px-3.5 py-2.5 rounded-xl bg-slate-900/80 border border-white/[0.1] text-white text-xs placeholder-slate-500 focus:outline-none focus:border-glacier-400 focus:ring-1 focus:ring-glacier-400 transition-colors leading-relaxed"
              ></textarea>
            </div>

            <!-- Form Actions -->
            <div class="flex items-center justify-end gap-3 pt-4 border-t border-white/[0.08]">
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
                class="flex items-center gap-2 px-6 py-2.5 rounded-xl bg-glacier-500 hover:bg-glacier-400 text-slate-950 text-xs font-bold shadow-md shadow-glacier-500/20 transition-all cursor-pointer disabled:opacity-50"
              >
                <Loader2 v-if="saving" class="w-4 h-4 animate-spin" />
                <Save v-else class="w-4 h-4" />
                <span>{{ saving ? 'Saving Changes...' : 'Save Changes' }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Shelf Select Modal -->
    <ShelfSelectModal
      v-if="showShelfModal && book"
      :book="book"
      @close="showShelfModal = false"
      @updated="$emit('shelf-updated', $event)"
    />

    <!-- Delete Confirmation Modal -->
    <div 
      v-if="showDeleteConfirm" 
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in"
      @click.self="showDeleteConfirm = false"
    >
      <div class="relative w-full max-w-sm bg-[#11131b] border border-white/[0.1] rounded-3xl p-6 text-center shadow-2xl shadow-black/90">
        <div class="w-12 h-12 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-400 flex items-center justify-center mx-auto mb-3">
          <Trash2 class="w-6 h-6" />
        </div>
        <h3 class="text-base font-bold text-white mb-1">Delete Book from Library?</h3>
        <p class="text-xs text-slate-400 mb-6 leading-relaxed">
          Are you sure you want to delete <span class="text-white font-medium">"{{ book?.title }}"</span>? This will permanently remove the file and all reading progress.
        </p>

        <div v-if="deleteError" class="mb-4 p-2.5 rounded-xl bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300">
          {{ deleteError }}
        </div>

        <div class="flex items-center justify-center gap-3">
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
            class="flex items-center gap-2 px-5 py-2 rounded-xl bg-rose-500 hover:bg-rose-400 text-white text-xs font-bold shadow-md shadow-rose-500/20 transition-all cursor-pointer disabled:opacity-50"
          >
            <Loader2 v-if="deleting" class="w-4 h-4 animate-spin" />
            <span>{{ deleting ? 'Deleting...' : 'Confirm Delete' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { 
  ArrowLeft, 
  ChevronRight, 
  Book, 
  BookOpen, 
  Download, 
  FileText, 
  Pencil, 
  Trash2, 
  Save, 
  Loader2, 
  AlertCircle, 
  Tag, 
  Bookmark, 
  Star, 
  X 
} from 'lucide-vue-next'
import { useAuth } from '../composables/useAuth'
import { 
  fetchBookDetail, 
  updateBookMetadata, 
  deleteBook, 
  fetchTags,
  setBookRating,
  deleteBookRating 
} from '../api/client'
import { useToast } from '../composables/useToast'
import ShelfSelectModal from './ShelfSelectModal.vue'

const props = defineProps({
  bookId: {
    type: [String, Number],
    required: true
  }
})

const emit = defineEmits(['read', 'filter-tag', 'filter-author', 'filter-series', 'shelf-updated', 'deleted'])

const router = useRouter()
const route = useRoute()
const { canRead, canDownload, canEdit, canDelete } = useAuth()
const { showToast } = useToast()

const book = ref(null)
const userRating = ref(0)
const avgRating = ref(0)
const hoverRating = ref(0)
const loading = ref(true)
const error = ref('')

const isEditing = ref(false)
const saving = ref(false)
const editError = ref('')
const showShelfModal = ref(false)
const showDeleteConfirm = ref(false)
const deleting = ref(false)
const deleteError = ref('')

const availableTags = ref([])
const tagInputText = ref('')
const showTagDropdown = ref(false)
const selectedTagIndex = ref(-1)
const tagInputRef = ref(null)

const editForm = reactive({
  title: '',
  authorsStr: '',
  tags: [],
  series: '',
  series_index: null,
  publisher: '',
  pub_date: '',
  language: '',
  description: ''
})

const formattedAuthorsList = computed(() => {
  if (!book.value || !book.value.authors) return ['Unknown Author']
  return book.value.authors.map(a => typeof a === 'string' ? a : a.name).filter(Boolean)
})

const filteredTagSuggestions = computed(() => {
  const currentTags = new Set(editForm.tags.map(t => t.toLowerCase()))
  const query = tagInputText.value.trim().toLowerCase()
  return availableTags.value
    .filter(t => !currentTags.has(t.name.toLowerCase()))
    .filter(t => !query || t.name.toLowerCase().includes(query))
    .slice(0, 8)
})

async function loadBookData() {
  loading.value = true
  error.value = ''
  try {
    const id = props.bookId || route.params.id
    const b = await fetchBookDetail(props.bookId)
    book.value = b
    userRating.value = b.user_rating || 0
    avgRating.value = b.avg_rating || 0
  } catch (err) {
    console.error('Failed to load book:', err)
    error.value = err.message || 'Book not found'
  } finally {
    loading.value = false
  }
}

async function loadAvailableTags() {
  try {
    const tags = await fetchTags()
    availableTags.value = tags || []
  } catch (err) {
    console.warn('Failed to load tags:', err)
  }
}

function goBack() {
  if (window.history.length > 1) {
    router.back()
  } else {
    router.push('/books')
  }
}

function navigateToTag(tagName) {
  emit('filter-tag', tagName)
  router.push(`/tags/${encodeURIComponent(tagName)}`)
}

function navigateToAuthor(authorName) {
  emit('filter-author', authorName)
  router.push(`/authors/${encodeURIComponent(authorName)}`)
}

function navigateToSeries(seriesName) {
  emit('filter-series', seriesName)
  router.push(`/series/${encodeURIComponent(seriesName)}`)
}

async function handleSetRating(star) {
  try {
    const res = await setBookRating(props.bookId, star)
    userRating.value = res.user_rating || star
    avgRating.value = res.avg_rating || star
    showToast(`Rated ${star} stars!`, 'success')
  } catch (err) {
    showToast('Failed to save rating', 'error')
  }
}

async function handleClearRating() {
  try {
    const res = await deleteBookRating(props.bookId)
    userRating.value = 0
    avgRating.value = res.avg_rating || 0
    showToast('Rating removed', 'info')
  } catch (err) {
    showToast('Failed to remove rating', 'error')
  }
}

function handleRead() {
  if (!canRead.value || !book.value) return
  emit('read', book.value)
  router.push(`/read/${book.value.id}`)
}

function formatAuthors(authors) {
  if (!authors || authors.length === 0) return 'Unknown Author'
  return authors.map(a => typeof a === 'string' ? a : a.name).filter(Boolean).join(', ')
}

function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  } catch {
    return dateStr
  }
}

function cleanDescription(desc) {
  if (!desc) return ''
  return desc.replace(/<[^>]*>?/gm, '').trim()
}

function startEditing() {
  if (!book.value) return
  isEditing.value = true
  editError.value = ''

  editForm.title = book.value.title || ''
  editForm.authorsStr = formattedAuthorsList.value.join(', ')
  editForm.tags = book.value.tags ? [...book.value.tags] : []
  editForm.series = book.value.series || ''
  editForm.series_index = book.value.series_index ?? null
  editForm.publisher = book.value.publisher || ''
  editForm.pub_date = book.value.pub_date || ''
  editForm.language = book.value.language || 'en'
  editForm.description = book.value.description || ''

  loadAvailableTags()
}

function cancelEditing() {
  isEditing.value = false
  editError.value = ''
}

function focusTagInput() {
  if (tagInputRef.value) {
    tagInputRef.value.focus()
  }
}

function addTag(tagName) {
  const trimmed = tagName.trim().replace(/^#+/, '')
  if (!trimmed) return
  const exists = editForm.tags.some(t => t.toLowerCase() === trimmed.toLowerCase())
  if (!exists) {
    editForm.tags.push(trimmed)
  }
  tagInputText.value = ''
  selectedTagIndex.value = -1
  nextTick(() => focusTagInput())
}

function removeTag(index) {
  editForm.tags.splice(index, 1)
}

function handleTagBlur() {
  setTimeout(() => {
    if (tagInputText.value.trim()) {
      addTag(tagInputText.value)
    }
    showTagDropdown.value = false
  }, 200)
}

function handleTagKeyDown(e) {
  if (e.key === 'Enter' || e.key === ',') {
    e.preventDefault()
    if (selectedTagIndex.value >= 0 && selectedTagIndex.value < filteredTagSuggestions.value.length) {
      addTag(filteredTagSuggestions.value[selectedTagIndex.value].name)
    } else if (tagInputText.value.trim()) {
      addTag(tagInputText.value)
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault()
    showTagDropdown.value = true
    if (filteredTagSuggestions.value.length > 0) {
      selectedTagIndex.value = (selectedTagIndex.value + 1) % filteredTagSuggestions.value.length
    }
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    showTagDropdown.value = true
    if (filteredTagSuggestions.value.length > 0) {
      selectedTagIndex.value = (selectedTagIndex.value - 1 + filteredTagSuggestions.value.length) % filteredTagSuggestions.value.length
    }
  } else if (e.key === 'Backspace' && !tagInputText.value && editForm.tags.length > 0) {
    removeTag(editForm.tags.length - 1)
  } else if (e.key === 'Escape') {
    showTagDropdown.value = false
  }
}

async function handleSaveMetadata() {
  if (!editForm.title.trim()) {
    editError.value = 'Title cannot be empty'
    return
  }

  saving.value = true
  editError.value = ''

  try {
    const authors = editForm.authorsStr
      .split(',')
      .map(a => a.trim())
      .filter(Boolean)

    const payload = {
      title: editForm.title.trim(),
      authors,
      tags: editForm.tags,
      series: editForm.series.trim(),
      series_index: editForm.series_index ? Number(editForm.series_index) : null,
      publisher: editForm.publisher.trim(),
      pub_date: editForm.pub_date.trim(),
      language: editForm.language.trim().toLowerCase(),
      description: editForm.description.trim()
    }

    const updated = await updateBookMetadata(book.value.id, payload)
    book.value = updated
    isEditing.value = false
    showToast('Metadata updated successfully', 'success')
  } catch (err) {
    console.error('Failed to save metadata:', err)
    editError.value = err.message || 'Failed to update metadata'
  } finally {
    saving.value = false
  }
}

async function handleDelete() {
  deleting.value = true
  deleteError.value = ''
  try {
    await deleteBook(book.value.id)
    showToast(`Deleted "${book.value.title}"`, 'info')
    emit('deleted', book.value.id)
    showDeleteConfirm.value = false
    router.replace('/books')
  } catch (err) {
    console.error('Failed to delete book:', err)
    deleteError.value = err.message || 'Failed to delete book'
  } finally {
    deleting.value = false
  }
}

watch(() => props.bookId, (newId) => {
  if (newId) {
    loadBookData()
  }
})

onMounted(() => {
  loadBookData()
})
</script>
