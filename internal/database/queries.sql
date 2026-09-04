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
INSERT INTO users (username, password_hash, role, display_name)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = ? LIMIT 1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = ? LIMIT 1;

-- name: ListUsers :many
SELECT id, username, role, display_name, created_at, updated_at
FROM users
ORDER BY id ASC;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;

-- name: CreateSession :one
INSERT INTO sessions (token, user_id, expires_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetSessionWithUser :one
SELECT s.token, s.user_id, s.expires_at, u.username, u.role, u.display_name
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

