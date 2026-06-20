package database

import (
	"testing"
)

func TestAddTopicWithNotes(t *testing.T) {
	setupTestDB(t)

	if _, err := AddTopicWithNotes("Merge Sort", "Divide and conquer, O(n log n)"); err != nil {
		t.Fatalf("AddTopicWithNotes: %v", err)
	}

	topics, err := GetAllTopics()
	if err != nil {
		t.Fatalf("GetAllTopics: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}
	if topics[0].Notes != "Divide and conquer, O(n log n)" {
		t.Errorf("expected notes %q, got %q", "Divide and conquer, O(n log n)", topics[0].Notes)
	}
}

func TestAddTopic_EmptyNotes(t *testing.T) {
	setupTestDB(t)

	AddTopic("No Notes Topic")
	topics, _ := GetAllTopics()

	if topics[0].Notes != "" {
		t.Errorf("expected empty notes, got %q", topics[0].Notes)
	}
}

func TestUpdateNotes(t *testing.T) {
	setupTestDB(t)

	AddTopic("Topic Without Notes")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := UpdateNotes(id, "Some useful notes"); err != nil {
		t.Fatalf("UpdateNotes: %v", err)
	}

	topics, _ = GetAllTopics()
	if topics[0].Notes != "Some useful notes" {
		t.Errorf("expected %q, got %q", "Some useful notes", topics[0].Notes)
	}
}

func TestAddTopicFullOptions(t *testing.T) {
	setupTestDB(t)

	pID, _ := GetOrCreateProject("Go")
	_, err := AddTopicFull("Goroutines", "Lightweight threads managed by the Go runtime", pID)
	if err != nil {
		t.Fatalf("AddTopicFull: %v", err)
	}

	topics, _ := GetAllTopics()
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}
	got := topics[0]
	if got.Topic != "Goroutines" {
		t.Errorf("topic: expected %q, got %q", "Goroutines", got.Topic)
	}
	if got.Notes != "Lightweight threads managed by the Go runtime" {
		t.Errorf("notes: expected %q, got %q", "Lightweight threads managed by the Go runtime", got.Notes)
	}
	if got.ProjectName != "Go" {
		t.Errorf("project: expected %q, got %q", "Go", got.ProjectName)
	}
}
