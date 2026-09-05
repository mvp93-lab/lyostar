package scanner

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/lyostar/lyostar/internal/database"
	"github.com/lyostar/lyostar/internal/epub"
	"github.com/lyostar/lyostar/internal/pdf"
)

// Stats tracks statistics of a scan run.
type Stats struct {
	TotalFiles int64 `json:"total_files"`
	Added      int64 `json:"added"`
	Skipped    int64 `json:"skipped"`
	Errors     int64 `json:"errors"`
}

var (
	// ErrBookAlreadyExists indicates the book SHA-256 is already indexed.
	ErrBookAlreadyExists = errors.New("book already exists in library")
	// ErrUnsupportedFormat indicates the file extension is not supported.
	ErrUnsupportedFormat = errors.New("unsupported book format")
)

// Scanner handles discovery and indexing of EPUB files in the books directory.
type Scanner struct {
	booksDir    string
	coversDir   string
	uploadsDir  string
	db          *database.DB
	workerCount int
	mu          sync.Mutex // ensures only one active scan run at a time
}

// New creates a new Scanner instance with bounded worker count.
func New(booksDir, coversDir string, db *database.DB, uploadsDir ...string) *Scanner {
	// Keep workers bounded (min 2, max 4) to stay strictly under RAM budget (< 30MB)
	workers := runtime.NumCPU()
	if workers < 2 {
		workers = 2
	}
	if workers > 4 {
		workers = 4
	}

	var upDir string
	if len(uploadsDir) > 0 {
		upDir = uploadsDir[0]
	}

	return &Scanner{
		booksDir:    booksDir,
		coversDir:   coversDir,
		uploadsDir:  upDir,
		db:          db,
		workerCount: workers,
	}
}

// Start initiates an initial scan run in the background.
func (s *Scanner) Start(ctx context.Context) error {
	go func() {
		log.Printf("[Scanner] Starting initial scan of %s...", s.booksDir)
		stats, err := s.Scan(ctx)
		if err != nil {
			log.Printf("[Scanner] Scan error: %v", err)
			return
		}
		log.Printf("[Scanner] Scan finished: %d discovered, %d added, %d skipped, %d errors",
			stats.TotalFiles, stats.Added, stats.Skipped, stats.Errors)
	}()
	return nil
}

// Scan performs a full scan of the books directory.
func (s *Scanner) Scan(ctx context.Context) (*Stats, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify books directory exists
	info, err := os.Stat(s.booksDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[Scanner] Books directory %s does not exist, skipping scan", s.booksDir)
			return &Stats{}, nil
		}
		return nil, fmt.Errorf("failed to access books directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("books directory %s is not a directory", s.booksDir)
	}

	stats := &Stats{}
	taskChan := make(chan string, 100)

	var wg sync.WaitGroup
	// Start fixed-size worker pool
	for i := 0; i < s.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range taskChan {
				if ctx.Err() != nil {
					return
				}
				s.processFile(ctx, filePath, stats)
			}
		}()
	}

	dirsToWalk := []string{s.booksDir}
	if s.uploadsDir != "" && s.uploadsDir != s.booksDir {
		if uinfo, err := os.Stat(s.uploadsDir); err == nil && uinfo.IsDir() {
			dirsToWalk = append(dirsToWalk, s.uploadsDir)
		}
	}

	// Walk configured directories
	for _, dir := range dirsToWalk {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Printf("[Scanner] Error accessing path %s: %v", path, err)
				atomic.AddInt64(&stats.Errors, 1)
				return nil
			}

			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Skip hidden files and directories
			name := d.Name()
			if strings.HasPrefix(name, ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				return nil
			}

			// Process .epub and .pdf files (case-insensitive)
			ext := strings.ToLower(filepath.Ext(name))
			if ext == ".epub" || ext == ".pdf" {
				atomic.AddInt64(&stats.TotalFiles, 1)
				taskChan <- path
			}

			return nil
		})
	}

	close(taskChan)
	wg.Wait()

	return stats, nil
}

// IndexFile indexes a single book file into the library immediately.
func (s *Scanner) IndexFile(ctx context.Context, filePath string) (*database.Book, error) {
	return s.indexFileInternal(ctx, filePath)
}

func (s *Scanner) processFile(ctx context.Context, filePath string, stats *Stats) {
	_, err := s.indexFileInternal(ctx, filePath)
	if err != nil {
		if errors.Is(err, ErrBookAlreadyExists) {
			atomic.AddInt64(&stats.Skipped, 1)
			return
		}
		log.Printf("[Scanner] Failed to process %s: %v", filePath, err)
		atomic.AddInt64(&stats.Errors, 1)
		return
	}
	atomic.AddInt64(&stats.Added, 1)
}

