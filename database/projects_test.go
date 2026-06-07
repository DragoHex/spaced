package database

import (
	"testing"
)

// ── Project CRUD ──────────────────────────────────────────────────────────────

func TestAddAndGetAllProjects(t *testing.T) {
	setupTestDB(t)

	if err := AddProject("DSA"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}
	if err := AddProject("Systems"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	projects, err := GetAllProjects()
	if err != nil {
		t.Fatalf("GetAllProjects: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	names := map[string]bool{projects[0].Name: true, projects[1].Name: true}
	if !names["DSA"] || !names["Systems"] {
		t.Errorf("unexpected projects: %+v", projects)
	}
}

func TestAddProject_DuplicateReturnsError(t *testing.T) {
	setupTestDB(t)

	if err := AddProject("DSA"); err != nil {
		t.Fatalf("first AddProject: %v", err)
	}
	if err := AddProject("DSA"); err == nil {
		t.Error("expected error on duplicate project name")
	}
}

func TestGetOrCreateProject(t *testing.T) {
	setupTestDB(t)

	id1, err := GetOrCreateProject("ML")
	if err != nil {
		t.Fatalf("GetOrCreateProject (create): %v", err)
	}
	if id1 == 0 {
		t.Error("expected non-zero id")
	}

	// Second call with same name should return the same id.
	id2, err := GetOrCreateProject("ML")
	if err != nil {
		t.Fatalf("GetOrCreateProject (get): %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected same id %d, got %d", id1, id2)
	}
}

func TestRenameProject(t *testing.T) {
	setupTestDB(t)

	AddProject("Old Name")
	projects, _ := GetAllProjects()
	id := projects[0].ID

	if err := RenameProject(id, "New Name"); err != nil {
		t.Fatalf("RenameProject: %v", err)
	}

	projects, _ = GetAllProjects()
	if projects[0].Name != "New Name" {
		t.Errorf("expected %q, got %q", "New Name", projects[0].Name)
	}
}

func TestDeleteProject(t *testing.T) {
	setupTestDB(t)

	AddProject("Temporary")
	projects, _ := GetAllProjects()
	id := projects[0].ID

	if err := DeleteProject(id); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}

	projects, _ = GetAllProjects()
	if len(projects) != 0 {
		t.Errorf("expected 0 projects, got %d", len(projects))
	}
}

// ── Topic ↔ Project linking ───────────────────────────────────────────────────

func TestAddTopicWithProject(t *testing.T) {
	setupTestDB(t)

	pID, _ := GetOrCreateProject("Algorithms")
	if err := AddTopicWithProject("Binary Search", pID); err != nil {
		t.Fatalf("AddTopicWithProject: %v", err)
	}

	topics, err := GetAllTopics()
	if err != nil {
		t.Fatalf("GetAllTopics: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("expected 1 topic, got %d", len(topics))
	}
	if topics[0].ProjectName != "Algorithms" {
		t.Errorf("expected project %q, got %q", "Algorithms", topics[0].ProjectName)
	}
	if topics[0].ProjectID == nil || *topics[0].ProjectID != pID {
		t.Errorf("expected project id %d, got %v", pID, topics[0].ProjectID)
	}
}

func TestAddTopic_NoProject_HasNilProjectID(t *testing.T) {
	setupTestDB(t)

	if err := AddTopic("No Project Topic"); err != nil {
		t.Fatalf("AddTopic: %v", err)
	}

	topics, _ := GetAllTopics()
	if topics[0].ProjectID != nil {
		t.Errorf("expected nil project id, got %v", topics[0].ProjectID)
	}
	if topics[0].ProjectName != "" {
		t.Errorf("expected empty project name, got %q", topics[0].ProjectName)
	}
}

func TestAssignTopicToProject(t *testing.T) {
	setupTestDB(t)

	AddTopic("Orphan Topic")
	topics, _ := GetAllTopics()
	tID := topics[0].ID

	pID, _ := GetOrCreateProject("Networks")
	if err := AssignTopicToProject(tID, pID); err != nil {
		t.Fatalf("AssignTopicToProject: %v", err)
	}

	topics, _ = GetAllTopics()
	if topics[0].ProjectName != "Networks" {
		t.Errorf("expected project %q, got %q", "Networks", topics[0].ProjectName)
	}
}

func TestDeleteProject_NullsTopicProjectID(t *testing.T) {
	setupTestDB(t)

	pID, _ := GetOrCreateProject("Temp Project")
	AddTopicWithProject("Topic In Project", pID)

	topics, _ := GetAllTopics()
	if topics[0].ProjectID == nil {
		t.Fatal("topic should have a project before delete")
	}

	DeleteProject(pID)

	topics, _ = GetAllTopics()
	if topics[0].ProjectID != nil {
		t.Errorf("topic project_id should be nil after project deleted, got %v", topics[0].ProjectID)
	}
}

func TestGetTopicsByProject(t *testing.T) {
	setupTestDB(t)

	p1, _ := GetOrCreateProject("Math")
	p2, _ := GetOrCreateProject("CS")

	AddTopicWithProject("Calculus", p1)
	AddTopicWithProject("Algebra", p1)
	AddTopicWithProject("Data Structures", p2)
	AddTopic("Unassigned")

	mathTopics, err := GetTopicsByProject(p1)
	if err != nil {
		t.Fatalf("GetTopicsByProject: %v", err)
	}
	if len(mathTopics) != 2 {
		t.Errorf("expected 2 Math topics, got %d", len(mathTopics))
	}

	csTopics, err := GetTopicsByProject(p2)
	if err != nil {
		t.Fatalf("GetTopicsByProject: %v", err)
	}
	if len(csTopics) != 1 {
		t.Errorf("expected 1 CS topic, got %d", len(csTopics))
	}
}
