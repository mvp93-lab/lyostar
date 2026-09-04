package pdf

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Metadata represents extracted metadata from a PDF file.
type Metadata struct {
	Title       string
	Authors     []string
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

// XMP metadata structures for Dublin Core
type xmpMeta struct {
	XMLName xml.Name `xml:"xmpmeta"`
	RDF     struct {
		Description struct {
			Title struct {
				Alt struct {
					Li []string `xml:"li"`
				} `xml:"Alt"`
			} `xml:"title"`
			Creator struct {
				Seq struct {
					Li []string `xml:"li"`
				} `xml:"Seq"`
			} `xml:"creator"`
			Description struct {
				Alt struct {
					Li []string `xml:"li"`
				} `xml:"Alt"`
			} `xml:"description"`
			Date       []string `xml:"date"`
			CreateDate string   `xml:"CreateDate"`
		} `xml:"Description"`
	} `xml:"RDF"`
}

var (
	rootRefRegex     = regexp.MustCompile(`/Root\s+(\d+)\s+(\d+)\s+R`)
	metadataRefRegex = regexp.MustCompile(`/Metadata\s+(\d+)\s+(\d+)\s+R`)
	infoRefRegex     = regexp.MustCompile(`/Info\s+(\d+)\s+(\d+)\s+R`)
	titleRegex       = regexp.MustCompile(`/Title\s*(\((?:\\.|[^)])*\)|<[0-9a-fA-F]+>)`)
	authorRegex      = regexp.MustCompile(`/Author\s*(\((?:\\.|[^)])*\)|<[0-9a-fA-F]+>)`)
	subjectRegex     = regexp.MustCompile(`/Subject\s*(\((?:\\.|[^)])*\)|<[0-9a-fA-F]+>)`)
	dateRegex        = regexp.MustCompile(`/CreationDate\s*(\((?:\\.|[^)])*\)|<[0-9a-fA-F]+>)`)
	xmpRegex         = regexp.MustCompile(`(?s)<x:xmpmeta[\s\S]*?</x:xmpmeta>`)
)

// Parse extracts metadata and cover thumbnail data from a PDF file.
func Parse(r io.ReaderAt, size int64) (*BookInfo, error) {
	if size < 8 {
		return nil, fmt.Errorf("file too small to be a PDF")
	}

	header := make([]byte, 8)
	if _, err := r.ReadAt(header, 0); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read PDF header: %w", err)
	}
	if !bytes.HasPrefix(header, []byte("%PDF-")) {
		return nil, fmt.Errorf("invalid PDF signature")
	}

	// Read trailer and tail (up to 128KB) to locate Info dictionary
	tailSize := int64(131072)
	if tailSize > size {
		tailSize = size
	}
	tailBuf := make([]byte, tailSize)
	if _, err := r.ReadAt(tailBuf, size-tailSize); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read PDF tail: %w", err)
	}

	var meta Metadata

	// 1. Check for Catalog /Root to find document-level /Metadata
	var docMetadataObj string
	if rootMatches := rootRefRegex.FindAllSubmatch(tailBuf, -1); len(rootMatches) > 0 {
		lastRoot := rootMatches[len(rootMatches)-1]
		catData, err := readObject(r, size, string(lastRoot[1]), string(lastRoot[2]))
		if err == nil {
			if metaRefMatch := metadataRefRegex.FindSubmatch(catData); len(metaRefMatch) > 1 {
				docMetadataObj = string(metaRefMatch[1])
				docMetadataGen := string(metaRefMatch[2])
				if xmpData, err := readObject(r, size, docMetadataObj, docMetadataGen); err == nil {
					extractXMPMetadata(xmpData, &meta)
				}
			}
		}
	}

	// 2. Check for Info object reference in trailer
	if infoMatches := infoRefRegex.FindAllSubmatch(tailBuf, -1); len(infoMatches) > 0 {
		lastMatch := infoMatches[len(infoMatches)-1]
		objNum := string(lastMatch[1])
		genNum := string(lastMatch[2])

		infoData, err := readObject(r, size, objNum, genNum)
		if err == nil && len(infoData) > 0 {
			extractInfoMetadata(infoData, &meta)
		}
	}

	// 3. Read first chunk (up to 4MB) to check cover images and fallback XMP if needed
	firstChunkSize := int64(4194304)
	if firstChunkSize > size {
		firstChunkSize = size
	}
	firstChunk := make([]byte, firstChunkSize)
	_, _ = r.ReadAt(firstChunk, 0)

	// 3. Extract cover image: find first embedded JPEG in first chunk
	var coverData []byte
	var coverExt string
	if img := findFirstJPEG(firstChunk); len(img) > 0 {
		coverData = img
		coverExt = "jpg"
	}

	return &BookInfo{
		Metadata:  meta,
		CoverData: coverData,
		CoverExt:  coverExt,
	}, nil
}

