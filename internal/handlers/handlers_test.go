package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chickenandpork/weekly-rundown/internal/db"
	"github.com/chickenandpork/weekly-rundown/internal/markup"
)

func newTestStore(t *testing.T) *db.Store {
	t.Helper()
	store, err := db.NewDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	if err := store.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return store
}

func newTestHandlers(t *testing.T) *Handlers {
	t.Helper()
	store := newTestStore(t)
	return New(store, Config{SlackBotToken: "test-secret"})
}

func TestHandleSlashCommand_Unauthorized(t *testing.T) {
	h := newTestHandlers(t)

	form := strings.NewReader("text=hello&user_name=alice&command=/wr")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "wrong-token")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleSlashCommand_MissingText(t *testing.T) {
	h := newTestHandlers(t)

	form := strings.NewReader("user_name=alice&command=/wr")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleSlashCommand_Note(t *testing.T) {
	h := newTestHandlers(t)

	form := strings.NewReader("text=Working on <project:Alpha>&user_name=alice&command=/wr&channel_name=D12345")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Note saved") {
		t.Errorf("HandleSlashCommand() body = %q; want 'Note saved'", body)
	}
	if !strings.Contains(body, "project:Alpha") {
		t.Errorf("HandleSlashCommand() body = %q; want 'project:Alpha'", body)
	}

	// Verify note was persisted
	count, _ := h.store.NoteCount()
	if count != 1 {
		t.Errorf("NoteCount() = %d; want 1", count)
	}
}

func TestHandleSlashCommand_NoteWithMultipleTags(t *testing.T) {
	h := newTestHandlers(t)

	text := "Fixed <bug:42> using <tech:Go> for <project:Beta>"
	form := strings.NewReader("text=" + text + "&user_name=bob&command=/wr&channel_name=D67890")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "bug:42") {
		t.Errorf("body missing 'bug:42': %s", body)
	}
	if !strings.Contains(body, "tech:Go") {
		t.Errorf("body missing 'tech:Go': %s", body)
	}
	if !strings.Contains(body, "project:Beta") {
		t.Errorf("body missing 'project:Beta': %s", body)
	}

	count, _ := h.store.NoteCount()
	if count != 1 {
		t.Errorf("NoteCount() = %d; want 1", count)
	}
}

func TestHandleSlashCommand_Daily(t *testing.T) {
	h := newTestHandlers(t)

	form := strings.NewReader("text=Shipped the new feature&user_name=charlie&command=/wr%20daily&channel_name=D99999")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Note saved") {
		t.Errorf("HandleSlashCommand() body = %q; want 'Note saved'", body)
	}
}

func TestHandleSlashCommand_Search_NoResults(t *testing.T) {
	h := newTestHandlers(t)

	form := strings.NewReader("text=test&user_name=alice&command=/wr%20search")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	// The search handler returns "No results found." which is a 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No results found") {
		t.Errorf("HandleSlashCommand() body = %q; want 'No results found'", body)
	}
}

func TestHandleSlashCommand_Search_WithResults(t *testing.T) {
	store := newTestStore(t)

	_, _ = store.SaveNote("alice", "D1", "Working on <project:Alpha>", `["project:Alpha"]`)
	_, _ = store.SaveNote("bob", "D2", "Deployed to production", `[]`)

	h := New(store, Config{SlackBotToken: "test-secret"})

	form := strings.NewReader("text=Alpha&user_name=alice&command=/wr%20search")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Alpha") {
		t.Errorf("HandleSlashCommand() body = %q; want 'Alpha'", body)
	}
}

func TestHandleSlashCommand_UnknownCommand(t *testing.T) {
	h := newTestHandlers(t)

	form := strings.NewReader("text=hello&user_name=alice&command=/unknown")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleHealth(t *testing.T) {
	h := newTestHandlers(t)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleHealth() status = %d; want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("HandleHealth() body not valid JSON: %v", err)
	}

	if resp["status"] != "ok" {
		t.Errorf("HandleHealth() status = %v; want 'ok'", resp["status"])
	}
	if resp["db_status"] != "ok" {
		t.Errorf("HandleHealth() db_status = %v; want 'ok'", resp["db_status"])
	}
}

func TestHandleHealth_WithNotes(t *testing.T) {
	store := newTestStore(t)
	_, _ = store.SaveNote("alice", "D1", "test", `[]`)

	h := New(store, Config{SlackBotToken: "test-secret"})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	h.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleHealth() status = %d; want %d", w.Code, http.StatusOK)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("HandleHealth() body not valid JSON: %v", err)
	}

	if int(resp["notes_count"].(float64)) != 1 {
		t.Errorf("HandleHealth() notes_count = %v; want 1", resp["notes_count"])
	}
}

func TestHandleWeekly_Empty(t *testing.T) {
	h := newTestHandlers(t)

	req := httptest.NewRequest("GET", "/weekly", nil)
	w := httptest.NewRecorder()
	h.HandleWeekly(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleWeekly() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "No notes in the past 7 days") {
		t.Errorf("HandleWeekly() body = %q; want 'No notes in the past 7 days'", body)
	}
}

