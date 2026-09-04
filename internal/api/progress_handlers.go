package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

// ReadingProgressResponse represents reading progress for a book.
type ReadingProgressResponse struct {
	UserID      int64   `json:"user_id"`
	BookID      int64   `json:"book_id"`
	Location    string  `json:"location"`
	Progress    float64 `json:"progress"`
	CurrentPage int64   `json:"current_page"`
	TotalPages  int64   `json:"total_pages"`
	IsFinished  bool    `json:"is_finished"`
	UpdatedAt   string  `json:"updated_at"`
}

// UpdateReadingProgressRequest is the payload for updating reading progress.
type UpdateReadingProgressRequest struct {
	Location    string  `json:"location"`
	Progress    float64 `json:"progress"`
	CurrentPage int64   `json:"current_page"`
	TotalPages  int64   `json:"total_pages"`
	IsFinished  bool    `json:"is_finished"`
}

// ContinueReadingItem represents a book in the user's continue-reading queue.
type ContinueReadingItem struct {
	BookID      int64    `json:"book_id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Format      string   `json:"format"`
	CoverURL    string   `json:"cover_url,omitempty"`
	Location    string   `json:"location"`
	Progress    float64  `json:"progress"`
	CurrentPage int64    `json:"current_page"`
	TotalPages  int64    `json:"total_pages"`
	UpdatedAt   string   `json:"updated_at"`
}

// RegisterProgressRoutes registers reading progress routes on the router.
func RegisterProgressRoutes(r chi.Router, db *database.Queries) {
	// GET /api/books/continue-reading
	r.Get("/continue-reading", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		limit := int64(12)
		if lStr := req.URL.Query().Get("limit"); lStr != "" {
			if l, err := strconv.ParseInt(lStr, 10, 64); err == nil && l > 0 && l <= 50 {
				limit = l
			}
		}

		rows, err := db.ListRecentProgressByUserID(req.Context(), database.ListRecentProgressByUserIDParams{
			UserID: user.ID,
			Limit:  limit,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query reading progress")
			return
		}

		items := make([]ContinueReadingItem, 0, len(rows))
		for _, row := range rows {
			var authors []string
			if namesStr, ok := row.AuthorNames.(string); ok && namesStr != "" {
				for _, a := range strings.Split(namesStr, ", ") {
					if trimmed := strings.TrimSpace(a); trimmed != "" {
						authors = append(authors, trimmed)
					}
				}
			}
			if len(authors) == 0 {
				authors = []string{"Unknown"}
			}

			var coverURL string
			if row.CoverPath != "" {
				coverURL = fmt.Sprintf("/api/books/%d/cover", row.BookID)
			}

			items = append(items, ContinueReadingItem{
				BookID:      row.BookID,
				Title:       row.Title,
				Authors:     authors,
				Format:      row.Format,
				CoverURL:    coverURL,
				Location:    row.Location,
				Progress:    row.Progress,
				CurrentPage: row.CurrentPage,
				TotalPages:  row.TotalPages,
				UpdatedAt:   row.UpdatedAt.Format(time.RFC3339),
			})
		}

		writeJSON(w, http.StatusOK, items)
	})

	// Routes under /api/books/{id}
	r.Route("/{id}/progress", func(pr chi.Router) {
		// GET /api/books/{id}/progress
		pr.Get("/", func(w http.ResponseWriter, req *http.Request) {
			user := auth.GetUser(req.Context())
			if user == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			bookID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid book id")
				return
			}

			p, err := db.GetProgress(req.Context(), database.GetProgressParams{
				UserID: user.ID,
				BookID: bookID,
			})
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					// Return empty progress
					writeJSON(w, http.StatusOK, ReadingProgressResponse{
						UserID:      user.ID,
						BookID:      bookID,
						Location:    "",
						Progress:    0,
						CurrentPage: 0,
						TotalPages:  0,
						IsFinished:  false,
						UpdatedAt:   "",
					})
					return
				}
				writeError(w, http.StatusInternalServerError, "database error")
				return
			}

			writeJSON(w, http.StatusOK, ReadingProgressResponse{
				UserID:      p.UserID,
				BookID:      p.BookID,
				Location:    p.Location,
				Progress:    p.Progress,
				CurrentPage: p.CurrentPage,
				TotalPages:  p.TotalPages,
				IsFinished:  p.IsFinished == 1,
				UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
			})
		})

		// PUT /api/books/{id}/progress
		pr.Put("/", func(w http.ResponseWriter, req *http.Request) {
			user := auth.GetUser(req.Context())
			if user == nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			bookID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid book id")
				return
			}

			var payload UpdateReadingProgressRequest
			if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json request")
				return
			}

			// Validate and clamp progress
			prog := payload.Progress
			if prog < 0 {
				prog = 0
			} else if prog > 1 {
				prog = 1
			}

			var isFinished int64
			if payload.IsFinished || prog >= 0.999 {
				isFinished = 1
			}

			p, err := db.UpsertProgress(req.Context(), database.UpsertProgressParams{
				UserID:      user.ID,
				BookID:      bookID,
				Location:    payload.Location,
				Progress:    prog,
				CurrentPage: payload.CurrentPage,
				TotalPages:  payload.TotalPages,
				IsFinished:  isFinished,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to update progress")
				return
			}

			writeJSON(w, http.StatusOK, ReadingProgressResponse{
				UserID:      p.UserID,
				BookID:      p.BookID,
				Location:    p.Location,
				Progress:    p.Progress,
				CurrentPage: p.CurrentPage,
				TotalPages:  p.TotalPages,
				IsFinished:  p.IsFinished == 1,
				UpdatedAt:   p.UpdatedAt.Format(time.RFC3339),
			})
		})
	})
}