// readObject attempts to find and read a PDF indirect object "N M obj ... endobj"
func readObject(r io.ReaderAt, size int64, objNum, genNum string) ([]byte, error) {
	needle := []byte(fmt.Sprintf("%s %s obj", objNum, genNum))
	needleLen := int64(len(needle))

	// Search in tail first (last 2MB), then from beginning (first 4MB)
	searchRegions := []struct {
		offset int64
		length int64
	}{
		{offset: max(0, size-2097152), length: min(2097152, size)},
		{offset: 0, length: min(4194304, size)},
	}

	for _, reg := range searchRegions {
		buf := make([]byte, reg.length)
		n, err := r.ReadAt(buf, reg.offset)
		if (err != nil && err != io.EOF) || n < int(needleLen) {
			continue
		}
		buf = buf[:n]

		idx := bytes.Index(buf, needle)
		if idx != -1 {
			start := idx + len(needle)
			end := bytes.Index(buf[start:], []byte("endobj"))
			if end != -1 {
				return buf[start : start+end], nil
			}
			return buf[start:], nil
		}
	}

	return nil, fmt.Errorf("object %s %s not found", objNum, genNum)
}

func extractInfoMetadata(data []byte, meta *Metadata) {
	if meta.Title == "" {
		if m := titleRegex.FindSubmatch(data); len(m) > 1 {
			meta.Title = decodePDFString(m[1])
		}
	}
	if len(meta.Authors) == 0 {
		if m := authorRegex.FindSubmatch(data); len(m) > 1 {
			author := decodePDFString(m[1])
			if author != "" {
				meta.Authors = []string{author}
			}
		}
	}
	if meta.Description == "" {
		if m := subjectRegex.FindSubmatch(data); len(m) > 1 {
			meta.Description = decodePDFString(m[1])
		}
	}
	if meta.PubDate == "" {
		if m := dateRegex.FindSubmatch(data); len(m) > 1 {
			meta.PubDate = parsePDFDate(decodePDFString(m[1]))
		}
	}
}

func extractXMPMetadata(data []byte, meta *Metadata) {
	xmpBlocks := xmpRegex.FindAll(data, -1)
	for i := len(xmpBlocks) - 1; i >= 0; i-- {
		var xmp xmpMeta
		if err := xml.Unmarshal(xmpBlocks[i], &xmp); err == nil {
			desc := xmp.RDF.Description
			if meta.Title == "" && len(desc.Title.Alt.Li) > 0 {
				meta.Title = strings.TrimSpace(desc.Title.Alt.Li[0])
			}
			if len(meta.Authors) == 0 && len(desc.Creator.Seq.Li) > 0 {
				for _, author := range desc.Creator.Seq.Li {
					trimmed := strings.TrimSpace(author)
					if trimmed != "" {
						meta.Authors = append(meta.Authors, trimmed)
					}
				}
			}
			if meta.Description == "" && len(desc.Description.Alt.Li) > 0 {
				meta.Description = strings.TrimSpace(desc.Description.Alt.Li[0])
			}
			if meta.PubDate == "" {
				if len(desc.Date) > 0 {
					meta.PubDate = strings.TrimSpace(desc.Date[0])
				} else if desc.CreateDate != "" {
					meta.PubDate = strings.TrimSpace(desc.CreateDate)
				}
			}
		}
	}
}

