package markup

import (
	"testing"
)

func TestParseBasicTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTags []Tag
	}{
		{
			name:  "project tag",
			input: "<project:Alpha>",
			wantTags: []Tag{
				{Type: "project", Value: "Alpha"},
			},
		},
		{
			name:  "bug tag",
			input: "<bug:123>",
			wantTags: []Tag{
				{Type: "bug", Value: "123"},
			},
		},
		{
			name:  "task tag",
			input: "<task:AUTH-456>",
			wantTags: []Tag{
				{Type: "task", Value: "AUTH-456"},
			},
		},
		{
			name:  "tech tag",
			input: "<tech:Golang>",
			wantTags: []Tag{
				{Type: "tech", Value: "Golang"},
			},
		},
		{
			name:    "multiple tags",
			input:   "<project:Alpha> working on auth <task:AUTH-123>",
			wantTags: []Tag{
				{Type: "project", Value: "Alpha"},
				{Type: "task", Value: "AUTH-123"},
			},
		},
		{
			name:    "mixed tags",
			input:   "<project:Beta> fixed <bug:42> using <tech:Go>",
			wantTags: []Tag{
				{Type: "project", Value: "Beta"},
				{Type: "bug", Value: "42"},
				{Type: "tech", Value: "Go"},
			},
		},
		{
			name:     "no tags",
			input:    "just plain text",
			wantTags: nil,
		},
		{
			name:     "empty string",
			input:    "",
			wantTags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.input)
			if len(tt.wantTags) == 0 && len(got) != 0 {
				t.Errorf("Parse(%q) = %v; want nil/empty", tt.input, got)
				return
			}
			if len(got) != len(tt.wantTags) {
				t.Errorf("Parse(%q) = %d tags; want %d", tt.input, len(got), len(tt.wantTags))
				for i := range got {
					t.Errorf("  got[%d] = %+v", i, got[i])
				}
				return
			}
			for i := range tt.wantTags {
				if got[i] != tt.wantTags[i] {
					t.Errorf("Parse(%q)[%d] = %+v; want %+v", tt.input, i, got[i], tt.wantTags[i])
				}
			}
		})
	}
}

func TestParseWhitespaceInTagValue(t *testing.T) {
	got := Parse("<project:Alpha Team>")
	if len(got) != 1 {
		t.Fatalf("Parse(<project:Alpha Team>) = %d tags; want 1", len(got))
	}
	if got[0].Value != "Alpha Team" {
		t.Errorf("Value = %q; want %q", got[0].Value, "Alpha Team")
	}
}

func TestTagListDedup(t *testing.T) {
	tags := []Tag{
		{Type: "project", Value: "Alpha"},
		{Type: "project", Value: "Alpha"},
		{Type: "tech", Value: "Go"},
	}
	list := TagList(tags)
	if len(list) != 2 {
		t.Errorf("TagList() = %d items; want 2", len(list))
	}
	seen := make(map[string]bool)
	for _, item := range list {
		if seen[item] {
			t.Errorf("TagList() contains duplicate: %q", item)
		}
		seen[item] = true
	}
}

func TestExtractByType(t *testing.T) {
	tags := []Tag{
		{Type: "project", Value: "Alpha"},
		{Type: "project", Value: "Beta"},
		{Type: "tech", Value: "Go"},
		{Type: "project", Value: "Gamma"},
	}

	got := ExtractByType(tags, "project")
	if len(got) != 3 {
		t.Errorf("ExtractByType(project) = %d; want 3", len(got))
	}

	got = ExtractByType(tags, "tech")
	if len(got) != 1 || got[0] != "Go" {
		t.Errorf("ExtractByType(tech) = %v; want [Go]", got)
	}

	got = ExtractByType(tags, "bug")
	if len(got) != 0 {
		t.Errorf("ExtractByType(bug) = %v; want empty", got)
	}
}
