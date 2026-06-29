package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Store is the SQLite database store.
type Store struct {
	*sql.DB
}

// Note represents a user note captured via /wr note or /wr daily.
type Note struct {
	ID        uint      `json:"id"`
	User      string    `json:"user"`
	Channel   string    `json:"channel"`
	Text      string    `json:"text"`
	Tags      string    `json:"tags"` // JSON array of tag strings
	CreatedAt time.Time `json:"created_at"`
}

// SearchResult represents a ranked search hit.
type SearchResult struct {
	ID            uint     `json:"id"`
	User          string   `json:"user"`
	Text          string   `json:"text"`
	Relevance     int      `json:"relevance"`
	CreatedAt     string   `json:"created_at"`
	MentionedTags []string `json:"mentioned_tags"`
}

func NewDB(dsn string) (*Store, error) {
	s, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return &Store{DB: s}, nil
}

func (s *Store) Migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user TEXT NOT NULL,
		channel TEXT,
		text TEXT NOT NULL,
		tags TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_notes_created ON notes(created_at);
	CREATE INDEX IF NOT EXISTS idx_notes_user ON notes(user);
	`
	_, err := s.Exec(schema)
	return err
}

// Search performs a full-text search across note text and tags.
func (s *Store) Search(query string, limit int) ([]SearchResult, error) {
	q := fmt.Sprintf("%%%s%%", query)
	rows, err := s.Query(
		`SELECT id, user, text, created_at, tags FROM notes 
		 WHERE text LIKE ? OR tags LIKE ? 
		 ORDER BY created_at DESC LIMIT ?`,
		q, q, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SearchResult, 0)
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.User, &n.Text, &n.CreatedAt, &n.Tags); err != nil {
			return nil, err
		}
		score := 0
		if n.Text == query {
			score = 100
		} else if n.Text != "" {
			score += 10
		}
		if n.Tags != "" {
			score += 5
		}
		var tags []string
		if err := json.Unmarshal([]byte(n.Tags), &tags); err != nil {
			tags = []string{n.Tags}
		}
		out = append(out, SearchResult{
			ID:            n.ID,
			User:          n.User,
			Text:          n.Text,
			Relevance:     score,
			CreatedAt:     n.CreatedAt.Format(time.RFC3339),
			MentionedTags: tags,
		})
	}
	return out, nil
}

// NotesByDateRange returns notes ordered by date.
func (s *Store) NotesByDateRange(start, end time.Time, limit int) ([]Note, error) {
	rows, err := s.Query(
		`SELECT id, user, channel, text, tags, created_at FROM notes 
		 WHERE created_at BETWEEN ? AND ? ORDER BY created_at ASC LIMIT ?`,
		start.Format(time.RFC3339), end.Format(time.RFC3339), limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.User, &n.Channel, &n.Text, &n.Tags, &n.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

// SaveNote creates a new note.
func (s *Store) SaveNote(user, channel, text, tags string) (uint, error) {
	res, err := s.Exec(
		`INSERT INTO notes (user, channel, text, tags, created_at) VALUES (?, ?, ?, ?, ?)`,
		user, channel, text, tags, time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint(id), nil
}

// NoteCount returns the total number of notes.
func (s *Store) NoteCount() (int64, error) {
	var count int64
	err := s.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	return count, err
}
