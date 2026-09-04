package config

import (
	"flag"
	"os"
	"strconv"
)

// Config holds runtime configuration options for Lyostar.
type Config struct {
	Port     int
	BooksDir string
	DataDir  string
}

// Load parses command line flags and environment variables.
// Priority: Command-line flags > Environment variables > Default values.
func Load() *Config {
	defaultPort := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			defaultPort = p
		}
	}

	defaultBooksDir := "./books"
	if envBooks := os.Getenv("BOOKS_DIR"); envBooks != "" {
		defaultBooksDir = envBooks
	}

	defaultDataDir := "./data"
	if envData := os.Getenv("DATA_DIR"); envData != "" {
		defaultDataDir = envData
	}

	port := flag.Int("port", defaultPort, "HTTP server listening port")
	books := flag.String("books", defaultBooksDir, "Path to directory containing ebook files (read-only)")
	data := flag.String("data", defaultDataDir, "Path to directory for application data and SQLite database")

	// Parse only if not already parsed (e.g., during tests)
	if !flag.Parsed() {
		flag.Parse()
	}

	return &Config{
		Port:     *port,
		BooksDir: *books,
		DataDir:  *data,
	}
}
