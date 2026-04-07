package store

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "nats-insight.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	slog.Info("database initialized", "path", dbPath)
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS connections (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			url TEXT NOT NULL DEFAULT 'nats://localhost:4222',
			auth_method TEXT NOT NULL DEFAULT 'none',
			username TEXT DEFAULT '',
			password TEXT DEFAULT '',
			token TEXT DEFAULT '',
			nkey TEXT DEFAULT '',
			creds_file TEXT DEFAULT '',
			monitor_url TEXT DEFAULT 'http://localhost:8222',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}
