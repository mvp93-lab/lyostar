package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestRouterHealth(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<html><body>Lyostar SPA</body></html>"),
		},
	}

	router := NewRouter(RouterConfig{
		StaticFS: mockFS,
		Version:  "1.0.0-test",
	})

	// Test GET /api/health
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

func TestRouterSPAFallback(t *testing.T) {
	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<html><body>Lyostar SPA</body></html>"),
		},
		"assets/app.js": &fstest.MapFile{
			Data: []byte("console.log('Lyostar');"),
		},
	}

	router := NewRouter(RouterConfig{
		StaticFS: mockFS,
		Version:  "1.0.0-test",
	})

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

	// 3. Unmatched API route returns 404 JSON, does not fallback to index.html
	reqAPI404 := httptest.NewRequest(http.MethodGet, "/api/unknown-endpoint", nil)
	recAPI404 := httptest.NewRecorder()
	router.ServeHTTP(recAPI404, reqAPI404)
	if recAPI404.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown api route, got %d", recAPI404.Code)
	}
}
