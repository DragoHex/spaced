package database

// Stats holds aggregate counts for the dashboard.
type Stats struct {
	Total      int
	Completed  int
	InProgress int
	Overdue    int
	DueToday   int
	Archived   int
	Projects   int
}

// GetStats returns aggregate counts across all topics.
func GetStats() (Stats, error) {
	var s Stats

	// Total active (non-archived) topics
	err := db.QueryRow(`SELECT COUNT(*) FROM topics WHERE archived = false`).Scan(&s.Total)
	if err != nil {
		return s, err
	}

	// Archived
	err = db.QueryRow(`SELECT COUNT(*) FROM topics WHERE archived = true`).Scan(&s.Archived)
	if err != nil {
		return s, err
	}

	// Completed (non-archived)
	err = db.QueryRow(`SELECT COUNT(*) FROM topics WHERE completed = true AND archived = false`).Scan(&s.Completed)
	if err != nil {
		return s, err
	}

	s.InProgress = s.Total - s.Completed

	// Due today (next_review_at <= now, not completed, not archived)
	now := nowFn()
	err = db.QueryRow(
		`SELECT COUNT(*) FROM topics WHERE next_review_at <= ? AND completed = false AND archived = false`,
		now,
	).Scan(&s.DueToday)
	if err != nil {
		return s, err
	}

	// Overdue means past next_review_at and not completed (same as due today in practice,
	// but we surface it as a distinct label so callers can decide how to present it).
	s.Overdue = s.DueToday

	// Project count
	err = db.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&s.Projects)
	if err != nil {
		return s, err
	}

	return s, nil
}
