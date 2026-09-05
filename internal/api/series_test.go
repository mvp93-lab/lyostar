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

func TestSeriesEndpointsAndFiltering(t *testing.T) {
	db, router, tempDir, booksDir, _ := setupBookManageTest(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	// Seed 3 books (2 in Dune series, 1 standalone)
	book1Path := filepath.Join(booksDir, "dune1.epub")
	book2Path := filepath.Join(booksDir, "dune2.epub")
	book3Path := filepath.Join(booksDir, "foundation.epub")
	_ = os.WriteFile(book1Path, []byte("d1"), 0644)
	_ = os.WriteFile(book2Path, []byte("d2"), 0644)
	_ = os.WriteFile(book3Path, []byte("f1"), 0644)

	_, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:       "Dune",
		FilePath:    book1Path,
		FileSha256:  "dune1_sha256",
		FileSize:    100,
		Format:      "epub",
		Series:      "Dune Chronicles",
		SeriesIndex: 1.0,
	})
	if err != nil {
		t.Fatalf("failed to create book1: %v", err)
	}

	_, err = db.CreateBook(ctx, database.CreateBookParams{
		Title:       "Dune Messiah",
		FilePath:    book2Path,
		FileSha256:  "dune2_sha256",
		FileSize:    110,
		Format:      "epub",
		Series:      "Dune Chronicles",
		SeriesIndex: 2.0,
	})
	if err != nil {
		t.Fatalf("failed to create book2: %v", err)
	}

	_, err = db.CreateBook(ctx, database.CreateBookParams{
		Title:       "Foundation",
		FilePath:    book3Path,
		FileSha256:  "found_sha256",
		FileSize:    120,
		Format:      "pdf",
		Series:      "",
		SeriesIndex: 0.0,
	})
	if err != nil {
		t.Fatalf("failed to create book3: %v", err)
	}

	_, userToken := createTestUserWithPerms(t, db, "seriesuser", auth.RoleReader, auth.Permissions{
		CanRead: true,
	})

	// 1. GET /api/series
	t.Run("GET /api/series lists series with book counts", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/series", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}

		var items []SeriesItem
		if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
			t.Fatalf("failed to decode series: %v", err)
		}

		if len(items) != 1 {
			t.Fatalf("expected 1 series, got %d", len(items))
		}
		if items[0].Name != "Dune Chronicles" || items[0].BookCount != 2 {
			t.Errorf("expected Dune Chronicles with 2 books, got %+v", items[0])
		}
	})

	// 2. GET /api/books?series=Dune%20Chronicles (sorted by series_index ASC)
	t.Run("GET /api/books?series=Dune Chronicles returns ordered volumes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/books?series=Dune%20Chronicles", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}

		var resp PaginatedResponse[BookListItem]
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total != 2 || len(resp.Items) != 2 {
			t.Fatalf("expected 2 books, got %d (total: %d)", len(resp.Items), resp.Total)
		}

		// Volume 1 should come before Volume 2
		if resp.Items[0].Title != "Dune" || resp.Items[0].SeriesIndex != 1.0 {
			t.Errorf("expected volume 1 to be Dune, got %+v", resp.Items[0])
		}
		if resp.Items[1].Title != "Dune Messiah" || resp.Items[1].SeriesIndex != 2.0 {
			t.Errorf("expected volume 2 to be Dune Messiah, got %+v", resp.Items[1])
		}
	})

	// 3. GET /api/books?format=pdf (Format filtering)
	t.Run("GET /api/books?format=pdf filters books by format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/books?format=pdf", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}

		var resp PaginatedResponse[BookListItem]
		if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Total != 1 || len(resp.Items) != 1 {
			t.Fatalf("expected 1 PDF book, got %d", len(resp.Items))
		}
		if resp.Items[0].Title != "Foundation" {
			t.Errorf("expected Foundation, got %s", resp.Items[0].Title)
		}
	})
}