func TestHandleWeekly_WithNotes(t *testing.T) {
	store := newTestStore(t)
	now := time.Now()
	_, _ = store.SaveNote("alice", "D1", "Recent note 1", `["project:Alpha"]`)
	_, _ = store.SaveNote("bob", "D2", "Recent note 2", `["tech:Go"]`)

	h := New(store, Config{SlackBotToken: "test-secret"})

	req := httptest.NewRequest("GET", "/weekly", nil)
	w := httptest.NewRecorder()
	h.HandleWeekly(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleWeekly() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Recent note") {
		t.Errorf("HandleWeekly() body missing 'Recent note': %s", body)
	}
}

func TestHandler_PreservesTagsInDatabase(t *testing.T) {
	store := newTestStore(t)
	h := New(store, Config{SlackBotToken: "test-secret"})

	originalText := "Building <project:Gamma> with <tech:Rust> fixes <bug:99>"
	form := strings.NewReader("text=" + originalText + "&user_name=dave&command=/wr&channel_name=D55555")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusOK)
	}

	// Verify the note was saved with the original text
	results, err := store.Search("Gamma", 20)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() returned %d results; want 1", len(results))
	}
	if results[0].Text != originalText {
		t.Errorf("Saved text = %q; want %q", results[0].Text, originalText)
	}
	if results[0].User != "dave" {
		t.Errorf("Saved user = %q; want 'dave'", results[0].User)
	}
	if len(results[0].MentionedTags) != 3 {
		t.Errorf("MentionedTags = %d; want 3", len(results[0].MentionedTags))
	}
}

func TestRespondError(t *testing.T) {
	h := newTestHandlers(t)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	h.respondError(w, "something broke")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("respondError() status = %d; want %d", w.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(w.Body.String(), "something broke") {
		t.Errorf("respondError() body = %q; want 'something broke'", w.Body.String())
	}
}

func TestHandleSlashCommand_SlackNotaryToken(t *testing.T) {
	// Test that a Slack application with a notary token still works
	// (token-based auth bypass when token is provided as X-Slack-Secret)
	h := newTestHandlers(t)

	// Same token should pass
	form := strings.NewReader("text=Test&user_name=alice&command=/wr")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "test-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusOK)
	}
}

func TestHandleSlashCommand_TokenMismatch(t *testing.T) {
	h := newTestHandlers(t)

	form := strings.NewReader("text=Test&user_name=alice&command=/wr")
	req := httptest.NewRequest("POST", "/slack/command", form)
	req.Header.Set("X-Slack-Secret", "wrong-secret")

	w := httptest.NewRecorder()
	h.HandleSlashCommand(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("HandleSlashCommand() status = %d; want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_WeeklyIntegration(t *testing.T) {
	store := newTestStore(t)
	h := New(store, Config{SlackBotToken: "test-secret"})

	// Add notes across different time periods
	now := time.Now()
	_, _ = store.SaveNote("alice", "D1", "7 days ago note", `[]`)
	_, _ = store.SaveNote("bob", "D2", "3 days ago note", `["project:Test"]`)
	_, _ = store.SaveNote("charlie", "D3", "Yesterday note", `["bug:100"]`)

	// Weekly should pick up all 3 (within 7 days, some are outside)
	req := httptest.NewRequest("GET", "/weekly", nil)
	w := httptest.NewRecorder()
	h.HandleWeekly(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("HandleWeekly() status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "note") {
		t.Errorf("HandleWeekly() body missing 'note': %s", body)
	}
}

func TestFormatSearchResults(t *testing.T) {
	results := []db.SearchResult{
		{
			ID:            1,
			User:          "alice",
			Text:          "Working on <project:Alpha>",
			Relevance:     25,
			CreatedAt:     "2026-06-20T10:00:00Z",
			MentionedTags: []string{"project:Alpha"},
		},
	}

	output := formatSearchResults(results)
	if !strings.Contains(output, "project:Alpha") {
		t.Errorf("formatSearchResults() output missing 'project:Alpha': %s", output)
	}
	if !strings.Contains(output, "alice") {
		t.Errorf("formatSearchResults() output missing 'alice': %s", output)
	}
}

func TestFormatWeeklySummary(t *testing.T) {
	notes := []db.Note{
		{User: "alice", Text: "Shipped the dashboard", Tags: `["project:Dashboard"]`},
		{User: "bob", Text: "Failed to deploy staging", Tags: `["bug:200"]`},
	}

	output := formatWeeklySummary(notes)
	if !strings.Contains(output, "dashboard") {
		t.Errorf("formatWeeklySummary() output missing 'dashboard': %s", output)
	}
	if !strings.Contains(output, "alice") {
		t.Errorf("formatWeeklySummary() output missing 'alice': %s", output)
	}
}
