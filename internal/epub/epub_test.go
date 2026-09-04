package epub

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/HugoSmits86/nativewebp"
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

func createTestJPEGImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, nil)
	return buf.Bytes()
}

func createEPUBArchive(files map[string][]byte) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		f, _ := zw.Create(name)
		_, _ = f.Write(content)
	}
	_ = zw.Close()
	return buf.Bytes()
}

func TestParseEPUB2WithCalibreMetadata(t *testing.T) {
	coverBytes := createTestImage(300, 450, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	containerXML := `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	opfXML := `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>The Hound of the Baskervilles</dc:title>
    <dc:creator opf:role="aut">Arthur Conan Doyle</dc:creator>
    <dc:creator opf:role="edt">John Watson</dc:creator>
    <dc:description>A classic mystery novel.</dc:description>
    <dc:publisher>George Newnes</dc:publisher>
    <dc:language>en</dc:language>
    <dc:date>1902</dc:date>
    <meta name="calibre:series" content="Sherlock Holmes"/>
    <meta name="calibre:series_index" content="5"/>
    <meta name="cover" content="cover-id"/>
  </metadata>
  <manifest>
    <item id="cover-id" href="images/cover.png" media-type="image/png"/>
    <item id="chap1" href="chap1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
</package>`

	files := map[string][]byte{
		"META-INF/container.xml": []byte(containerXML),
		"content.opf":            []byte(opfXML),
		"images/cover.png":       coverBytes,
	}

	archive := createEPUBArchive(files)

	info, err := Parse(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("failed to parse EPUB 2: %v", err)
	}

	m := info.Metadata
	if m.Title != "The Hound of the Baskervilles" {
		t.Errorf("expected title 'The Hound of the Baskervilles', got %q", m.Title)
	}
	if len(m.Authors) != 2 || m.Authors[0] != "Arthur Conan Doyle" || m.Authors[1] != "John Watson" {
		t.Errorf("unexpected authors: %+v", m.Authors)
	}
	if m.Description != "A classic mystery novel." {
		t.Errorf("unexpected description: %q", m.Description)
	}
	if m.Publisher != "George Newnes" {
		t.Errorf("unexpected publisher: %q", m.Publisher)
	}
	if m.Language != "en" {
		t.Errorf("unexpected language: %q", m.Language)
	}
	if m.PubDate != "1902" {
		t.Errorf("unexpected pubdate: %q", m.PubDate)
	}
	if m.Series != "Sherlock Holmes" || m.SeriesIndex != 5 {
		t.Errorf("expected series 'Sherlock Holmes' index 5, got %q (%.1f)", m.Series, m.SeriesIndex)
	}
	if len(info.CoverData) == 0 {
		t.Errorf("expected cover data, got empty")
	}
	if info.CoverExt != ".png" {
		t.Errorf("expected cover extension .png, got %q", info.CoverExt)
	}
}

func TestParseEPUB3Standard(t *testing.T) {
	coverBytes := createTestJPEGImage(400, 600)

	containerXML := `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="EPUB/package.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	opfXML := `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Dune</dc:title>
    <dc:creator>Frank Herbert</dc:creator>
    <dc:description>Sci-fi masterpiece.</dc:description>
    <dc:language>en</dc:language>
    <meta property="belongs-to-collection" id="c01">Dune Chronicles</meta>
    <meta refines="#c01" property="group-position">1.5</meta>
  </metadata>
  <manifest>
    <item id="my-cover" href="cover.jpg" media-type="image/jpeg" properties="cover-image"/>
  </manifest>
</package>`

	files := map[string][]byte{
		"META-INF/container.xml": []byte(containerXML),
		"EPUB/package.opf":       []byte(opfXML),
		"EPUB/cover.jpg":         coverBytes,
	}

	archive := createEPUBArchive(files)

	info, err := Parse(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("failed to parse EPUB 3: %v", err)
	}

	m := info.Metadata
	if m.Title != "Dune" {
		t.Errorf("expected title 'Dune', got %q", m.Title)
	}
	if len(m.Authors) != 1 || m.Authors[0] != "Frank Herbert" {
		t.Errorf("unexpected authors: %+v", m.Authors)
	}
	if m.Series != "Dune Chronicles" || m.SeriesIndex != 1.5 {
		t.Errorf("expected series 'Dune Chronicles' index 1.5, got %q (%.1f)", m.Series, m.SeriesIndex)
	}
	if len(info.CoverData) == 0 {
		t.Errorf("expected cover data, got empty")
	}
	if info.CoverExt != ".jpg" {
		t.Errorf("expected cover extension .jpg, got %q", info.CoverExt)
	}
}

func TestSaveCoverThumbnail(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "lyostar-thumb-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a large image: 1000 x 1500
	largeImgBytes := createTestImage(1000, 1500, color.RGBA{R: 0, G: 128, B: 255, A: 255})

	destWebP := filepath.Join(tempDir, "covers", "test_cover.webp")
	if err := SaveCoverThumbnail(largeImgBytes, destWebP, 400); err != nil {
		t.Fatalf("failed to save cover thumbnail: %v", err)
	}

	// Verify file was created
	f, err := os.Open(destWebP)
	if err != nil {
		t.Fatalf("failed to open generated webp file: %v", err)
	}
	defer f.Close()

	// Decode WebP
	cfg, err := nativewebp.DecodeConfig(f)
	if err != nil {
		t.Fatalf("failed to decode webp config: %v", err)
	}

	if cfg.Width != 400 {
		t.Errorf("expected thumbnail width 400, got %d", cfg.Width)
	}
	if cfg.Height != 600 {
		t.Errorf("expected thumbnail height 600, got %d", cfg.Height)
	}
}
