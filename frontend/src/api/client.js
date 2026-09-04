export async function fetchBooks({ page = 1, limit = 24 } = {}) {
  const res = await fetch(`/api/books?page=${page}&limit=${limit}`)
  if (!res.ok) throw new Error(`Failed to fetch books: ${res.statusText}`)
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
