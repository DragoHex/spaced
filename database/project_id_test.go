package database

import "testing"

func TestGetProjectByID_Found(t *testing.T) {
	setupTestDB(t)

	AddProjectWithDescription("DSA", "Data Structures & Algorithms")
	projects, _ := GetAllProjects()
	id := projects[0].ID

	p, err := GetProjectByID(id)
	if err != nil {
		t.Fatalf("GetProjectByID: %v", err)
	}
	if p.Name != "DSA" {
		t.Errorf("expected name 'DSA', got %q", p.Name)
	}
	if p.Description != "Data Structures & Algorithms" {
		t.Errorf("expected description, got %q", p.Description)
	}
}

func TestGetProjectByID_NotFound(t *testing.T) {
	setupTestDB(t)

	_, err := GetProjectByID(999)
	if err == nil {
		t.Error("expected error for non-existent project ID")
	}
}
