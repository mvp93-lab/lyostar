-- name: CreateBook :one
INSERT INTO books (
    title, file_path, file_sha256, file_size, format,
    description, publisher, language, pub_date, series,
    series_index, cover_path
) VALUES (
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?
)
RETURNING *;

-- name: GetBookByID :one
SELECT * FROM books
WHERE id = ? LIMIT 1;

-- name: GetBookBySHA256 :one
SELECT * FROM books
WHERE file_sha256 = ? LIMIT 1;

-- name: GetBookByFilePath :one
SELECT * FROM books
WHERE file_path = ? LIMIT 1;

-- name: ListBooks :many
SELECT * FROM books
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountBooks :one
SELECT COUNT(*) FROM books;

-- name: DeleteBook :exec
DELETE FROM books
WHERE id = ?;

-- name: UpdateBookMetadata :one
UPDATE books
SET title = ?,
    description = ?,
    publisher = ?,
    language = ?,
    pub_date = ?,
    series = ?,
    series_index = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: CreateAuthor :one
INSERT INTO authors (name)
VALUES (?)
ON CONFLICT(name) DO UPDATE SET name=excluded.name
RETURNING *;

-- name: GetAuthorByID :one
SELECT * FROM authors
WHERE id = ? LIMIT 1;

-- name: GetAuthorByName :one
SELECT * FROM authors
WHERE name = ? LIMIT 1;

-- name: ListAuthors :many
SELECT * FROM authors
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: AddBookAuthor :exec
INSERT INTO book_authors (book_id, author_id, role)
VALUES (?, ?, ?)
ON CONFLICT(book_id, author_id) DO UPDATE SET role=excluded.role;

-- name: ClearBookAuthors :exec
DELETE FROM book_authors
WHERE book_id = ?;

-- name: GetAuthorsForBook :many
SELECT a.*, ba.role
FROM authors a
JOIN book_authors ba ON a.id = ba.author_id
WHERE ba.book_id = ?
ORDER BY a.name ASC;

-- name: CreateTag :one
INSERT INTO tags (name)
VALUES (?)
ON CONFLICT(name) DO UPDATE SET name=excluded.name
RETURNING *;

-- name: GetTagByID :one
SELECT * FROM tags
WHERE id = ? LIMIT 1;

-- name: GetTagByName :one
SELECT * FROM tags
WHERE name = ? LIMIT 1;

-- name: ListTags :many
SELECT t.id, t.name, t.created_at, COUNT(bt.book_id) as book_count
FROM tags t
LEFT JOIN book_tags bt ON t.id = bt.tag_id
GROUP BY t.id
ORDER BY book_count DESC, t.name ASC;

-- name: AddBookTag :exec
INSERT INTO book_tags (book_id, tag_id)
VALUES (?, ?)
ON CONFLICT(book_id, tag_id) DO NOTHING;

-- name: ClearBookTags :exec
DELETE FROM book_tags
WHERE book_id = ?;

-- name: GetTagsForBook :many
SELECT t.*
FROM tags t
JOIN book_tags bt ON t.id = bt.tag_id
WHERE bt.book_id = ?
ORDER BY t.name ASC;

-- name: ListBooksWithAuthors :many
SELECT
    b.id, b.title, b.file_path, b.file_sha256, b.file_size, b.format,
    b.description, b.publisher, b.language, b.pub_date, b.series,
    b.series_index, b.cover_path, b.created_at, b.updated_at,
    coalesce(GROUP_CONCAT(a.name, ', '), '') as author_names
FROM books b
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
GROUP BY b.id
ORDER BY b.id DESC
LIMIT ? OFFSET ?;

-- name: SearchBooksFTS :many
SELECT b.*
FROM books b
JOIN books_fts ON b.id = books_fts.rowid
WHERE books_fts.fulltext MATCH ?
ORDER BY books_fts.rank
LIMIT ? OFFSET ?;

-- name: SearchBooksFTSWithAuthors :many
SELECT
    b.id, b.title, b.file_path, b.file_sha256, b.file_size, b.format,
    b.description, b.publisher, b.language, b.pub_date, b.series,
    b.series_index, b.cover_path, b.created_at, b.updated_at,
    coalesce(GROUP_CONCAT(a.name, ', '), '') as author_names
FROM books b
JOIN books_fts ON b.id = books_fts.rowid
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
WHERE books_fts.fulltext MATCH ?
GROUP BY b.id
ORDER BY books_fts.rank
LIMIT ? OFFSET ?;

