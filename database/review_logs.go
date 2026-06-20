package database

import "fmt"

// ActivityDay holds per-day counts for the progress chart.
type ActivityDay struct {
	Date     string `json:"date"`
	Added    int    `json:"added"`
	Reviewed int    `json:"reviewed"`
}

func createReviewLogsTable() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS review_logs (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		topic_id   INTEGER NOT NULL,
		reviewed_at TIMESTAMP NOT NULL
	);`)
	if err != nil {
		panic(fmt.Sprintf("create review_logs table: %v", err))
	}
}

// LogReview records a review event for the given topic at the current time.
func LogReview(topicID int64) error {
	_, err := db.Exec(
		`INSERT INTO review_logs (topic_id, reviewed_at) VALUES (?, ?)`,
		topicID, nowFn(),
	)
	return err
}

// GetTopicActivity returns per-day added and reviewed counts for the last N days
// (including today), always returning a contiguous date series with no gaps.
func GetTopicActivity(days int) ([]ActivityDay, error) {
	now := nowFn()

	// Build a complete date range so the chart has no gaps.
	result := make([]ActivityDay, days)
	dateIdx := make(map[string]int, days)
	for i := 0; i < days; i++ {
		d := now.AddDate(0, 0, -(days - 1 - i)).Format("2006-01-02")
		result[i] = ActivityDay{Date: d}
		dateIdx[d] = i
	}

	startDay := now.AddDate(0, 0, -(days - 1)).Format("2006-01-02")

	// Topics added per day.
	rows, err := db.Query(`
		SELECT date(created_at) AS d, COUNT(*)
		FROM topics
		WHERE date(created_at) >= ?
		GROUP BY date(created_at)`, startDay)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var d string
		var cnt int
		if err := rows.Scan(&d, &cnt); err != nil {
			return nil, err
		}
		if idx, ok := dateIdx[d]; ok {
			result[idx].Added = cnt
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Reviews completed per day.
	rows2, err := db.Query(`
		SELECT date(reviewed_at) AS d, COUNT(*)
		FROM review_logs
		WHERE date(reviewed_at) >= ?
		GROUP BY date(reviewed_at)`, startDay)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var d string
		var cnt int
		if err := rows2.Scan(&d, &cnt); err != nil {
			return nil, err
		}
		if idx, ok := dateIdx[d]; ok {
			result[idx].Reviewed = cnt
		}
	}
	return result, rows2.Err()
}
