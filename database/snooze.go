package database

import "fmt"

// SnoozeTopic pushes the next_review_at date forward by the given number of days
// without advancing the review cycle.
func SnoozeTopic(id int64, days int) error {
	if days <= 0 {
		return fmt.Errorf("snooze days must be a positive integer, got %d", days)
	}

	var currentNext string
	if err := db.QueryRow(`SELECT next_review_at FROM topics WHERE id = ?`, id).
		Scan(&currentNext); err != nil {
		return err
	}

	_, err := db.Exec(
		`UPDATE topics SET next_review_at = datetime(next_review_at, ? || ' days') WHERE id = ?`,
		days, id,
	)
	return err
}
