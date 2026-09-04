package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

func TestReadingProgressFlow(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "lyostar-progress-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create user
	pwHash, _ := auth.HashPassword("secretpass")
	user, err := db.CreateUser(t.Context(), database.CreateUserParams{
		Username:     "tester",
		PasswordHash: pwHash,
		Role:         auth.RoleReader,
		DisplayName:  "Test Reader",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create book
	book, err := db.CreateBook(t.Context(), database.CreateBookParams{
		Title:       "Test Book",
		FilePath:    filepath.Join(tmpDir, "test.epub"),
		FileSha256:  "sha-progress-1",
		FileSize:    1234,
		Format:      "epub",
		Description: "A test book",
		Publisher:   "Lyostar",
		Language:    "en",
	})
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	// Setup router
	handler := NewRouter(RouterConfig{
		DB:      db,
		Version: "test",
	})

	// Login to get session cookie
	loginBody, _ := json.Marshal(map[string]string{
		"username": "tester",
		"password": "secretpass",
	})
	reqLogin := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	handler.ServeHTTP(wLogin, reqLogin)

	cookies := wLogin.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == auth.CookieName {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatalf("expected session cookie after login")
	}

	// 1. GET /api/books/{id}/progress initially -> returns 0 progress
	reqProg1 := httptest.NewRequest("GET", "/api/books/1/progress", nil)
	reqProg1.AddCookie(sessionCookie)
	wProg1 := httptest.NewRecorder()
	handler.ServeHTTP(wProg1, reqProg1)

	if wProg1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", wProg1.Code, wProg1.Body.String())
	}
	var progRes ReadingProgressResponse
	if err := json.Unmarshal(wProg1.Body.Bytes(), &progRes); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if progRes.Progress != 0 || progRes.IsFinished {
		t.Fatalf("expected 0 progress, got %f", progRes.Progress)
	}

	// 2. PUT /api/books/{id}/progress -> update to 45%
	updateBody, _ := json.Marshal(UpdateReadingProgressRequest{
		Location:    "epubcfi(/6/4[chap01]!/4/2/10)",
		Progress:    0.45,
		CurrentPage: 45,
		TotalPages:  100,
		IsFinished:  false,
	})
	reqUpdate := httptest.NewRequest("PUT", "/api/books/1/progress", bytes.NewReader(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	reqUpdate.AddCookie(sessionCookie)
	wUpdate := httptest.NewRecorder()
	handler.ServeHTTP(wUpdate, reqUpdate)

	if wUpdate.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on update, got %d: %s", wUpdate.Code, wUpdate.Body.String())
	}

	// 3. GET /api/books/continue-reading -> should return the book
	reqCR := httptest.NewRequest("GET", "/api/books/continue-reading", nil)
	reqCR.AddCookie(sessionCookie)
	wCR := httptest.NewRecorder()
	handler.ServeHTTP(wCR, reqCR)

	if wCR.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on continue-reading, got %d: %s", wCR.Code, wCR.Body.String())
	}
	var crItems []ContinueReadingItem
	if err := json.Unmarshal(wCR.Body.Bytes(), &crItems); err != nil {
		t.Fatalf("failed to decode continue-reading items: %v", err)
	}
	if len(crItems) != 1 {
		t.Fatalf("expected 1 item in continue-reading, got %d", len(crItems))
	}
	if crItems[0].BookID != book.ID || crItems[0].Progress != 0.45 {
		t.Fatalf("unexpected continue reading item: %+v", crItems[0])
	}

	// 4. GET /api/books list -> includes user progress
	reqBooks := httptest.NewRequest("GET", "/api/books", nil)
	reqBooks.AddCookie(sessionCookie)
	wBooks := httptest.NewRecorder()
	handler.ServeHTTP(wBooks, reqBooks)

	if wBooks.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", wBooks.Code, wBooks.Body.String())
	}
	var booksRes PaginatedResponse[BookListItem]
	if err := json.Unmarshal(wBooks.Body.Bytes(), &booksRes); err != nil {
		t.Fatalf("failed to decode books response: %v", err)
	}
	if len(booksRes.Items) != 1 || booksRes.Items[0].Progress != 0.45 {
		t.Fatalf("expected book progress 0.45 in list, got %+v", booksRes.Items[0])
	}
	_ = user
}
