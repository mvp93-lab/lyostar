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
	Tags        []string `json:"tags"`
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
	Tags        []string          `json:"tags"`
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

// UpdateBookMetadataRequest represents payload for updating book metadata.
type UpdateBookMetadataRequest struct {
	Title       string   `json:"title"`
	Authors     []string `json:"authors"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Publisher   string   `json:"publisher"`
	Language    string   `json:"language"`
	PubDate     string   `json:"pub_date"`
	Series      string   `json:"series"`
	SeriesIndex float64  `json:"series_index"`
}

// TagItem represents a tag with its book count.
type TagItem struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	BookCount int64  `json:"book_count"`
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

		// GET /api/tags (List all tags with book counts)
		api.Get("/tags", func(w http.ResponseWriter, r *http.Request) {
			tagRows, err := cfg.DB.ListTags(r.Context())
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to query tags")
				return
			}

			items := make([]TagItem, 0, len(tagRows))
			for _, tr := range tagRows {
				items = append(items, TagItem{
					ID:        tr.ID,
					Name:      tr.Name,
					BookCount: tr.BookCount,
				})
			}

			writeJSON(w, http.StatusOK, items)
		})

		// Protected endpoints (Require active authentication)
		api.Group(func(protected chi.Router) {
			protected.Use(RequireAuth)

			// Books endpoints
			protected.Route("/books", func(books chi.Router) {
				// Register reading progress routes (/continue-reading, /{id}/progress)
				RegisterProgressRoutes(books, cfg.DB.Queries)

				// Register bookmarks and highlights routes (/{id}/bookmarks, /{id}/highlights)
				RegisterBookmarkAndHighlightRoutes(books, cfg.DB.Queries)

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

					tagRows, _ := cfg.DB.GetTagsForBook(r.Context(), book.ID)
					tags := make([]string, 0, len(tagRows))
					for _, tr := range tagRows {
						tags = append(tags, tr.Name)
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
						Tags:        tags,
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
					tagFilter := strings.TrimSpace(r.URL.Query().Get("tag"))

					userID := int64(0)
					if u := auth.GetUser(r.Context()); u != nil {
						userID = u.ID
					}

					var items []BookListItem
					var total int64

					if tagFilter != "" {
						rows, err := cfg.DB.ListBooksByTagWithAuthorsAndProgress(r.Context(), database.ListBooksByTagWithAuthorsAndProgressParams{
							UserID: userID,
							Name:   tagFilter,
							Limit:  int64(limit),
							Offset: int64(offset),
						})
						if err != nil {
							writeError(w, http.StatusInternalServerError, "failed to query books by tag")
							return
						}

						count, err := cfg.DB.CountBooksByTag(r.Context(), tagFilter)
						if err != nil {
							writeError(w, http.StatusInternalServerError, "failed to count books by tag")
							return
						}
						total = count

						items = make([]BookListItem, 0, len(rows))
						for _, row := range rows {
							items = append(items, toBookListItem(
								row.ID, row.Title, row.AuthorNames, row.TagNames, row.Description,
								row.Publisher, row.Language, row.PubDate, row.Series,
								row.SeriesIndex, row.FileSize, row.Format, row.CoverPath,
								row.CreatedAt, row.UpdatedAt,
								row.UserProgress, row.UserIsFinished == 1,
							))
						}
					} else {
						rows, err := cfg.DB.ListBooksWithAuthorsAndProgress(r.Context(), database.ListBooksWithAuthorsAndProgressParams{
							UserID: userID,
							Limit:  int64(limit),
							Offset: int64(offset),
						})
						if err != nil {
							writeError(w, http.StatusInternalServerError, "failed to query books")
							return
						}

						count, err := cfg.DB.CountBooks(r.Context())
						if err != nil {
							writeError(w, http.StatusInternalServerError, "failed to count books")
							return
						}
						total = count

						items = make([]BookListItem, 0, len(rows))
						for _, row := range rows {
							items = append(items, toBookListItem(
								row.ID, row.Title, row.AuthorNames, row.TagNames, row.Description,
								row.Publisher, row.Language, row.PubDate, row.Series,
								row.SeriesIndex, row.FileSize, row.Format, row.CoverPath,
								row.CreatedAt, row.UpdatedAt,
								row.UserProgress, row.UserIsFinished == 1,
							))
						}
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

					tagRows, _ := cfg.DB.GetTagsForBook(r.Context(), id)
					tags := make([]string, 0, len(tagRows))
					for _, tr := range tagRows {
						tags = append(tags, tr.Name)
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
						Tags:        tags,
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

				// PUT /api/books/{id} (Update book metadata, requires can_edit)
				book.With(RequireEdit).Put("/", func(w http.ResponseWriter, r *http.Request) {
					id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
					if err != nil {
						writeError(w, http.StatusBadRequest, "invalid book id")
						return
					}

					_, err = cfg.DB.GetBookByID(r.Context(), id)
					if err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							writeError(w, http.StatusNotFound, "book not found")
							return
						}
						writeError(w, http.StatusInternalServerError, "database error")
						return
					}

					var req UpdateBookMetadataRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						writeError(w, http.StatusBadRequest, "invalid json payload")
						return
					}

					req.Title = strings.TrimSpace(req.Title)
					if req.Title == "" {
						writeError(w, http.StatusBadRequest, "title cannot be empty")
						return
					}

					updatedBook, err := cfg.DB.UpdateBookMetadata(r.Context(), database.UpdateBookMetadataParams{
						ID:          id,
						Title:       req.Title,
						Description: strings.TrimSpace(req.Description),
						Publisher:   strings.TrimSpace(req.Publisher),
						Language:    strings.TrimSpace(req.Language),
						PubDate:     strings.TrimSpace(req.PubDate),
						Series:      strings.TrimSpace(req.Series),
						SeriesIndex: req.SeriesIndex,
					})
					if err != nil {
						writeError(w, http.StatusInternalServerError, "failed to update book metadata")
						return
					}

					if req.Authors != nil {
						if err := cfg.DB.ClearBookAuthors(r.Context(), id); err != nil {
							writeError(w, http.StatusInternalServerError, "failed to update book authors")
							return
						}
						for _, authorName := range req.Authors {
							trimmed := strings.TrimSpace(authorName)
							if trimmed == "" {
								continue
							}
							author, err := cfg.DB.CreateAuthor(r.Context(), trimmed)
							if err != nil {
								continue
							}
							_ = cfg.DB.AddBookAuthor(r.Context(), database.AddBookAuthorParams{
								BookID:   id,
								AuthorID: author.ID,
								Role:     "aut",
							})
						}
					}

					if req.Tags != nil {
						if err := cfg.DB.ClearBookTags(r.Context(), id); err != nil {
							writeError(w, http.StatusInternalServerError, "failed to update book tags")
							return
						}
						for _, tagName := range req.Tags {
							trimmed := strings.TrimSpace(tagName)
							if trimmed == "" {
								continue
							}
							tag, err := cfg.DB.CreateTag(r.Context(), trimmed)
							if err != nil {
								continue
							}
							_ = cfg.DB.AddBookTag(r.Context(), database.AddBookTagParams{
								BookID: id,
								TagID:  tag.ID,
							})
						}
					}

					tagRows, _ := cfg.DB.GetTagsForBook(r.Context(), id)
					tags := make([]string, 0, len(tagRows))
					for _, tr := range tagRows {
						tags = append(tags, tr.Name)
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

					hasCover := updatedBook.CoverPath != ""
					var coverURL string
					if hasCover {
						coverURL = fmt.Sprintf("/api/books/%d/cover", updatedBook.ID)
					}

					var userProgress float64
					var isFinished bool
					if u := auth.GetUser(r.Context()); u != nil {
						if prog, err := cfg.DB.GetProgress(r.Context(), database.GetProgressParams{UserID: u.ID, BookID: updatedBook.ID}); err == nil {
							userProgress = prog.Progress
							isFinished = prog.IsFinished == 1
						}
					}

					writeJSON(w, http.StatusOK, BookDetailResponse{
						ID:          updatedBook.ID,
						Title:       updatedBook.Title,
						Authors:     authors,
						Tags:        tags,
						Description: updatedBook.Description,
						Publisher:   updatedBook.Publisher,
						Language:    updatedBook.Language,
						PubDate:     updatedBook.PubDate,
						Series:      updatedBook.Series,
						SeriesIndex: updatedBook.SeriesIndex,
						FileSize:    updatedBook.FileSize,
						Format:      updatedBook.Format,
						HasCover:    hasCover,
						CoverURL:    coverURL,
						FileURL:     fmt.Sprintf("/api/books/%d/file", updatedBook.ID),
						DownloadURL: fmt.Sprintf("/api/books/%d/download", updatedBook.ID),
						Progress:    userProgress,
						IsFinished:  isFinished,
						CreatedAt:   formatTime(updatedBook.CreatedAt),
						UpdatedAt:   formatTime(updatedBook.UpdatedAt),
					})
				})

				// DELETE /api/books/{id} (Delete book from library, requires can_delete)
				book.With(RequireDelete).Delete("/", func(w http.ResponseWriter, r *http.Request) {
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

					if err := cfg.DB.DeleteBook(r.Context(), id); err != nil {
						writeError(w, http.StatusInternalServerError, "failed to delete book")
						return
					}

					// Remove cover thumbnail cache file if exists
					if b.CoverPath != "" {
						_ = os.Remove(b.CoverPath)
					}

					// Strictly read-only policy:
					// ONLY remove file on disk if it was uploaded to UploadsDir!
					// Files in /books are strictly read-only and preserved on disk.
					if cfg.UploadsDir != "" && b.FilePath != "" {
						rel, err := filepath.Rel(cfg.UploadsDir, b.FilePath)
						if err == nil && !strings.HasPrefix(rel, "..") && !filepath.IsAbs(rel) {
							_ = os.Remove(b.FilePath)
						}
					}

					writeJSON(w, http.StatusOK, map[string]any{
						"status":  "deleted",
						"id":      id,
						"message": "Book deleted successfully",
					})
				})

				// GET /api/books/{id}/shelves: List shelf IDs this book belongs to for current user
				book.Get("/shelves", func(w http.ResponseWriter, r *http.Request) {
					user := auth.GetUser(r.Context())
					if user == nil {
						writeError(w, http.StatusUnauthorized, "unauthorized")
						return
					}

					id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
					if err != nil {
						writeError(w, http.StatusBadRequest, "invalid book id")
						return
					}

					shelfIDs, err := cfg.DB.GetBookShelfIDsForUser(r.Context(), database.GetBookShelfIDsForUserParams{
						BookID: id,
						UserID: user.ID,
					})
					if err != nil {
						writeError(w, http.StatusInternalServerError, "failed to query book shelves")
						return
					}
					if shelfIDs == nil {
						shelfIDs = []int64{}
					}

					writeJSON(w, http.StatusOK, shelfIDs)
				})

				// PUT /api/books/{id}/shelves: Batch update user shelves this book belongs to
				book.Put("/shelves", func(w http.ResponseWriter, r *http.Request) {
					user := auth.GetUser(r.Context())
					if user == nil {
						writeError(w, http.StatusUnauthorized, "unauthorized")
						return
					}

					id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
					if err != nil {
						writeError(w, http.StatusBadRequest, "invalid book id")
						return
					}

					if _, err := cfg.DB.GetBookByID(r.Context(), id); err != nil {
						if errors.Is(err, sql.ErrNoRows) {
							writeError(w, http.StatusNotFound, "book not found")
							return
						}
						writeError(w, http.StatusInternalServerError, "failed to query book")
						return
					}

					var body UpdateBookShelvesRequest
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						writeError(w, http.StatusBadRequest, "invalid request body")
						return
					}

					userShelves, err := cfg.DB.ListShelvesForUser(r.Context(), user.ID)
					if err != nil {
						writeError(w, http.StatusInternalServerError, "failed to list shelves")
						return
					}

					ownedMap := make(map[int64]bool)
					for _, s := range userShelves {
						if s.UserID == user.ID {
							ownedMap[s.ID] = true
						}
					}

					targetSet := make(map[int64]bool)
					for _, sid := range body.ShelfIDs {
						if ownedMap[sid] {
							targetSet[sid] = true
						}
					}

					for sid := range ownedMap {
						if targetSet[sid] {
							_ = cfg.DB.AddBookToShelf(r.Context(), database.AddBookToShelfParams{
								ShelfID: sid,
								BookID:  id,
							})
						} else {
							_ = cfg.DB.RemoveBookFromShelf(r.Context(), database.RemoveBookFromShelfParams{
								ShelfID: sid,
								BookID:  id,
							})
						}
					}

					shelfIDs, _ := cfg.DB.GetBookShelfIDsForUser(r.Context(), database.GetBookShelfIDsForUserParams{
						BookID: id,
						UserID: user.ID,
					})
					if shelfIDs == nil {
						shelfIDs = []int64{}
					}

					writeJSON(w, http.StatusOK, shelfIDs)
				})
			})

			// Shelves endpoints (/api/shelves/*)
			protected.Route("/shelves", func(shelves chi.Router) {
				RegisterShelfRoutes(shelves, cfg.DB.Queries)
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
					row.ID, row.Title, row.AuthorNames, row.TagNames, row.Description,
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
	id int64, title string, authorNames any, tagNames any, description, publisher, language, pubDate, series string,
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

	var tags []string
	if tagsStr, ok := tagNames.(string); ok && tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ", ") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tags = append(tags, trimmed)
			}
		}
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
		Tags:        tags,
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
