package api

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
	"time"

	"github.com/lyostar/lyostar/internal/auth"
	"github.com/lyostar/lyostar/internal/database"
	"github.com/lyostar/lyostar/internal/scanner"
)

func createUploadTestEPUB(title, author string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 150))
	for y := 0; y < 150; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 50, B: 50, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	_ = png.Encode(&imgBuf, img)

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
    <meta name="cover" content="cover-id"/>
  </metadata>
  <manifest>
    <item id="cover-id" href="cover.png" media-type="image/png"/>
    <item id="chap1" href="chap1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
</package>`, title, author)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	f, _ := zw.Create("META-INF/container.xml")
	_, _ = f.Write([]byte(containerXML))

	f, _ = zw.Create("content.opf")
	_, _ = f.Write([]byte(opfXML))

	f, _ = zw.Create("cover.png")
	_, _ = f.Write(imgBuf.Bytes())

	_ = zw.Close()
	return buf.Bytes()
}

func setupUploadRouter(t *testing.T) (*database.DB, *scanner.Scanner, http.Handler, string, string) {
	tempDir, err := os.MkdirTemp("", "lyostar-upload-test-*")
	if err != nil {
		t.Fatalf("failed creating temp dir: %v", err)
	}

	booksDir := filepath.Join(tempDir, "books")
	dataDir := filepath.Join(tempDir, "data")
	coversDir := filepath.Join(dataDir, "cache", "covers")
	uploadsDir := filepath.Join(dataDir, "uploads")

	_ = os.MkdirAll(booksDir, 0755)
	_ = os.MkdirAll(coversDir, 0755)
	_ = os.MkdirAll(uploadsDir, 0755)

	dbPath := filepath.Join(dataDir, "app.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	sc := scanner.New(booksDir, coversDir, db, uploadsDir)

	mockFS := fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte("<html><body>Lyostar SPA</body></html>"),
		},
	}

	router := NewRouter(RouterConfig{
		DB:         db,
		Scanner:    sc,
		UploadsDir: uploadsDir,
		StaticFS:   mockFS,
		Version:    "1.0.0-test",
	})

	return db, sc, router, tempDir, uploadsDir
}

func makeUploadRequest(t *testing.T, filename string, content []byte) (*http.Request, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatalf("failed to copy content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/books/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, writer.FormDataContentType()
}

func TestUploadBook(t *testing.T) {
	db, _, router, tempDir, uploadsDir := setupUploadRouter(t)
	defer os.RemoveAll(tempDir)
	defer db.Close()

	ctx := context.Background()

	// Create admin user (can_upload = 1)
	adminHash, _ := auth.HashPassword("AdminPass123")
	adminUser, err := db.CreateUser(ctx, database.CreateUserParams{
		Username:     "admin",
		PasswordHash: adminHash,
		Role:         auth.RoleAdmin,
		CanRead:      1,
		CanDownload:  1,
		CanUpload:    1,
		CanEdit:      1,
		CanDelete:    1,
	})
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}

	// Create reader user with can_upload = 0
	readerHash, _ := auth.HashPassword("ReaderPass123")
	readerUser, err := db.CreateUser(ctx, database.CreateUserParams{
		Username:     "reader",
		PasswordHash: readerHash,
		Role:         auth.RoleReader,
		CanRead:      1,
		CanDownload:  1,
		CanUpload:    0,
		CanEdit:      0,
		CanDelete:    0,
	})
	if err != nil {
		t.Fatalf("failed to create reader user: %v", err)
	}

	adminSession, _ := auth.GenerateToken()
	_, _ = db.CreateSession(ctx, database.CreateSessionParams{
		Token:     adminSession,
		UserID:    adminUser.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	readerSession, _ := auth.GenerateToken()
	_, _ = db.CreateSession(ctx, database.CreateSessionParams{
		Token:     readerSession,
		UserID:    readerUser.ID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})

	epubData := createUploadTestEPUB("Uploaded Masterpiece", "Famous Author")

	// 1. Unauthenticated upload -> 401
	{
		req, _ := makeUploadRequest(t, "test.epub", epubData)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for unauthenticated upload, got %d", rec.Code)
		}
	}

	// 2. Reader without can_upload -> 403 Forbidden
	{
		req, _ := makeUploadRequest(t, "test.epub", epubData)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: readerSession})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for reader without can_upload, got %d", rec.Code)
		}
	}

	// 3. Invalid file format (e.g. .txt) -> 400 Bad Request
	{
		req, _ := makeUploadRequest(t, "notes.txt", []byte("plain text content"))
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: adminSession})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid format, got %d: %s", rec.Code, rec.Body.String())
		}
	}

	// 4. Successful upload by Admin -> 201 Created
	var uploadedBook BookDetailResponse
	{
		req, _ := makeUploadRequest(t, "masterpiece.epub", epubData)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: adminSession})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201 for valid upload, got %d: %s", rec.Code, rec.Body.String())
		}

		if err := json.NewDecoder(rec.Body).Decode(&uploadedBook); err != nil {
			t.Fatalf("failed to decode upload response: %v", err)
		}

		if uploadedBook.Title != "Uploaded Masterpiece" {
			t.Errorf("expected title 'Uploaded Masterpiece', got '%s'", uploadedBook.Title)
		}
		if len(uploadedBook.Authors) == 0 || uploadedBook.Authors[0].Name != "Famous Author" {
			t.Errorf("expected author 'Famous Author', got %+v", uploadedBook.Authors)
		}
		if uploadedBook.Format != "epub" {
			t.Errorf("expected format 'epub', got '%s'", uploadedBook.Format)
		}
		if !uploadedBook.HasCover {
			t.Errorf("expected book to have cover")
		}

		// Verify file exists in uploadsDir
		uploadFiles, err := os.ReadDir(uploadsDir)
		if err != nil || len(uploadFiles) != 1 {
			t.Fatalf("expected 1 file in uploadsDir, got %d (err: %v)", len(uploadFiles), err)
		}
	}

	// 5. Uploading the exact same book again -> 409 Conflict
	{
		req, _ := makeUploadRequest(t, "duplicate.epub", epubData)
		req.AddCookie(&http.Cookie{Name: auth.CookieName, Value: adminSession})
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409 for duplicate upload, got %d: %s", rec.Code, rec.Body.String())
		}

		// Verify duplicate file was cleaned up from uploadsDir
		uploadFiles, err := os.ReadDir(uploadsDir)
		if err != nil || len(uploadFiles) != 1 {
			t.Fatalf("expected still only 1 file in uploadsDir after duplicate cleanup, got %d", len(uploadFiles))
		}
	}
}
