package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

func TestCustomShelvesEndpoints(t *testing.T) {
	db, router, tempDir, _, _ := setupBookManageTest(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	_, _ = createTestUserWithPerms(t, db, "admin_user", auth.RoleAdmin, auth.Permissions{
		CanRead: true, CanDownload: true, CanUpload: true, CanEdit: true, CanDelete: true,
	})

	_, token2 := createTestUserWithPerms(t, db, "reader_user", auth.RoleReader, auth.Permissions{
		CanRead: true, CanDownload: true,
	})

	// Seed 2 books
	book1, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:      "Dune",
		FilePath:   "/books/dune.epub",
		FileSha256: "sha256_dune",
		FileSize:   1024,
		Format:     "epub",
	})
	if err != nil {
		t.Fatalf("failed to create book1: %v", err)
	}

	book2, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:      "Neuromancer",
		FilePath:   "/books/neuromancer.epub",
		FileSha256: "sha256_neuromancer",
		FileSize:   2048,
		Format:     "epub",
	})
	if err != nil {
		t.Fatalf("failed to create book2: %v", err)
	}

	// 1. Unauthenticated request to /api/shelves must return 401
	{
		req := httptest.NewRequest(http.MethodGet, "/api/shelves", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 Unauthorized for unauth request, got %d", rec.Code)
		}
	}

	// 2. User2 creates a shelf "Sci-Fi Favorites"
	var shelfID int64
	{
		body, _ := json.Marshal(CreateShelfRequest{
			Name:        "Sci-Fi Favorites",
			Description: "My top favorite science fiction novels",
			IsPublic:    false,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/shelves", bytes.NewReader(body))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		var created ShelfItem
		if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
			t.Fatalf("failed to decode created shelf: %v", err)
		}
		if created.Name != "Sci-Fi Favorites" || !created.IsOwner {
			t.Fatalf("unexpected shelf fields: %+v", created)
		}
		shelfID = created.ID
	}

	// 3. User2 creates duplicate shelf name -> 409 Conflict
	{
		body, _ := json.Marshal(CreateShelfRequest{
			Name: "Sci-Fi Favorites",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/shelves", bytes.NewReader(body))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409 Conflict for duplicate shelf name, got %d", rec.Code)
		}
	}

	// 4. User2 adds book1 to shelf
	{
		body, _ := json.Marshal(AddBookToShelfRequest{BookID: book1.ID})
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/shelves/%d/books", shelfID), bytes.NewReader(body))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// 5. User2 queries books in shelf
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/shelves/%d/books", shelfID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var resp PaginatedResponse[BookListItem]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode shelf books: %v", err)
		}
		if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].ID != book1.ID {
			t.Fatalf("unexpected shelf books response: %+v", resp)
		}
	}

	// 6. Check book1's shelves endpoint: GET /api/books/{id}/shelves
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/shelves", book1.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var sids []int64
		if err := json.NewDecoder(rec.Body).Decode(&sids); err != nil {
			t.Fatalf("failed to decode shelf ids: %v", err)
		}
		if len(sids) != 1 || sids[0] != shelfID {
			t.Fatalf("expected [ %d ], got %v", shelfID, sids)
		}
	}

	// 7. Batch update shelves for book2 via PUT /api/books/{id}/shelves
	{
		body, _ := json.Marshal(UpdateBookShelvesRequest{
			ShelfIDs: []int64{shelfID},
		})
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/shelves", book2.ID), bytes.NewReader(body))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// Now shelf should have 2 books
	{
		count, err := db.CountBooksByShelf(t.Context(), shelfID)
		if err != nil || count != 2 {
			t.Fatalf("expected 2 books in shelf, got %d (err: %v)", count, err)
		}
	}

	// 8. Remove book1 from shelf via DELETE /api/shelves/{id}/books/{bookId}
	{
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/shelves/%d/books/%d", shelfID, book1.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		count, _ := db.CountBooksByShelf(t.Context(), shelfID)
		if count != 1 {
			t.Fatalf("expected 1 book remaining in shelf, got %d", count)
		}
	}

	// 9. Update shelf details
	{
		body, _ := json.Marshal(UpdateShelfRequest{
			Name:        "Cyberpunk & Sci-Fi",
			Description: "Updated description",
			IsPublic:    true,
		})
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/shelves/%d", shelfID), bytes.NewReader(body))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// 10. Delete shelf
	{
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/shelves/%d", shelfID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// Sách book2 vẫn nguyên vẹn sau khi xóa kệ
	{
		b, err := db.GetBookByID(t.Context(), book2.ID)
		if err != nil || b.Title != "Neuromancer" {
			t.Fatalf("book should remain intact after shelf deletion")
		}
	}
}
