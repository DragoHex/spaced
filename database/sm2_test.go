package database

import (
	"testing"
	"time"
)

// SM-2 algorithm tests.
// Quality scale: 0-5 where >=3 means correct recall, <3 means reset.
// interval_days starts at 0 (topic never reviewed).

func TestSM2NextInterval_FirstReview_PerfectRecall(t *testing.T) {
	// prevInterval=0 → first ever review → next interval = 1 day
	interval, ef := sm2NextInterval(0, 2.5, 5)
	if interval != 1 {
		t.Errorf("first review interval: expected 1, got %d", interval)
	}
	// EF for q=5: 2.5 + (0.1 - 0*(0.08+0*0.02)) = 2.6
	if ef < 2.59 || ef > 2.61 {
		t.Errorf("first review EF for q=5: expected ~2.6, got %f", ef)
	}
}

func TestSM2NextInterval_SecondReview_PerfectRecall(t *testing.T) {
	// prevInterval=1 → second review → next interval = 6 days
	interval, _ := sm2NextInterval(1, 2.5, 5)
	if interval != 6 {
		t.Errorf("second review interval: expected 6, got %d", interval)
	}
}

func TestSM2NextInterval_ThirdAndBeyond(t *testing.T) {
	// prevInterval=6, EF=2.5, q=4 → EF stays 2.5 → interval = round(6 * 2.5) = 15
	interval, _ := sm2NextInterval(6, 2.5, 4)
	if interval != 15 {
		t.Errorf("third review interval: expected 15, got %d", interval)
	}

	// prevInterval=15, EF=2.5, q=5 → EF → 2.6 → interval = round(15 * 2.6) = 39
	interval2, _ := sm2NextInterval(15, 2.5, 5)
	if interval2 < 38 || interval2 > 40 {
		t.Errorf("fourth review interval: expected ~39, got %d", interval2)
	}
}

func TestSM2NextInterval_LowQuality_Resets(t *testing.T) {
	// quality < 3 resets interval to 1 regardless of prevInterval
	interval, ef := sm2NextInterval(15, 2.5, 2)
	if interval != 1 {
		t.Errorf("low quality should reset interval to 1, got %d", interval)
	}
	// EF should decrease
	if ef >= 2.5 {
		t.Errorf("low quality should decrease EF, got %f", ef)
	}
}

func TestSM2NextInterval_EFFloor(t *testing.T) {
	// EF should not drop below 1.3 even with worst quality
	_, ef := sm2NextInterval(1, 1.31, 0)
	if ef < 1.3 {
		t.Errorf("EF floor: expected >= 1.3, got %f", ef)
	}
}

func TestMarkTopicDoneWithQuality_FirstReview(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return base }
	defer func() { nowFn = time.Now }()

	AddTopic("SM2 Topic")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	nextDate, err := MarkTopicDoneWithQuality(id, 5)
	if err != nil {
		t.Fatalf("MarkTopicDoneWithQuality: %v", err)
	}

	// First interval = 1 day
	expected := base.AddDate(0, 0, 1)
	diff := nextDate.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("first SM2 review: expected next date %v, got %v", expected, nextDate)
	}

	topics, _ = GetAllTopics()
	if topics[0].EasinessFactor < 2.59 {
		t.Errorf("EF should increase after perfect recall, got %f", topics[0].EasinessFactor)
	}
	if topics[0].IntervalDays != 1 {
		t.Errorf("interval_days should be 1 after first review, got %d", topics[0].IntervalDays)
	}
}

func TestMarkTopicDoneWithQuality_SecondReview(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return base }
	defer func() { nowFn = time.Now }()

	AddTopic("SM2 Two")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	MarkTopicDoneWithQuality(id, 5) // first review → interval=1

	nextDate, err := MarkTopicDoneWithQuality(id, 5) // second → interval=6
	if err != nil {
		t.Fatalf("second MarkTopicDoneWithQuality: %v", err)
	}

	expected := base.AddDate(0, 0, 6)
	diff := nextDate.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("second SM2 review: expected %v, got %v", expected, nextDate)
	}
}

func TestMarkTopicDoneWithQuality_LowQuality_ResetsInterval(t *testing.T) {
	setupTestDB(t)

	base := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	nowFn = func() time.Time { return base }
	defer func() { nowFn = time.Now }()

	AddTopic("Hard Topic")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	MarkTopicDoneWithQuality(id, 5) // first: interval=1
	MarkTopicDoneWithQuality(id, 5) // second: interval=6

	// Low quality → reset to 1 day
	nextDate, err := MarkTopicDoneWithQuality(id, 1)
	if err != nil {
		t.Fatalf("MarkTopicDoneWithQuality low quality: %v", err)
	}

	expected := base.AddDate(0, 0, 1)
	diff := nextDate.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > time.Second {
		t.Errorf("low quality: expected next date %v (reset to 1 day), got %v", expected, nextDate)
	}
}

func TestMarkTopicDoneWithQuality_InvalidQuality(t *testing.T) {
	setupTestDB(t)

	AddTopic("Quality Test")
	topics, _ := GetAllTopics()
	id := topics[0].ID

	if _, err := MarkTopicDoneWithQuality(id, -1); err == nil {
		t.Error("expected error for quality -1")
	}
	if _, err := MarkTopicDoneWithQuality(id, 6); err == nil {
		t.Error("expected error for quality 6")
	}
}
