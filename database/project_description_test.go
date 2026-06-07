package database

import (
	"strings"
	"testing"
)

// ── Project.Description field ─────────────────────────────────────────────────

func TestAddProject_DefaultDescriptionIsEmpty(t *testing.T) {
	setupTestDB(t)

	if err := AddProject("DSA"); err != nil {
		t.Fatalf("AddProject: %v", err)
	}

	projects, err := GetAllProjects()
	if err != nil {
		t.Fatalf("GetAllProjects: %v", err)
	}
	if projects[0].Description != "" {
		t.Errorf("expected empty description, got %q", projects[0].Description)
	}
}

func TestAddProjectWithDescription(t *testing.T) {
	setupTestDB(t)

	desc := "Data Structures and Algorithms for interviews"
	if err := AddProjectWithDescription("DSA", desc); err != nil {
		t.Fatalf("AddProjectWithDescription: %v", err)
	}

	projects, err := GetAllProjects()
	if err != nil {
		t.Fatalf("GetAllProjects: %v", err)
	}
	if projects[0].Description != desc {
		t.Errorf("expected description %q, got %q", desc, projects[0].Description)
	}
}

func TestUpdateProjectDescription(t *testing.T) {
	setupTestDB(t)

	AddProject("Systems")
	projects, _ := GetAllProjects()
	id := projects[0].ID

	newDesc := "Operating systems, networking, distributed systems"
	if err := UpdateProjectDescription(id, newDesc); err != nil {
		t.Fatalf("UpdateProjectDescription: %v", err)
	}

	projects, _ = GetAllProjects()
	if projects[0].Description != newDesc {
		t.Errorf("expected %q, got %q", newDesc, projects[0].Description)
	}
}

func TestGetOrCreateProject_DescriptionEmpty(t *testing.T) {
	setupTestDB(t)

	// GetOrCreateProject creates without a description.
	_, err := GetOrCreateProject("ML")
	if err != nil {
		t.Fatalf("GetOrCreateProject: %v", err)
	}

	projects, _ := GetAllProjects()
	if projects[0].Description != "" {
		t.Errorf("auto-created project should have empty description, got %q", projects[0].Description)
	}
}

// ── Export includes description ───────────────────────────────────────────────

func TestTopicGroup_CarriesProjectDescription(t *testing.T) {
	setupTestDB(t)

	pID, _ := GetOrCreateProject("Go")
	UpdateProjectDescription(pID, "Go programming language topics")
	AddTopicWithProject("Goroutines", pID)

	groups, err := GetTopicsGroupedByProject()
	if err != nil {
		t.Fatalf("GetTopicsGroupedByProject: %v", err)
	}

	goGroup := findGroup(groups, "Go")
	if goGroup == nil {
		t.Fatal("expected Go group")
	}
	if goGroup.ProjectDescription != "Go programming language topics" {
		t.Errorf("expected description in group, got %q", goGroup.ProjectDescription)
	}
}

func TestRenderMarkdown_IncludesProjectDescription(t *testing.T) {
	setupTestDB(t)

	pID, _ := GetOrCreateProject("Rust")
	UpdateProjectDescription(pID, "Systems programming with memory safety")
	AddTopicWithProject("Ownership", pID)

	groups, _ := GetTopicsGroupedByProject()
	out := RenderMarkdown(groups)

	if !strings.Contains(out, "Systems programming with memory safety") {
		t.Errorf("markdown should include project description, got:\n%s", out)
	}
}

func TestRenderMarkdown_NoDescriptionShownWhenEmpty(t *testing.T) {
	setupTestDB(t)

	pID, _ := GetOrCreateProject("Math")
	// No description set.
	AddTopicWithProject("Calculus", pID)

	groups, _ := GetTopicsGroupedByProject()
	out := RenderMarkdown(groups)

	// Should still have the heading but no empty description line.
	if !strings.Contains(out, "## Math") {
		t.Errorf("markdown should have project heading, got:\n%s", out)
	}
}

func TestRenderCSV_IncludesProjectDescription(t *testing.T) {
	setupTestDB(t)

	pID, _ := GetOrCreateProject("Algorithms")
	UpdateProjectDescription(pID, "Core algorithm techniques")
	AddTopicWithProject("Binary Search", pID)

	groups, _ := GetTopicsGroupedByProject()
	out := RenderCSV(groups)

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if !strings.Contains(lines[0], "Project Description") {
		t.Errorf("CSV header should contain 'Project Description', got %q", lines[0])
	}
	if !strings.Contains(out, "Core algorithm techniques") {
		t.Errorf("CSV should contain project description, got:\n%s", out)
	}
}

func TestRenderCSV_UnassignedTopicHasEmptyDescription(t *testing.T) {
	setupTestDB(t)

	AddTopic("No Project")

	groups, _ := GetTopicsGroupedByProject()
	out := RenderCSV(groups)

	if !strings.Contains(out, "No Project") {
		t.Errorf("CSV should contain unassigned topic, got:\n%s", out)
	}
}
