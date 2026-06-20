package database

import (
	"database/sql"
	"fmt"
)

// Project represents a grouping of related topics.
type Project struct {
	ID          int64
	Name        string
	Description string
}

// ── Schema ────────────────────────────────────────────────────────────────────

func createProjectsTable() {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS projects (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL DEFAULT ''
	);`)
	if err != nil {
		panic(fmt.Sprintf("create projects table: %v", err))
	}

	// Idempotent migrations for production DBs created before this version.
	db.Exec(`ALTER TABLE topics ADD COLUMN project_id INTEGER REFERENCES projects(id)`)
	db.Exec(`ALTER TABLE projects ADD COLUMN description TEXT NOT NULL DEFAULT ''`)
}

// ── Project CRUD ──────────────────────────────────────────────────────────────

// AddProject creates a new project with an empty description.
// Returns an error if the name already exists.
func AddProject(name string) error {
	_, err := db.Exec(`INSERT INTO projects (name, description) VALUES (?, '')`, name)
	return err
}

// AddProjectWithDescription creates a new project with the given description.
func AddProjectWithDescription(name, description string) error {
	_, err := db.Exec(`INSERT INTO projects (name, description) VALUES (?, ?)`, name, description)
	return err
}

// GetAllProjects returns all projects ordered by name.
func GetAllProjects() ([]Project, error) {
	rows, err := db.Query(`SELECT id, name, description FROM projects ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

// GetOrCreateProject returns the id of the project with the given name,
// creating it (with an empty description) if it doesn't exist.
func GetOrCreateProject(name string) (int64, error) {
	var id int64
	err := db.QueryRow(`SELECT id FROM projects WHERE name = ?`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	result, err := db.Exec(`INSERT INTO projects (name, description) VALUES (?, '')`, name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetProjectByID returns the project with the given id.
func GetProjectByID(id int64) (Project, error) {
	var p Project
	err := db.QueryRow(`SELECT id, name, description FROM projects WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.Description)
	return p, err
}

// RenameProject changes the name of an existing project.
func RenameProject(id int64, newName string) error {
	_, err := db.Exec(`UPDATE projects SET name = ? WHERE id = ?`, newName, id)
	return err
}

// UpdateProjectDescription sets the description for an existing project.
func UpdateProjectDescription(id int64, description string) error {
	_, err := db.Exec(`UPDATE projects SET description = ? WHERE id = ?`, description, id)
	return err
}

// DeleteProject removes a project and sets project_id to NULL on its topics.
func DeleteProject(id int64) error {
	if _, err := db.Exec(`UPDATE topics SET project_id = NULL WHERE project_id = ?`, id); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}

// ── Topic ↔ Project ───────────────────────────────────────────────────────────

// AddTopicWithProject inserts a topic and associates it with a project.
func AddTopicWithProject(topic string, projectID int64) (int64, error) {
	now := nowFn()
	res, err := db.Exec(
		`INSERT INTO topics (topic, created_at, next_review_at, review_cycle, completed, archived, project_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		topic, now, now, 0, false, false, projectID,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// AssignTopicToProject updates a topic's project association.
func AssignTopicToProject(topicID, projectID int64) error {
	_, err := db.Exec(`UPDATE topics SET project_id = ? WHERE id = ?`, projectID, topicID)
	return err
}

// GetTopicsByProject returns all topics belonging to the given project.
func GetTopicsByProject(projectID int64) ([]Topic, error) {
	rows, err := db.Query(`
		SELECT t.id, t.topic, t.notes, t.created_at, t.next_review_at, t.review_cycle,
		       t.completed, t.archived, t.parked, t.easiness_factor, t.interval_days, t.project_id, p.name
		FROM topics t
		LEFT JOIN projects p ON t.project_id = p.id
		WHERE t.project_id = ?
		ORDER BY t.completed ASC, t.created_at DESC`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTopics(rows)
}
