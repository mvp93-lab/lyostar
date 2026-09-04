package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

func TestAuthFlowAndProtection(t *testing.T) {
	db, _, router, tempDir := setupTestDBAndRouter(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	// 1. Initial status: Setup required (0 users)
	{
		req := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for status, got %d", rec.Code)
		}

		var status AuthStatusResponse
		if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
			t.Fatalf("failed to decode status: %v", err)
		}
		if !status.SetupRequired {
			t.Errorf("expected setup_required to be true initially")
		}
		if status.Authenticated {
			t.Errorf("expected authenticated to be false initially")
		}
	}

	// 2. Accessing protected endpoint before login returns 401 Unauthorized
	{
		req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for /api/books when unauthenticated, got %d", rec.Code)
		}
	}

	// 3. First-run setup: Create Admin account
	var adminCookie *http.Cookie
	{
		setupPayload := SetupRequest{
			Username:    "admin",
			Password:    "SuperSecret123",
			DisplayName: "Master Librarian",
		}
		body, _ := json.Marshal(setupPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 for /api/auth/setup, got %d: %s", rec.Code, rec.Body.String())
		}

		cookies := rec.Result().Cookies()
		for _, c := range cookies {
			if c.Name == auth.CookieName {
				adminCookie = c
				break
			}
		}
		if adminCookie == nil {
			t.Fatalf("expected session cookie to be set on setup")
		}
	}

	// 4. Repeated setup is now forbidden (403)
	{
		setupPayload := SetupRequest{
			Username: "admin2",
			Password: "AnotherPassword123",
		}
		body, _ := json.Marshal(setupPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403 on repeated setup, got %d", rec.Code)
		}
	}

	// 5. Test GET /api/auth/me with admin cookie
	{
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.AddCookie(adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for /api/auth/me, got %d", rec.Code)
		}

		var me auth.CurrentUser
		if err := json.NewDecoder(rec.Body).Decode(&me); err != nil {
			t.Fatalf("failed to decode /api/auth/me: %v", err)
		}
		if me.Username != "admin" || me.Role != auth.RoleAdmin {
			t.Errorf("unexpected user from /api/auth/me: %+v", me)
		}
	}

	// 6. Admin creates a new reader user
	var readerCookie *http.Cookie
	{
		createPayload := CreateUserRequest{
			Username:    "reader1",
			Password:    "ReaderPass123",
			Role:        auth.RoleReader,
			DisplayName: "Book Lover",
		}
		body, _ := json.Marshal(createPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 creating user, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// 7. Reader logs in
	{
		loginPayload := LoginRequest{
			Username: "reader1",
			Password: "ReaderPass123",
		}
		body, _ := json.Marshal(loginPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for login, got %d: %s", rec.Code, rec.Body.String())
		}

		cookies := rec.Result().Cookies()
		for _, c := range cookies {
			if c.Name == auth.CookieName {
				readerCookie = c
				break
			}
		}
		if readerCookie == nil {
			t.Fatalf("expected session cookie for reader login")
		}
	}

	// 8. Reader can access /api/books
	{
		req := httptest.NewRequest(http.MethodGet, "/api/books", nil)
		req.AddCookie(readerCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200 for reader accessing /api/books, got %d", rec.Code)
		}
	}

	// 9. Reader CANNOT trigger scan (403 Forbidden)
	{
		req := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
		req.AddCookie(readerCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403 for reader accessing /api/scan, got %d", rec.Code)
		}
	}

	// 10. Admin CAN trigger scan (202 Accepted)
	{
		req := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
		req.AddCookie(adminCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusAccepted {
			t.Errorf("expected 202 for admin accessing /api/scan, got %d", rec.Code)
		}
	}

	// 11. Logout clears session
	{
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.AddCookie(readerCookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for logout, got %d", rec.Code)
		}

		// Trying to use old reader cookie now fails (401)
		req2 := httptest.NewRequest(http.MethodGet, "/api/books", nil)
		req2.AddCookie(readerCookie)
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 after logout, got %d", rec2.Code)
		}
	}
}

func TestGranularUserPermissions(t *testing.T) {
	db, _, router, tempDir := setupTestDBAndRouter(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := t.Context()

	// 1. Create Admin
	adminPw, _ := auth.HashPassword("AdminPass123")
	adminUser, err := db.CreateUser(ctx, database.CreateUserParams{
		Username:     "admin_perm",
		PasswordHash: adminPw,
		Role:         auth.RoleAdmin,
		DisplayName:  "Admin Perm",
		CanRead:      1,
		CanDownload:  1,
		CanUpload:    1,
		CanEdit:      1,
		CanDelete:    1,
	})
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	adminToken := "admin-perm-token"
	_, _ = db.CreateSession(ctx, database.CreateSessionParams{
		Token:     adminToken,
		UserID:    adminUser.ID,
		ExpiresAt: time.Now().Add(time.Hour),
	})
	adminCookie := &http.Cookie{Name: "lyostar_session", Value: adminToken}

	// 2. Create restricted reader with can_download=false via POST /api/users
	createPayload := CreateUserRequest{
		Username:    "restricted_reader",
		Password:    "ReaderPass123",
		Role:        "reader",
		DisplayName: "Restricted Reader",
		Permissions: &auth.Permissions{
			CanRead:     true,
			CanDownload: false,
			CanUpload:   false,
			CanEdit:     false,
			CanDelete:   false,
		},
	}
	body, _ := json.Marshal(createPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(adminCookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for creating restricted user, got %d: %s", rec.Code, rec.Body.String())
	}

	var createdUser UserItem
	if err := json.NewDecoder(rec.Body).Decode(&createdUser); err != nil {
		t.Fatalf("failed to decode created user: %v", err)
	}
	if !createdUser.Permissions.CanRead || createdUser.Permissions.CanDownload {
		t.Errorf("unexpected permissions: %+v", createdUser.Permissions)
	}

	// 3. Login as restricted reader
	loginPayload := LoginRequest{
		Username: "restricted_reader",
		Password: "ReaderPass123",
	}
	body, _ = json.Marshal(loginPayload)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()
	router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for reader login, got %d", loginRec.Code)
	}

	var readerCookie *http.Cookie
	for _, c := range loginRec.Result().Cookies() {
		if c.Name == auth.CookieName {
			readerCookie = c
			break
		}
	}
	if readerCookie == nil {
		t.Fatalf("expected reader session cookie")
	}

	// 4. Seed a book file
	bookPath := tempDir + "/test_perm.epub"
	_ = os.WriteFile(bookPath, []byte("fake epub data"), 0644)
	book, err := db.CreateBook(ctx, database.CreateBookParams{
		Title:      "Perm Test Book",
		FilePath:   bookPath,
		FileSha256: "sha256perm123",
		FileSize:   14,
		Format:     "epub",
	})
	if err != nil {
		t.Fatalf("failed to create book: %v", err)
	}

	// 5. Restricted reader requests /api/books/{id}/file -> Allowed (200)
	{
		fileReq := httptest.NewRequest(http.MethodGet, "/api/books/"+strconv.FormatInt(book.ID, 10)+"/file", nil)
		fileReq.AddCookie(readerCookie)
		fileRec := httptest.NewRecorder()
		router.ServeHTTP(fileRec, fileReq)

		if fileRec.Code != http.StatusOK {
			t.Errorf("expected 200 for reading book with can_read=true, got %d", fileRec.Code)
		}
	}

	// 6. Restricted reader requests /api/books/{id}/download -> Forbidden (403)
	{
		dlReq := httptest.NewRequest(http.MethodGet, "/api/books/"+strconv.FormatInt(book.ID, 10)+"/download", nil)
		dlReq.AddCookie(readerCookie)
		dlRec := httptest.NewRecorder()
		router.ServeHTTP(dlRec, dlReq)

		if dlRec.Code != http.StatusForbidden {
			t.Errorf("expected 403 for downloading book with can_download=false, got %d", dlRec.Code)
		}
	}

	// 7. Admin updates reader to can_download=true and can_read=false via PUT /api/users/{id}
	{
		updatePayload := UpdateUserRequest{
			DisplayName: "Updated Reader",
			Role:        "reader",
			Permissions: &auth.Permissions{
				CanRead:     false,
				CanDownload: true,
			},
		}
		upBody, _ := json.Marshal(updatePayload)
		putReq := httptest.NewRequest(http.MethodPut, "/api/users/"+strconv.FormatInt(createdUser.ID, 10), bytes.NewReader(upBody))
		putReq.Header.Set("Content-Type", "application/json")
		putReq.AddCookie(adminCookie)
		putRec := httptest.NewRecorder()
		router.ServeHTTP(putRec, putReq)

		if putRec.Code != http.StatusOK {
			t.Fatalf("expected 200 for updating user permissions, got %d: %s", putRec.Code, putRec.Body.String())
		}
	}

	// 8. Now reader requests download -> Allowed (200)
	{
		dlReq := httptest.NewRequest(http.MethodGet, "/api/books/"+strconv.FormatInt(book.ID, 10)+"/download", nil)
		dlReq.AddCookie(readerCookie)
		dlRec := httptest.NewRecorder()
		router.ServeHTTP(dlRec, dlReq)

		if dlRec.Code != http.StatusOK {
			t.Errorf("expected 200 for download after permission granted, got %d", dlRec.Code)
		}
	}

	// 9. Now reader requests file reading -> Forbidden (403)
	{
		fileReq := httptest.NewRequest(http.MethodGet, "/api/books/"+strconv.FormatInt(book.ID, 10)+"/file", nil)
		fileReq.AddCookie(readerCookie)
		fileRec := httptest.NewRecorder()
		router.ServeHTTP(fileRec, fileReq)

		if fileRec.Code != http.StatusForbidden {
			t.Errorf("expected 403 for reading file after can_read revoked, got %d", fileRec.Code)
		}
	}
}

