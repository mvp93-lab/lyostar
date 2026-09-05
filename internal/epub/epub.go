package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
)

// Metadata represents extracted Dublin Core and series metadata from an EPUB.
type Metadata struct {
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

// BookInfo wraps the extracted metadata and optional cover image data.
type BookInfo struct {
	Metadata  Metadata
	CoverData []byte
	CoverExt  string
}

// containerXML maps to META-INF/container.xml
type containerXML struct {
	Rootfiles []struct {
		FullPath  string `xml:"full-path,attr"`
		MediaType string `xml:"media-type,attr"`
	} `xml:"rootfiles>rootfile"`
}

// opfPackage represents the package document (.opf)
type opfPackage struct {
	Metadata struct {
		Titles       []string `xml:"title"`
		Creators     []string `xml:"creator"`
		Subjects     []string `xml:"subject"`
		Descriptions []string `xml:"description"`
		Publishers   []string `xml:"publisher"`
		Languages    []string `xml:"language"`
		Dates        []string `xml:"date"`
		Metas        []struct {
			Name     string `xml:"name,attr"`
			Content  string `xml:"content,attr"`
			Property string `xml:"property,attr"`
			ID       string `xml:"id,attr"`
			Refines  string `xml:"refines,attr"`
			Value    string `xml:",chardata"`
		} `xml:"meta"`
	} `xml:"metadata"`
	Manifest struct {
		Items []struct {
			ID         string `xml:"id,attr"`
			Href       string `xml:"href,attr"`
			MediaType  string `xml:"media-type,attr"`
			Properties string `xml:"properties,attr"`
		} `xml:"item"`
	} `xml:"manifest"`
}

// ParseFile opens an EPUB file and extracts its metadata and cover image directly
// in-memory without extracting the archive to disk.
func ParseFile(filePath string) (*BookInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open epub file: %w", err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat epub file: %w", err)
	}

	return Parse(file, fi.Size())
}

// Parse reads an EPUB archive from any io.ReaderAt and extracts metadata and cover.
func Parse(r io.ReaderAt, size int64) (*BookInfo, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, fmt.Errorf("invalid epub zip archive: %w", err)
	}

	// 1. Locate and parse META-INF/container.xml
	containerFile := findZipFile(zr, "META-INF/container.xml")
	if containerFile == nil {
		return nil, fmt.Errorf("META-INF/container.xml not found in epub")
	}

	containerData, err := readZipEntry(containerFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read container.xml: %w", err)
	}

	var container containerXML
	if err := xml.Unmarshal(containerData, &container); err != nil {
		return nil, fmt.Errorf("failed to parse container.xml: %w", err)
	}

	var opfPath string
	for _, rf := range container.Rootfiles {
		if rf.FullPath != "" {
			opfPath = rf.FullPath
			break
		}
	}
	if opfPath == "" {
		return nil, fmt.Errorf("no rootfile specified in container.xml")
	}

	// 2. Locate and parse the .opf package document
	opfPath = path.Clean(opfPath)
	opfFile := findZipFile(zr, opfPath)
	if opfFile == nil {
		return nil, fmt.Errorf("opf file not found at %s", opfPath)
	}

	opfData, err := readZipEntry(opfFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read opf file: %w", err)
	}

	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse opf package: %w", err)
	}

	// 3. Extract Dublin Core & Series Metadata
	meta := extractMetadata(&pkg)

	// 4. Extract Cover Image
	opfDir := path.Dir(opfPath)
	coverBytes, coverExt := extractCover(zr, &pkg, opfDir)

	return &BookInfo{
		Metadata:  meta,
		CoverData: coverBytes,
		CoverExt:  coverExt,
	}, nil
}

