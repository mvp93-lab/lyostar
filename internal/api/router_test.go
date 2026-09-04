package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/lyostar/lyostar/internal/database"
	"github.com/lyostar/lyostar/internal/scanner"
)

func setupTestDBAndRouter(t *testing.T) (*database.DB, *scanner.Scanner, http.Handler, string) {
	tempDir, err := os.MkdirTemp("", "lyostar-api-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	booksDir := filepath.Join(tempDir, "books")
	dataDir := filepath.Join(tempDir, "data")
	coversDir := filepath.Join(dataDir, "cache", "covers")

	_ = os.MkdirAll(booksDir, 0755)
	_ = os.MkdirAll(coversDir, 0755)

	dbPath := filepath.Join(dataDir, "app.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	sc := scanner.New(booksDir, coversDir, db)

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<html><body>Lyostar SPA</body></html>"),
		},
	}

	router := NewRouter(RouterConfig{
		DB:       db,
		Scanner:  sc,
		StaticFS: mockFS,
		Version:  "1.0.0-test",
	})

	return db, sc, router, tempDir
}

func TestRouterHealth(t *testing.T) {
	db, sc, router, tempDir := setupTestDBAndRouter(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()
	_ = sc

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" || resp.Version != "1.0.0-test" {
		t.Errorf("unexpected health response: %+v", resp)
	}
}

func TestRouterBooksEndpoints(t *testing.T) {
	db, _, router, tempDir := setupTestDBAndRouter(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	// Seed dummy book and dummy files
	fakeCoverPath := filepath.Join(tempDir, "data", "cache", "covers", "fake_cover.webp")
	_ = os.WriteFile(fakeCoverPath, []byte("RIFF....WEBPVP8 ..."), 0644)

	fakeEpubPath := filepath.Join(tempDir, "books", "test_book.epub")
	_ = os.WriteFile(fakeEpubPath, []byte("PK\x03\x04fake-epub-content"), 0644)

	author, err := db.CreateAuthor(ctx, "Sherlock Holmes Author")
	if err != nil {
		t.Fatalf("failed to create author: %v", err)
	}

	book, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:       "The Sign of the Four",
		FilePath:    fakeEpubPath,
		FileSha256:  "abcd1234efgh5678",
		FileSize:    1024,
		Format:      "epub",
		Description: "A thrilling adventure",
		Publisher:   "Spencer Blackett",
		Language:    "en",
		PubDate:     "1890",
		Series:      "Sherlock Holmes",
		SeriesIndex: 2,
		CoverPath:   fakeCoverPath,
	})
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	_ = db.AddBookAuthor(ctx, database.AddBookAuthorParams{
		BookID:   book.ID,
		AuthorID: author.ID,
		Role:     "aut",
	})

	// 1. Test GET /api/books
	{
		req := httptest.NewRequest(http.MethodGet, "/api/books?page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for /api/books, got %d: %s", rec.Code, rec.Body.String())
		}

		var res PaginatedResponse[BookListItem]
		if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
			t.Fatalf("failed to decode /api/books response: %v", err)
		}

		if res.Total != 1 || len(res.Items) != 1 {
			t.Fatalf("expected 1 item and total 1, got total %d, items %d", res.Total, len(res.Items))
		}
		if res.Items[0].Title != "The Sign of the Four" {
			t.Errorf("expected book title 'The Sign of the Four', got %q", res.Items[0].Title)
		}
		if len(res.Items[0].Authors) != 1 || res.Items[0].Authors[0] != "Sherlock Holmes Author" {
			t.Errorf("unexpected authors: %+v", res.Items[0].Authors)
		}
		if !res.Items[0].HasCover || res.Items[0].CoverURL == "" {
			t.Errorf("expected cover url, got %+v", res.Items[0])
		}
	}

	// 2. Test GET /api/books/{id}
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d", book.ID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for book detail, got %d", rec.Code)
		}

		var detail BookDetailResponse
		if err := json.NewDecoder(rec.Body).Decode(&detail); err != nil {
			t.Fatalf("failed to decode detail response: %v", err)
		}
		if detail.Title != "The Sign of the Four" {
			t.Errorf("unexpected title in detail: %q", detail.Title)
		}
		if len(detail.Authors) != 1 || detail.Authors[0].Name != "Sherlock Holmes Author" {
			t.Errorf("unexpected author in detail: %+v", detail.Authors)
		}

		// 404 for invalid ID
		req404 := httptest.NewRequest(http.MethodGet, "/api/books/9999", nil)
		rec404 := httptest.NewRecorder()
		router.ServeHTTP(rec404, req404)
		if rec404.Code != http.StatusNotFound {
			t.Errorf("expected 404 for non-existent book, got %d", rec404.Code)
		}
	}

	// 3. Test GET /api/books/{id}/cover
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/cover", book.ID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for cover, got %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "image/webp" {
			t.Errorf("expected Content-Type image/webp, got %q", ct)
		}
	}

	// 4. Test GET /api/books/{id}/file
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/file", book.ID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for file, got %d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/epub+zip" {
			t.Errorf("expected Content-Type application/epub+zip, got %q", ct)
		}
	}

	// 5. Test GET /api/search
	{
		req := httptest.NewRequest(http.MethodGet, "/api/search?q=Sign", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for search, got %d", rec.Code)
		}

		var searchRes PaginatedResponse[BookListItem]
		if err := json.NewDecoder(rec.Body).Decode(&searchRes); err != nil {
			t.Fatalf("failed to decode search response: %v", err)
		}
		if searchRes.Total != 1 || len(searchRes.Items) != 1 {
			t.Fatalf("expected 1 search result, got %d items (total: %d)", len(searchRes.Items), searchRes.Total)
		}

		// Empty search query
		reqEmpty := httptest.NewRequest(http.MethodGet, "/api/search?q=", nil)
		recEmpty := httptest.NewRecorder()
		router.ServeHTTP(recEmpty, reqEmpty)
		if recEmpty.Code != http.StatusOK {
			t.Errorf("expected 200 for empty search, got %d", recEmpty.Code)
		}
	}

	// 6. Test POST /api/scan
	{
		req := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Fatalf("expected 202 for /api/scan, got %d", rec.Code)
		}
	}
}

func TestRouterSPAFallback(t *testing.T) {
	db, _, router, tempDir := setupTestDBAndRouter(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	// 1. Root route serves index.html
	reqRoot := httptest.NewRequest(http.MethodGet, "/", nil)
	recRoot := httptest.NewRecorder()
	router.ServeHTTP(recRoot, reqRoot)
	if recRoot.Code != http.StatusOK {
		t.Errorf("expected 200 for root, got %d", recRoot.Code)
	}

	// 2. Unmatched non-API route falls back to index.html (SPA routing)
	reqSPA := httptest.NewRequest(http.MethodGet, "/books/42/read", nil)
	recSPA := httptest.NewRecorder()
	router.ServeHTTP(recSPA, reqSPA)
	if recSPA.Code != http.StatusOK {
		t.Errorf("expected 200 for SPA route, got %d", recSPA.Code)
	}
	if recSPA.Body.String() != "<html><body>Lyostar SPA</body></html>" {
		t.Errorf("expected index.html body, got %q", recSPA.Body.String())
	}

	// 3. Unmatched API route returns 404 JSON
	reqAPI404 := httptest.NewRequest(http.MethodGet, "/api/unknown-endpoint", nil)
	recAPI404 := httptest.NewRecorder()
	router.ServeHTTP(recAPI404, reqAPI404)
	if recAPI404.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown api route, got %d", recAPI404.Code)
	}
}
