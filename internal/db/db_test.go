package db

import (
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	store, err := NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer store.Close()

	if store.DB == nil {
		t.Fatal("NewDB() returned nil DB")
	}
}

func TestMigrate(t *testing.T) {
	store, err := NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
}

func TestSaveNote(t *testing.T) {
	store, err := NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	id, err := store.SaveNote("alice", "D12345", "Working on <project:Alpha>", `["project:Alpha"]`)
	if err != nil {
		t.Fatalf("SaveNote() error = %v", err)
	}
	if id == 0 {
		t.Fatal("SaveNote() returned id 0")
	}

	// Save another note and verify ID increments
	id2, err := store.SaveNote("bob", "D67890", "Fixed <bug:42>", `["bug:42"]`)
	if err != nil {
		t.Fatalf("SaveNote() error = %v", err)
	}
	if id2 <= id {
		t.Errorf("Second SaveNote() returned id %d; want > %d", id2, id)
	}
}

func TestNoteCount(t *testing.T) {
	store, err := NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	count, err := store.NoteCount()
	if err != nil {
		t.Fatalf("NoteCount() error = %v", err)
	}
	if count != 0 {
		t.Errorf("NoteCount() = %d; want 0", count)
	}

	_, err = store.SaveNote("alice", "D1", "note 1", `[]`)
	if err != nil {
		t.Fatalf("SaveNote() error = %v", err)
	}

	count, err = store.NoteCount()
	if err != nil {
		t.Fatalf("NoteCount() error = %v", err)
	}
	if count != 1 {
		t.Errorf("NoteCount() = %d; want 1", count)
	}
}

func TestSearch(t *testing.T) {
	store, err := NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// Pre-seed notes
	_, _ = store.SaveNote("alice", "D1", "<project:Alpha> fixing auth", `["project:Alpha"]`)
	_, _ = store.SaveNote("bob", "D2", "Deployed to production", `[]`)
	_, _ = store.SaveNote("charlie", "D3", "<task:AUTH-123> related to auth", `["task:AUTH-123"]`)

	tests := []struct {
		name     string
		query    string
		wantMin  int
		wantMax  int
	}{
		{
			name:    "search by project name",
			query:   "Alpha",
			wantMin: 1,
			wantMax: 1,
		},
		{
			name:    "search by text keyword",
			query:   "production",
			wantMin: 1,
			wantMax: 1,
		},
		{
			name:    "search broad keyword",
			query:   "auth",
			wantMin: 1,
			wantMax: 2,
		},
		{
			name:    "search no match",
			query:   "nonexistent123",
			wantMin: 0,
			wantMax: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := store.Search(tt.query, 20)
			if err != nil {
				t.Fatalf("Search() error = %v", err)
			}
			if len(results) < tt.wantMin || len(results) > tt.wantMax {
				t.Errorf("Search(%q) = %d results; want between %d and %d", tt.query, len(results), tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestSearchResultStructure(t *testing.T) {
	store, err := NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	_, _ = store.SaveNote("alice", "D1", "<project:Alpha> test note", `["project:Alpha"]`)

	results, err := store.Search("Alpha", 20)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search() returned no results")
	}

	r := results[0]
	if r.User != "alice" {
		t.Errorf("result.User = %q; want %q", r.User, "alice")
	}
	if r.Relevance <= 0 {
		t.Errorf("result.Relevance = %d; want > 0", r.Relevance)
	}
	if len(r.MentionedTags) == 0 {
		t.Error("result.MentionedTags is empty; expected tags")
	}
}

func TestNotesByDateRange(t *testing.T) {
	store, err := NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	end := time.Now()
	start := end.Add(-10 * 24 * time.Hour) // 10 days ago

	_, _ = store.SaveNote("alice", "D1", "old note", `[]`)
	_, _ = store.SaveNote("bob", "D2", "recent note", `[]`)

	notes, err := store.NotesByDateRange(start, end, 50)
	if err != nil {
		t.Fatalf("NotesByDateRange() error = %v", err)
	}
	if len(notes) != 2 {
		t.Errorf("NotesByDateRange() = %d notes; want 2", len(notes))
	}
}
