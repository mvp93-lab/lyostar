package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

func TestTagsEndpointsAndFiltering(t *testing.T) {
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

	// Create Tags
	tagSciFi, err := db.CreateTag(ctx, "Science Fiction")
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}
	tagSpace, err := db.CreateTag(ctx, "Space Opera")
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}
	tagClassic, err := db.CreateTag(ctx, "Classic")
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	// Link tags:
	// Dune: Science Fiction, Space Opera
	_ = db.AddBookTag(ctx, database.AddBookTagParams{BookID: book1.ID, TagID: tagSciFi.ID})
	_ = db.AddBookTag(ctx, database.AddBookTagParams{BookID: book1.ID, TagID: tagSpace.ID})
	// Foundation: Science Fiction, Classic
	_ = db.AddBookTag(ctx, database.AddBookTagParams{BookID: book2.ID, TagID: tagSciFi.ID})
	_ = db.AddBookTag(ctx, database.AddBookTagParams{BookID: book2.ID, TagID: tagClassic.ID})

	// Create a test user with session
	_, userToken := createTestUserWithPerms(t, db, "tagreader", auth.RoleReader, auth.Permissions{
		CanRead:   true,
		CanEdit:   true,
		CanDelete: true,
	})

	// 1. GET /api/tags
	{
		req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for GET /api/tags, got %d", rec.Code)
		}

		var tags []TagItem
		if err := json.NewDecoder(rec.Body).Decode(&tags); err != nil {
			t.Fatalf("failed to decode tags: %v", err)
		}

		if len(tags) != 3 {
			t.Fatalf("expected 3 tags, got %d", len(tags))
		}

		// Most popular should be Science Fiction (count 2)
		if tags[0].Name != "Science Fiction" || tags[0].BookCount != 2 {
			t.Errorf("expected Science Fiction with 2 books as top tag, got: %+v", tags[0])
		}
	}

	// 2. GET /api/books?tag=Space%20Opera (should return only Dune)
	{
		req := httptest.NewRequest(http.MethodGet, "/api/books?tag=Space%20Opera", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var resp PaginatedResponse[BookListItem]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode books response: %v", err)
		}

		if resp.Total != 1 || len(resp.Items) != 1 {
			t.Fatalf("expected 1 book for tag 'Space Opera', got %d (items: %d)", resp.Total, len(resp.Items))
		}
		if resp.Items[0].Title != "Dune" {
			t.Errorf("expected book 'Dune', got %q", resp.Items[0].Title)
		}
		// Verify tags are included in BookListItem
		if len(resp.Items[0].Tags) != 2 {
			t.Errorf("expected 2 tags on Dune, got: %+v", resp.Items[0].Tags)
		}
	}

	// 3. GET /api/books?tag=Science%20Fiction (should return both books)
	{
		req := httptest.NewRequest(http.MethodGet, "/api/books?tag=Science%20Fiction", nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rec.Code)
		}

		var resp PaginatedResponse[BookListItem]
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode books response: %v", err)
		}

		if resp.Total != 2 || len(resp.Items) != 2 {
			t.Fatalf("expected 2 books for tag 'Science Fiction', got %d", resp.Total)
		}
	}

	// 4. PUT /api/books/{id} (Update tags)
	{
		updateBody, _ := json.Marshal(UpdateBookMetadataRequest{
			Title: "Dune Messiah",
			Tags:  []string{"Science Fiction", "Political Intrigue"},
		})
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d", book1.ID), bytes.NewReader(updateBody))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for update, got %d: %s", rec.Code, rec.Body.String())
		}

		var detail BookDetailResponse
		if err := json.NewDecoder(rec.Body).Decode(&detail); err != nil {
			t.Fatalf("failed to decode detail: %v", err)
		}

		if len(detail.Tags) != 2 {
			t.Errorf("expected 2 tags after update, got: %+v", detail.Tags)
		}

		// Verify GET /api/books?tag=Political%20Intrigue finds Dune Messiah
		filterReq := httptest.NewRequest(http.MethodGet, "/api/books?tag=Political%20Intrigue", nil)
		filterReq.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		filterRec := httptest.NewRecorder()
		router.ServeHTTP(filterRec, filterReq)
		var filterResp PaginatedResponse[BookListItem]
		_ = json.NewDecoder(filterRec.Body).Decode(&filterResp)
		if filterResp.Total != 1 || len(filterResp.Items) != 1 || filterResp.Items[0].Title != "Dune Messiah" {
			t.Errorf("expected to find Dune Messiah with new tag, got: %+v", filterResp)
		}
	}
}
