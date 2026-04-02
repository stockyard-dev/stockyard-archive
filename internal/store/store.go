package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type DB struct{ db *sql.DB }

type Collection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon,omitempty"`
	CreatedAt string `json:"created_at"`
	ClipCount int    `json:"clip_count"`
}

type Clip struct {
	ID           string   `json:"id"`
	URL          string   `json:"url,omitempty"`
	Title        string   `json:"title"`
	Content      string   `json:"content,omitempty"`
	Excerpt      string   `json:"excerpt,omitempty"`
	Author       string   `json:"author,omitempty"`
	SiteName     string   `json:"site_name,omitempty"`
	CollectionID string   `json:"collection_id,omitempty"`
	Status       string   `json:"status"` // inbox, read, archived
	Favorite     bool     `json:"favorite"`
	Tags         []string `json:"tags"`
	ReadTime     int      `json:"read_time,omitempty"` // estimated minutes
	CreatedAt    string   `json:"created_at"`
	ReadAt       string   `json:"read_at,omitempty"`
	NoteCount    int      `json:"note_count"`
}

type Annotation struct {
	ID        string `json:"id"`
	ClipID    string `json:"clip_id"`
	Highlight string `json:"highlight,omitempty"` // selected text
	Note      string `json:"note"`
	Color     string `json:"color,omitempty"`
	CreatedAt string `json:"created_at"`
}

type ClipFilter struct {
	CollectionID string
	Status       string
	Tag          string
	Favorite     string // "true" or ""
	Search       string
	SortBy       string // created, title, read_time
	Limit        int
	Offset       int
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
	for _, q := range []string{
		`CREATE TABLE IF NOT EXISTS collections (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			icon TEXT DEFAULT '📁',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS clips (
			id TEXT PRIMARY KEY,
			url TEXT DEFAULT '',
			title TEXT NOT NULL,
			content TEXT DEFAULT '',
			excerpt TEXT DEFAULT '',
			author TEXT DEFAULT '',
			site_name TEXT DEFAULT '',
			collection_id TEXT DEFAULT '',
			status TEXT DEFAULT 'inbox',
			favorite INTEGER DEFAULT 0,
			tags_json TEXT DEFAULT '[]',
			read_time INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now')),
			read_at TEXT DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS annotations (
			id TEXT PRIMARY KEY,
			clip_id TEXT NOT NULL REFERENCES clips(id) ON DELETE CASCADE,
			highlight TEXT DEFAULT '',
			note TEXT DEFAULT '',
			color TEXT DEFAULT '#d4a843',
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_clips_collection ON clips(collection_id)`,
		`CREATE INDEX IF NOT EXISTS idx_clips_status ON clips(status)`,
		`CREATE INDEX IF NOT EXISTS idx_clips_favorite ON clips(favorite)`,
		`CREATE INDEX IF NOT EXISTS idx_annotations_clip ON annotations(clip_id)`,
	} {
		if _, err := db.Exec(q); err != nil {
			return nil, fmt.Errorf("migrate: %w", err)
		}
	}
	return &DB{db: db}, nil
}

func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }

func estimateReadTime(content string) int {
	words := len(strings.Fields(content))
	minutes := words / 200 // ~200 wpm average
	if minutes < 1 && words > 0 {
		return 1
	}
	return minutes
}

// ── Collections ──

func (d *DB) CreateCollection(c *Collection) error {
	c.ID = genID()
	c.CreatedAt = now()
	if c.Icon == "" {
		c.Icon = "📁"
	}
	_, err := d.db.Exec(`INSERT INTO collections (id,name,icon,created_at) VALUES (?,?,?,?)`,
		c.ID, c.Name, c.Icon, c.CreatedAt)
	return err
}

func (d *DB) GetCollection(id string) *Collection {
	var c Collection
	if err := d.db.QueryRow(`SELECT id,name,icon,created_at FROM collections WHERE id=?`, id).Scan(
		&c.ID, &c.Name, &c.Icon, &c.CreatedAt); err != nil {
		return nil
	}
	d.db.QueryRow(`SELECT COUNT(*) FROM clips WHERE collection_id=?`, id).Scan(&c.ClipCount)
	return &c
}

func (d *DB) ListCollections() []Collection {
	rows, err := d.db.Query(`SELECT id,name,icon,created_at FROM collections ORDER BY name ASC`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.CreatedAt); err != nil {
			continue
		}
		d.db.QueryRow(`SELECT COUNT(*) FROM clips WHERE collection_id=?`, c.ID).Scan(&c.ClipCount)
		out = append(out, c)
	}
	return out
}

