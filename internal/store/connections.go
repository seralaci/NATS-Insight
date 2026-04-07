package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type Connection struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	AuthMethod string `json:"authMethod"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	Token      string `json:"token,omitempty"`
	NKey       string `json:"nkey,omitempty"`
	CredsFile  string `json:"credsFile,omitempty"`
	MonitorURL string `json:"monitorUrl"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

var ErrNotFound = errors.New("not found")

func (s *Store) ListConnections() ([]Connection, error) {
	rows, err := s.db.Query(`
		SELECT id, name, url, auth_method, username, password, token, nkey, creds_file, monitor_url, created_at, updated_at
		FROM connections ORDER BY name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query connections: %w", err)
	}
	defer rows.Close()

	var conns []Connection
	for rows.Next() {
		var c Connection
		if err := rows.Scan(&c.ID, &c.Name, &c.URL, &c.AuthMethod, &c.Username, &c.Password, &c.Token, &c.NKey, &c.CredsFile, &c.MonitorURL, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan connection: %w", err)
		}
		conns = append(conns, c)
	}
	if conns == nil {
		conns = []Connection{}
	}
	return conns, nil
}

func (s *Store) GetConnection(id string) (*Connection, error) {
	var c Connection
	err := s.db.QueryRow(`
		SELECT id, name, url, auth_method, username, password, token, nkey, creds_file, monitor_url, created_at, updated_at
		FROM connections WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.URL, &c.AuthMethod, &c.Username, &c.Password, &c.Token, &c.NKey, &c.CredsFile, &c.MonitorURL, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query connection: %w", err)
	}
	return &c, nil
}

func (s *Store) CreateConnection(c *Connection) error {
	c.ID = generateID()
	now := time.Now().UTC().Format(time.RFC3339)
	c.CreatedAt = now
	c.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO connections (id, name, url, auth_method, username, password, token, nkey, creds_file, monitor_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, c.ID, c.Name, c.URL, c.AuthMethod, c.Username, c.Password, c.Token, c.NKey, c.CredsFile, c.MonitorURL, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert connection: %w", err)
	}
	return nil
}

func (s *Store) UpdateConnection(c *Connection) error {
	c.UpdatedAt = time.Now().UTC().Format(time.RFC3339)

	result, err := s.db.Exec(`
		UPDATE connections SET name=?, url=?, auth_method=?, username=?, password=?, token=?, nkey=?, creds_file=?, monitor_url=?, updated_at=?
		WHERE id=?
	`, c.Name, c.URL, c.AuthMethod, c.Username, c.Password, c.Token, c.NKey, c.CredsFile, c.MonitorURL, c.UpdatedAt, c.ID)
	if err != nil {
		return fmt.Errorf("update connection: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteConnection(id string) error {
	result, err := s.db.Exec("DELETE FROM connections WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func generateID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return hex.EncodeToString(b)
}
