package database

// ParkTopic removes a topic from the revision cycle without losing its cycle progress.
func ParkTopic(id int64) error {
	_, err := db.Exec(`UPDATE topics SET parked = true WHERE id = ?`, id)
	return err
}

// OnboardTopic returns a parked topic to the revision cycle, scheduling it as immediately due.
// If cycle is non-nil the topic's review_cycle is set to that value first; otherwise the
// existing review_cycle is preserved. completed is always reset to false.
func OnboardTopic(id int64, cycle *int64) error {
	now := nowFn()
	if cycle != nil {
		_, err := db.Exec(
			`UPDATE topics SET parked = false, completed = false, review_cycle = ?, next_review_at = ? WHERE id = ?`,
			*cycle, now, id,
		)
		return err
	}
	_, err := db.Exec(
		`UPDATE topics SET parked = false, completed = false, next_review_at = ? WHERE id = ?`,
		now, id,
	)
	return err
}
