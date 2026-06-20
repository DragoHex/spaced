package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Topic represents a spaced-repetition topic row.
type Topic struct {
	ID              int64
	Topic           string
	Notes           string
	CreatedAt       time.Time
	NextReviewAt    time.Time
	ReviewCycle     int64
	Completed       bool
	Archived        bool
	Parked          bool
	EasinessFactor  float64
	IntervalDays    int64
	ProjectID       *int64
	ProjectName     string
}

var db *sql.DB

// cycleDays maps review cycle → absolute days from createdAt for next_review_at.
// Cycle 0 = day 0 (immediate), then 1, 4, 11, 25, 55, 115.
var cycleDays = []int{0, 1, 4, 11, 25, 55, 115}

// GetReviewDay returns the day-from-creation label for a given review cycle.
// Cycle 3 → "Day 11", cycle 4 → "Day 25", etc.
func GetReviewDay(cycle int64) int {
	if cycle < 0 || int(cycle) >= len(cycleDays) {
		return 0
	}
	return cycleDays[cycle]
}

func resolveDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "spaced.db"
	}
	dir := filepath.Join(home, ".spaced")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "spaced.db"
	}
	return filepath.Join(dir, "spaced.db")
}

// InitDB opens the production DB at ~/.spaced/spaced.db.
func InitDB() {
	InitDBWithPath(resolveDBPath())
}

// InitDBWithPath opens a DB at the given path. Pass ":memory:" in tests.
func InitDBWithPath(path string) {
	if db != nil {
		db.Close()
	}
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	createTables()
}

func createTables() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS topics (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		topic            TEXT NOT NULL,
		notes            TEXT NOT NULL DEFAULT '',
		created_at       TIMESTAMP NOT NULL,
		next_review_at   TIMESTAMP NOT NULL,
		review_cycle     INT NOT NULL,
		completed        BOOLEAN NOT NULL,
		archived         BOOLEAN NOT NULL,
		parked           BOOLEAN NOT NULL DEFAULT 0,
		easiness_factor  REAL NOT NULL DEFAULT 2.5,
		interval_days    INT NOT NULL DEFAULT 0,
		project_id       INTEGER REFERENCES projects(id)
	);`)
	if err != nil {
		log.Fatal(err)
	}
	createProjectsTable()
	createReviewLogsTable()
	migrateParkedColumn()
}

func migrateParkedColumn() {
	_, err := db.Exec(`ALTER TABLE topics ADD COLUMN parked BOOLEAN NOT NULL DEFAULT 0`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		log.Fatal("migrate parked column:", err)
	}
}

// nowFn can be swapped in tests for deterministic time.
var nowFn = func() time.Time { return time.Now() }

// scanTopics reads a "topics LEFT JOIN projects" result set into []Topic.
// Expected column order: id, topic, notes, created_at, next_review_at,
// review_cycle, completed, archived, parked, easiness_factor, interval_days,
// project_id, project_name.
func scanTopics(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]Topic, error) {
	var topics []Topic
	for rows.Next() {
		var t Topic
		var projectID sql.NullInt64
		var projectName sql.NullString
		if err := rows.Scan(
			&t.ID, &t.Topic, &t.Notes, &t.CreatedAt, &t.NextReviewAt,
			&t.ReviewCycle, &t.Completed, &t.Archived, &t.Parked,
			&t.EasinessFactor, &t.IntervalDays,
			&projectID, &projectName,
		); err != nil {
			return nil, err
		}
		if projectID.Valid {
			v := projectID.Int64
			t.ProjectID = &v
		}
		t.ProjectName = projectName.String
		topics = append(topics, t)
	}
	return topics, rows.Err()
}

// ── Topic operations ──────────────────────────────────────────────────────────

func AddTopic(topic string) error {
	now := nowFn()
	_, err := db.Exec(
		`INSERT INTO topics (topic, created_at, next_review_at, review_cycle, completed, archived)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		topic, now, now, 0, false, false,
	)
	return err
}

func GetTopicsToReview() ([]Topic, error) {
	rows, err := db.Query(`
		SELECT t.id, t.topic, t.notes, t.created_at, t.next_review_at, t.review_cycle,
		       t.completed, t.archived, t.parked, t.easiness_factor, t.interval_days, t.project_id, p.name
		FROM topics t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.next_review_at <= ? AND t.completed = false AND t.archived = false AND t.parked = false`,
		nowFn(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTopics(rows)
}

func MarkTopicDone(id int64) (nextReviewAt time.Time, err error) {
	var cycle int64
	var createdAt time.Time
	err = db.QueryRow(`SELECT review_cycle, created_at FROM topics WHERE id = ?`, id).
		Scan(&cycle, &createdAt)
	if err != nil {
		return
	}

	if cycle == 6 {
		_, err = db.Exec(`UPDATE topics SET completed = true WHERE id = ?`, id)
		return
	}
	if cycle < 0 || cycle > 6 {
		err = fmt.Errorf("invalid review cycle: %d", cycle)
		return
	}

	// Schedule on the absolute day from creation. If that date is already past
	// (overdue review), fall back to tomorrow so we never schedule into the past.
	newCycle := cycle + 1
	nextReviewAt = createdAt.AddDate(0, 0, cycleDays[newCycle])
	if tomorrow := nowFn().AddDate(0, 0, 1); nextReviewAt.Before(tomorrow) {
		nextReviewAt = tomorrow
	}
	_, err = db.Exec(
		`UPDATE topics SET next_review_at = ?, review_cycle = ? WHERE id = ?`,
		nextReviewAt, newCycle, id,
	)
	return
}

func GetAllTopics() ([]Topic, error) {
	rows, err := db.Query(`
		SELECT t.id, t.topic, t.notes, t.created_at, t.next_review_at, t.review_cycle,
		       t.completed, t.archived, t.parked, t.easiness_factor, t.interval_days, t.project_id, p.name
		FROM topics t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.archived = false
		ORDER BY t.completed ASC, t.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTopics(rows)
}

func ArchiveTopic(id int64) error {
	_, err := db.Exec(`UPDATE topics SET archived = true WHERE id = ?`, id)
	return err
}

func UnarchiveTopic(id int64) error {
	_, err := db.Exec(`UPDATE topics SET archived = false WHERE id = ?`, id)
	return err
}

func ModifyTopic(id int64, newTopic string) error {
	_, err := db.Exec(`UPDATE topics SET topic = ? WHERE id = ?`, newTopic, id)
	return err
}

func UpdateTopicReviewCycle(id int64, newCycle int64) error {
	if newCycle < 0 || newCycle > 6 {
		return fmt.Errorf("invalid review cycle %d: must be 0–6", newCycle)
	}
	var createdAt time.Time
	if err := db.QueryRow(`SELECT created_at FROM topics WHERE id = ?`, id).
		Scan(&createdAt); err != nil {
		return err
	}
	nextDate := createdAt.AddDate(0, 0, cycleDays[newCycle])
	_, err := db.Exec(
		`UPDATE topics SET next_review_at = ?, review_cycle = ? WHERE id = ?`,
		nextDate, newCycle, id,
	)
	return err
}

func DeleteTopic(id int64) error {
	_, err := db.Exec(`DELETE FROM topics WHERE id = ?`, id)
	return err
}

func GetTopicStatus(id int64) (completed bool, archived bool, err error) {
	err = db.QueryRow(`SELECT completed, archived FROM topics WHERE id = ?`, id).
		Scan(&completed, &archived)
	return
}
