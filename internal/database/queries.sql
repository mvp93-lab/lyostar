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

-- name: SearchBooksFTS :many
SELECT b.*
FROM books b
JOIN books_fts ON b.id = books_fts.rowid
WHERE books_fts.fulltext MATCH ?
ORDER BY books_fts.rank
LIMIT ? OFFSET ?;
