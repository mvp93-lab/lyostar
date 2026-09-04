package scanner

import (
	"context"
)

// Scanner handles discovery and indexing of EPUB files in the books directory.
type Scanner struct {
	booksDir string
}

// New creates a new Scanner instance.
func New(booksDir string) *Scanner {
	return &Scanner{booksDir: booksDir}
}

// Start begins the initial file walk and background monitoring.
func (s *Scanner) Start(ctx context.Context) error {
	// Future background worker pool for scanning books
	return nil
}
