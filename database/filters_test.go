package database

import (
	"testing"
	"time"
)

// TopicFilter options for GetTopicsFiltered.

func TestGetTopicsFiltered_Overdue(t *testing.T) {
	setupTestDB(t)

	// Set nowFn to a known time so we can control "now".
	base := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return base }
	defer func() { nowFn = time.Now }()

	// Add a topic that is overdue (next_review_at in past).
	AddTopic("Overdue Topic")

	// Advance nowFn forward 5 days so the topic is overdue.
	nowFn = func() time.Time { return base.AddDate(0, 0, 5) }

	// Add a future topic (manually set next_review_at to 10 days from now).
	AddTopic("Future Topic")
	topics, _ := GetAllTopics()
	for _, tp := range topics {
		if tp.Topic == "Future Topic" {
			db.Exec(`UPDATE topics SET next_review_at = ? WHERE id = ?`,
				base.AddDate(0, 0, 15), tp.ID)
		}
	}

	results, err := GetTopicsFiltered(TopicFilter{Overdue: true})
	if err != nil {
		t.Fatalf("GetTopicsFiltered overdue: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 overdue topic, got %d", len(results))
	}
	if results[0].Topic != "Overdue Topic" {
		t.Errorf("expected 'Overdue Topic', got %q", results[0].Topic)
	}
}

func TestGetTopicsFiltered_Completed(t *testing.T) {
	setupTestDB(t)

	AddTopic("Done Topic")
	AddTopic("Pending Topic")

	topics, _ := GetAllTopics()
	for _, tp := range topics {
		if tp.Topic == "Done Topic" {
			for i := 0; i < 5; i++ {
				MarkTopicDone(tp.ID)
			}
		}
	}

	results, err := GetTopicsFiltered(TopicFilter{Completed: true})
	if err != nil {
		t.Fatalf("GetTopicsFiltered completed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 completed topic, got %d", len(results))
	}
	if results[0].Topic != "Done Topic" {
		t.Errorf("expected 'Done Topic', got %q", results[0].Topic)
	}
}

func TestGetTopicsFiltered_IncludeArchived(t *testing.T) {
	setupTestDB(t)

	AddTopic("Archived Topic")
	AddTopic("Active Topic")

	topics, _ := GetAllTopics()
	for _, tp := range topics {
		if tp.Topic == "Archived Topic" {
			ArchiveTopic(tp.ID)
		}
	}

	// Default list should exclude archived.
	allTopics, _ := GetAllTopics()
	for _, tp := range allTopics {
		if tp.Topic == "Archived Topic" {
			t.Error("archived topic should not appear in GetAllTopics")
		}
	}

	// With IncludeArchived flag it should appear.
	results, err := GetTopicsFiltered(TopicFilter{IncludeArchived: true})
	if err != nil {
		t.Fatalf("GetTopicsFiltered archived: %v", err)
	}
	found := false
	for _, r := range results {
		if r.Topic == "Archived Topic" {
			found = true
		}
	}
	if !found {
		t.Error("expected archived topic in IncludeArchived results")
	}
}

func TestGetTopicsFiltered_ProjectFilter(t *testing.T) {
	setupTestDB(t)

	p1, _ := GetOrCreateProject("Math")
	p2, _ := GetOrCreateProject("CS")
	AddTopicWithProject("Calculus", p1)
	AddTopicWithProject("Algorithms", p2)
	AddTopic("No Project")

	results, err := GetTopicsFiltered(TopicFilter{ProjectID: &p1})
	if err != nil {
		t.Fatalf("GetTopicsFiltered by project: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 Math topic, got %d", len(results))
	}
	if results[0].Topic != "Calculus" {
		t.Errorf("expected 'Calculus', got %q", results[0].Topic)
	}
}

func TestGetTopicsFiltered_NoFilters_ExcludesArchived(t *testing.T) {
	setupTestDB(t)

	AddTopic("Visible")
	AddTopic("Hidden")

	topics, _ := GetAllTopics()
	for _, tp := range topics {
		if tp.Topic == "Hidden" {
			ArchiveTopic(tp.ID)
		}
	}

	results, err := GetTopicsFiltered(TopicFilter{})
	if err != nil {
		t.Fatalf("GetTopicsFiltered: %v", err)
	}
	for _, r := range results {
		if r.Archived {
			t.Error("archived topic should not appear without IncludeArchived flag")
		}
	}
	if len(results) != 1 {
		t.Errorf("expected 1 visible topic, got %d", len(results))
	}
}
