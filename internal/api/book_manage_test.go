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
	"testing/fstest"
	"time"

	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
	"github.com/lyostar/lyostar/internal/scanner"
)

func setupBookManageTest(t *testing.T) (*database.DB, http.Handler, string, string, string) {
	tempDir, err := os.MkdirTemp("", "lyostar-manage-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	booksDir := filepath.Join(tempDir, "books")
	dataDir := filepath.Join(tempDir, "data")
	uploadsDir := filepath.Join(dataDir, "uploads")
	coversDir := filepath.Join(dataDir, "cache", "covers")

	_ = os.MkdirAll(booksDir, 0755)
	_ = os.MkdirAll(uploadsDir, 0755)
	_ = os.MkdirAll(coversDir, 0755)

	dbPath := filepath.Join(dataDir, "app.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	sc := scanner.New(booksDir, coversDir, db, uploadsDir)

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<html><body>Lyostar SPA</body></html>"),
		},
	}

	router := NewRouter(RouterConfig{
		DB:         db,
		Scanner:    sc,
		UploadsDir: uploadsDir,
		StaticFS:   mockFS,
		Version:    "1.0.0-test",
	})

	return db, router, tempDir, booksDir, uploadsDir
}

func createTestUserWithPerms(t *testing.T, db *database.DB, username string, role string, perms auth.Permissions) (int64, string) {
	ctx := context.Background()
	hash, err := auth.HashPassword("secret123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	canReadVal := int64(0)
	if perms.CanRead {
		canReadVal = 1
	}
	canDownloadVal := int64(0)
	if perms.CanDownload {
		canDownloadVal = 1
	}
	canUploadVal := int64(0)
	if perms.CanUpload {
		canUploadVal = 1
	}
	canEditVal := int64(0)
	if perms.CanEdit {
		canEditVal = 1
	}
	canDeleteVal := int64(0)
	if perms.CanDelete {
		canDeleteVal = 1
	}

	u, err := db.CreateUser(ctx, database.CreateUserParams{
		Username:     username,
		PasswordHash: hash,
		Role:         role,
		DisplayName:  username,
		CanRead:      canReadVal,
		CanDownload:  canDownloadVal,
		CanUpload:    canUploadVal,
		CanEdit:      canEditVal,
		CanDelete:    canDeleteVal,
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create session token
	token, _ := auth.GenerateToken()
	_, err = db.CreateSession(ctx, database.CreateSessionParams{
		Token:     token,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	return u.ID, token
}

func TestUpdateBookMetadata(t *testing.T) {
	db, router, tempDir, booksDir, _ := setupBookManageTest(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	// Seed book in database
	bookPath := filepath.Join(booksDir, "sample.epub")
	_ = os.WriteFile(bookPath, []byte("dummy"), 0644)

	book, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:       "Original Title",
		FilePath:    bookPath,
		FileSha256:  "sample_sha256",
		FileSize:    100,
		Format:      "epub",
		Description: "Original Description",
		Publisher:   "Old Publisher",
		Language:    "en",
		PubDate:     "2020",
		Series:      "Old Series",
		SeriesIndex: 1.0,
	})
	if err != nil {
		t.Fatalf("failed to seed book: %v", err)
	}

	// Create user with can_edit = false
	_, noEditToken := createTestUserWithPerms(t, db, "noedit", auth.RoleReader, auth.Permissions{
		CanRead:   true,
		CanEdit:   false,
		CanDelete: false,
	})

	// Create user with can_edit = true
	_, editToken := createTestUserWithPerms(t, db, "editor", auth.RoleReader, auth.Permissions{
		CanRead:   true,
		CanEdit:   true,
		CanDelete: false,
	})

	updateBody := UpdateBookMetadataRequest{
		Title:       "Refactored Code",
		Authors:     []string{"Martin Fowler", "Kent Beck"},
		Description: "Improving the Design of Existing Code",
		Publisher:   "Addison-Wesley",
		Language:    "en",
		PubDate:     "2018-11-20",
		Series:      "Signature Series",
		SeriesIndex: 2.0,
	}
	bodyBytes, _ := json.Marshal(updateBody)

	// 1. Unauthorized (no token)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d", book.ID), bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
	}

	// 2. Forbidden (noEditToken lacks can_edit)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d", book.ID), bytes.NewReader(bodyBytes))
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: noEditToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", rec.Code)
	}

	// 3. Bad request (empty title)
	badBody, _ := json.Marshal(UpdateBookMetadataRequest{Title: "   "})
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d", book.ID), bytes.NewReader(badBody))
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: editToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for empty title, got %d", rec.Code)
	}

	// 4. Success with editToken
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/books/%d", book.ID), bytes.NewReader(bodyBytes))
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: editToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp BookDetailResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Title != "Refactored Code" {
		t.Errorf("expected updated title 'Refactored Code', got %q", resp.Title)
	}
	if len(resp.Authors) != 2 || resp.Authors[0].Name != "Kent Beck" || resp.Authors[1].Name != "Martin Fowler" {
		t.Errorf("unexpected authors: %+v", resp.Authors)
	}
	if resp.Series != "Signature Series" || resp.SeriesIndex != 2.0 {
		t.Errorf("unexpected series: %s #%f", resp.Series, resp.SeriesIndex)
	}
	if resp.Publisher != "Addison-Wesley" || resp.PubDate != "2018-11-20" {
		t.Errorf("unexpected pub info: %s %s", resp.Publisher, resp.PubDate)
	}

	// 5. Verify FTS5 search reflects the update
	searchReq := httptest.NewRequest(http.MethodGet, "/api/search?q=Refactored", nil)
	searchReq.AddCookie(&http.Cookie{Name: auth.CookieName, Value: editToken})
	searchRec := httptest.NewRecorder()
	router.ServeHTTP(searchRec, searchReq)
	if searchRec.Code != http.StatusOK {
		t.Fatalf("expected search 200, got %d", searchRec.Code)
	}
	var searchResp PaginatedResponse[BookListItem]
	if err := json.NewDecoder(searchRec.Body).Decode(&searchResp); err != nil {
		t.Fatalf("failed to decode search response: %v", err)
	}
	if searchResp.Total != 1 || len(searchResp.Items) != 1 || searchResp.Items[0].Title != "Refactored Code" {
		t.Errorf("expected search to find updated book, got: %+v", searchResp)
	}
}

func TestDeleteBook(t *testing.T) {
	db, router, tempDir, booksDir, uploadsDir := setupBookManageTest(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	// Seed read-only library book in booksDir
	readOnlyFilePath := filepath.Join(booksDir, "readonly.epub")
	_ = os.WriteFile(readOnlyFilePath, []byte("read-only content"), 0644)

	roBook, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:       "Read Only Book",
		FilePath:    readOnlyFilePath,
		FileSha256:  "ro_sha256",
		FileSize:    50,
		Format:      "epub",
		Description: "Cannot be deleted from disk",
	})
	if err != nil {
		t.Fatalf("failed to seed roBook: %v", err)
	}

	// Seed uploaded book in uploadsDir
	uploadedFilePath := filepath.Join(uploadsDir, "upload_123.pdf")
	_ = os.WriteFile(uploadedFilePath, []byte("uploaded file content"), 0644)

	fakeCoverPath := filepath.Join(tempDir, "data", "cache", "covers", "up_sha256.webp")
	_ = os.WriteFile(fakeCoverPath, []byte("webp-data"), 0644)

	upBook, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:       "Uploaded Book",
		FilePath:    uploadedFilePath,
		FileSha256:  "up_sha256",
		FileSize:    80,
		Format:      "pdf",
		Description: "Can be deleted from disk",
		CoverPath:   fakeCoverPath,
	})
	if err != nil {
		t.Fatalf("failed to seed upBook: %v", err)
	}

	// User without can_delete
	_, noDelToken := createTestUserWithPerms(t, db, "nodel", auth.RoleReader, auth.Permissions{
		CanRead:   true,
		CanDelete: false,
	})

	// User with can_delete
	_, delToken := createTestUserWithPerms(t, db, "deleter", auth.RoleReader, auth.Permissions{
		CanRead:   true,
		CanDelete: true,
	})

	// 1. Unauthorized
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d", upBook.ID), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
	}

	// 2. Forbidden (noDelToken)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d", upBook.ID), nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: noDelToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden, got %d", rec.Code)
	}

	// 3. Delete uploaded book (file and cover should be deleted on disk)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d", upBook.ID), nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: delToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify DB record deleted
	if _, err := db.GetBookByID(ctx, upBook.ID); err == nil {
		t.Errorf("expected book %d to be deleted from database", upBook.ID)
	}

	// Verify physical upload file removed
	if _, err := os.Stat(uploadedFilePath); !os.IsNotExist(err) {
		t.Errorf("expected uploaded file to be removed from disk, but it still exists")
	}

	// Verify cover thumbnail removed
	if _, err := os.Stat(fakeCoverPath); !os.IsNotExist(err) {
		t.Errorf("expected cover thumbnail to be removed from disk, but it still exists")
	}

	// 4. Delete read-only book from /books (DB record removed, but file preserved on disk!)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/books/%d", roBook.ID), nil)
	req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: delToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify DB record deleted
	if _, err := db.GetBookByID(ctx, roBook.ID); err == nil {
		t.Errorf("expected book %d to be deleted from database", roBook.ID)
	}

	// Verify read-only file PRESERVED
	if _, err := os.Stat(readOnlyFilePath); err != nil {
		t.Errorf("STRICTLY READ-ONLY violation: /books file was removed or missing: %v", err)
	}
}
