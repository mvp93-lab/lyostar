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

// Scanner handles discovery and indexing of EPUB files in the books directory.
type Scanner struct {
	booksDir    string
	coversDir   string
	db          *database.DB
	workerCount int
	mu          sync.Mutex // ensures only one active scan run at a time
}

// New creates a new Scanner instance with bounded worker count.
func New(booksDir, coversDir string, db *database.DB) *Scanner {
	// Keep workers bounded (min 2, max 4) to stay strictly under RAM budget (< 30MB)
	workers := runtime.NumCPU()
	if workers < 2 {
		workers = 2
	}
	if workers > 4 {
		workers = 4
	}

	return &Scanner{
		booksDir:    booksDir,
		coversDir:   coversDir,
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

	// Walk the books directory in STRICTLY READ-ONLY mode
	walkErr := filepath.WalkDir(s.booksDir, func(path string, d fs.DirEntry, err error) error {
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

	close(taskChan)
	wg.Wait()

	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return stats, walkErr
	}

	return stats, nil
}

func (s *Scanner) processFile(ctx context.Context, filePath string, stats *Stats) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[Scanner] Failed to open file %s: %v", filePath, err)
		atomic.AddInt64(&stats.Errors, 1)
		return
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Printf("[Scanner] Failed to stat file %s: %v", filePath, err)
		atomic.AddInt64(&stats.Errors, 1)
		return
	}
	fileSize := fi.Size()

	// Compute SHA-256 checksum
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Printf("[Scanner] Failed to compute sha256 for %s: %v", filePath, err)
		atomic.AddInt64(&stats.Errors, 1)
		return
	}
	fileSha256 := hex.EncodeToString(hasher.Sum(nil))

	// Check if book already exists in SQLite
	existing, err := s.db.GetBookBySHA256(ctx, fileSha256)
	if err == nil && existing.ID != 0 {
		// Book already indexed
		atomic.AddInt64(&stats.Skipped, 1)
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[Scanner] Database lookup error for sha256 %s: %v", fileSha256, err)
		atomic.AddInt64(&stats.Errors, 1)
		return
	}

	// Reset file pointer to beginning
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Printf("[Scanner] Failed to rewind file %s: %v", filePath, err)
		atomic.AddInt64(&stats.Errors, 1)
		return
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var format string
	var meta struct {
		Title       string
		Authors     []string
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
			log.Printf("[Scanner] Failed to parse EPUB %s: %v", filePath, err)
			atomic.AddInt64(&stats.Errors, 1)
			return
		}
		meta = struct {
			Title       string
			Authors     []string
			Description string
			Publisher   string
			Language    string
			PubDate     string
			Series      string
			SeriesIndex float64
		}{
			Title:       bookInfo.Metadata.Title,
			Authors:     bookInfo.Metadata.Authors,
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
			log.Printf("[Scanner] Failed to parse PDF %s: %v", filePath, err)
			atomic.AddInt64(&stats.Errors, 1)
			return
		}
		meta = struct {
			Title       string
			Authors     []string
			Description string
			Publisher   string
			Language    string
			PubDate     string
			Series      string
			SeriesIndex float64
		}{
			Title:       bookInfo.Metadata.Title,
			Authors:     bookInfo.Metadata.Authors,
			Description: bookInfo.Metadata.Description,
			Publisher:   bookInfo.Metadata.Publisher,
			Language:    bookInfo.Metadata.Language,
			PubDate:     bookInfo.Metadata.PubDate,
			Series:      bookInfo.Metadata.Series,
			SeriesIndex: bookInfo.Metadata.SeriesIndex,
		}
		coverData = bookInfo.CoverData

	default:
		log.Printf("[Scanner] Unsupported format for file %s", filePath)
		atomic.AddInt64(&stats.Errors, 1)
		return
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
		log.Printf("[Scanner] Failed to insert book %s into database: %v", title, err)
		atomic.AddInt64(&stats.Errors, 1)
		return
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

	atomic.AddInt64(&stats.Added, 1)
}
