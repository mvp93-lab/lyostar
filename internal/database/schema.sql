CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    file_path TEXT NOT NULL UNIQUE,
    file_sha256 TEXT NOT NULL UNIQUE,
    file_size INTEGER NOT NULL,
    format TEXT NOT NULL DEFAULT 'epub',
    description TEXT NOT NULL DEFAULT '',
    publisher TEXT NOT NULL DEFAULT '',
    language TEXT NOT NULL DEFAULT '',
    pub_date TEXT NOT NULL DEFAULT '',
    series TEXT NOT NULL DEFAULT '',
    series_index REAL NOT NULL DEFAULT 0,
    cover_path TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS authors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS book_authors (
    book_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'aut',
    PRIMARY KEY (book_id, author_id),
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES authors(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_books_sha256 ON books(file_sha256);
CREATE INDEX IF NOT EXISTS idx_books_file_path ON books(file_path);
CREATE INDEX IF NOT EXISTS idx_book_authors_author_id ON book_authors(author_id);

-- FTS5 virtual table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS books_fts USING fts5(
    fulltext,
    title,
    series,
    description,
    content='books',
    content_rowid='id'
);

-- Triggers for FTS5 synchronization with books table
CREATE TRIGGER IF NOT EXISTS books_ai AFTER INSERT ON books BEGIN
    INSERT INTO books_fts(rowid, fulltext, title, series, description)
    VALUES (
        new.id,
        new.title || ' ' || coalesce(new.series, '') || ' ' || coalesce(new.description, ''),
        new.title,
        new.series,
        new.description
    );
END;

CREATE TRIGGER IF NOT EXISTS books_ad AFTER DELETE ON books BEGIN
    INSERT INTO books_fts(books_fts, rowid, fulltext, title, series, description)
    VALUES (
        'delete',
        old.id,
        old.title || ' ' || coalesce(old.series, '') || ' ' || coalesce(old.description, ''),
        old.title,
        old.series,
        old.description
    );
END;

CREATE TRIGGER IF NOT EXISTS books_au AFTER UPDATE ON books BEGIN
    INSERT INTO books_fts(books_fts, rowid, fulltext, title, series, description)
    VALUES (
        'delete',
        old.id,
        old.title || ' ' || coalesce(old.series, '') || ' ' || coalesce(old.description, ''),
        old.title,
        old.series,
        old.description
    );
    INSERT INTO books_fts(rowid, fulltext, title, series, description)
    VALUES (
        new.id,
        new.title || ' ' || coalesce(new.series, '') || ' ' || coalesce(new.description, ''),
        new.title,
        new.series,
        new.description
    );
END;
