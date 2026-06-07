package database

// AddTopicWithNotes inserts a topic with an initial notes string.
func AddTopicWithNotes(topic, notes string) error {
	now := nowFn()
	_, err := db.Exec(
		`INSERT INTO topics (topic, notes, created_at, next_review_at, review_cycle, completed, archived)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		topic, notes, now, now, 0, false, false,
	)
	return err
}

// AddTopicFull inserts a topic with notes and a project association.
func AddTopicFull(topic, notes string, projectID int64) error {
	now := nowFn()
	_, err := db.Exec(
		`INSERT INTO topics (topic, notes, created_at, next_review_at, review_cycle, completed, archived, project_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		topic, notes, now, now, 0, false, false, projectID,
	)
	return err
}

// UpdateNotes replaces the notes field for a topic.
func UpdateNotes(id int64, notes string) error {
	_, err := db.Exec(`UPDATE topics SET notes = ? WHERE id = ?`, notes, id)
	return err
}
