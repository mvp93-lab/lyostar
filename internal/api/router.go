package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
	"github.com/lyostar/lyostar/internal/scanner"
)

// RouterConfig holds dependencies for the HTTP router.
type RouterConfig struct {
	DB         *database.DB
	Scanner    *scanner.Scanner
	UploadsDir string
	StaticFS   fs.FS
	Version    string
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
	DownloadURL string   `json:"download_url"`
	Progress    float64  `json:"progress"`
	IsFinished  bool     `json:"is_finished"`
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
	DownloadURL string            `json:"download_url"`
	Progress    float64           `json:"progress"`
	IsFinished  bool              `json:"is_finished"`
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
	r.Use(cfg.AuthMiddleware)

	version := cfg.Version
	if version == "" {
		version = "0.1.0-dev"
	}

	// API routes
	r.Route("/api", func(api chi.Router) {
		// Auth routes (/api/auth/* and /api/users/*)
		cfg.RegisterAuthRoutes(api)

		api.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, HealthResponse{
				Status:  "ok",
				Version: version,
			})
		})

		// Trigger scanner (Admin only)
		api.With(RequireAuth, RequireAdmin).Post("/scan", func(w http.ResponseWriter, r *http.Request) {
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

		// Protected endpoints (Require active authentication)
		api.Group(func(protected chi.Router) {
			protected.Use(RequireAuth)

			// Books endpoints
			protected.Route("/books", func(books chi.Router) {
				// Register reading progress routes (/continue-reading, /{id}/progress)
				RegisterProgressRoutes(books, cfg.DB.Queries)

				// POST /api/books/upload (Upload ebook, requires can_upload)
				books.With(RequireUpload).Post("/upload", func(w http.ResponseWriter, r *http.Request) {
					if cfg.UploadsDir == "" {
						writeError(w, http.StatusInternalServerError, "uploads directory not configured")
						return
					}

					// Limit upload body to 100MB
					r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
					if err := r.ParseMultipartForm(100 << 20); err != nil {
						writeError(w, http.StatusBadRequest, "file too large or invalid multipart form (max 100MB)")
						return
					}

					file, header, err := r.FormFile("file")
					if err != nil {
						writeError(w, http.StatusBadRequest, "missing 'file' in upload form")
						return
					}
					defer file.Close()

					ext := strings.ToLower(filepath.Ext(header.Filename))
					if ext != ".epub" && ext != ".pdf" {
						writeError(w, http.StatusBadRequest, "unsupported file format; only .epub and .pdf are supported")
						return
					}

					origBase := filepath.Base(header.Filename)
					if origBase == "." || origBase == "/" || origBase == "" {
						origBase = "upload" + ext
					}

					destName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), origBase)
					destPath := filepath.Join(cfg.UploadsDir, destName)

					out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
					if err != nil {
						writeError(w, http.StatusInternalServerError, "failed to save uploaded file")
						return
					}

					_, copyErr := io.Copy(out, file)
					closeErr := out.Close()
					if copyErr != nil || closeErr != nil {
						_ = os.Remove(destPath)
						writeError(w, http.StatusInternalServerError, "failed to write uploaded file")
						return
					}

					book, err := cfg.Scanner.IndexFile(r.Context(), destPath)
					if err != nil {
						_ = os.Remove(destPath)
						if errors.Is(err, scanner.ErrBookAlreadyExists) {
							writeError(w, http.StatusConflict, "this book already exists in the library")
							return
						}
						writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse book: %v", err))
						return
					}

					authorRows, _ := cfg.DB.GetAuthorsForBook(r.Context(), book.ID)
					authors := make([]AuthorRoleItem, 0, len(authorRows))
					for _, ar := range authorRows {
						authors = append(authors, AuthorRoleItem{
							ID:   ar.ID,
							Name: ar.Name,
							Role: ar.Role,
						})
					}

					hasCover := book.CoverPath != ""
					var coverURL string
					if hasCover {
						coverURL = fmt.Sprintf("/api/books/%d/cover", book.ID)
					}

					writeJSON(w, http.StatusCreated, BookDetailResponse{
						ID:          book.ID,
						Title:       book.Title,
						Authors:     authors,
						Description: book.Description,
						Publisher:   book.Publisher,
						Language:    book.Language,
						PubDate:     book.PubDate,
						Series:      book.Series,
						SeriesIndex: book.SeriesIndex,
						FileSize:    book.FileSize,
						Format:      book.Format,
						HasCover:    hasCover,
						CoverURL:    coverURL,
						FileURL:     fmt.Sprintf("/api/books/%d/file", book.ID),
						DownloadURL: fmt.Sprintf("/api/books/%d/download", book.ID),
						Progress:    0,
						IsFinished:  false,
						CreatedAt:   formatTime(book.CreatedAt),
						UpdatedAt:   formatTime(book.UpdatedAt),
					})
				})

				// GET /api/books
				books.Get("/", func(w http.ResponseWriter, r *http.Request) {
					page, limit := parsePagination(r)
					offset := (page - 1) * limit

					userID := int64(0)
					if u := auth.GetUser(r.Context()); u != nil {
						userID = u.ID
					}

					rows, err := cfg.DB.ListBooksWithAuthorsAndProgress(r.Context(), database.ListBooksWithAuthorsAndProgressParams{
						UserID: userID,
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
							row.UserProgress, row.UserIsFinished == 1,
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

					var userProgress float64
					var isFinished bool
					if u := auth.GetUser(r.Context()); u != nil {
						if prog, err := cfg.DB.GetProgress(r.Context(), database.GetProgressParams{UserID: u.ID, BookID: b.ID}); err == nil {
							userProgress = prog.Progress
							isFinished = prog.IsFinished == 1
						}
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
						DownloadURL: fmt.Sprintf("/api/books/%d/download", b.ID),
						Progress:    userProgress,
						IsFinished:  isFinished,
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

				// GET /api/books/{id}/file (Read in browser, requires can_read)
				book.With(RequireRead).Get("/file", func(w http.ResponseWriter, r *http.Request) {
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

				// GET /api/books/{id}/download (Direct file download, requires can_download)
				book.With(RequireDownload).Get("/download", func(w http.ResponseWriter, r *http.Request) {
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
					w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
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

			userID := int64(0)
			if u := auth.GetUser(r.Context()); u != nil {
				userID = u.ID
			}

			rows, err := cfg.DB.SearchBooksFTSWithAuthorsAndProgress(r.Context(), database.SearchBooksFTSWithAuthorsAndProgressParams{
				UserID:   userID,
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
					row.UserProgress, row.UserIsFinished == 1,
				))
			}

			writeJSON(w, http.StatusOK, PaginatedResponse[BookListItem]{
				Items: items,
				Page:  page,
				Limit: limit,
				Total: total,
			})
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
	progress float64, isFinished bool,
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
		DownloadURL: fmt.Sprintf("/api/books/%d/download", id),
		Progress:    progress,
		IsFinished:  isFinished,
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
