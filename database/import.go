package database

import (
	"encoding/csv"
	"fmt"
	"strings"
	"time"
)

// ImportGroup is the parsed representation of one project section from an
// exported file. Topics contains the lightweight import-only view of each row.
type ImportGroup struct {
	ProjectName        string
	ProjectDescription string
	Topics             []ImportTopic
}

// ImportTopic holds only the fields needed when importing a topic from a file.
type ImportTopic struct {
	Topic string
	Notes string
}

// ── CSV import ────────────────────────────────────────────────────────────────

// expectedCSVHeader is the canonical header produced by RenderCSV.
var expectedCSVColumns = []string{
	"ID", "Topic", "Project", "Project Description", "Notes",
	"Cycle", "Created", "Next Review", "Status",
}

// ParseCSVImport reads a CSV string (as produced by RenderCSV / the template)
// and returns the data grouped by project.
func ParseCSVImport(content string) ([]ImportGroup, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("CSV content is empty")
	}

	r := csv.NewReader(strings.NewReader(content))
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV parse error: %w", err)
	}
	if len(records) < 1 {
		return nil, fmt.Errorf("CSV has no rows")
	}

	// Validate header.
	header := records[0]
	if len(header) < len(expectedCSVColumns) {
		return nil, fmt.Errorf("CSV header has %d columns, expected at least %d",
			len(header), len(expectedCSVColumns))
	}
	for i, want := range expectedCSVColumns {
		if !strings.EqualFold(strings.TrimSpace(header[i]), want) {
			return nil, fmt.Errorf("CSV column %d: expected %q, got %q", i, want, header[i])
		}
	}

	// col indices
	const (
		colTopic   = 1
		colProject = 2
		colDesc    = 3
		colNotes   = 4
	)

	groupMap := make(map[string]*ImportGroup)
	var order []string

	for _, row := range records[1:] {
		if len(row) < len(expectedCSVColumns) {
			continue
		}
		topic := strings.TrimSpace(row[colTopic])
		projectName := strings.TrimSpace(row[colProject])
		desc := strings.TrimSpace(row[colDesc])
		notes := strings.TrimSpace(row[colNotes])

		if _, ok := groupMap[projectName]; !ok {
			groupMap[projectName] = &ImportGroup{
				ProjectName:        projectName,
				ProjectDescription: desc,
			}
			order = append(order, projectName)
		}
		// Use first non-empty description encountered for the project.
		if groupMap[projectName].ProjectDescription == "" && desc != "" {
			groupMap[projectName].ProjectDescription = desc
		}
		groupMap[projectName].Topics = append(groupMap[projectName].Topics, ImportTopic{
			Topic: topic,
			Notes: notes,
		})
	}

	groups := make([]ImportGroup, 0, len(order))
	for _, key := range order {
		groups = append(groups, *groupMap[key])
	}
	return groups, nil
}

// ── Markdown import ───────────────────────────────────────────────────────────

// ParseMarkdownImport reads a Markdown string (as produced by RenderMarkdown)
// and returns the data grouped by project.
func ParseMarkdownImport(content string) ([]ImportGroup, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("markdown content is empty")
	}

	var groups []ImportGroup
	var current *ImportGroup
	headerSeen := false // have we passed the table header row in the current section?

	lines := strings.Split(content, "\n")
	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")

		// Project heading: ## Name
		if strings.HasPrefix(line, "## ") {
			name := strings.TrimPrefix(line, "## ")
			groups = append(groups, ImportGroup{ProjectName: strings.TrimSpace(name)})
			current = &groups[len(groups)-1]
			headerSeen = false
			continue
		}

		if current == nil {
			continue
		}

		// Description line: _some text_ (single-line italic, right after ## heading)
		if strings.HasPrefix(line, "_") && strings.HasSuffix(line, "_") && len(line) > 2 {
			text := line[1 : len(line)-1]
			// Ignore the "Generated: …" line at document level (current is set, but
			// we haven't seen any topics yet and this is NOT after a ## heading).
			if current.ProjectDescription == "" && len(current.Topics) == 0 {
				current.ProjectDescription = text
			}
			continue
		}

		// Table separator: |----|...|
		if strings.HasPrefix(line, "|") && strings.Contains(line, "---") {
			continue
		}

		// Table row: | ... |
		if strings.HasPrefix(line, "|") {
			cells := splitTableRow(line)
			if len(cells) < 6 {
				continue
			}
			// Header row detection — second cell (index 1) equals "Topic"
			if strings.EqualFold(cells[1], "Topic") {
				headerSeen = true
				continue
			}
			if !headerSeen {
				continue
			}
			topic := cells[1]
			notes := ""
			if len(cells) >= 6 {
				notes = cells[5]
			}
			// Unescape pipe characters escaped by the renderer.
			notes = strings.ReplaceAll(notes, `\|`, "|")
			current.Topics = append(current.Topics, ImportTopic{
				Topic: topic,
				Notes: notes,
			})
		}
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("no project sections (## headings) found in markdown")
	}
	return groups, nil
}

// splitTableRow splits a markdown table row like "| a | b | c |" into
// trimmed cell values, correctly handling escaped pipes (\|).
func splitTableRow(line string) []string {
	// Replace escaped pipes with a placeholder so they aren't treated as
	// column delimiters, then restore them after splitting.
	const placeholder = "\x00"
	line = strings.ReplaceAll(line, `\|`, placeholder)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(strings.ReplaceAll(p, placeholder, `\|`))
	}
	return cells
}

// ── ImportGroups ──────────────────────────────────────────────────────────────

