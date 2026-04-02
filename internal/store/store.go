package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct { db *sql.DB }

type Clip struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	URL          string   `json:"url"`
	Content      string   `json:"content"`
	Tags         string   `json:"tags"`
	CreatedAt    string   `json:"created_at"`
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "archive.db") + "?_journal_mode=WAL&_busy_timeout=5000"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS clips (
			id TEXT PRIMARY KEY,\n\t\t\ttitle TEXT DEFAULT '',\n\t\t\turl TEXT DEFAULT '',\n\t\t\tcontent TEXT DEFAULT '',\n\t\t\ttags TEXT DEFAULT '',
			created_at TEXT DEFAULT (datetime('now'))
		)`)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }

func genID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }

func (d *DB) Create(e *Clip) error {
	e.ID = genID()
	e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := d.db.Exec(`INSERT INTO clips (id, title, url, content, tags, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		e.ID, e.Title, e.URL, e.Content, e.Tags, e.CreatedAt)
	return err
}

func (d *DB) Get(id string) *Clip {
	row := d.db.QueryRow(`SELECT id, title, url, content, tags, created_at FROM clips WHERE id=?`, id)
	var e Clip
	if err := row.Scan(&e.ID, &e.Title, &e.URL, &e.Content, &e.Tags, &e.CreatedAt); err != nil {
		return nil
	}
	return &e
}

func (d *DB) List() []Clip {
	rows, err := d.db.Query(`SELECT id, title, url, content, tags, created_at FROM clips ORDER BY created_at DESC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var result []Clip
	for rows.Next() {
		var e Clip
		if err := rows.Scan(&e.ID, &e.Title, &e.URL, &e.Content, &e.Tags, &e.CreatedAt); err != nil {
			continue
		}
		result = append(result, e)
	}
	return result
}

func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM clips WHERE id=?`, id)
	return err
}

func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM clips`).Scan(&n)
	return n
}
