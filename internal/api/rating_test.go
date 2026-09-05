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

func TestRatingEndpoints(t *testing.T) {
	db, router, tempDir, booksDir, _ := setupBookManageTest(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	bookPath := filepath.Join(booksDir, "book.epub")
	_ = os.WriteFile(bookPath, []byte("epub"), 0644)

	book, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:      "Hyperion",
		FilePath:   bookPath,
		FileSha256: "hyp_sha256",
		FileSize:   100,
		Format:     "epub",
	})
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	_, userToken := createTestUserWithPerms(t, db, "rater", auth.RoleReader, auth.Permissions{
		CanRead: true,
	})

	// 1. PUT /api/books/{id}/rating (Rate 5 stars)
	t.Run("PUT /api/books/{id}/rating sets rating", func(t *testing.T) {
		body, _ := json.Marshal(RatingRequest{Rating: 5})
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d/rating", book.ID), bytes.NewReader(body))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rr.Code, rr.Body.String())
		}

		var resp map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)

		if resp["user_rating"] != float64(5) {
			t.Errorf("expected user_rating 5, got %v", resp["user_rating"])
		}
		if resp["avg_rating"] != float64(5) {
			t.Errorf("expected avg_rating 5, got %v", resp["avg_rating"])
		}
	})

	// 2. GET /api/books/{id} reflects user_rating and avg_rating
	t.Run("GET /api/books/{id} returns rating metadata", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/books/%d", book.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}

		var detail BookDetailResponse
		_ = json.Unmarshal(rr.Body.Bytes(), &detail)

		if detail.UserRating != 5 {
			t.Errorf("expected detail.UserRating = 5, got %d", detail.UserRating)
		}
		if detail.AvgRating != 5.0 {
			t.Errorf("expected detail.AvgRating = 5.0, got %f", detail.AvgRating)
		}
	})

	// 3. DELETE /api/books/{id}/rating (Remove rating)
	t.Run("DELETE /api/books/{id}/rating clears rating", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d/rating", book.ID), nil)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: userToken})
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", rr.Code)
		}

		var resp map[string]any
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)

		if resp["user_rating"] != float64(0) {
			t.Errorf("expected user_rating 0, got %v", resp["user_rating"])
		}
	})
}
