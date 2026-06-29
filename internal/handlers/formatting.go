package handlers

import (
	"fmt"

	"github.com/chickenandpork/weekly-rundown/internal/db"
	"github.com/chickenandpork/weekly-rundown/internal/markup"
)

func formatSearchResults(results []db.SearchResult) string {
	out := fmt.Sprintf("Found %d results:\n\n", len(results))
	for i, r := range results {
		out += fmt.Sprintf("### %d. %s (@%s, %s)\n", i+1, r.Text, r.User, r.CreatedAt)
		if r.Relevance > 0 {
			out += fmt.Sprintf("Relevance: %d\n", r.Relevance)
		}
		if len(r.MentionedTags) > 0 {
			out += fmt.Sprintf("Tags: %s\n", r.MentionedTags)
		}
		out += "\n"
	}
	return out
}

func formatWeeklySummary(notes []db.Note) string {
	if len(notes) == 0 {
		return "No notes in the past 7 days."
	}

	byUser := make(map[string][]db.Note)
	for _, n := range notes {
		byUser[n.User] = append(byUser[n.User], n)
	}

	out := fmt.Sprintf("## Weekly Rundown (%d notes)\n\n", len(notes))
	for user, userNotes := range byUser {
		out += fmt.Sprintf("### @%s (%d notes)\n", user, len(userNotes))
		for _, n := range userNotes {
			out += fmt.Sprintf("- [%s] %s\n", n.CreatedAt.Format("2006-01-02 15:04"), n.Text)
			tags := markup.Parse(n.Text)
			if len(tags) > 0 {
				for _, t := range tags {
					out += fmt.Sprintf("  - <%s:%s>\n", t.Type, t.Value)
				}
			}
		}
		out += "\n"
	}
	return out
}
