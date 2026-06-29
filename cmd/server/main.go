package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chickenandpork/weekly-rundown/internal/db"
	"github.com/chickenandpork/weekly-rundown/internal/handlers"
	"github.com/gorilla/mux"
)

func main() {
	dbPath := getEnv("DB_PATH", "data/weekly-rundown.db")
	port := getEnv("PORT", "8080")
	slackBotToken := getEnv("SLACK_BOT_TOKEN", "")

	// Ensure data directory exists
	os.MkdirAll("data", 0755)

	// Initialize database
	sqlDB, err := db.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := sqlDB.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	cfg := handlers.Config{
		SlackBotToken: slackBotToken,
	}
	h := handlers.New(sqlDB, cfg)

	// Record start time for health checks
	startTime := time.Now().UTC().Format(time.RFC3339)

	// Setup router
	r := mux.NewRouter()
	r.HandleFunc("/slack/command", h.HandleSlashCommand).Methods("POST")
	r.HandleFunc("/health", h.HandleHealth).Methods("GET")
	r.HandleFunc("/weekly", h.HandleWeekly).Methods("GET")

	// Health check endpoint with uptime
	r.HandleFunc("/health/uptime", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"start_time": startTime,
			"uptime":     time.Since(timeMustParse(startTime)).Round(time.Second).String(),
		})
	}).Methods("GET")

	log.Printf("Starting weekly-rundown on :%s (db=%s, start=%s)", port, dbPath, startTime)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func timeMustParse(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(fmt.Sprintf("failed to parse time: %v", err))
	}
	return t
}
