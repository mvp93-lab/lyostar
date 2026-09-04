package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseInitAndPragmas(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lyostar-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Verify journal mode is WAL
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode;").Scan(&journalMode); err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected journal_mode 'wal', got %q", journalMode)
	}

	// Verify foreign keys is ON (1)
	var foreignKeys int
	if err := db.QueryRow("PRAGMA foreign_keys;").Scan(&foreignKeys); err != nil {
		t.Fatalf("failed to query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("expected foreign_keys 1, got %d", foreignKeys)
	}

	// Verify CRUD using sqlc
	ctx := context.Background()
	author, err := db.CreateAuthor(ctx, "Arthur Conan Doyle")
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	book, err := db.CreateBook(ctx, CreateBookParams{
		Title:       "A Study in Scarlet",
		FilePath:    "/books/study_in_scarlet.epub",
		FileSha256:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		FileSize:    123456,
		Format:      "epub",
		Description: "The first Sherlock Holmes adventure.",
		Publisher:   "Ward Lock & Co",
		Language:    "en",
		PubDate:     "1887",
		Series:      "Sherlock Holmes",
		SeriesIndex: 1,
		CoverPath:   "/data/cache/covers/e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855.webp",
	})
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	if err := db.AddBookAuthor(ctx, AddBookAuthorParams{
		BookID:   book.ID,
		AuthorID: author.ID,
		Role:     "aut",
	}); err != nil {
		t.Fatalf("failed to link book author: %v", err)
	}

	// Test FTS5 search
	results, err := db.SearchBooksFTS(ctx, SearchBooksFTSParams{
		Fulltext: "Sherlock",
		Limit:    10,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("failed to execute FTS5 search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 FTS5 result, got %d", len(results))
	}
	if results[0].Title != "A Study in Scarlet" {
		t.Errorf("expected book title 'A Study in Scarlet', got %q", results[0].Title)
	}
}