func extractMetadata(pkg *opfPackage) Metadata {
	var m Metadata

	// Title
	for _, t := range pkg.Metadata.Titles {
		trimmed := strings.TrimSpace(t)
		if trimmed != "" {
			m.Title = trimmed
			break
		}
	}

	// Authors
	for _, a := range pkg.Metadata.Creators {
		trimmed := strings.TrimSpace(a)
		if trimmed != "" {
			m.Authors = append(m.Authors, trimmed)
		}
	}

	// Tags / Subjects (<dc:subject>)
	seenTags := make(map[string]bool)
	for _, s := range pkg.Metadata.Subjects {
		parts := strings.FieldsFunc(s, func(r rune) bool {
			return r == ',' || r == ';'
		})
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" && !seenTags[strings.ToLower(trimmed)] {
				seenTags[strings.ToLower(trimmed)] = true
				m.Tags = append(m.Tags, trimmed)
			}
		}
	}

	// Description
	for _, d := range pkg.Metadata.Descriptions {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			m.Description = trimmed
			break
		}
	}

	// Publisher
	for _, p := range pkg.Metadata.Publishers {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			m.Publisher = trimmed
			break
		}
	}

	// Language
	for _, l := range pkg.Metadata.Languages {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			m.Language = trimmed
			break
		}
	}

	// PubDate
	for _, d := range pkg.Metadata.Dates {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			m.PubDate = trimmed
			break
		}
	}

	// Series & Series Index (Calibre & EPUB 3)
	var collectionID string
	for _, meta := range pkg.Metadata.Metas {
		name := strings.ToLower(strings.TrimSpace(meta.Name))
		prop := strings.ToLower(strings.TrimSpace(meta.Property))

		// Calibre format
		if name == "calibre:series" && meta.Content != "" {
			m.Series = strings.TrimSpace(meta.Content)
		} else if name == "calibre:series_index" && meta.Content != "" {
			if idx, err := strconv.ParseFloat(strings.TrimSpace(meta.Content), 64); err == nil {
				m.SeriesIndex = idx
			}
		}

		// EPUB 3 format
		if prop == "belongs-to-collection" {
			val := strings.TrimSpace(meta.Value)
			if val == "" {
				val = strings.TrimSpace(meta.Content)
			}
			if val != "" {
				m.Series = val
				collectionID = meta.ID
			}
		}
	}

	// If EPUB 3 series index is refined
	if collectionID != "" {
		for _, meta := range pkg.Metadata.Metas {
			if meta.Refines == "#"+collectionID && strings.ToLower(meta.Property) == "group-position" {
				val := strings.TrimSpace(meta.Value)
				if val == "" {
					val = strings.TrimSpace(meta.Content)
				}
				if idx, err := strconv.ParseFloat(val, 64); err == nil {
					m.SeriesIndex = idx
				}
			}
		}
	}

	return m
}

func extractCover(zr *zip.Reader, pkg *opfPackage, opfDir string) ([]byte, string) {
	// Find cover item in manifest
	var targetHref string

	// Strategy 1: EPUB 3 properties="cover-image"
	for _, item := range pkg.Manifest.Items {
		props := strings.Fields(item.Properties)
		for _, p := range props {
			if p == "cover-image" {
				targetHref = item.Href
				break
			}
		}
		if targetHref != "" {
			break
		}
	}

	// Strategy 2: EPUB 2 <meta name="cover" content="<item-id>"/>
	if targetHref == "" {
		var coverID string
		for _, meta := range pkg.Metadata.Metas {
			if strings.ToLower(meta.Name) == "cover" && meta.Content != "" {
				coverID = strings.TrimSpace(meta.Content)
				break
			}
		}
		if coverID != "" {
			for _, item := range pkg.Manifest.Items {
				if item.ID == coverID {
					targetHref = item.Href
					break
				}
			}
		}
	}

	// Strategy 3: Item ID is "cover" or "cover-image" with image media-type
	if targetHref == "" {
		for _, item := range pkg.Manifest.Items {
			id := strings.ToLower(item.ID)
			if (id == "cover" || id == "cover-image") && strings.HasPrefix(item.MediaType, "image/") {
				targetHref = item.Href
				break
			}
		}
	}

	// Strategy 4: Item href contains "cover" and is an image
	if targetHref == "" {
		for _, item := range pkg.Manifest.Items {
			lowerHref := strings.ToLower(item.Href)
			if strings.Contains(lowerHref, "cover") && strings.HasPrefix(item.MediaType, "image/") {
				targetHref = item.Href
				break
			}
		}
	}

	if targetHref == "" {
		return nil, ""
	}

	// Decode URL-encoded href (e.g. "images/My%20Cover.jpg")
	unescapedHref, err := url.PathUnescape(targetHref)
	if err == nil {
		targetHref = unescapedHref
	}

	fullPath := path.Clean(path.Join(opfDir, targetHref))
	coverFile := findZipFile(zr, fullPath)
	if coverFile == nil {
		// Fallback: try matching basename if full path was not found
		coverBase := path.Base(targetHref)
		coverFile = findZipFile(zr, coverBase)
	}

	if coverFile == nil {
		return nil, ""
	}

	data, err := readZipEntry(coverFile)
	if err != nil {
		return nil, ""
	}

	ext := strings.ToLower(path.Ext(coverFile.Name))
	return data, ext
}

func findZipFile(zr *zip.Reader, targetPath string) *zip.File {
	normalized := strings.TrimPrefix(path.Clean(targetPath), "/")
	for _, f := range zr.File {
		cleanName := strings.TrimPrefix(path.Clean(f.Name), "/")
		if strings.EqualFold(cleanName, normalized) {
			return f
		}
	}
	return nil
}

func readZipEntry(zf *zip.File) ([]byte, error) {
	rc, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
