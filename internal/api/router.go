package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lyostar/lyostar/internal/database"
	"github.com/lyostar/lyostar/internal/scanner"
)

// RouterConfig holds dependencies for the HTTP router.
type RouterConfig struct {
	DB       *database.DB
	Scanner  *scanner.Scanner
	StaticFS fs.FS
	Version  string
}

// HealthResponse represents the health check payload.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// BookListItem represents book item in list and search responses.
type BookListItem struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Description string   `json:"description"`
	Publisher   string   `json:"publisher"`
	Language    string   `json:"language"`
	PubDate     string   `json:"pub_date"`
	Series      string   `json:"series"`
	SeriesIndex float64  `json:"series_index"`
	FileSize    int64    `json:"file_size"`
	Format      string   `json:"format"`
	HasCover    bool     `json:"has_cover"`
	CoverURL    string   `json:"cover_url,omitempty"`
	FileURL     string   `json:"file_url"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// BookDetailResponse represents full details of a book.
type BookDetailResponse struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Authors     []AuthorRoleItem  `json:"authors"`
	Description string            `json:"description"`
	Publisher   string            `json:"publisher"`
	Language    string            `json:"language"`
	PubDate     string            `json:"pub_date"`
	Series      string            `json:"series"`
	SeriesIndex float64           `json:"series_index"`
	FileSize    int64             `json:"file_size"`
	Format      string            `json:"format"`
	HasCover    bool              `json:"has_cover"`
	CoverURL    string            `json:"cover_url,omitempty"`
	FileURL     string            `json:"file_url"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// AuthorRoleItem represents an author with their contribution role.
type AuthorRoleItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// PaginatedResponse wraps list items with pagination metadata.
type PaginatedResponse[T any] struct {
	Items []T   `json:"items"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
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
			writeJSON(w, http.StatusOK, HealthResponse{
				Status:  "ok",
				Version: version,
			})
		})

		// Trigger scanner
		api.Post("/scan", func(w http.ResponseWriter, r *http.Request) {
			if cfg.Scanner == nil {
				writeError(w, http.StatusInternalServerError, "scanner not configured")
				return
			}
			_ = cfg.Scanner.Start(r.Context())
			writeJSON(w, http.StatusAccepted, map[string]string{
				"status":  "scanning",
				"message": "Scan started in background",
			})
		})

		// Books endpoints
		api.Route("/books", func(books chi.Router) {
			// GET /api/books
			books.Get("/", func(w http.ResponseWriter, r *http.Request) {
				page, limit := parsePagination(r)
				offset := (page - 1) * limit

				rows, err := cfg.DB.ListBooksWithAuthors(r.Context(), database.ListBooksWithAuthorsParams{
					Limit:  int64(limit),
					Offset: int64(offset),
				})
				if err != nil {
					writeError(w, http.StatusInternalServerError, "failed to query books")
					return
				}

				total, err := cfg.DB.CountBooks(r.Context())
				if err != nil {
					writeError(w, http.StatusInternalServerError, "failed to count books")
					return
				}

				items := make([]BookListItem, 0, len(rows))
				for _, row := range rows {
					items = append(items, toBookListItem(
						row.ID, row.Title, row.AuthorNames, row.Description,
						row.Publisher, row.Language, row.PubDate, row.Series,
						row.SeriesIndex, row.FileSize, row.Format, row.CoverPath,
						row.CreatedAt, row.UpdatedAt,
					))
				}

				writeJSON(w, http.StatusOK, PaginatedResponse[BookListItem]{
					Items: items,
					Page:  page,
					Limit: limit,
					Total: total,
				})
			})

			// Specific book routes
			books.Route("/{id}", func(book chi.Router) {
				// GET /api/books/{id}
				book.Get("/", func(w http.ResponseWriter, r *http.Request) {
					id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
					if err != nil {
						writeError(w, http.StatusBadRequest, "invalid book id")
						return
					}

					b, err := cfg.DB.GetBookByID(r.Context(), id)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							writeError(w, http.StatusNotFound, "book not found")
							return
						}
						writeError(w, http.StatusInternalServerError, "database error")
						return
					}

					authorRows, err := cfg.DB.GetAuthorsForBook(r.Context(), id)
					if err != nil {
						writeError(w, http.StatusInternalServerError, "failed to fetch authors")
						return
					}

					authors := make([]AuthorRoleItem, 0, len(authorRows))
					for _, ar := range authorRows {
						authors = append(authors, AuthorRoleItem{
							ID:   ar.ID,
							Name: ar.Name,
							Role: ar.Role,
						})
					}

					hasCover := b.CoverPath != ""
					var coverURL string
					if hasCover {
						coverURL = fmt.Sprintf("/api/books/%d/cover", b.ID)
					}

					writeJSON(w, http.StatusOK, BookDetailResponse{
						ID:          b.ID,
						Title:       b.Title,
						Authors:     authors,
						Description: b.Description,
						Publisher:   b.Publisher,
						Language:    b.Language,
						PubDate:     b.PubDate,
						Series:      b.Series,
						SeriesIndex: b.SeriesIndex,
						FileSize:    b.FileSize,
						Format:      b.Format,
						HasCover:    hasCover,
						CoverURL:    coverURL,
						FileURL:     fmt.Sprintf("/api/books/%d/file", b.ID),
						CreatedAt:   formatTime(b.CreatedAt),
						UpdatedAt:   formatTime(b.UpdatedAt),
					})
				})

				// GET /api/books/{id}/cover
				book.Get("/cover", func(w http.ResponseWriter, r *http.Request) {
					id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
					if err != nil {
						writeError(w, http.StatusBadRequest, "invalid book id")
						return
					}

					b, err := cfg.DB.GetBookByID(r.Context(), id)
					if err != nil || b.CoverPath == "" {
						writeError(w, http.StatusNotFound, "cover not found")
						return
					}

					if _, err := os.Stat(b.CoverPath); err != nil {
						writeError(w, http.StatusNotFound, "cover file not found")
						return
					}

					w.Header().Set("Cache-Control", "public, max-age=86400")
					w.Header().Set("Content-Type", "image/webp")
					http.ServeFile(w, r, b.CoverPath)
				})

				// GET /api/books/{id}/file
				book.Get("/file", func(w http.ResponseWriter, r *http.Request) {
					id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
					if err != nil {
						writeError(w, http.StatusBadRequest, "invalid book id")
						return
					}

					b, err := cfg.DB.GetBookByID(r.Context(), id)
					if err != nil {
						writeError(w, http.StatusNotFound, "book not found")
						return
					}

					if _, err := os.Stat(b.FilePath); err != nil {
						writeError(w, http.StatusNotFound, "book file not found")
						return
					}

					filename := filepath.Base(b.FilePath)
					switch strings.ToLower(b.Format) {
					case "pdf":
						w.Header().Set("Content-Type", "application/pdf")
					default:
						w.Header().Set("Content-Type", "application/epub+zip")
					}
					w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
					http.ServeFile(w, r, b.FilePath)
				})
			})
		})

		// GET /api/search
		api.Get("/search", func(w http.ResponseWriter, r *http.Request) {
			rawQuery := strings.TrimSpace(r.URL.Query().Get("q"))
			if rawQuery == "" {
				writeJSON(w, http.StatusOK, PaginatedResponse[BookListItem]{
					Items: []BookListItem{},
					Page:  1,
					Limit: 20,
					Total: 0,
				})
				return
			}

			ftsQuery := sanitizeFTS5Query(rawQuery)
			if ftsQuery == "" {
				writeJSON(w, http.StatusOK, PaginatedResponse[BookListItem]{
					Items: []BookListItem{},
					Page:  1,
					Limit: 20,
					Total: 0,
				})
				return
			}

			page, limit := parsePagination(r)
			offset := (page - 1) * limit

			rows, err := cfg.DB.SearchBooksFTSWithAuthors(r.Context(), database.SearchBooksFTSWithAuthorsParams{
				Fulltext: ftsQuery,
				Limit:    int64(limit),
				Offset:   int64(offset),
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "search query failed")
				return
			}

			total, err := cfg.DB.CountSearchBooksFTS(r.Context(), ftsQuery)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to count search results")
				return
			}

			items := make([]BookListItem, 0, len(rows))
			for _, row := range rows {
				items = append(items, toBookListItem(
					row.ID, row.Title, row.AuthorNames, row.Description,
					row.Publisher, row.Language, row.PubDate, row.Series,
					row.SeriesIndex, row.FileSize, row.Format, row.CoverPath,
					row.CreatedAt, row.UpdatedAt,
				))
			}

			writeJSON(w, http.StatusOK, PaginatedResponse[BookListItem]{
				Items: items,
				Page:  page,
				Limit: limit,
				Total: total,
			})
		})

		// 404 handler specifically for unhandled /api routes
		api.NotFound(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, http.StatusNotFound, "endpoint not found")
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

func parsePagination(r *http.Request) (int, int) {
	page := 1
	limit := 20

	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && l > 0 {
		if l > 100 {
			l = 100
		}
		limit = l
	}

	return page, limit
}

func sanitizeFTS5Query(q string) string {
	terms := strings.Fields(q)
	if len(terms) == 0 {
		return ""
	}

	var sanitized []string
	for _, t := range terms {
		clean := strings.Map(func(r rune) rune {
			if strings.ContainsRune(`"()[]{}:*^'\/`, r) {
				return -1
			}
			return r
		}, t)
		if clean != "" {
			sanitized = append(sanitized, clean+"*")
		}
	}

	return strings.Join(sanitized, " ")
}

func toBookListItem(
	id int64, title string, authorNames any, description, publisher, language, pubDate, series string,
	seriesIndex float64, fileSize int64, format, coverPath string,
	createdAt, updatedAt any,
) BookListItem {
	var authors []string
	if namesStr, ok := authorNames.(string); ok && namesStr != "" {
		for _, a := range strings.Split(namesStr, ", ") {
			if trimmed := strings.TrimSpace(a); trimmed != "" {
				authors = append(authors, trimmed)
			}
		}
	}
	if len(authors) == 0 {
		authors = []string{"Unknown"}
	}

	hasCover := coverPath != ""
	var coverURL string
	if hasCover {
		coverURL = fmt.Sprintf("/api/books/%d/cover", id)
	}

	return BookListItem{
		ID:          id,
		Title:       title,
		Authors:     authors,
		Description: description,
		Publisher:   publisher,
		Language:    language,
		PubDate:     pubDate,
		Series:      series,
		SeriesIndex: seriesIndex,
		FileSize:    fileSize,
		Format:      format,
		HasCover:    hasCover,
		CoverURL:    coverURL,
		FileURL:     fmt.Sprintf("/api/books/%d/file", id),
		CreatedAt:   formatTime(createdAt),
		UpdatedAt:   formatTime(updatedAt),
	}
}

func formatTime(val any) string {
	switch v := val.(type) {
	case time.Time:
		return v.Format(time.RFC3339)
	case string:
		return v
	default:
		return ""
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{
		"error": msg,
	})
}
