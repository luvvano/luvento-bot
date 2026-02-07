package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

type Group struct {
	ID        int64
	ChatID    int64
	Title     string
	AddedBy   int64
	CreatedAt string
}

func New(dbPath string) (*Storage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *Storage) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS groups (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER UNIQUE NOT NULL,
		title TEXT,
		added_by INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_groups_chat_id ON groups(chat_id);
	`

	_, err := s.db.Exec(schema)
	return err
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) AddGroup(chatID int64, title string, addedBy int64) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO groups (chat_id, title, added_by) VALUES (?, ?, ?)",
		chatID, title, addedBy,
	)
	return err
}

func (s *Storage) RemoveGroup(chatID int64) error {
	_, err := s.db.Exec("DELETE FROM groups WHERE chat_id = ?", chatID)
	return err
}

func (s *Storage) GetAllGroups() ([]Group, error) {
	rows, err := s.db.Query("SELECT id, chat_id, title, added_by, created_at FROM groups")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.ChatID, &g.Title, &g.AddedBy, &g.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}

	return groups, rows.Err()
}

func (s *Storage) IsGroupRegistered(chatID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM groups WHERE chat_id = ?", chatID).Scan(&count)
	return count > 0, err
}
