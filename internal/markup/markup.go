package markup

import (
	"regexp"
	"strings"
)

// Tag types supported by the markup syntax.
const (
	TagProject = "project"
	TagBug     = "bug"
	TagTask    = "task"
	TagTech    = "tech"
)

// Tag represents a single parsed tag.
type Tag struct {
	Type  string
	Value string
}

// Regex patterns for identifier markup.
// Formats: <project:NAME>, <bug:123>, <task:ABC-456>, <tech:Golang>
var tagPattern = regexp.MustCompile(`<(\w+):([^>]+)>`)

// Parse extracts all structured tags from the note text.
func Parse(text string) []Tag {
	matches := tagPattern.FindAllStringSubmatch(text, -1)
	tags := make([]Tag, 0, len(matches))
	for _, m := range matches {
		if len(m) == 3 {
			tags = append(tags, Tag{
				Type:  m[1],
				Value: strings.TrimSpace(m[2]),
			})
		}
	}
	return tags
}

// TagList returns a deduplicated sorted list of tag values.
func TagList(tags []Tag) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, t := range tags {
		key := t.Type + ":" + t.Value
		if !seen[key] {
			seen[key] = true
			result = append(result, key)
		}
	}
	return result
}

// ExtractByType filters tags by type.
func ExtractByType(tags []Tag, t string) []string {
	result := make([]string, 0)
	for _, tag := range tags {
		if tag.Type == t {
			result = append(result, tag.Value)
		}
	}
	return result
}
