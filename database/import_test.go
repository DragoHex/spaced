package database

import (
	"strings"
	"testing"
)

// ── CSV import ────────────────────────────────────────────────────────────────

func TestParseCSVImport_ValidInput(t *testing.T) {
	csv := `ID,Topic,Project,Project Description,Notes,Cycle,Created,Next Review,Status
1,Binary Search,DSA,Data Structures,O(log n),Day 3,2026-06-01,2026-06-08,In Progress
2,Merge Sort,DSA,Data Structures,Divide and conquer,Day 1,2026-06-01,2026-06-07,Completed
3,TCP/IP,Systems,,Three-way handshake,Day 8,2026-06-01,2026-06-09,In Progress`

	groups, err := ParseCSVImport(csv)
	if err != nil {
		t.Fatalf("ParseCSVImport: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups (DSA, Systems), got %d", len(groups))
	}

	dsa := findImportGroup(groups, "DSA")
	if dsa == nil {
		t.Fatal("expected DSA group")
	}
	if dsa.ProjectDescription != "Data Structures" {
		t.Errorf("expected DSA description 'Data Structures', got %q", dsa.ProjectDescription)
	}
	if len(dsa.Topics) != 2 {
		t.Errorf("expected 2 DSA topics, got %d", len(dsa.Topics))
	}
	if dsa.Topics[0].Topic != "Binary Search" {
		t.Errorf("expected 'Binary Search', got %q", dsa.Topics[0].Topic)
	}
	if dsa.Topics[0].Notes != "O(log n)" {
		t.Errorf("expected notes 'O(log n)', got %q", dsa.Topics[0].Notes)
	}
}

