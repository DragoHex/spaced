package database

import (
	"testing"
	"time"
)

func TestGetStats_BasicCounts(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return base }
	defer func() { nowFn = time.Now }()

	p1, _ := GetOrCreateProject("Math")
	p2, _ := GetOrCreateProject("CS")

	AddTopicWithProject("Calculus", p1)
	AddTopicWithProject("Algorithms", p2)
	AddTopicWithProject("Data Structures", p2)
	AddTopic("Unassigned")

	stats, err := GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}

	if stats.Total != 4 {
		t.Errorf("total: expected 4, got %d", stats.Total)
	}
	if stats.Completed != 0 {
		t.Errorf("completed: expected 0, got %d", stats.Completed)
	}
	if stats.Projects != 2 {
		t.Errorf("projects: expected 2, got %d", stats.Projects)
	}
	if stats.InProgress != 4 {
		t.Errorf("in progress: expected 4, got %d", stats.InProgress)
	}
}

func TestGetStats_CountsCompleted(t *testing.T) {
	setupTestDB(t)

	AddTopic("A")
	AddTopic("B")

	topics, _ := GetAllTopics()
	for _, tp := range topics {
		if tp.Topic == "A" {
			for i := 0; i < 5; i++ {
				MarkTopicDone(tp.ID)
			}
		}
	}

	stats, _ := GetStats()
	if stats.Completed != 1 {
		t.Errorf("expected 1 completed, got %d", stats.Completed)
	}
	if stats.Total != 2 {
		t.Errorf("expected total 2, got %d", stats.Total)
	}
}

func TestGetStats_Overdue(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return base }
	defer func() { nowFn = time.Now }()

	AddTopic("Overdue")

	// Advance time so the topic is overdue.
	nowFn = func() time.Time { return base.AddDate(0, 0, 5) }

	stats, err := GetStats()
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if stats.Overdue != 1 {
		t.Errorf("expected 1 overdue, got %d", stats.Overdue)
	}
	if stats.DueToday != 1 {
		t.Errorf("expected 1 due today (also overdue counts as due), got %d", stats.DueToday)
	}
}

func TestGetStats_ArchivedExcluded(t *testing.T) {
	setupTestDB(t)

	AddTopic("Visible")
	AddTopic("Hidden")

	topics, _ := GetAllTopics()
	for _, tp := range topics {
		if tp.Topic == "Hidden" {
			ArchiveTopic(tp.ID)
		}
	}

	stats, _ := GetStats()
	// Archived topics should not be included in total
	if stats.Total != 1 {
		t.Errorf("expected total 1 (archived excluded), got %d", stats.Total)
	}
}
