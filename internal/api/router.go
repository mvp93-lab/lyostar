package api

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lyostar/lyostar/internal/database"
)

// RouterConfig holds dependencies for the HTTP router.
type RouterConfig struct {
	DB       *database.DB
	StaticFS fs.FS
	Version  string
}

// HealthResponse represents the health check payload.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// NewRouter constructs and configures the HTTP router.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	version := cfg.Version
	if version == "" {
		version = "0.1.0-dev"
	}

	// API routes
	r.Route("/api", func(api chi.Router) {
		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(HealthResponse{
				Status:  "ok",
				Version: version,
			})
		})

		// 404 handler specifically for unhandled /api routes
		api.NotFound(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "endpoint not found",
			})
		})
	})

	// Serve SPA static frontend if filesystem is provided
	if cfg.StaticFS != nil {
		fileServer := http.FileServer(http.FS(cfg.StaticFS))
		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			cleanPath := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
			if cleanPath == "" {
				cleanPath = "index.html"
			}

			// Check if file exists in the embedded static filesystem
			f, err := cfg.StaticFS.Open(cleanPath)
			if err != nil {
				// File does not exist: fallback to index.html for client-side SPA routing
				indexFile, indexErr := cfg.StaticFS.Open("index.html")
				if indexErr != nil {
					http.NotFound(w, r)
					return
				}
				_ = indexFile.Close()

				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}
			_ = f.Close()

			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}
