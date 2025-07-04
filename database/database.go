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
	_, err := db.Exec("INSERT INTO topics (topic, created_at, next_review_at, review_cycle, completed, archived) VALUES (?, ?, ?, ?, ?, ?)",
		topic, time.Now(), time.Now().AddDate(0, 0, 3), 1, false, false)
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
	err := db.QueryRow("SELECT review_cycle FROM topics WHERE id = ?", id).Scan(&reviewCycle)
	if err != nil {
		return err
	}

	var nextReviewDate time.Time
	var completed bool
	switch reviewCycle.Int64 {
	case 1:
		nextReviewDate = time.Now().AddDate(0, 0, 8)
	case 2:
		nextReviewDate = time.Now().AddDate(0, 0, 15)
	case 3:
		nextReviewDate = time.Now().AddDate(0, 0, 30)
	case 4:
		completed = true
	default:
		return fmt.Errorf("invalid review cycle: %d", reviewCycle.Int64)
	}

	if completed {
		_, err = db.Exec("UPDATE topics SET completed = true WHERE id = ?", id)
	} else {
		_, err = db.Exec("UPDATE topics SET next_review_at = ?, review_cycle = ? WHERE id = ?", nextReviewDate, reviewCycle.Int64+1, id)
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

func ModifyTopic(id int64, newTopic string) error {
	_, err := db.Exec("UPDATE topics SET topic = ? WHERE id = ?", newTopic, id)
	return err
}

func DeleteTopic(id int64) error {
	_, err := db.Exec("DELETE FROM topics WHERE id = ?", id)
	return err
}

func GetReviewDay(cycle int64) int {
	switch cycle {
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
