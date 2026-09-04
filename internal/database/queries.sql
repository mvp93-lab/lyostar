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
    coalesce(GROUP_CONCAT(a.name, ', '), '') as author_names,
    coalesce(rp.progress, 0.0) as user_progress,
    coalesce(rp.is_finished, 0) as user_is_finished
FROM books b
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
LEFT JOIN reading_progress rp ON b.id = rp.book_id AND rp.user_id = ?
GROUP BY b.id
ORDER BY b.id DESC
LIMIT ? OFFSET ?;

-- name: SearchBooksFTSWithAuthorsAndProgress :many
SELECT
    b.id, b.title, b.file_path, b.file_sha256, b.file_size, b.format,
    b.description, b.publisher, b.language, b.pub_date, b.series,
    b.series_index, b.cover_path, b.created_at, b.updated_at,
    coalesce(GROUP_CONCAT(a.name, ', '), '') as author_names,
    coalesce(rp.progress, 0.0) as user_progress,
    coalesce(rp.is_finished, 0) as user_is_finished
FROM books b
JOIN books_fts ON b.id = books_fts.rowid
LEFT JOIN book_authors ba ON b.id = ba.book_id
LEFT JOIN authors a ON ba.author_id = a.id
LEFT JOIN reading_progress rp ON b.id = rp.book_id AND rp.user_id = ?
WHERE books_fts.fulltext MATCH ?
GROUP BY b.id
ORDER BY books_fts.rank
LIMIT ? OFFSET ?;

