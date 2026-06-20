package database

import "time"

// Stats holds aggregate counts for the dashboard.
type Stats struct {
	Total      int
	Completed  int
	InProgress int
	Parked     int
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

	// Parked (non-archived)
	err = db.QueryRow(`SELECT COUNT(*) FROM topics WHERE parked = true AND archived = false`).Scan(&s.Parked)
	if err != nil {
		return s, err
	}

	s.InProgress = s.Total - s.Completed - s.Parked

	now := nowFn()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// DueToday: all items needing review today or earlier (includes overdue)
	err = db.QueryRow(
		`SELECT COUNT(*) FROM topics WHERE next_review_at <= ? AND completed = false AND archived = false AND parked = false`,
		endOfDay,
	).Scan(&s.DueToday)
	if err != nil {
		return s, err
	}

	// Overdue: strictly past due (before start of today)
	err = db.QueryRow(
		`SELECT COUNT(*) FROM topics WHERE next_review_at < ? AND completed = false AND archived = false AND parked = false`,
		startOfDay,
	).Scan(&s.Overdue)
	if err != nil {
		return s, err
	}

	// Project count
	err = db.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&s.Projects)
	if err != nil {
		return s, err
	}

	return s, nil
}