-- name: CountSearchBooksFTS :one
SELECT COUNT(*)
FROM books b
JOIN books_fts ON b.id = books_fts.rowid
WHERE books_fts.fulltext MATCH ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CreateUser :one
INSERT INTO users (
    username, password_hash, role, display_name,
    can_read, can_download, can_upload, can_edit, can_delete
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT id, username, role, display_name,
       can_read, can_download, can_upload, can_edit, can_delete,
       created_at, updated_at
FROM users
ORDER BY id ASC;

-- name: UpdateUser :one
UPDATE users
SET display_name = ?,
    role = ?,
    can_read = ?,
    can_download = ?,
    can_upload = ?,
    can_edit = ?,
    can_delete = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password_hash = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;

-- name: CreateSession :one
INSERT INTO sessions (token, user_id, expires_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetSessionWithUser :one
SELECT s.token, s.user_id, s.expires_at, u.username, u.role, u.display_name,
       u.can_read, u.can_download, u.can_upload, u.can_edit, u.can_delete
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.token = ? AND s.expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= CURRENT_TIMESTAMP;

-- name: GetProgress :one
SELECT * FROM reading_progress
WHERE user_id = ? AND book_id = ? LIMIT 1;

-- name: UpsertProgress :one
INSERT INTO reading_progress (
    user_id, book_id, location, progress, current_page, total_pages, is_finished, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
ON CONFLICT(user_id, book_id) DO UPDATE SET
    location = excluded.location,
    progress = excluded.progress,
    current_page = excluded.current_page,
    total_pages = excluded.total_pages,
    is_finished = excluded.is_finished,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: ListRecentProgressByUserID :many
SELECT
    rp.user_id, rp.book_id, rp.location, rp.progress, rp.current_page, rp.total_pages, rp.is_finished, rp.updated_at,
    b.title, b.file_path, b.file_size, b.format, b.cover_path,
    coalesce(GROUP_CONCAT(a.name, ', '), '') as author_names
FROM reading_progress rp
JOIN books b ON rp.book_id = b.id
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
WHERE rp.user_id = ? AND rp.is_finished = 0 AND rp.progress > 0
GROUP BY b.id
ORDER BY rp.updated_at DESC
LIMIT ?;

-- name: ListBooksWithAuthorsAndProgress :many
SELECT
    b.id, b.title, b.file_path, b.file_sha256, b.file_size, b.format,
    b.description, b.publisher, b.language, b.pub_date, b.series,
    b.series_index, b.cover_path, b.created_at, b.updated_at,
    coalesce((SELECT GROUP_CONCAT(a.name, ', ') FROM book_authors ba JOIN authors a ON ba.author_id = a.id WHERE ba.book_id = b.id), '') as author_names,
    coalesce((SELECT GROUP_CONCAT(t.name, ', ') FROM book_tags bt JOIN tags t ON bt.tag_id = t.id WHERE bt.book_id = b.id), '') as tag_names,
    coalesce(rp.progress, 0.0) as user_progress,
    coalesce(rp.is_finished, 0) as user_is_finished
FROM books b
LEFT JOIN reading_progress rp ON b.id = rp.book_id AND rp.user_id = ?
GROUP BY b.id
ORDER BY b.id DESC
LIMIT ? OFFSET ?;

-- name: SearchBooksFTSWithAuthorsAndProgress :many
SELECT
    b.id, b.title, b.file_path, b.file_sha256, b.file_size, b.format,
    b.description, b.publisher, b.language, b.pub_date, b.series,
    b.series_index, b.cover_path, b.created_at, b.updated_at,
    coalesce((SELECT GROUP_CONCAT(a.name, ', ') FROM book_authors ba JOIN authors a ON ba.author_id = a.id WHERE ba.book_id = b.id), '') as author_names,
    coalesce((SELECT GROUP_CONCAT(t.name, ', ') FROM book_tags bt JOIN tags t ON bt.tag_id = t.id WHERE bt.book_id = b.id), '') as tag_names,
    coalesce(rp.progress, 0.0) as user_progress,
    coalesce(rp.is_finished, 0) as user_is_finished
FROM books b
JOIN books_fts ON b.id = books_fts.rowid
LEFT JOIN reading_progress rp ON b.id = rp.book_id AND rp.user_id = ?
WHERE books_fts.fulltext MATCH ?
GROUP BY b.id
ORDER BY books_fts.rank
LIMIT ? OFFSET ?;

-- name: ListBooksByTagWithAuthorsAndProgress :many
SELECT
    b.id, b.title, b.file_path, b.file_sha256, b.file_size, b.format,
    b.description, b.publisher, b.language, b.pub_date, b.series,
    b.series_index, b.cover_path, b.created_at, b.updated_at,
    coalesce((SELECT GROUP_CONCAT(a.name, ', ') FROM book_authors ba JOIN authors a ON ba.author_id = a.id WHERE ba.book_id = b.id), '') as author_names,
    coalesce((SELECT GROUP_CONCAT(t.name, ', ') FROM book_tags bt JOIN tags t ON bt.tag_id = t.id WHERE bt.book_id = b.id), '') as tag_names,
    coalesce(rp.progress, 0.0) as user_progress,
    coalesce(rp.is_finished, 0) as user_is_finished
FROM books b
JOIN book_tags bt ON b.id = bt.book_id
JOIN tags t ON bt.tag_id = t.id
LEFT JOIN reading_progress rp ON b.id = rp.book_id AND rp.user_id = ?
WHERE t.name = ?
GROUP BY b.id
ORDER BY b.id DESC
LIMIT ? OFFSET ?;

-- name: CountBooksByTag :one
SELECT COUNT(DISTINCT b.id)
FROM books b
JOIN book_tags bt ON b.id = bt.book_id
JOIN tags t ON bt.tag_id = t.id
WHERE t.name = ?;

-- name: CreateShelf :one
INSERT INTO shelves (user_id, name, description, is_public)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetShelfByID :one
SELECT * FROM shelves
WHERE id = ? LIMIT 1;

-- name: GetShelfByNameAndUser :one
SELECT * FROM shelves
WHERE user_id = ? AND name = ? LIMIT 1;

-- name: ListShelvesForUser :many
SELECT 
    s.id, s.user_id, s.name, s.description, s.is_public, s.created_at, s.updated_at,
    u.username as owner_username,
    u.display_name as owner_display_name,
    COUNT(sb.book_id) as book_count
FROM shelves s
JOIN users u ON s.user_id = u.id
LEFT JOIN shelf_books sb ON s.id = sb.shelf_id
WHERE s.user_id = ? OR s.is_public = 1
GROUP BY s.id
ORDER BY s.is_public ASC, s.name ASC;

-- name: UpdateShelf :one
UPDATE shelves
SET name = ?,
    description = ?,
    is_public = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: DeleteShelf :exec
DELETE FROM shelves
WHERE id = ? AND user_id = ?;

-- name: AddBookToShelf :exec
INSERT OR IGNORE INTO shelf_books (shelf_id, book_id)
VALUES (?, ?);

-- name: RemoveBookFromShelf :exec
DELETE FROM shelf_books
WHERE shelf_id = ? AND book_id = ?;

-- name: GetBookShelfIDsForUser :many
SELECT sb.shelf_id
FROM shelf_books sb
JOIN shelves s ON sb.shelf_id = s.id
WHERE sb.book_id = ? AND s.user_id = ?;

-- name: ListBooksByShelfWithAuthorsAndProgress :many
SELECT
    b.id, b.title, b.file_path, b.file_sha256, b.file_size, b.format,
    b.description, b.publisher, b.language, b.pub_date, b.series,
    b.series_index, b.cover_path, b.created_at, b.updated_at,
    coalesce((SELECT GROUP_CONCAT(a.name, ', ') FROM book_authors ba JOIN authors a ON ba.author_id = a.id WHERE ba.book_id = b.id), '') as author_names,
    coalesce((SELECT GROUP_CONCAT(t.name, ', ') FROM book_tags bt JOIN tags t ON bt.tag_id = t.id WHERE bt.book_id = b.id), '') as tag_names,
    coalesce(rp.progress, 0.0) as user_progress,
    coalesce(rp.is_finished, 0) as user_is_finished,
    sb.added_at as shelf_added_at
FROM shelf_books sb
JOIN books b ON sb.book_id = b.id
LEFT JOIN reading_progress rp ON b.id = rp.book_id AND rp.user_id = ?
WHERE sb.shelf_id = ?
ORDER BY sb.added_at DESC
LIMIT ? OFFSET ?;

-- name: CountBooksByShelf :one
SELECT COUNT(*)
FROM shelf_books
WHERE shelf_id = ?;

-- Bookmarks queries
-- name: ListBookmarksForUserAndBook :many
SELECT * FROM bookmarks
WHERE user_id = ? AND book_id = ?
ORDER BY created_at DESC;

-- name: GetBookmarkByID :one
SELECT * FROM bookmarks
WHERE id = ? LIMIT 1;

-- name: CreateBookmark :one
INSERT INTO bookmarks (
    user_id, book_id, title, location, progress, created_at
) VALUES (
    ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: DeleteBookmark :exec
DELETE FROM bookmarks
WHERE id = ? AND user_id = ?;

-- Highlights & Notes queries
-- name: ListHighlightsForUserAndBook :many
SELECT * FROM highlights
WHERE user_id = ? AND book_id = ?
ORDER BY created_at DESC;

-- name: GetHighlightByID :one
SELECT * FROM highlights
WHERE id = ? LIMIT 1;

-- name: CreateHighlight :one
INSERT INTO highlights (
    user_id, book_id, location, selected_text, note, color, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
RETURNING *;

-- name: UpdateHighlight :one
UPDATE highlights
SET note = ?,
    color = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND user_id = ?
RETURNING *;

-- name: DeleteHighlight :exec
DELETE FROM highlights
WHERE id = ? AND user_id = ?;


