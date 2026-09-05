package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
)

// ShelfItem represents a custom collection / shelf with metadata.
type ShelfItem struct {
	ID               int64  `json:"id"`
	UserID           int64  `json:"user_id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	IsPublic         bool   `json:"is_public"`
	BookCount        int64  `json:"book_count"`
	OwnerUsername    string `json:"owner_username,omitempty"`
	OwnerDisplayName string `json:"owner_display_name,omitempty"`
	IsOwner          bool   `json:"is_owner"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// CreateShelfRequest is the payload for POST /api/shelves.
type CreateShelfRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// UpdateShelfRequest is the payload for PUT /api/shelves/{id}.
type UpdateShelfRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// AddBookToShelfRequest is the payload for POST /api/shelves/{id}/books.
type AddBookToShelfRequest struct {
	BookID int64 `json:"book_id"`
}

// UpdateBookShelvesRequest is the payload for PUT /api/books/{id}/shelves.
type UpdateBookShelvesRequest struct {
	ShelfIDs []int64 `json:"shelf_ids"`
}

// RegisterShelfRoutes registers routes under /api/shelves.
func RegisterShelfRoutes(r chi.Router, db *database.Queries) {
	// GET /api/shelves: List all shelves available to current user
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		rows, err := db.ListShelvesForUser(req.Context(), user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query shelves")
			return
		}

		items := make([]ShelfItem, 0, len(rows))
		for _, s := range rows {
			isOwner := s.UserID == user.ID
			items = append(items, ShelfItem{
				ID:               s.ID,
				UserID:           s.UserID,
				Name:             s.Name,
				Description:      s.Description,
				IsPublic:         s.IsPublic == 1,
				BookCount:        s.BookCount,
				OwnerUsername:    s.OwnerUsername,
				OwnerDisplayName: s.OwnerDisplayName,
				IsOwner:          isOwner,
				CreatedAt:        formatTime(s.CreatedAt),
				UpdatedAt:        formatTime(s.UpdatedAt),
			})
		}

		writeJSON(w, http.StatusOK, items)
	})

	// POST /api/shelves: Create a new shelf
	r.Post("/", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body CreateShelfRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "shelf name cannot be empty")
			return
		}

		var isPublic int64
		if body.IsPublic {
			isPublic = 1
		}

		shelf, err := db.CreateShelf(req.Context(), database.CreateShelfParams{
			UserID:      user.ID,
			Name:        name,
			Description: strings.TrimSpace(body.Description),
			IsPublic:    isPublic,
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				writeError(w, http.StatusConflict, fmt.Sprintf("a shelf named %q already exists", name))
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to create shelf")
			return
		}

		writeJSON(w, http.StatusCreated, ShelfItem{
			ID:               shelf.ID,
			UserID:           shelf.UserID,
			Name:             shelf.Name,
			Description:      shelf.Description,
			IsPublic:         shelf.IsPublic == 1,
			BookCount:        0,
			OwnerUsername:    user.Username,
			OwnerDisplayName: user.DisplayName,
			IsOwner:          true,
			CreatedAt:        formatTime(shelf.CreatedAt),
			UpdatedAt:        formatTime(shelf.UpdatedAt),
		})
	})

	// GET /api/shelves/{id}: Get shelf details
	r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		shelfID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shelf id")
			return
		}

		shelf, err := db.GetShelfByID(req.Context(), shelfID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "shelf not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query shelf")
			return
		}

		// Verify permission: owner, public shelf, or admin
		if shelf.UserID != user.ID && shelf.IsPublic == 0 && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		count, _ := db.CountBooksByShelf(req.Context(), shelf.ID)

		writeJSON(w, http.StatusOK, ShelfItem{
			ID:          shelf.ID,
			UserID:      shelf.UserID,
			Name:        shelf.Name,
			Description: shelf.Description,
			IsPublic:    shelf.IsPublic == 1,
			BookCount:   count,
			IsOwner:     shelf.UserID == user.ID,
			CreatedAt:   formatTime(shelf.CreatedAt),
			UpdatedAt:   formatTime(shelf.UpdatedAt),
		})
	})

	// PUT /api/shelves/{id}: Update shelf
	r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		shelfID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shelf id")
			return
		}

		existing, err := db.GetShelfByID(req.Context(), shelfID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "shelf not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query shelf")
			return
		}

		// Only owner or admin can edit
		if existing.UserID != user.ID && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "only the shelf owner can edit this shelf")
			return
		}

		var body UpdateShelfRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		name := strings.TrimSpace(body.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "shelf name cannot be empty")
			return
		}

		var isPublic int64
		if body.IsPublic {
			isPublic = 1
		}

		updated, err := db.UpdateShelf(req.Context(), database.UpdateShelfParams{
			Name:        name,
			Description: strings.TrimSpace(body.Description),
			IsPublic:    isPublic,
			ID:          shelfID,
			UserID:      existing.UserID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				writeError(w, http.StatusConflict, fmt.Sprintf("a shelf named %q already exists", name))
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to update shelf")
			return
		}

		count, _ := db.CountBooksByShelf(req.Context(), updated.ID)

		writeJSON(w, http.StatusOK, ShelfItem{
			ID:          updated.ID,
			UserID:      updated.UserID,
			Name:        updated.Name,
			Description: updated.Description,
			IsPublic:    updated.IsPublic == 1,
			BookCount:   count,
			IsOwner:     updated.UserID == user.ID,
			CreatedAt:   formatTime(updated.CreatedAt),
			UpdatedAt:   formatTime(updated.UpdatedAt),
		})
	})

	// DELETE /api/shelves/{id}: Delete shelf
	r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		shelfID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shelf id")
			return
		}

		existing, err := db.GetShelfByID(req.Context(), shelfID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "shelf not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query shelf")
			return
		}

		// Only owner or admin can delete
		if existing.UserID != user.ID && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "only the shelf owner can delete this shelf")
			return
		}

		if err := db.DeleteShelf(req.Context(), database.DeleteShelfParams{
			ID:     shelfID,
			UserID: existing.UserID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to delete shelf")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "shelf deleted successfully"})
	})

	// GET /api/shelves/{id}/books: List books in shelf
	r.Get("/{id}/books", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		shelfID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shelf id")
			return
		}

		shelf, err := db.GetShelfByID(req.Context(), shelfID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "shelf not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query shelf")
			return
		}

		// Verify view permission
		if shelf.UserID != user.ID && shelf.IsPublic == 0 && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}

		page, _ := strconv.Atoi(req.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(req.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 24
		}
		offset := (page - 1) * limit

		total, err := db.CountBooksByShelf(req.Context(), shelfID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to count shelf books")
			return
		}

		rows, err := db.ListBooksByShelfWithAuthorsAndProgress(req.Context(), database.ListBooksByShelfWithAuthorsAndProgressParams{
			UserID:  user.ID,
			ShelfID: shelfID,
			Limit:   int64(limit),
			Offset:  int64(offset),
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to query shelf books")
			return
		}

		items := make([]BookListItem, 0, len(rows))
		for _, b := range rows {
			items = append(items, toBookListItem(
				b.ID,
				b.Title,
				b.AuthorNames,
				b.TagNames,
				b.Description,
				b.Publisher,
				b.Language,
				b.PubDate,
				b.Series,
				b.SeriesIndex,
				b.FileSize,
				b.Format,
				b.CoverPath,
				b.CreatedAt,
				b.UpdatedAt,
				b.UserProgress,
				b.UserIsFinished == 1,
			))
		}

		writeJSON(w, http.StatusOK, PaginatedResponse[BookListItem]{
			Items: items,
			Page:  page,
			Limit: limit,
			Total: total,
		})
	})

	// POST /api/shelves/{id}/books: Add book to shelf
	r.Post("/{id}/books", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		shelfID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shelf id")
			return
		}

		shelf, err := db.GetShelfByID(req.Context(), shelfID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "shelf not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query shelf")
			return
		}

		// Only owner or admin can add books to shelf
		if shelf.UserID != user.ID && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "only the shelf owner can add books to this shelf")
			return
		}

		var body AddBookToShelfRequest
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil || body.BookID <= 0 {
			writeError(w, http.StatusBadRequest, "valid book_id is required")
			return
		}

		// Verify book exists
		if _, err := db.GetBookByID(req.Context(), body.BookID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "book not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to verify book")
			return
		}

		if err := db.AddBookToShelf(req.Context(), database.AddBookToShelfParams{
			ShelfID: shelfID,
			BookID:  body.BookID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to add book to shelf")
			return
		}

		writeJSON(w, http.StatusCreated, map[string]string{"message": "book added to shelf"})
	})

	// DELETE /api/shelves/{id}/books/{bookId}: Remove book from shelf
	r.Delete("/{id}/books/{bookId}", func(w http.ResponseWriter, req *http.Request) {
		user := auth.GetUser(req.Context())
		if user == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		shelfID, err := strconv.ParseInt(chi.URLParam(req, "id"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid shelf id")
			return
		}

		bookID, err := strconv.ParseInt(chi.URLParam(req, "bookId"), 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid book id")
			return
		}

		shelf, err := db.GetShelfByID(req.Context(), shelfID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "shelf not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "failed to query shelf")
			return
		}

		// Only owner or admin can remove books
		if shelf.UserID != user.ID && user.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "only the shelf owner can remove books from this shelf")
			return
		}

		if err := db.RemoveBookFromShelf(req.Context(), database.RemoveBookFromShelfParams{
			ShelfID: shelfID,
			BookID:  bookID,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to remove book from shelf")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"message": "book removed from shelf"})
	})
}
