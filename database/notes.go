package database

import "time"

// AddTopicWithNotes inserts a topic with an initial notes string.
func AddTopicWithNotes(topic, notes string) (int64, error) {
	now := nowFn()
	res, err := db.Exec(
		`INSERT INTO topics (topic, notes, created_at, next_review_at, review_cycle, completed, archived)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		topic, notes, now, now, 0, false, false,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// AddTopicFull inserts a topic with notes and a project association.
func AddTopicFull(topic, notes string, projectID int64) (int64, error) {
	now := nowFn()
	return addTopicFullWithDate(topic, notes, projectID, now)
}

// AddTopicFullWithDate inserts a topic with an explicit initial review date.
func AddTopicFullWithDate(topic, notes string, projectID int64, reviewAt time.Time) (int64, error) {
	return addTopicFullWithDate(topic, notes, projectID, reviewAt)
}

func addTopicFullWithDate(topic, notes string, projectID int64, reviewAt time.Time) (int64, error) {
	now := nowFn()
	res, err := db.Exec(
		`INSERT INTO topics (topic, notes, created_at, next_review_at, review_cycle, completed, archived, project_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		topic, notes, now, reviewAt, 0, false, false, projectID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateNotes replaces the notes field for a topic.
func UpdateNotes(id int64, notes string) error {
	_, err := db.Exec(`UPDATE topics SET notes = ? WHERE id = ?`, notes, id)
	return err
}
