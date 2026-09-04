package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// DB wraps the SQLite sql.DB connection and sqlc generated Queries.
type DB struct {
	*sql.DB
	*Queries
}

// Open initializes and configures a SQLite connection with required PRAGMAs,
// connection limits, and automatic schema migration.
func Open(dbPath string) (*DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
	}

	// Connect using pure-Go modernc.org/sqlite driver with DSN PRAGMAs
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)", dbPath)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Enforce single connection limit to prevent SQLite write-lock concurrency issues
	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	// Explicitly apply PRAGMAs to guarantee strict adherence to AGENTS.md directives
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA synchronous=NORMAL;",
	}
	for _, pragma := range pragmas {
		if _, err := conn.Exec(pragma); err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed executing pragma %q: %w", pragma, err)
		}
	}

	// Run initial schema migration automatically if tables do not exist
	if _, err := conn.Exec(schemaSQL); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to apply initial schema migration: %w", err)
	}

	queries := New(conn)

	return &DB{
		DB:      conn,
		Queries: queries,
	}, nil
}
