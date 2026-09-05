package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

func TestAuthorsEndpointsAndFiltering(t *testing.T) {
	db, router, tempDir, booksDir, _ := setupBookManageTest(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	// Seed 2 books
	book1Path := filepath.Join(booksDir, "book1.epub")
	book2Path := filepath.Join(booksDir, "book2.epub")
	_ = os.WriteFile(book1Path, []byte("b1"), 0644)
	_ = os.WriteFile(book2Path, []byte("b2"), 0644)

	book1, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:      "Dune",
		FilePath:   book1Path,
		FileSha256: "dune_sha256",
		FileSize:   100,
		Format:     "epub",
	})
	if err != nil {
		t.Fatalf("failed to create book1: %v", err)
	}

	book2, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:      "Foundation",
		FilePath:   book2Path,
		FileSha256: "foundation_sha256",
		FileSize:   120,
		Format:     "epub",
	})
	if err != nil {
		t.Fatalf("failed to create book2: %v", err)
	}

	// Create Authors
	authorFrank, err := db.CreateAuthor(ctx, "Frank Herbert")
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}
	authorIsaac, err := db.CreateAuthor(ctx, "Isaac Asimov")
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	// Link Authors: Book1 -> Frank Herbert, Book2 -> Isaac Asimov
	_ = db.AddBookAuthor(ctx, database.AddBookAuthorParams{BookID: book1.ID, AuthorID: authorFrank.ID, Role: "aut"})
	_ = db.AddBookAuthor(ctx, database.AddBookAuthorParams{BookID: book2.ID, AuthorID: authorIsaac.ID, Role: "aut"})

	// Create a test user with session
	_, userToken := createTestUserWithPerms(t, db, "authorreader", auth.RoleReader, auth.Permissions{
		CanRead: true,
	})

	// 1. Test GET /api/authors
	t.Run("GET /api/authors returns list with book_count", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/authors", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var items []AuthorItem
		if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
			t.Fatalf("failed to unmarshal authors: %v", err)
		}

		if len(items) != 2 {
			t.Fatalf("expected 2 authors, got %d", len(items))
		}

		// Should be alphabetically ordered
		if items[0].Name != "Frank Herbert" || items[0].BookCount != 1 {
			t.Errorf("expected Frank Herbert with 1 book, got %+v", items[0])
		}
		if items[1].Name != "Isaac Asimov" || items[1].BookCount != 1 {
			t.Errorf("expected Isaac Asimov with 1 book, got %+v", items[1])
		}
	})

	// 2. Test GET /api/books?author=Frank%20Herbert
	t.Run("GET /api/books?author=Frank Herbert filters accurately", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/books?author=Frank%20Herbert", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp PaginatedResponse[BookListItem]
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal books: %v", err)
		}

		if resp.Total != 1 || len(resp.Items) != 1 {
			t.Fatalf("expected 1 book for Frank Herbert, got %d (total: %d)", len(resp.Items), resp.Total)
		}

		if resp.Items[0].Title != "Dune" {
			t.Errorf("expected Dune, got %s", resp.Items[0].Title)
		}
	})
}
