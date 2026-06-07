package database

import (
	"strings"
	"testing"
)

func TestExportTopicsGroupedByProject(t *testing.T) {
	setupTestDB(t)

	p1, _ := GetOrCreateProject("DSA")
	p2, _ := GetOrCreateProject("Systems")

	AddTopicFull("Binary Search", "O(log n) search on sorted array", p1)
	AddTopicFull("Merge Sort", "Divide and conquer", p1)
	AddTopicFull("TCP/IP", "Connection-oriented protocol", p2)
	AddTopic("Unassigned Topic")

	groups, err := GetTopicsGroupedByProject()
	if err != nil {
		t.Fatalf("GetTopicsGroupedByProject: %v", err)
	}

	// Should have 3 groups: DSA, Systems, "" (unassigned)
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d: %v", len(groups), groupNames(groups))
	}

	dsaGroup := findGroup(groups, "DSA")
	if dsaGroup == nil {
		t.Fatal("expected DSA group")
	}
	if len(dsaGroup.Topics) != 2 {
		t.Errorf("expected 2 DSA topics, got %d", len(dsaGroup.Topics))
	}

	sysGroup := findGroup(groups, "Systems")
	if sysGroup == nil {
		t.Fatal("expected Systems group")
	}
	if len(sysGroup.Topics) != 1 {
		t.Errorf("expected 1 Systems topic, got %d", len(sysGroup.Topics))
	}

	unassigned := findGroup(groups, "")
	if unassigned == nil {
		t.Fatal("expected unassigned group")
	}
	if len(unassigned.Topics) != 1 {
		t.Errorf("expected 1 unassigned topic, got %d", len(unassigned.Topics))
	}
}

func TestExportGroupsIncludeCompletedAndArchived(t *testing.T) {
	setupTestDB(t)

	AddTopic("Active")
	AddTopic("Will Be Archived")

	topics, _ := GetAllTopics()
	for _, tp := range topics {
		if tp.Topic == "Will Be Archived" {
			ArchiveTopic(tp.ID)
		}
	}

	groups, err := GetTopicsGroupedByProject()
	if err != nil {
		t.Fatalf("GetTopicsGroupedByProject: %v", err)
	}

	total := 0
	for _, g := range groups {
		total += len(g.Topics)
	}
	// Export includes ALL topics including archived
	if total != 2 {
		t.Errorf("export should include archived topics; expected 2, got %d", total)
	}
}

func TestRenderMarkdown(t *testing.T) {
	setupTestDB(t)

	p, _ := GetOrCreateProject("Go")
	AddTopicFull("Goroutines", "Lightweight threads", p)

	groups, _ := GetTopicsGroupedByProject()
	out := RenderMarkdown(groups)

	if !strings.Contains(out, "# Spaced Repetition Export") {
		t.Error("markdown should have title")
	}
	if !strings.Contains(out, "## Go") {
		t.Error("markdown should have project heading")
	}
	if !strings.Contains(out, "Goroutines") {
		t.Error("markdown should contain topic name")
	}
	if !strings.Contains(out, "Lightweight threads") {
		t.Error("markdown should contain notes")
	}
}

func TestRenderCSV(t *testing.T) {
	setupTestDB(t)

	p, _ := GetOrCreateProject("Go")
	AddTopicFull("Channels", "Used for goroutine communication", p)

	groups, _ := GetTopicsGroupedByProject()
	out := RenderCSV(groups)

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least header + 1 data row, got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[0], "ID,Topic,Project") {
		t.Errorf("first line should be header, got %q", lines[0])
	}
	if !strings.Contains(out, "Channels") {
		t.Error("CSV should contain topic name")
	}
	if !strings.Contains(out, "Go") {
		t.Error("CSV should contain project name")
	}
}

// helpers

type projectGroup = TopicGroup

func groupNames(groups []TopicGroup) []string {
	var names []string
	for _, g := range groups {
		names = append(names, g.ProjectName)
	}
	return names
}

func findGroup(groups []TopicGroup, name string) *TopicGroup {
	for i, g := range groups {
		if g.ProjectName == name {
			return &groups[i]
		}
	}
	return nil
}
