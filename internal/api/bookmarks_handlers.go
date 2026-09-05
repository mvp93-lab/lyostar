package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

// BookmarkResponse represents a bookmark returned to the client.
type BookmarkResponse struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	BookID    int64   `json:"book_id"`
	Title     string  `json:"title"`
	Location  string  `json:"location"`
	Progress  float64 `json:"progress"`
	CreatedAt string  `json:"created_at"`
}

// CreateBookmarkRequest is the payload to create a new bookmark.
type CreateBookmarkRequest struct {
	Title    string  `json:"title"`
	Location string  `json:"location"`
	Progress float64 `json:"progress"`
}

// HighlightResponse represents a highlight/note returned to the client.
type HighlightResponse struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	BookID       int64  `json:"book_id"`
	Location     string `json:"location"`
	SelectedText string `json:"selected_text"`
	Note         string `json:"note"`
	Color        string `json:"color"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CreateHighlightRequest is the payload to create a new highlight or note.
type CreateHighlightRequest struct {
	Location     string `json:"location"`
	SelectedText string `json:"selected_text"`
	Note         string `json:"note"`
	Color        string `json:"color"`
}

// UpdateHighlightRequest is the payload to update an existing highlight or note.
type UpdateHighlightRequest struct {
	Note  string `json:"note"`
	Color string `json:"color"`
}

// RegisterBookmarkAndHighlightRoutes registers bookmarks and highlights routes on the router.
func RegisterBookmarkAndHighlightRoutes(r chi.Router, db *database.Queries) {
	// ==================== BOOKMARKS ====================

	// GET /api/books/{id}/bookmarks
	r.Get("/{id}/bookmarks", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !user.Permissions.CanRead && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "reading permission required")
			return
		}

		bookID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid book id")
			return
		}

		rows, err := db.ListBookmarksForUserAndBook(req.Context(), database.ListBookmarksForUserAndBookParams{
			UserID: user.ID,
			BookID: bookID,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query bookmarks")
			return
		}

		res := make([]BookmarkResponse, 0, len(rows))
		for _, b := range rows {
			res = append(res, BookmarkResponse{
				ID:        b.ID,
				UserID:    b.UserID,
				BookID:    b.BookID,
				Title:     b.Title,
				Location:  b.Location,
				Progress:  b.Progress,
				CreatedAt: b.CreatedAt.Format(time.RFC3339),
			})
		}

		writeJSON(w, http.StatusOK, res)
	})

	// POST /api/books/{id}/bookmarks
	r.Post("/{id}/bookmarks", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !user.Permissions.CanRead && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "reading permission required")
			return
		}

		bookID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid book id")
			return
		}

		var payload CreateBookmarkRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		location := strings.TrimSpace(payload.Location)
		if location == "" {
			writeError(w, http.StatusBadRequest, "location is required")
			return
		}

		title := strings.TrimSpace(payload.Title)
		if title == "" {
			title = "Bookmark"
		}

		progress := payload.Progress
		if progress < 0 {
			progress = 0
		} else if progress > 1 {
			progress = 1
		}

		created, err := db.CreateBookmark(req.Context(), database.CreateBookmarkParams{
			UserID:   user.ID,
			BookID:   bookID,
			Title:    title,
			Location: location,
			Progress: progress,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create bookmark")
			return
		}

		writeJSON(w, http.StatusCreated, BookmarkResponse{
			ID:        created.ID,
			UserID:    created.UserID,
			BookID:    created.BookID,
			Title:     created.Title,
			Location:  created.Location,
			Progress:  created.Progress,
			CreatedAt: created.CreatedAt.Format(time.RFC3339),
		})
	})

	// DELETE /api/books/{id}/bookmarks/{bookmarkId}
	r.Delete("/{id}/bookmarks/{bookmarkId}", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		bookmarkID, err := strconv.ParseInt(chi.URLParam(req, "bookmarkId"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid bookmark id")
			return
		}

		// Verify existence and ownership
		b, err := db.GetBookmarkByID(req.Context(), bookmarkID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "bookmark not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query bookmark")
			return
		}

		if b.UserID != user.ID && user.Role != "admin" {
			writeError(w, http.StatusForbidden, "not authorized to delete this bookmark")
			return
		}

		if err := db.DeleteBookmark(req.Context(), database.DeleteBookmarkParams{
			ID:     bookmarkID,
			UserID: b.UserID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete bookmark")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "bookmark deleted"})
	})

	// ==================== HIGHLIGHTS & NOTES ====================

	// GET /api/books/{id}/highlights
	r.Get("/{id}/highlights", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !user.Permissions.CanRead && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "reading permission required")
			return
		}

		bookID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid book id")
			return
		}

		rows, err := db.ListHighlightsForUserAndBook(req.Context(), database.ListHighlightsForUserAndBookParams{
			UserID: user.ID,
			BookID: bookID,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query highlights")
			return
		}

		res := make([]HighlightResponse, 0, len(rows))
		for _, h := range rows {
			res = append(res, HighlightResponse{
				ID:           h.ID,
				UserID:       h.UserID,
				BookID:       h.BookID,
				Location:     h.Location,
				SelectedText: h.SelectedText,
				Note:         h.Note,
				Color:        h.Color,
				CreatedAt:    h.CreatedAt.Format(time.RFC3339),
				UpdatedAt:    h.UpdatedAt.Format(time.RFC3339),
			})
		}

		writeJSON(w, http.StatusOK, res)
	})

	// POST /api/books/{id}/highlights
	r.Post("/{id}/highlights", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !user.Permissions.CanRead && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "reading permission required")
			return
		}

		bookID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid book id")
			return
		}

		var payload CreateHighlightRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		location := strings.TrimSpace(payload.Location)
		if location == "" {
			writeError(w, http.StatusBadRequest, "location is required")
			return
		}

		selectedText := strings.TrimSpace(payload.SelectedText)
		note := strings.TrimSpace(payload.Note)
		if selectedText == "" && note == "" {
			writeError(w, http.StatusBadRequest, "selected text or note is required")
			return
		}

		color := strings.ToLower(strings.TrimSpace(payload.Color))
		switch color {
		case "yellow", "green", "blue", "pink":
			// valid
		default:
			color = "yellow"
		}

		created, err := db.CreateHighlight(req.Context(), database.CreateHighlightParams{
			UserID:       user.ID,
			BookID:       bookID,
			Location:     location,
			SelectedText: selectedText,
			Note:         note,
			Color:        color,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create highlight")
			return
		}

		writeJSON(w, http.StatusCreated, HighlightResponse{
			ID:           created.ID,
			UserID:       created.UserID,
			BookID:       created.BookID,
			Location:     created.Location,
			SelectedText: created.SelectedText,
			Note:         created.Note,
			Color:        created.Color,
			CreatedAt:    created.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    created.UpdatedAt.Format(time.RFC3339),
		})
	})

	// PUT /api/books/{id}/highlights/{highlightId}
	r.Put("/{id}/highlights/{highlightId}", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		highlightID, err := strconv.ParseInt(chi.URLParam(req, "highlightId"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid highlight id")
			return
		}

		h, err := db.GetHighlightByID(req.Context(), highlightID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "highlight not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query highlight")
			return
		}

		if h.UserID != user.ID && user.Role != "admin" {
			writeError(w, http.StatusForbidden, "not authorized to update this highlight")
			return
		}

		var payload UpdateHighlightRequest
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		color := strings.ToLower(strings.TrimSpace(payload.Color))
		if color == "" {
			color = h.Color
		} else {
			switch color {
			case "yellow", "green", "blue", "pink":
				// valid
			default:
				color = h.Color
			}
		}

		updated, err := db.UpdateHighlight(req.Context(), database.UpdateHighlightParams{
			ID:     highlightID,
			UserID: h.UserID,
			Note:   strings.TrimSpace(payload.Note),
			Color:  color,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to update highlight")
			return
		}

		writeJSON(w, http.StatusOK, HighlightResponse{
			ID:           updated.ID,
			UserID:       updated.UserID,
			BookID:       updated.BookID,
			Location:     updated.Location,
			SelectedText: updated.SelectedText,
			Note:         updated.Note,
			Color:        updated.Color,
			CreatedAt:    updated.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    updated.UpdatedAt.Format(time.RFC3339),
		})
	})

	// DELETE /api/books/{id}/highlights/{highlightId}
	r.Delete("/{id}/highlights/{highlightId}", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		highlightID, err := strconv.ParseInt(chi.URLParam(req, "highlightId"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid highlight id")
			return
		}

		h, err := db.GetHighlightByID(req.Context(), highlightID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "highlight not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query highlight")
			return
		}

		if h.UserID != user.ID && user.Role != "admin" {
			writeError(w, http.StatusForbidden, "not authorized to delete this highlight")
			return
		}

		if err := db.DeleteHighlight(req.Context(), database.DeleteHighlightParams{
			ID:     highlightID,
			UserID: h.UserID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete highlight")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "highlight deleted"})
	})
}
