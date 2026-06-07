package database

import (
	"testing"
	"time"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	InitDBWithPath(":memory:")
	t.Cleanup(func() {
		if db != nil {
			db.Close()
			db = nil
		}
	})
}

// ── Topic CRUD ────────────────────────────────────────────────────────────────

func TestAddAndGetAllTopics(t *testing.T) {
	setupTestDB(t)

	if err := AddTopic("Learn Go"); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	topics, err := GetAllTopics()
	if err != nil {
		t.Fatalf("GetAllTopics: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}

	got := topics[0]
	if got.Topic != "Learn Go" {
		t.Errorf("topic text: expected %q, got %q", "Learn Go", got.Topic)
	}
	if got.ReviewCycle != 0 {
		t.Errorf("review cycle: expected 0, got %d", got.ReviewCycle)
	}
	if got.Completed {
		t.Error("new topic should not be completed")
	}
	if got.Archived {
		t.Error("new topic should not be archived")
	}
}

func TestGetTopicsToReview_NewTopicIsDue(t *testing.T) {
	setupTestDB(t)

	if err := AddTopic("Due Now"); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	topics, err := GetTopicsToReview()
	if err != nil {
		t.Fatalf("GetTopicsToReview: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic to review, got %d", len(topics))
	}
	if topics[0].Topic != "Due Now" {
		t.Errorf("expected %q, got %q", "Due Now", topics[0].Topic)
	}
}

func TestGetTopicsToReview_CompletedNotIncluded(t *testing.T) {
	setupTestDB(t)

	AddTopic("Complete Me")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	for i := 0; i < 5; i++ {
		MarkTopicDone(id) //nolint:errcheck
	}

	due, err := GetTopicsToReview()
	if err != nil {
		t.Fatalf("GetTopicsToReview: %v", err)
	}
	if len(due) != 0 {
		t.Errorf("completed topic should not appear in review list, got %d", len(due))
	}
}

func TestMarkTopicDone_ProgressesThroughAllCycles(t *testing.T) {
	setupTestDB(t)

	AddTopic("Cycle Test")
	topics, _ := GetAllTopics()
	id := topics[0].ID
	createdAt := topics[0].CreatedAt

	cases := []struct {
		wantCycle       int64
		wantDaysFromCreated int
	}{
		{1, 2},
		{2, 7},
		{3, 14},
		{4, 29},
	}

	for _, c := range cases {
		if _, err := MarkTopicDone(id); err != nil {
			t.Fatalf("MarkTopicDone → cycle %d: %v", c.wantCycle, err)
		}
		topics, _ = GetAllTopics()
		got := topics[0]

		if got.ReviewCycle != c.wantCycle {
			t.Errorf("review cycle: expected %d, got %d", c.wantCycle, got.ReviewCycle)
		}

		wantDate := createdAt.AddDate(0, 0, c.wantDaysFromCreated)
		diff := got.NextReviewAt.Sub(wantDate)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Errorf("next_review_at for cycle %d: expected %v, got %v (diff %v)",
				c.wantCycle, wantDate, got.NextReviewAt, diff)
		}
	}

	// Fifth call marks as completed
	if _, err := MarkTopicDone(id); err != nil {
		t.Fatalf("final MarkTopicDone: %v", err)
	}
	topics, _ = GetAllTopics()
	if !topics[0].Completed {
		t.Error("expected topic to be completed after all cycles")
	}
}

func TestArchiveAndUnarchiveTopic(t *testing.T) {
	setupTestDB(t)

	AddTopic("Archive Test")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := ArchiveTopic(id); err != nil {
		t.Fatalf("ArchiveTopic: %v", err)
	}
	_, archived, err := GetTopicStatus(id)
	if err != nil {
		t.Fatalf("GetTopicStatus after archive: %v", err)
	}
	if !archived {
		t.Error("expected topic to be archived")
	}

	if err := UnarchiveTopic(id); err != nil {
		t.Fatalf("UnarchiveTopic: %v", err)
	}
	_, archived, err = GetTopicStatus(id)
	if err != nil {
		t.Fatalf("GetTopicStatus after unarchive: %v", err)
	}
	if archived {
		t.Error("expected topic to be unarchived")
	}
}

func TestModifyTopic(t *testing.T) {
	setupTestDB(t)

	AddTopic("Original Text")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := ModifyTopic(id, "Modified Text"); err != nil {
		t.Fatalf("ModifyTopic: %v", err)
	}

	topics, _ = GetAllTopics()
	if topics[0].Topic != "Modified Text" {
		t.Errorf("expected %q, got %q", "Modified Text", topics[0].Topic)
	}
}

func TestDeleteTopic(t *testing.T) {
	setupTestDB(t)

	AddTopic("Delete Me")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := DeleteTopic(id); err != nil {
		t.Fatalf("DeleteTopic: %v", err)
	}

	topics, _ = GetAllTopics()
	if len(topics) != 0 {
		t.Errorf("expected 0 topics after delete, got %d", len(topics))
	}
}

func TestUpdateTopicReviewCycle_Valid(t *testing.T) {
	setupTestDB(t)

	AddTopic("Cycle Override")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := UpdateTopicReviewCycle(id, 3); err != nil {
		t.Fatalf("UpdateTopicReviewCycle(3): %v", err)
	}

	topics, _ = GetAllTopics()
	if topics[0].ReviewCycle != 3 {
		t.Errorf("expected cycle 3, got %d", topics[0].ReviewCycle)
	}
}

func TestUpdateTopicReviewCycle_Invalid(t *testing.T) {
	setupTestDB(t)

	AddTopic("Bad Cycle")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := UpdateTopicReviewCycle(id, 99); err == nil {
		t.Error("expected error for invalid review cycle 99")
	}
	if err := UpdateTopicReviewCycle(id, -1); err == nil {
		t.Error("expected error for invalid review cycle -1")
	}
}

// ── GetReviewDay ──────────────────────────────────────────────────────────────

func TestGetReviewDay(t *testing.T) {
	cases := []struct {
		cycle int64
		want  int
	}{
		{0, 1},
		{1, 3},
		{2, 8},
		{3, 15},
		{4, 30},
	}
	for _, c := range cases {
		if got := GetReviewDay(c.cycle); got != c.want {
			t.Errorf("GetReviewDay(%d): expected %d, got %d", c.cycle, c.want, got)
		}
	}
}

// ── DB path resolution ────────────────────────────────────────────────────────

func TestResolveDBPath_ContainsSpaced(t *testing.T) {
	path := resolveDBPath()
	if path == "" {
		t.Fatal("resolveDBPath returned empty string")
	}
	// Should contain ".spaced" directory in the path
	if !containsSubstring(path, ".spaced") {
		t.Errorf("expected path to contain '.spaced', got %q", path)
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsAt(s, sub))
}

func containsAt(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
