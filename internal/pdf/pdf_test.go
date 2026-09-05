package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMinimalPDF(t *testing.T) {
	// A valid minimal PDF with an Info object
	pdfContent := []byte(`%PDF-1.4
1 0 obj
<<
  /Type /Catalog
  /Pages 2 0 R
>>
endobj
2 0 obj
<<
  /Type /Pages
  /Kids [3 0 R]
  /Count 1
>>
endobj
3 0 obj
<<
  /Type /Page
  /Parent 2 0 R
  /MediaBox [0 0 612 792]
>>
endobj
4 0 obj
<<
  /Title (Sample PDF Book)
  /Author (Jane Doe)
  /Subject (A test book for unit tests)
  /Keywords (Go, Programming, Backend)
  /CreationDate (D:20230515120000Z)
>>
endobj
trailer
<<
  /Root 1 0 R
  /Info 4 0 R
>>
%%EOF`)

	r := bytes.NewReader(pdfContent)
	info, err := Parse(r, int64(len(pdfContent)))
	if err != nil {
		t.Fatalf("Parse returned unexpected error: %v", err)
	}

	if info.Metadata.Title != "Sample PDF Book" {
		t.Errorf("expected Title 'Sample PDF Book', got '%s'", info.Metadata.Title)
	}
	if len(info.Metadata.Authors) == 0 || info.Metadata.Authors[0] != "Jane Doe" {
		t.Errorf("expected Author 'Jane Doe', got '%v'", info.Metadata.Authors)
	}
	if len(info.Metadata.Tags) != 3 || info.Metadata.Tags[0] != "Go" || info.Metadata.Tags[1] != "Programming" || info.Metadata.Tags[2] != "Backend" {
		t.Errorf("expected 3 tags [Go, Programming, Backend], got '%v'", info.Metadata.Tags)
	}
	if info.Metadata.Description != "A test book for unit tests" {
		t.Errorf("expected Description 'A test book for unit tests', got '%s'", info.Metadata.Description)
	}
	if info.Metadata.PubDate != "2023-05-15" {
		t.Errorf("expected PubDate '2023-05-15', got '%s'", info.Metadata.PubDate)
	}
}

func TestParseProGoPDFIfPresent(t *testing.T) {
	pdfPath := filepath.Join("..", "..", "books", "Pro_Go.pdf")
	f, err := os.Open(pdfPath)
	if err != nil {
		t.Skip("books/Pro_Go.pdf not found, skipping integration check")
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	info, err := Parse(f, fi.Size())
	if err != nil {
		t.Fatalf("failed to parse Pro_Go.pdf: %v", err)
	}

	t.Logf("Parsed Pro_Go.pdf: Title=%q, Authors=%v, PubDate=%q, CoverSize=%d bytes",
		info.Metadata.Title, info.Metadata.Authors, info.Metadata.PubDate, len(info.CoverData))

	if len(info.CoverData) == 0 {
		t.Logf("Notice: Pro_Go.pdf cover image was not found or not raster JPEG")
	} else {
		t.Logf("Success: extracted cover image (%d bytes)", len(info.CoverData))
	}
}