func (d *DB) UpdateCollection(id string, c *Collection) error {
	_, err := d.db.Exec(`UPDATE collections SET name=?,icon=? WHERE id=?`, c.Name, c.Icon, id)
	return err
}

func (d *DB) DeleteCollection(id string) error {
	d.db.Exec(`UPDATE clips SET collection_id='' WHERE collection_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM collections WHERE id=?`, id)
	return err
}

// ── Clips ──

func (d *DB) SaveClip(c *Clip) error {
	c.ID = genID()
	c.CreatedAt = now()
	if c.Status == "" {
		c.Status = "inbox"
	}
	if c.Tags == nil {
		c.Tags = []string{}
	}
	c.ReadTime = estimateReadTime(c.Content)
	// Generate excerpt from content if not provided
	if c.Excerpt == "" && c.Content != "" {
		words := strings.Fields(c.Content)
		if len(words) > 30 {
			c.Excerpt = strings.Join(words[:30], " ") + "..."
		} else {
			c.Excerpt = c.Content
		}
	}
	tj, _ := json.Marshal(c.Tags)
	fav := 0
	if c.Favorite {
		fav = 1
	}
	_, err := d.db.Exec(`INSERT INTO clips (id,url,title,content,excerpt,author,site_name,collection_id,status,favorite,tags_json,read_time,created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.ID, c.URL, c.Title, c.Content, c.Excerpt, c.Author, c.SiteName, c.CollectionID, c.Status, fav, string(tj), c.ReadTime, c.CreatedAt)
	return err
}

func (d *DB) hydrateClip(c *Clip) {
	d.db.QueryRow(`SELECT COUNT(*) FROM annotations WHERE clip_id=?`, c.ID).Scan(&c.NoteCount)
}

func (d *DB) scanClip(s interface{ Scan(...any) error }) *Clip {
	var c Clip
	var tj string
	var fav int
	if err := s.Scan(&c.ID, &c.URL, &c.Title, &c.Content, &c.Excerpt, &c.Author, &c.SiteName, &c.CollectionID, &c.Status, &fav, &tj, &c.ReadTime, &c.CreatedAt, &c.ReadAt); err != nil {
		return nil
	}
	json.Unmarshal([]byte(tj), &c.Tags)
	if c.Tags == nil {
		c.Tags = []string{}
	}
	c.Favorite = fav == 1
	d.hydrateClip(&c)
	return &c
}

const clipCols = `id,url,title,content,excerpt,author,site_name,collection_id,status,favorite,tags_json,read_time,created_at,read_at`

func (d *DB) GetClip(id string) *Clip {
	return d.scanClip(d.db.QueryRow(`SELECT `+clipCols+` FROM clips WHERE id=?`, id))
}

func (d *DB) ListClips(f ClipFilter) ([]Clip, int) {
	where := []string{"1=1"}
	args := []any{}
	if f.CollectionID != "" {
		where = append(where, "collection_id=?")
		args = append(args, f.CollectionID)
	}
	if f.Status != "" && f.Status != "all" {
		where = append(where, "status=?")
		args = append(args, f.Status)
	}
	if f.Tag != "" {
		where = append(where, `tags_json LIKE ?`)
		args = append(args, `%"`+f.Tag+`"%`)
	}
	if f.Favorite == "true" {
		where = append(where, "favorite=1")
	}
	if f.Search != "" {
		where = append(where, "(title LIKE ? OR content LIKE ? OR url LIKE ?)")
		s := "%" + f.Search + "%"
		args = append(args, s, s, s)
	}
	w := strings.Join(where, " AND ")
	var total int
	d.db.QueryRow("SELECT COUNT(*) FROM clips WHERE "+w, args...).Scan(&total)

	order := "created_at"
	switch f.SortBy {
	case "title":
		order = "title"
	case "read_time":
		order = "read_time"
	}
	if f.Limit <= 0 {
		f.Limit = 50
	}
	q := fmt.Sprintf("SELECT %s FROM clips WHERE %s ORDER BY favorite DESC, %s DESC LIMIT ? OFFSET ?", clipCols, w, order)
	args = append(args, f.Limit, f.Offset)
	rows, err := d.db.Query(q, args...)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()
	var out []Clip
	for rows.Next() {
		if c := d.scanClip(rows); c != nil {
			out = append(out, *c)
		}
	}
	return out, total
}

func (d *DB) UpdateClip(id string, c *Clip) error {
	tj, _ := json.Marshal(c.Tags)
	fav := 0
	if c.Favorite {
		fav = 1
	}
	c.ReadTime = estimateReadTime(c.Content)
	_, err := d.db.Exec(`UPDATE clips SET url=?,title=?,content=?,excerpt=?,author=?,site_name=?,collection_id=?,tags_json=?,favorite=?,read_time=? WHERE id=?`,
		c.URL, c.Title, c.Content, c.Excerpt, c.Author, c.SiteName, c.CollectionID, string(tj), fav, c.ReadTime, id)
	return err
}

func (d *DB) SetStatus(id, status string) error {
	t := now()
	readAt := ""
	if status == "read" {
		readAt = t
	}
	_, err := d.db.Exec(`UPDATE clips SET status=?,read_at=? WHERE id=?`, status, readAt, id)
	return err
}

func (d *DB) ToggleFavorite(id string) error {
	_, err := d.db.Exec(`UPDATE clips SET favorite=1-favorite WHERE id=?`, id)
	return err
}

func (d *DB) DeleteClip(id string) error {
	d.db.Exec(`DELETE FROM annotations WHERE clip_id=?`, id)
	_, err := d.db.Exec(`DELETE FROM clips WHERE id=?`, id)
	return err
}

// ── Annotations ──

func (d *DB) CreateAnnotation(a *Annotation) error {
	a.ID = genID()
	a.CreatedAt = now()
	if a.Color == "" {
		a.Color = "#d4a843"
	}
	_, err := d.db.Exec(`INSERT INTO annotations (id,clip_id,highlight,note,color,created_at) VALUES (?,?,?,?,?,?)`,
		a.ID, a.ClipID, a.Highlight, a.Note, a.Color, a.CreatedAt)
	return err
}

func (d *DB) ListAnnotations(clipID string) []Annotation {
	rows, err := d.db.Query(`SELECT id,clip_id,highlight,note,color,created_at FROM annotations WHERE clip_id=? ORDER BY created_at ASC`, clipID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []Annotation
	for rows.Next() {
		var a Annotation
		if err := rows.Scan(&a.ID, &a.ClipID, &a.Highlight, &a.Note, &a.Color, &a.CreatedAt); err != nil {
			continue
		}
		out = append(out, a)
	}
	return out
}

func (d *DB) DeleteAnnotation(id string) error {
	_, err := d.db.Exec(`DELETE FROM annotations WHERE id=?`, id)
	return err
}

// ── Tags ──

func (d *DB) AllTags() []string {
	rows, err := d.db.Query(`SELECT DISTINCT tags_json FROM clips WHERE tags_json != '[]'`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	seen := map[string]bool{}
	for rows.Next() {
		var j string
		rows.Scan(&j)
		var tags []string
		json.Unmarshal([]byte(j), &tags)
		for _, t := range tags {
			seen[t] = true
		}
	}
	var out []string
	for t := range seen {
		out = append(out, t)
	}
	return out
}

// ── Stats ──

type Stats struct {
	Total       int `json:"total"`
	Inbox       int `json:"inbox"`
	Read        int `json:"read"`
	Archived    int `json:"archived"`
	Favorites   int `json:"favorites"`
	Collections int `json:"collections"`
	Annotations int `json:"annotations"`
	Tags        int `json:"tags"`
	TotalReadMin int `json:"total_read_min"`
}

func (d *DB) Stats() Stats {
	var s Stats
	d.db.QueryRow(`SELECT COUNT(*) FROM clips`).Scan(&s.Total)
	d.db.QueryRow(`SELECT COUNT(*) FROM clips WHERE status='inbox'`).Scan(&s.Inbox)
	d.db.QueryRow(`SELECT COUNT(*) FROM clips WHERE status='read'`).Scan(&s.Read)
	d.db.QueryRow(`SELECT COUNT(*) FROM clips WHERE status='archived'`).Scan(&s.Archived)
	d.db.QueryRow(`SELECT COUNT(*) FROM clips WHERE favorite=1`).Scan(&s.Favorites)
	d.db.QueryRow(`SELECT COUNT(*) FROM collections`).Scan(&s.Collections)
	d.db.QueryRow(`SELECT COUNT(*) FROM annotations`).Scan(&s.Annotations)
	s.Tags = len(d.AllTags())
	d.db.QueryRow(`SELECT COALESCE(SUM(read_time),0) FROM clips`).Scan(&s.TotalReadMin)
	return s
}
