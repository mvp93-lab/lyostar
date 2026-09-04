package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/lyostar/lyostar/internal/auth"
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
