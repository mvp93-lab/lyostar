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
