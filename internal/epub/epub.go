package epub

// Metadata represents the extracted EPUB Dublin Core metadata.
type Metadata struct {
	Title       string
	Description string
	Publisher   string
	Language    string
	PubDate     string
	Series      string
	SeriesIndex string
	Authors     []string
}
