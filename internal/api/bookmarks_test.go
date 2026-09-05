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

func TestBookmarksAndHighlightsEndpoints(t *testing.T) {
	db, router, tempDir, _, _ := setupBookManageTest(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	_, token1 := createTestUserWithPerms(t, db, "reader_one", auth.RoleReader, auth.Permissions{
		CanRead: true, CanDownload: true,
	})

	_, token2 := createTestUserWithPerms(t, db, "reader_two", auth.RoleReader, auth.Permissions{
		CanRead: true, CanDownload: true,
	})

	_, tokenNoRead := createTestUserWithPerms(t, db, "reader_noread", auth.RoleReader, auth.Permissions{
		CanRead: false,
	})

	// Seed test book
	book, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:      "Hyperion",
		FilePath:   "/books/hyperion.epub",
		FileSha256: "sha256_hyperion",
		FileSize:   1024,
		Format:     "epub",
	})
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	// 1. Unauthenticated request returns 401
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/bookmarks", book.ID), nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	}

	// 2. User without can_read returns 403
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/bookmarks", book.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: tokenNoRead})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for user without can_read, got %d", rec.Code)
		}
	}

	// 3. Create bookmark for reader_one
	var bookmarkID int64
	{
		body := `{"title":"Chapter 1: The Priest's Tale","location":"epubcfi(/6/4[chap1]!/4/2/10)","progress":0.15}`
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/books/%d/bookmarks", book.ID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token1})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		var created BookmarkResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if created.Title != "Chapter 1: The Priest's Tale" || created.Progress != 0.15 {
			t.Fatalf("unexpected bookmark data: %+v", created)
		}
		bookmarkID = created.ID
	}

	// 4. List bookmarks for reader_one
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/bookmarks", book.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token1})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var bookmarks []BookmarkResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &bookmarks); err != nil {
			t.Fatalf("failed to parse bookmarks: %v", err)
		}
		if len(bookmarks) != 1 || bookmarks[0].ID != bookmarkID {
			t.Fatalf("expected 1 bookmark with ID %d, got %v", bookmarkID, bookmarks)
		}
	}

	// 5. Reader two lists bookmarks (should be empty for reader_two)
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/bookmarks", book.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		var bookmarks []BookmarkResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &bookmarks)
		if len(bookmarks) != 0 {
			t.Fatalf("expected 0 bookmarks for reader_two, got %d", len(bookmarks))
		}
	}

	// 6. Reader two tries to delete reader one's bookmark -> 403 Forbidden
	{
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d/bookmarks/%d", book.ID, bookmarkID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden when deleting other user's bookmark, got %d", rec.Code)
		}
	}

	// 7. Reader one deletes their bookmark -> 200 OK
	{
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d/bookmarks/%d", book.ID, bookmarkID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token1})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
	}

	// ==================== HIGHLIGHTS & NOTES TESTS ====================

	var highlightID int64
	// 8. Create highlight
	{
		body := `{"location":"epubcfi(/6/6!/4/2/4)","selected_text":"The Hegemony of Man spans hundreds of worlds.","note":"Key world-building concept","color":"blue"}`
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/books/%d/highlights", book.ID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token1})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", rec.Code, rec.Body.String())
		}

		var created HighlightResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
			t.Fatalf("failed to parse highlight: %v", err)
		}
		if created.Color != "blue" || created.Note != "Key world-building concept" {
			t.Fatalf("unexpected highlight data: %+v", created)
		}
		highlightID = created.ID
	}

	// 9. List highlights
	{
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d/highlights", book.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token1})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var highlights []HighlightResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &highlights); err != nil {
			t.Fatalf("failed to parse highlights: %v", err)
		}
		if len(highlights) != 1 || highlights[0].ID != highlightID {
			t.Fatalf("expected 1 highlight, got %v", highlights)
		}
	}

	// 10. Update highlight note and color
	{
		body := `{"note":"Updated thought on Hegemony","color":"green"}`
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/highlights/%d", book.ID, highlightID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token1})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}

		var updated HighlightResponse
		_ = json.Unmarshal(rec.Body.Bytes(), &updated)
		if updated.Color != "green" || updated.Note != "Updated thought on Hegemony" {
			t.Fatalf("unexpected updated highlight: %+v", updated)
		}
	}

	// 11. Reader two cannot update reader one's highlight -> 403
	{
		body := `{"note":"Hacked note","color":"pink"}`
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/highlights/%d", book.ID, highlightID), bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token2})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %d", rec.Code)
		}
	}

	// 12. Delete highlight
	{
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d/highlights/%d", book.ID, highlightID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: token1})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}
	}
}