// decodePDFString parses literal strings (...) or hex strings <...>
func decodePDFString(raw []byte) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) < 2 {
		return ""
	}

	var content []byte
	if raw[0] == '(' && raw[len(raw)-1] == ')' {
		content = unescapePDFLiteral(raw[1 : len(raw)-1])
	} else if raw[0] == '<' && raw[len(raw)-1] == '>' {
		hexStr := strings.ReplaceAll(string(raw[1:len(raw)-1]), " ", "")
		if len(hexStr)%2 != 0 {
			hexStr += "0"
		}
		var err error
		content, err = hex.DecodeString(hexStr)
		if err != nil {
			return ""
		}
	} else {
		return ""
	}

	// Check for UTF-16BE Byte Order Mark (\xfe\xff)
	if len(content) >= 2 && content[0] == 0xFE && content[1] == 0xFF {
		return decodeUTF16BE(content[2:])
	}

	// Check for UTF-8 or ASCII / PDFDocEncoding
	return strings.ToValidUTF8(string(content), "")
}

func unescapePDFLiteral(data []byte) []byte {
	var out []byte
	n := len(data)
	for i := 0; i < n; i++ {
		if data[i] == '\\' && i+1 < n {
			i++
			switch data[i] {
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case 'b':
				out = append(out, '\b')
			case 'f':
				out = append(out, '\f')
			case '(', ')', '\\':
				out = append(out, data[i])
			default:
				// Octal digit
				if data[i] >= '0' && data[i] <= '7' {
					oct := []byte{data[i]}
					for j := 0; j < 2 && i+1 < n && data[i+1] >= '0' && data[i+1] <= '7'; j++ {
						i++
						oct = append(oct, data[i])
					}
					val, _ := strconv.ParseInt(string(oct), 8, 32)
					out = append(out, byte(val))
				} else {
					out = append(out, data[i])
				}
			}
		} else {
			out = append(out, data[i])
		}
	}
	return out
}

func decodeUTF16BE(b []byte) string {
	if len(b)%2 != 0 {
		b = b[:len(b)-1]
	}
	u16s := make([]uint16, len(b)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = uint16(b[i*2])<<8 | uint16(b[i*2+1])
	}
	return strings.TrimSpace(string(utf16.Decode(u16s)))
}

// parsePDFDate parses PDF date format "D:YYYYMMDDHHmmSSOHH'mm'"
func parsePDFDate(raw string) string {
	s := strings.TrimPrefix(raw, "D:")
	if len(s) >= 4 {
		year := s[:4]
		if len(s) >= 6 {
			month := s[4:6]
			if len(s) >= 8 {
				day := s[6:8]
				return fmt.Sprintf("%s-%s-%s", year, month, day)
			}
			return fmt.Sprintf("%s-%s", year, month)
		}
		return year
	}
	return raw
}

// findFirstJPEG searches for a valid embedded JPEG in data
func findFirstJPEG(data []byte) []byte {
	soi := []byte{0xFF, 0xD8, 0xFF}
	eoi := []byte{0xFF, 0xD9}

	offset := 0
	for {
		idx := bytes.Index(data[offset:], soi)
		if idx == -1 {
			break
		}
		start := offset + idx
		endIdx := bytes.Index(data[start:], eoi)
		if endIdx == -1 {
			offset = start + 3
			continue
		}

		end := start + endIdx + 2
		candidate := data[start:end]

		// Validate image header and dimensions
		cfg, _, err := image.DecodeConfig(bytes.NewReader(candidate))
		if err == nil && cfg.Width >= 100 && cfg.Height >= 100 {
			return candidate
		}

		offset = start + 3
	}

	return nil
}
