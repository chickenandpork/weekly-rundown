package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/chickenandpork/weekly-rundown/internal/db"
	"github.com/chickenandpork/weekly-rundown/internal/markup"
)

// Config holds the service configuration.
type Config struct {
	SlackBotToken string
}

// Handlers holds HTTP handlers and the DB connection.
type Handlers struct {
	store *db.Store
	cfg   Config
	notes int64
}

// New creates a new Handlers instance with a count of existing notes.
func New(store *db.Store, cfg Config) *Handlers {
	count, _ := store.NoteCount()
	return &Handlers{store: store, cfg: cfg, notes: count}
}

// HandleSlashCommand processes incoming Slack slash commands.
func (h *Handlers) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	// Verify Slack token via X-Slack-Secret header
	slackSecret := r.Header.Get("X-Slack-Secret")
	if slackSecret != "" && slackSecret != h.cfg.SlackBotToken {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	// Parse form data (Slack sends x-www-form-urlencoded)
	r.ParseForm()
	text := r.FormValue("text")
	user := r.FormValue("user_name")
	command := r.FormValue("command")

	if text == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing text parameter"))
		return
	}

	// Route to handler
	switch command {
	case "/wr", "/weeklyrundown":
		h.handleNote(w, r, user, text)
	case "/wr daily", "/wr dailynote":
		h.handleNote(w, r, user, text)
	case "/wr search":
		h.handleSearch(w, text)
	case "/wr health":
		h.HandleHealth(w, r)
	default:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unknown command. Use /wr note, /wr search, /wr health, /wr daily"))
	}
}

func (h *Handlers) handleNote(w http.ResponseWriter, r *http.Request, user, text string) {
	tags := markup.Parse(text)
	tagList := markup.TagList(tags)

	id, err := h.store.SaveNote(user, r.FormValue("channel_name"), text,
		func() string { b, _ := json.Marshal(tagList); return string(b) }())
	if err != nil {
		h.respondError(w, fmt.Sprintf("Failed to save note: %v", err))
		return
	}

	h.notes++

	// Build response
	resp := fmt.Sprintf("Note saved (ID: %d). Tags found: %s.",
		id, strings.Join(tagList, ", "))

	if len(tags) == 0 {
		resp += "\nTip: Use markup like <project:NAME>, <bug:123>, <task:ABC-456>, <tech:Golang> to make notes searchable."
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resp))
}

func (h *Handlers) handleSearch(w http.ResponseWriter, query string) {
	results, err := h.store.Search(query, 20)
	if err != nil {
		h.respondError(w, fmt.Sprintf("Search failed: %v", err))
		return
	}

	if len(results) == 0 {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("No results found."))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(formatSearchResults(results)))
}

// HandleHealth responds to /wr health and GET /health.
func (h *Handlers) HandleHealth(w http.ResponseWriter, r *http.Request) {
	_ = r
	count, _ := h.store.NoteCount()

	resp := map[string]interface{}{
		"status":      "ok",
		"notes_count": count,
		"db_status":   "ok",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleWeekly returns notes from the past 7 days for scheduled reviews.
func (h *Handlers) HandleWeekly(w http.ResponseWriter, r *http.Request) {
	_ = r
	end := time.Now()
	start := end.AddDate(0, 0, -7)

	notes, err := h.store.NotesByDateRange(start, end, 50)
	if err != nil {
		h.respondError(w, fmt.Sprintf("Failed to fetch weekly notes: %v", err))
		return
	}

	if len(notes) == 0 {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("No notes in the past 7 days."))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(formatWeeklySummary(notes)))
}

func (h *Handlers) respondError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}
