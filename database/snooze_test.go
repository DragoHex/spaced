package database

import (
	"testing"
	"time"
)

func TestSnoozeTopic_PushesReviewDate(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return base }
	defer func() { nowFn = time.Now }()

	AddTopic("Snooze Me")
	topics, _ := GetAllTopics()
	id := topics[0].ID
	originalNext := topics[0].NextReviewAt

	if err := SnoozeTopic(id, 3); err != nil {
		t.Fatalf("SnoozeTopic: %v", err)
	}

	topics, _ = GetAllTopics()
	newNext := topics[0].NextReviewAt

	wantNext := originalNext.AddDate(0, 0, 3)
	diff := newNext.Sub(wantNext)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("expected next_review_at %v, got %v", wantNext, newNext)
	}
}

func TestSnoozeTopic_DoesNotAdvanceCycle(t *testing.T) {
	setupTestDB(t)

	AddTopic("Cycle Stable")
	topics, _ := GetAllTopics()
	id := topics[0].ID
	originalCycle := topics[0].ReviewCycle

	SnoozeTopic(id, 5)

	topics, _ = GetAllTopics()
	if topics[0].ReviewCycle != originalCycle {
		t.Errorf("review cycle should not change: expected %d, got %d",
			originalCycle, topics[0].ReviewCycle)
	}
}

func TestSnoozeTopic_ZeroDaysReturnsError(t *testing.T) {
	setupTestDB(t)

	AddTopic("Zero Snooze")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := SnoozeTopic(id, 0); err == nil {
		t.Error("expected error for snooze of 0 days")
	}
}

func TestSnoozeTopic_NegativeDaysReturnsError(t *testing.T) {
	setupTestDB(t)

	AddTopic("Negative Snooze")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if err := SnoozeTopic(id, -2); err == nil {
		t.Error("expected error for snooze of negative days")
	}
}