// ImportGroups persists the parsed groups into the database.
// Topics with blank names are skipped. Projects with empty names are not created
// and their topics are stored without a project association.
// Returns the count of topics successfully imported.
func ImportGroups(groups []ImportGroup) (int, error) {
	count := 0
	for _, g := range groups {
		var projectID int64

		if strings.TrimSpace(g.ProjectName) != "" {
			var err error
			if g.ProjectDescription != "" {
				// Try to get existing project first, then create with description.
				projectID, err = GetOrCreateProject(g.ProjectName)
				if err != nil {
					return count, fmt.Errorf("project %q: %w", g.ProjectName, err)
				}
				// Update description if we just created it or it's empty.
				UpdateProjectDescription(projectID, g.ProjectDescription)
			} else {
				projectID, err = GetOrCreateProject(g.ProjectName)
				if err != nil {
					return count, fmt.Errorf("project %q: %w", g.ProjectName, err)
				}
			}
		}

		for _, t := range g.Topics {
			if strings.TrimSpace(t.Topic) == "" {
				continue
			}
			var err error
			if projectID != 0 {
				err = AddTopicFull(t.Topic, t.Notes, projectID)
			} else {
				err = AddTopicWithNotes(t.Topic, t.Notes)
			}
			if err != nil {
				return count, fmt.Errorf("topic %q: %w", t.Topic, err)
			}
			count++
		}
	}
	return count, nil
}

// ── Templates ─────────────────────────────────────────────────────────────────

// RenderMarkdownTemplate returns a complete, import-ready markdown example
// with realistic mock data across two projects.
func RenderMarkdownTemplate() string {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	groups := []TopicGroup{
		{
			ProjectName:        "Algorithms",
			ProjectDescription: "Core algorithmic problem-solving patterns",
			Topics: []Topic{
				{
					ID: 1, Topic: "Binary Search",
					Notes:        "Divide search space in half each iteration. O(log n).",
					ReviewCycle:  2, NextReviewAt: base.AddDate(0, 0, 15),
					CreatedAt: base, Completed: false, Archived: false,
				},
				{
					ID: 2, Topic: "Merge Sort",
					Notes:        "Divide and conquer. Stable sort. O(n log n) time, O(n) space.",
					ReviewCycle:  3, NextReviewAt: base.AddDate(0, 0, 22),
					CreatedAt: base, Completed: false, Archived: false,
				},
				{
					ID: 3, Topic: "Dynamic Programming",
					Notes:        "Break into overlapping subproblems, memoize. Bottom-up or top-down.",
					ReviewCycle:  4, NextReviewAt: base.AddDate(0, 0, 30),
					CreatedAt: base, Completed: true, Archived: false,
				},
			},
		},
		{
			ProjectName:        "Systems",
			ProjectDescription: "Operating systems, networking, distributed systems",
			Topics: []Topic{
				{
					ID: 4, Topic: "TCP/IP Three-Way Handshake",
					Notes:        "SYN → SYN-ACK → ACK. Establishes reliable connection.",
					ReviewCycle:  1, NextReviewAt: base.AddDate(0, 0, 3),
					CreatedAt: base, Completed: false, Archived: false,
				},
				{
					ID: 5, Topic: "CAP Theorem",
					Notes:        "Consistency, Availability, Partition-tolerance — pick two.",
					ReviewCycle:  2, NextReviewAt: base.AddDate(0, 0, 8),
					CreatedAt: base, Completed: false, Archived: true,
				},
				{
					ID: 6, Topic: "Virtual Memory",
					Notes:        "",
					ReviewCycle:  0, NextReviewAt: base,
					CreatedAt: base, Completed: false, Archived: false,
				},
			},
		},
	}
	return RenderMarkdown(groups)
}

// RenderCSVTemplate returns a complete, import-ready CSV example with realistic
// mock data across two projects.
func RenderCSVTemplate() string {
	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	groups := []TopicGroup{
		{
			ProjectName:        "Algorithms",
			ProjectDescription: "Core algorithmic problem-solving patterns",
			Topics: []Topic{
				{ID: 1, Topic: "Binary Search", Notes: "Divide search space. O(log n).", ReviewCycle: 2, CreatedAt: base, NextReviewAt: base.AddDate(0, 0, 15)},
				{ID: 2, Topic: "Merge Sort", Notes: "Stable sort. O(n log n).", ReviewCycle: 3, CreatedAt: base, NextReviewAt: base.AddDate(0, 0, 22)},
				{ID: 3, Topic: "Dynamic Programming", Notes: "Memoize overlapping subproblems.", ReviewCycle: 4, CreatedAt: base, NextReviewAt: base.AddDate(0, 0, 30), Completed: true},
			},
		},
		{
			ProjectName:        "Systems",
			ProjectDescription: "OS, networking, distributed systems",
			Topics: []Topic{
				{ID: 4, Topic: "TCP/IP Handshake", Notes: "SYN, SYN-ACK, ACK.", ReviewCycle: 1, CreatedAt: base, NextReviewAt: base.AddDate(0, 0, 3)},
				{ID: 5, Topic: "CAP Theorem", Notes: "Consistency, Availability, Partition-tolerance.", ReviewCycle: 2, CreatedAt: base, NextReviewAt: base.AddDate(0, 0, 8), Archived: true},
				{ID: 6, Topic: "Virtual Memory", Notes: "", ReviewCycle: 0, CreatedAt: base, NextReviewAt: base},
			},
		},
	}
	return RenderCSV(groups)
}
