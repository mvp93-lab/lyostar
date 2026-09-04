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

	// Apply migrations for existing databases
	if err := migrateSchema(conn); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	queries := New(conn)

	return &DB{
		DB:      conn,
		Queries: queries,
	}, nil
}

func migrateSchema(conn *sql.DB) error {
	// Check columns in users table
	rows, err := conn.Query("PRAGMA table_info(users)")
	if err != nil {
		return err
	}
	defer rows.Close()

	existingCols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
			existingCols[name] = true
		}
	}

	colsToAdd := []struct {
		name string
		def  string
	}{
		{"can_read", "INTEGER NOT NULL DEFAULT 1"},
		{"can_download", "INTEGER NOT NULL DEFAULT 1"},
		{"can_upload", "INTEGER NOT NULL DEFAULT 0"},
		{"can_edit", "INTEGER NOT NULL DEFAULT 0"},
		{"can_delete", "INTEGER NOT NULL DEFAULT 0"},
	}

	for _, col := range colsToAdd {
		if !existingCols[col.name] {
			alterSQL := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s;", col.name, col.def)
			if _, err := conn.Exec(alterSQL); err != nil {
				return fmt.Errorf("failed adding column %s: %w", col.name, err)
			}
		}
	}
	return nil
}
