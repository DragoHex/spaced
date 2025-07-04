package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

const dbName = "spaced.db"

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", dbName)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	createTopicsTable := `
	CREATE TABLE IF NOT EXISTS topics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		topic TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL,
		next_review_at TIMESTAMP NOT NULL,
		review_cycle INT NOT NULL,
		completed BOOLEAN NOT NULL,
		archived BOOLEAN NOT NULL
	);`

	_, err := db.Exec(createTopicsTable)
	if err != nil {
		log.Fatal(err)
	}
}

func AddTopic(topic string) error {
	now := time.Now()
	_, err := db.Exec("INSERT INTO topics (topic, created_at, next_review_at, review_cycle, completed, archived) VALUES (?, ?, ?, ?, ?, ?)",
		topic, now, now, 0, false, false)
	return err
}

func GetTopicsToReview() ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, topic, review_cycle FROM topics WHERE next_review_at <= ? AND completed = false AND archived = false", time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []map[string]interface{}
	for rows.Next() {
		var id, reviewCycle sql.NullInt64
		var topicStr sql.NullString
		if err := rows.Scan(&id, &topicStr, &reviewCycle); err != nil {
			return nil, err
		}
		topics = append(topics, map[string]interface{}{"id": id.Int64, "topic": topicStr.String, "review_cycle": reviewCycle.Int64})
	}
	return topics, nil
}

func MarkTopicDone(id int64) error {
	var reviewCycle sql.NullInt64
	var createdAt time.Time
	err := db.QueryRow("SELECT review_cycle, created_at FROM topics WHERE id = ?", id).Scan(&reviewCycle, &createdAt)
	if err != nil {
		return err
	}

	var nextReviewDate time.Time
	var completed bool
	var newReviewCycle int64

	switch reviewCycle.Int64 {
	case 0:
		newReviewCycle = 1
		nextReviewDate = createdAt.AddDate(0, 0, 2) // Day 1 + 2 days = Day 3
	case 1:
		newReviewCycle = 2
		nextReviewDate = createdAt.AddDate(0, 0, 7) // Day 3 + 5 days = Day 8 (relative to created_at: Day 1 + 7 days)
	case 2:
		newReviewCycle = 3
		nextReviewDate = createdAt.AddDate(0, 0, 14) // Day 8 + 7 days = Day 15 (relative to created_at: Day 1 + 14 days)
	case 3:
		newReviewCycle = 4
		nextReviewDate = createdAt.AddDate(0, 0, 29) // Day 15 + 14 days = Day 30 (relative to created_at: Day 1 + 29 days)
	case 4:
		completed = true
	default:
		return fmt.Errorf("invalid review cycle: %d", reviewCycle.Int64)
	}

	if completed {
		_, err = db.Exec("UPDATE topics SET completed = true WHERE id = ?", id)
	} else {
		_, err = db.Exec("UPDATE topics SET next_review_at = ?, review_cycle = ? WHERE id = ?", nextReviewDate, newReviewCycle, id)
	}
	return err
}

func GetAllTopics() ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, topic, created_at, next_review_at, review_cycle, completed, archived FROM topics ORDER BY completed ASC, created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topics []map[string]interface{}
	for rows.Next() {
		var id, reviewCycle sql.NullInt64
		var topicStr sql.NullString
		var createdAt, nextReviewAt sql.NullTime
		var completed, archived sql.NullBool

		if err := rows.Scan(&id, &topicStr, &createdAt, &nextReviewAt, &reviewCycle, &completed, &archived); err != nil {
			return nil, err
		}
		topics = append(topics, map[string]interface{}{
			"id":             id.Int64,
			"topic":          topicStr.String,
			"created_at":     createdAt.Time,
			"next_review_at": nextReviewAt.Time,
			"review_cycle":   reviewCycle.Int64,
			"completed":      completed.Bool,
			"archived":       archived.Bool,
		})
	}
	return topics, nil
}

func ArchiveTopic(id int64) error {
	_, err := db.Exec("UPDATE topics SET archived = true WHERE id = ?", id)
	return err
}

func UnarchiveTopic(id int64) error {
	_, err := db.Exec("UPDATE topics SET archived = false WHERE id = ?", id)
	return err
}

func ModifyTopic(id int64, newTopic string) error {
	_, err := db.Exec("UPDATE topics SET topic = ? WHERE id = ?", newTopic, id)
	return err
}

func UpdateTopicReviewCycle(id int64, newReviewCycle int64) error {
	var createdAt time.Time
	err := db.QueryRow("SELECT created_at FROM topics WHERE id = ?", id).Scan(&createdAt)
	if err != nil {
		return err
	}

	var nextReviewDate time.Time
	switch newReviewCycle {
	case 0:
		nextReviewDate = createdAt // Day 1
	case 1:
		nextReviewDate = createdAt.AddDate(0, 0, 2) // Day 3
	case 2:
		nextReviewDate = createdAt.AddDate(0, 0, 7) // Day 8
	case 3:
		nextReviewDate = createdAt.AddDate(0, 0, 14) // Day 15
	case 4:
		nextReviewDate = createdAt.AddDate(0, 0, 29) // Day 30
	default:
		return fmt.Errorf("invalid review cycle: %d. Must be 0, 1, 2, 3, or 4.", newReviewCycle)
	}
	_, err = db.Exec("UPDATE topics SET next_review_at = ?, review_cycle = ? WHERE id = ?", nextReviewDate, newReviewCycle, id)
	return err
}

func DeleteTopic(id int64) error {
	_, err := db.Exec("DELETE FROM topics WHERE id = ?", id)
	return err
}

func GetTopicStatus(id int64) (completed bool, archived bool, err error) {
	err = db.QueryRow("SELECT completed, archived FROM topics WHERE id = ?", id).Scan(&completed, &archived)
	return
}

func GetReviewDay(cycle int64) int {
	switch cycle {
	case 0:
		return 1
	case 1:
		return 3
	case 2:
		return 8
	case 3:
		return 15
	case 4:
		return 30
	default:
		return 0
	}
}
