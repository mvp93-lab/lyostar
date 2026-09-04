package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/lyostar/lyostar/frontend"
	"github.com/lyostar/lyostar/internal/api"
	"github.com/lyostar/lyostar/internal/config"
	"github.com/lyostar/lyostar/internal/database"
)

func main() {
	cfg := config.Load()

	log.Printf("[Lyostar] Starting Lyostar ebook server...")
	log.Printf("[Lyostar] Config: Port=%d, BooksDir=%s, DataDir=%s", cfg.Port, cfg.BooksDir, cfg.DataDir)

	// Ensure application directories exist
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Fatalf("[Lyostar] Failed to create data directory: %v", err)
	}
	cacheCoversDir := filepath.Join(cfg.DataDir, "cache", "covers")
	if err := os.MkdirAll(cacheCoversDir, 0755); err != nil {
		log.Fatalf("[Lyostar] Failed to create cache/covers directory: %v", err)
	}

	// Initialize SQLite database
	dbPath := filepath.Join(cfg.DataDir, "app.db")
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("[Lyostar] Failed to initialize database at %s: %v", dbPath, err)
	}
	defer db.Close()
	log.Printf("[Lyostar] Database initialized at %s (WAL mode, pure Go)", dbPath)

	// Obtain embedded frontend static assets
	distFS, err := frontend.DistFS()
	if err != nil {
		log.Fatalf("[Lyostar] Failed to load embedded frontend: %v", err)
	}

	// Create HTTP router
	router := api.NewRouter(api.RouterConfig{
		DB:       db,
		StaticFS: distFS,
		Version:  "0.1.0-dev",
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown on SIGINT and SIGTERM
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[Lyostar] Server listening on http://localhost:%d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		log.Fatalf("[Lyostar] Server error: %v", err)
	case <-shutdownCtx.Done():
		log.Printf("[Lyostar] Shutdown signal received, shutting down gracefully...")
		stop()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[Lyostar] Server shutdown error: %v", err)
			_ = server.Close()
		}
		log.Printf("[Lyostar] Server stopped cleanly")
	}
}