func TestParseCSVImport_EmptyProjectNameTreatedAsLiteral(t *testing.T) {
	csv := `ID,Topic,Project,Project Description,Notes,Cycle,Created,Next Review,Status
1,Orphan Topic,,,,Day 1,2026-06-01,2026-06-01,In Progress`

	groups, err := ParseCSVImport(csv)
	if err != nil {
		t.Fatalf("ParseCSVImport: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].ProjectName != "" {
		t.Errorf("empty project name should be preserved, got %q", groups[0].ProjectName)
	}
}

func TestParseCSVImport_InvalidHeader(t *testing.T) {
	csv := `Name,Topic,Blah
1,test,stuff`

	_, err := ParseCSVImport(csv)
	if err == nil {
		t.Error("expected error for invalid CSV header")
	}
}

func TestParseCSVImport_EmptyInput(t *testing.T) {
	_, err := ParseCSVImport("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseCSVImport_QuotedFields(t *testing.T) {
	csv := `ID,Topic,Project,Project Description,Notes,Cycle,Created,Next Review,Status
1,"Topic, with comma",DSA,"Desc, quoted","Notes with ""quotes""",Day 1,2026-06-01,2026-06-01,In Progress`

	groups, err := ParseCSVImport(csv)
	if err != nil {
		t.Fatalf("ParseCSVImport: %v", err)
	}
	if len(groups) == 0 || len(groups[0].Topics) == 0 {
		t.Fatal("expected at least 1 topic")
	}
	topic := groups[0].Topics[0]
	if topic.Topic != "Topic, with comma" {
		t.Errorf("expected quoted topic name, got %q", topic.Topic)
	}
	if topic.Notes != `Notes with "quotes"` {
		t.Errorf("expected unescaped quotes in notes, got %q", topic.Notes)
	}
}

// ── Markdown import ───────────────────────────────────────────────────────────

func TestParseMarkdownImport_ValidInput(t *testing.T) {
	md := `# Spaced Repetition Export

_Generated: 2026-06-07_

## Algorithms

_Core algorithmic patterns_

| ID | Topic | Cycle | Next Review | Status | Notes |
|----|-------|-------|-------------|--------|-------|
| 1 | Binary Search | Day 3 | 2026-06-08 | In Progress | O(log n) search |
| 2 | Merge Sort | Day 8 | 2026-06-15 | Completed | Stable, O(n log n) |

## Systems

| ID | Topic | Cycle | Next Review | Status | Notes |
|----|-------|-------|-------------|--------|-------|
| 3 | TCP/IP | Day 1 | 2026-06-08 | In Progress | Three-way handshake |
`

	groups, err := ParseMarkdownImport(md)
	if err != nil {
		t.Fatalf("ParseMarkdownImport: %v", err)
	}

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	algo := findImportGroup(groups, "Algorithms")
	if algo == nil {
		t.Fatal("expected Algorithms group")
	}
	if algo.ProjectDescription != "Core algorithmic patterns" {
		t.Errorf("expected description 'Core algorithmic patterns', got %q", algo.ProjectDescription)
	}
	if len(algo.Topics) != 2 {
		t.Errorf("expected 2 Algorithms topics, got %d", len(algo.Topics))
	}

	sys := findImportGroup(groups, "Systems")
	if sys == nil {
		t.Fatal("expected Systems group")
	}
	if sys.ProjectDescription != "" {
		t.Errorf("Systems has no description, expected empty, got %q", sys.ProjectDescription)
	}
	if len(sys.Topics) != 1 {
		t.Errorf("expected 1 Systems topic, got %d", len(sys.Topics))
	}
	if sys.Topics[0].Topic != "TCP/IP" {
		t.Errorf("expected 'TCP/IP', got %q", sys.Topics[0].Topic)
	}
	if sys.Topics[0].Notes != "Three-way handshake" {
		t.Errorf("expected 'Three-way handshake', got %q", sys.Topics[0].Notes)
	}
}

func TestParseMarkdownImport_UnassignedGroup(t *testing.T) {
	md := `# Spaced Repetition Export

_Generated: 2026-06-07_

## Unassigned

| ID | Topic | Cycle | Next Review | Status | Notes |
|----|-------|-------|-------------|--------|-------|
| 5 | Orphan Topic | Day 1 | 2026-06-01 | In Progress |  |
`
	groups, err := ParseMarkdownImport(md)
	if err != nil {
		t.Fatalf("ParseMarkdownImport: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].ProjectName != "Unassigned" {
		t.Errorf("expected 'Unassigned', got %q", groups[0].ProjectName)
	}
	if len(groups[0].Topics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(groups[0].Topics))
	}
}

func TestParseMarkdownImport_EmptyInput(t *testing.T) {
	_, err := ParseMarkdownImport("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseMarkdownImport_NotesPipeEscape(t *testing.T) {
	md := `# Spaced Repetition Export

_Generated: 2026-06-07_

## Math

| ID | Topic | Cycle | Next Review | Status | Notes |
|----|-------|-------|-------------|--------|-------|
| 1 | Sets | Day 1 | 2026-06-01 | In Progress | A \| B means union |
`
	groups, err := ParseMarkdownImport(md)
	if err != nil {
		t.Fatalf("ParseMarkdownImport: %v", err)
	}
	if len(groups) == 0 || len(groups[0].Topics) == 0 {
		t.Fatal("expected at least 1 topic")
	}
	// Escaped pipe should be unescaped on import
	notes := groups[0].Topics[0].Notes
	if !strings.Contains(notes, "|") {
		t.Errorf("expected unescaped pipe in notes, got %q", notes)
	}
}

// ── ImportGroups ──────────────────────────────────────────────────────────────

func TestImportGroups_CreatesTopicsAndProjects(t *testing.T) {
	setupTestDB(t)

	groups := []ImportGroup{
		{
			ProjectName:        "Go",
			ProjectDescription: "Go programming",
			Topics: []ImportTopic{
				{Topic: "Goroutines", Notes: "Lightweight threads"},
				{Topic: "Channels", Notes: "Communication between goroutines"},
			},
		},
		{
			ProjectName: "Rust",
			Topics: []ImportTopic{
				{Topic: "Ownership", Notes: "Memory safety without GC"},
			},
		},
	}

	count, err := ImportGroups(groups)
	if err != nil {
		t.Fatalf("ImportGroups: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 imported topics, got %d", count)
	}

	projects, _ := GetAllProjects()
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}

	topics, _ := GetAllTopics()
	if len(topics) != 3 {
		t.Errorf("expected 3 topics in DB, got %d", len(topics))
	}

	// Check description preserved
	goTopics, _ := GetTopicsByProject(findProjectID(projects, "Go"))
	if len(goTopics) != 2 {
		t.Errorf("expected 2 Go topics, got %d", len(goTopics))
	}

	goProj := findProjectByName(projects, "Go")
	if goProj == nil || goProj.Description != "Go programming" {
		t.Errorf("expected Go project description 'Go programming', got %v", goProj)
	}
}

func TestImportGroups_SkipsBlankTopicNames(t *testing.T) {
	setupTestDB(t)

	groups := []ImportGroup{
		{
			ProjectName: "Test",
			Topics: []ImportTopic{
				{Topic: "Valid Topic", Notes: ""},
				{Topic: "", Notes: "No topic name"},
				{Topic: "  ", Notes: "Whitespace only"},
			},
		},
	}

	count, err := ImportGroups(groups)
	if err != nil {
		t.Fatalf("ImportGroups: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 imported topic (blank names skipped), got %d", count)
	}
}

func TestImportGroups_EmptyProjectNameSkipsProjectCreation(t *testing.T) {
	setupTestDB(t)

	groups := []ImportGroup{
		{
			ProjectName: "",
			Topics:      []ImportTopic{{Topic: "Floating Topic", Notes: ""}},
		},
	}

	count, err := ImportGroups(groups)
	if err != nil {
		t.Fatalf("ImportGroups: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 imported topic, got %d", count)
	}

	projects, _ := GetAllProjects()
	if len(projects) != 0 {
		t.Errorf("expected no projects created for empty name, got %d", len(projects))
	}

	topics, _ := GetAllTopics()
	if topics[0].ProjectID != nil {
		t.Errorf("topic with empty project should have nil project_id")
	}
}

// ── Template rendering ────────────────────────────────────────────────────────

func TestRenderMarkdownTemplate_ContainsMockData(t *testing.T) {
	out := RenderMarkdownTemplate()

	if !strings.Contains(out, "# Spaced Repetition Export") {
		t.Error("template should have title")
	}
	if !strings.Contains(out, "## ") {
		t.Error("template should have at least one project heading")
	}
	// Should have at least two project sections
	count := strings.Count(out, "## ")
	if count < 2 {
		t.Errorf("template should have at least 2 projects, found %d", count)
	}
	// Should have table rows
	if !strings.Contains(out, "| ") {
		t.Error("template should have markdown table rows")
	}
}

func TestRenderCSVTemplate_ContainsMockData(t *testing.T) {
	out := RenderCSVTemplate()

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 4 {
		t.Fatalf("template should have header + at least 3 data rows, got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[0], "ID,Topic,Project") {
		t.Errorf("first line should be header, got %q", lines[0])
	}
	// Should have rows from at least 2 different projects
	projects := map[string]bool{}
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, ",", 4)
		if len(fields) >= 3 {
			projects[fields[2]] = true
		}
	}
	if len(projects) < 2 {
		t.Errorf("template CSV should have topics from at least 2 projects, got %d", len(projects))
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func findImportGroup(groups []ImportGroup, name string) *ImportGroup {
	for i := range groups {
		if groups[i].ProjectName == name {
			return &groups[i]
		}
	}
	return nil
}

func findProjectID(projects []Project, name string) int64 {
	for _, p := range projects {
		if p.Name == name {
			return p.ID
		}
	}
	return 0
}

func findProjectByName(projects []Project, name string) *Project {
	for i := range projects {
		if projects[i].Name == name {
			return &projects[i]
		}
	}
	return nil
}
