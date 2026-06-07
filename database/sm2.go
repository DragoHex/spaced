package database

import (
	"fmt"
	"math"
	"time"
)

const (
	defaultEF = 2.5
	minEF     = 1.3
)

// sm2NextInterval computes the next review interval (in days) and the updated
// easiness factor using the SM-2 algorithm.
//
//   - prevInterval: the previous interval in days (0 = topic never reviewed)
//   - ef:           current easiness factor (default 2.5)
//   - quality:      recall quality 0–5 (0=blackout, 5=perfect)
//
// SM-2 ladder: 0 → 1 → 6 → round(prevInterval * EF) → …
func sm2NextInterval(prevInterval int64, ef float64, quality int) (int64, float64) {
	newEF := ef + (0.1 - float64(5-quality)*(0.08+float64(5-quality)*0.02))
	if newEF < minEF {
		newEF = minEF
	}

	var nextInterval int64
	if quality < 3 {
		nextInterval = 1
	} else {
		switch prevInterval {
		case 0:
			nextInterval = 1
		case 1:
			nextInterval = 6
		default:
			nextInterval = int64(math.Round(float64(prevInterval) * newEF))
		}
	}

	return nextInterval, newEF
}

// MarkTopicDoneWithQuality advances a topic using the SM-2 algorithm.
// quality must be 0–5 (0=total blackout, 5=perfect recall).
// Returns the computed next review date.
func MarkTopicDoneWithQuality(id int64, quality int) (time.Time, error) {
	if quality < 0 || quality > 5 {
		return time.Time{}, fmt.Errorf("quality must be 0–5, got %d", quality)
	}

	var ef float64
	var intervalDays int64
	if err := db.QueryRow(
		`SELECT easiness_factor, interval_days FROM topics WHERE id = ?`, id,
	).Scan(&ef, &intervalDays); err != nil {
		return time.Time{}, err
	}

	nextInterval, newEF := sm2NextInterval(intervalDays, ef, quality)

	reviewDate := nowFn().AddDate(0, 0, int(nextInterval))

	_, err := db.Exec(
		`UPDATE topics
		 SET next_review_at = ?, interval_days = ?, easiness_factor = ?,
		     review_cycle = review_cycle + 1
		 WHERE id = ?`,
		reviewDate, nextInterval, newEF, id,
	)
	return reviewDate, err
}