func (s *Scanner) indexFileInternal(ctx context.Context, filePath string) (*database.Book, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}
	fileSize := fi.Size()

	// Compute SHA-256 checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to compute sha256 for %s: %w", filePath, err)
	}
	fileSha256 := hex.EncodeToString(hasher.Sum(nil))

	// Check if book already exists in SQLite
	existing, err := s.db.GetBookBySHA256(ctx, fileSha256)
	if err == nil && existing.ID != 0 {
		return nil, ErrBookAlreadyExists
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("database lookup error for sha256 %s: %w", fileSha256, err)
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to rewind file %s: %w", filePath, err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var format string
	var meta struct {
		Title       string
		Authors     []string
		Tags        []string
		Description string
		Publisher   string
		Language    string
		PubDate     string
		Series      string
		SeriesIndex float64
	}
	var coverData []byte

	switch ext {
	case ".epub":
		format = "epub"
		bookInfo, err := epub.Parse(file, fileSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EPUB %s: %w", filePath, err)
		}
		meta = struct {
			Title       string
			Authors     []string
			Tags        []string
			Description string
			Publisher   string
			Language    string
			PubDate     string
			Series      string
			SeriesIndex float64
		}{
			Title:       bookInfo.Metadata.Title,
			Authors:     bookInfo.Metadata.Authors,
			Tags:        bookInfo.Metadata.Tags,
			Description: bookInfo.Metadata.Description,
			Publisher:   bookInfo.Metadata.Publisher,
			Language:    bookInfo.Metadata.Language,
			PubDate:     bookInfo.Metadata.PubDate,
			Series:      bookInfo.Metadata.Series,
			SeriesIndex: bookInfo.Metadata.SeriesIndex,
		}
		coverData = bookInfo.CoverData

	case ".pdf":
		format = "pdf"
		bookInfo, err := pdf.Parse(file, fileSize)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PDF %s: %w", filePath, err)
		}
		meta = struct {
			Title       string
			Authors     []string
			Tags        []string
			Description string
			Publisher   string
			Language    string
			PubDate     string
			Series      string
			SeriesIndex float64
		}{
			Title:       bookInfo.Metadata.Title,
			Authors:     bookInfo.Metadata.Authors,
			Tags:        bookInfo.Metadata.Tags,
			Description: bookInfo.Metadata.Description,
			Publisher:   bookInfo.Metadata.Publisher,
			Language:    bookInfo.Metadata.Language,
			PubDate:     bookInfo.Metadata.PubDate,
			Series:      bookInfo.Metadata.Series,
			SeriesIndex: bookInfo.Metadata.SeriesIndex,
		}
		coverData = bookInfo.CoverData

	default:
		return nil, ErrUnsupportedFormat
	}

	// Title fallback: if empty in metadata, use filename without extension
	title := strings.TrimSpace(meta.Title)
	if title == "" {
		title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		title = strings.ReplaceAll(title, "_", " ")
	}

	// Handle cover image extraction and downscaling
	var coverPath string
	if len(coverData) > 0 {
		coverFilename := fileSha256 + ".webp"
		fullCoverPath := filepath.Join(s.coversDir, coverFilename)

		// Check if cover already exists in cache
		if _, err := os.Stat(fullCoverPath); os.IsNotExist(err) {
			if err := epub.SaveCoverThumbnail(coverData, fullCoverPath, epub.DefaultMaxCoverWidth); err != nil {
				log.Printf("[Scanner] Failed to save cover thumbnail for %s: %v", filePath, err)
			} else {
				coverPath = fullCoverPath
			}
		} else {
			coverPath = fullCoverPath
		}
	}

	// Insert book record into database
	createdBook, err := s.db.CreateBook(ctx, database.CreateBookParams{
		Title:       title,
		FilePath:    filePath,
		FileSha256:  fileSha256,
		FileSize:    fileSize,
		Format:      format,
		Description: meta.Description,
		Publisher:   meta.Publisher,
		Language:    meta.Language,
		PubDate:     meta.PubDate,
		Series:      meta.Series,
		SeriesIndex: meta.SeriesIndex,
		CoverPath:   coverPath,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to insert book %s into database: %w", title, err)
	}

	// Link authors
	authors := meta.Authors
	if len(authors) == 0 {
		authors = []string{"Unknown"}
	}

	for _, authorName := range authors {
		authorName = strings.TrimSpace(authorName)
		if authorName == "" {
			continue
		}
		author, err := s.db.CreateAuthor(ctx, authorName)
		if err != nil {
			log.Printf("[Scanner] Failed to create/get author %s: %v", authorName, err)
			continue
		}

		if err := s.db.AddBookAuthor(ctx, database.AddBookAuthorParams{
			BookID:   createdBook.ID,
			AuthorID: author.ID,
			Role:     "aut",
		}); err != nil {
			log.Printf("[Scanner] Failed to link author %s to book %d: %v", authorName, createdBook.ID, err)
		}
	}

	// Link tags
	for _, tagName := range meta.Tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		tag, err := s.db.CreateTag(ctx, tagName)
		if err != nil {
			log.Printf("[Scanner] Failed to create/get tag %s: %v", tagName, err)
			continue
		}

		if err := s.db.AddBookTag(ctx, database.AddBookTagParams{
			BookID: createdBook.ID,
			TagID:  tag.ID,
		}); err != nil {
			log.Printf("[Scanner] Failed to link tag %s to book %d: %v", tagName, createdBook.ID, err)
		}
	}

	return &createdBook, nil
}
