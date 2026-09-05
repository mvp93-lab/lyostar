export async function fetchBooks({ page = 1, limit = 24, tag = '', author = '' } = {}) {
  let url = `/api/books?page=${page}&limit=${limit}`
  if (tag) {
    url += `&tag=${encodeURIComponent(tag)}`
  }
  if (author) {
    url += `&author=${encodeURIComponent(author)}`
  }
  const res = await fetch(url)
  if (!res.ok) throw new Error(`Failed to fetch books: ${res.statusText}`)
  return res.json()
}

export async function fetchTags() {
  const res = await fetch('/api/tags')
  if (!res.ok) throw new Error(`Failed to fetch tags: ${res.statusText}`)
  return res.json()
}

export async function fetchAuthors() {
  const res = await fetch('/api/authors')
  if (!res.ok) throw new Error(`Failed to fetch authors: ${res.statusText}`)
  return res.json()
}

export async function searchBooks({ q = '', page = 1, limit = 24 } = {}) {
  const res = await fetch(`/api/search?q=${encodeURIComponent(q)}&page=${page}&limit=${limit}`)
  if (!res.ok) throw new Error(`Search failed: ${res.statusText}`)
  return res.json()
}

export async function fetchBookDetail(id) {
  const res = await fetch(`/api/books/${id}`)
  if (!res.ok) throw new Error(`Book not found`)
  return res.json()
}

export async function triggerScan() {
  const res = await fetch('/api/scan', { method: 'POST' })
  if (!res.ok) throw new Error(`Failed to trigger scan`)
  return res.json()
}

export async function fetchBookProgress(bookId) {
  const res = await fetch(`/api/books/${bookId}/progress`)
  if (!res.ok) throw new Error(`Failed to fetch reading progress`)
  return res.json()
}

export async function saveBookProgress(bookId, { location = '', progress = 0, currentPage = 0, totalPages = 0, isFinished = false } = {}) {
  const res = await fetch(`/api/books/${bookId}/progress`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      location,
      progress,
      current_page: currentPage,
      total_pages: totalPages,
      is_finished: isFinished
    })
  })
  if (!res.ok) throw new Error(`Failed to save reading progress`)
  return res.json()
}

export async function fetchContinueReading(limit = 12) {
  const res = await fetch(`/api/books/continue-reading?limit=${limit}`)
  if (!res.ok) throw new Error(`Failed to fetch continue reading list`)
  return res.json()
}

export async function uploadBook(file) {
  const formData = new FormData()
  formData.append('file', file)

  const res = await fetch('/api/books/upload', {
    method: 'POST',
    body: formData
  })

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || `Upload failed with status ${res.status}`)
  }
  return data
}

export async function updateBookMetadata(bookId, payload) {
  const res = await fetch(`/api/books/${bookId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  })

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || `Update failed with status ${res.status}`)
  }
  return data
}

export async function deleteBook(bookId) {
  const res = await fetch(`/api/books/${bookId}`, {
    method: 'DELETE'
  })

  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error(data.error || `Delete failed with status ${res.status}`)
  }
  return data
}

export async function fetchShelves() {
  const res = await fetch('/api/shelves')
  if (!res.ok) throw new Error(`Failed to fetch shelves`)
  return res.json()
}

export async function createShelf({ name, description = '', is_public = false }) {
  const res = await fetch('/api/shelves', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, description, is_public })
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || 'Failed to create shelf')
  return data
}

export async function updateShelf(id, { name, description = '', is_public = false }) {
  const res = await fetch(`/api/shelves/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, description, is_public })
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || 'Failed to update shelf')
  return data
}

export async function deleteShelf(id) {
  const res = await fetch(`/api/shelves/${id}`, {
    method: 'DELETE'
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || 'Failed to delete shelf')
  return data
}

export async function fetchShelfBooks(shelfId, { page = 1, limit = 24 } = {}) {
  const res = await fetch(`/api/shelves/${shelfId}/books?page=${page}&limit=${limit}`)
  if (!res.ok) throw new Error('Failed to fetch books in shelf')
  return res.json()
}

export async function addBookToShelf(shelfId, bookId) {
  const res = await fetch(`/api/shelves/${shelfId}/books`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ book_id: bookId })
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || 'Failed to add book to shelf')
  return data
}

export async function removeBookFromShelf(shelfId, bookId) {
  const res = await fetch(`/api/shelves/${shelfId}/books/${bookId}`, {
    method: 'DELETE'
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || 'Failed to remove book from shelf')
  return data
}

export async function fetchBookShelves(bookId) {
  const res = await fetch(`/api/books/${bookId}/shelves`)
  if (!res.ok) throw new Error('Failed to fetch book shelves')
  return res.json()
}

export async function updateBookShelves(bookId, shelfIds) {
  const res = await fetch(`/api/books/${bookId}/shelves`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ shelf_ids: shelfIds })
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || 'Failed to update book shelves')
  return data
}

// Bookmarks API
export async function fetchBookmarks(bookId) {
  const res = await fetch(`/api/books/${bookId}/bookmarks`)
  if (!res.ok) throw new Error('Failed to fetch bookmarks')
  return res.json()
}

export async function createBookmark(bookId, data) {
  const res = await fetch(`/api/books/${bookId}/bookmarks`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(json.error || 'Failed to create bookmark')
  return json
}

export async function deleteBookmark(bookId, bookmarkId) {
  const res = await fetch(`/api/books/${bookId}/bookmarks/${bookmarkId}`, {
    method: 'DELETE'
  })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(json.error || 'Failed to delete bookmark')
  return json
}

// Highlights & Notes API
export async function fetchHighlights(bookId) {
  const res = await fetch(`/api/books/${bookId}/highlights`)
  if (!res.ok) throw new Error('Failed to fetch highlights')
  return res.json()
}

export async function createHighlight(bookId, data) {
  const res = await fetch(`/api/books/${bookId}/highlights`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(json.error || 'Failed to create highlight')
  return json
}

export async function updateHighlight(bookId, highlightId, data) {
  const res = await fetch(`/api/books/${bookId}/highlights/${highlightId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(json.error || 'Failed to update highlight')
  return json
}

export async function deleteHighlight(bookId, highlightId) {
  const res = await fetch(`/api/books/${bookId}/highlights/${highlightId}`, {
    method: 'DELETE'
  })
  const json = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(json.error || 'Failed to delete highlight')
  return json
}


