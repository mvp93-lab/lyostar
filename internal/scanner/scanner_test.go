package scanner

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/lyostar/lyostar/internal/database"
)

func createTestImage(width, height int, col color.Color) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, col)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func createTestEPUB(title, author, series string, seriesIndex float64) []byte {
	coverBytes := createTestImage(200, 300, color.RGBA{R: 50, G: 100, B: 200, A: 255})

	containerXML := `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	opfXML := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:description>Test Description</dc:description>
    <dc:language>en</dc:language>
    <meta name="calibre:series" content="%s"/>
    <meta name="calibre:series_index" content="%.1f"/>
    <meta name="cover" content="cover-id"/>
  </metadata>
  <manifest>
    <item id="cover-id" href="cover.png" media-type="image/png"/>
    <item id="chap1" href="chap1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
</package>`, title, author, series, seriesIndex)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, _ := zw.Create("META-INF/container.xml")
	_, _ = f.Write([]byte(containerXML))

	f, _ = zw.Create("content.opf")
	_, _ = f.Write([]byte(opfXML))

	f, _ = zw.Create("cover.png")
	_, _ = f.Write(coverBytes)

	_ = zw.Close()
	return buf.Bytes()
}

func TestScanner(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lyostar-scanner-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	booksDir := filepath.Join(tempDir, "books")
	dataDir := filepath.Join(tempDir, "data")
	coversDir := filepath.Join(dataDir, "cache", "covers")

	if err := os.MkdirAll(booksDir, 0755); err != nil {
		t.Fatalf("failed to create books dir: %v", err)
	}
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		t.Fatalf("failed to create covers dir: %v", err)
	}

	// Create test EPUB files in booksDir
	epub1 := createTestEPUB("Book One", "Author Alpha", "Alpha Series", 1)
	epub2 := createTestEPUB("Book Two", "Author Beta", "Beta Series", 2)
	epub3 := createTestEPUB("Book Three", "Author Gamma", "Gamma Series", 3)

	if err := os.WriteFile(filepath.Join(booksDir, "book1.epub"), epub1, 0644); err != nil {
		t.Fatalf("failed to write book1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(booksDir, "book2.epub"), epub2, 0644); err != nil {
		t.Fatalf("failed to write book2: %v", err)
	}

	// Nested book
	nestedDir := filepath.Join(booksDir, "nested", "folder")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "book3.epub"), epub3, 0644); err != nil {
		t.Fatalf("failed to write book3: %v", err)
	}

	// Non-epub and hidden files (should be ignored)
	_ = os.WriteFile(filepath.Join(booksDir, "readme.txt"), []byte("not an epub"), 0644)
	_ = os.WriteFile(filepath.Join(booksDir, ".hidden.epub"), epub1, 0644)

	// Snapshot directory state before scan to verify read-only property
	booksBefore, err := os.ReadDir(booksDir)
	if err != nil {
		t.Fatalf("failed to read books dir: %v", err)
	}

	// Initialize database
	dbPath := filepath.Join(dataDir, "app.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	s := New(booksDir, coversDir, db)

	// 1. First scan
	stats, err := s.Scan(ctx)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if stats.TotalFiles != 3 {
		t.Errorf("expected 3 total files, got %d", stats.TotalFiles)
	}
	if stats.Added != 3 {
		t.Errorf("expected 3 added books, got %d", stats.Added)
	}
	if stats.Skipped != 0 {
		t.Errorf("expected 0 skipped books, got %d", stats.Skipped)
	}
	if stats.Errors != 0 {
		t.Errorf("expected 0 errors, got %d", stats.Errors)
	}

	// Verify books in DB
	books, err := db.ListBooks(ctx, database.ListBooksParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("failed to query books: %v", err)
	}
	if len(books) != 3 {
		t.Fatalf("expected 3 books in DB, got %d", len(books))
	}

	for _, b := range books {
		if b.CoverPath == "" {
			t.Errorf("expected book %s to have a cover path", b.Title)
		} else {
			if _, err := os.Stat(b.CoverPath); err != nil {
				t.Errorf("cover file %s does not exist on disk: %v", b.CoverPath, err)
			}
		}

		authors, err := db.GetAuthorsForBook(ctx, b.ID)
		if err != nil || len(authors) == 0 {
			t.Errorf("expected authors for book %s, got: %v", b.Title, err)
		}
	}

	// 2. Second scan (idempotent: should skip all)
	stats2, err := s.Scan(ctx)
	if err != nil {
		t.Fatalf("second scan failed: %v", err)
	}
	if stats2.TotalFiles != 3 {
		t.Errorf("second scan: expected 3 total files, got %d", stats2.TotalFiles)
	}
	if stats2.Added != 0 {
		t.Errorf("second scan: expected 0 added, got %d", stats2.Added)
	}
	if stats2.Skipped != 3 {
		t.Errorf("second scan: expected 3 skipped, got %d", stats2.Skipped)
	}
	if stats2.Errors != 0 {
		t.Errorf("second scan: expected 0 errors, got %d", stats2.Errors)
	}

	// 3. Verify booksDir was STRICTLY READ-ONLY
	booksAfter, err := os.ReadDir(booksDir)
	if err != nil {
		t.Fatalf("failed to read books dir after scan: %v", err)
	}
	if len(booksBefore) != len(booksAfter) {
		t.Errorf("books directory modified! Before count: %d, after count: %d",
			len(booksBefore), len(booksAfter))
	}
}
