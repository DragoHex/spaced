package database

import (
	"fmt"
	"strings"
	"time"
)

// TopicGroup holds topics belonging to one project (or no project).
type TopicGroup struct {
	ProjectName        string
	ProjectDescription string
	Topics             []Topic
}

// GetTopicsGroupedByProject returns ALL topics (including archived/completed)
// grouped by project. Topics without a project are in a group with ProjectName "".
func GetTopicsGroupedByProject() ([]TopicGroup, error) {
	// Build a description lookup by project name.
	descByName := map[string]string{}
	if projects, err := GetAllProjects(); err == nil {
		for _, p := range projects {
			descByName[p.Name] = p.Description
		}
	}

	rows, err := db.Query(`
		SELECT t.id, t.topic, t.notes, t.created_at, t.next_review_at, t.review_cycle,
		       t.completed, t.archived, t.easiness_factor, t.interval_days, t.project_id, p.name
		FROM topics t
		LEFT JOIN projects p ON t.project_id = p.id
		ORDER BY COALESCE(p.name, ''), t.created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	topics, err := scanTopics(rows)
	if err != nil {
		return nil, err
	}

	// Group topics preserving insertion order.
	groupMap := make(map[string]*TopicGroup)
	var order []string
	for _, t := range topics {
		key := t.ProjectName
		if _, ok := groupMap[key]; !ok {
			groupMap[key] = &TopicGroup{
				ProjectName:        key,
				ProjectDescription: descByName[key],
			}
			order = append(order, key)
		}
		groupMap[key].Topics = append(groupMap[key].Topics, t)
	}

	groups := make([]TopicGroup, 0, len(order))
	for _, key := range order {
		groups = append(groups, *groupMap[key])
	}
	return groups, nil
}

// RenderMarkdown produces a Markdown export of the topic groups.
func RenderMarkdown(groups []TopicGroup) string {
	var b strings.Builder
	b.WriteString("# Spaced Repetition Export\n\n")
	b.WriteString(fmt.Sprintf("_Generated: %s_\n\n", time.Now().Format("2006-01-02")))

	for _, g := range groups {
		heading := g.ProjectName
		if heading == "" {
			heading = "Unassigned"
		}
		b.WriteString(fmt.Sprintf("## %s\n\n", heading))
		if g.ProjectDescription != "" {
			b.WriteString(fmt.Sprintf("_%s_\n\n", g.ProjectDescription))
		}
		b.WriteString("| ID | Topic | Cycle | Next Review | Status | Notes |\n")
		b.WriteString("|----|-------|-------|-------------|--------|-------|\n")

		for _, t := range g.Topics {
			status := "In Progress"
			if t.Completed {
				status = "Completed"
			} else if t.Archived {
				status = "Archived"
			}
			notes := strings.ReplaceAll(t.Notes, "|", "\\|")
			b.WriteString(fmt.Sprintf("| %d | %s | Day %d | %s | %s | %s |\n",
				t.ID,
				t.Topic,
				GetReviewDay(t.ReviewCycle),
				t.NextReviewAt.Format("2006-01-02"),
				status,
				notes,
			))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// RenderCSV produces a CSV export of the topic groups.
func RenderCSV(groups []TopicGroup) string {
	var b strings.Builder
	b.WriteString("ID,Topic,Project,Project Description,Notes,Cycle,Created,Next Review,Status\n")

	for _, g := range groups {
		for _, t := range g.Topics {
			status := "In Progress"
			if t.Completed {
				status = "Completed"
			} else if t.Archived {
				status = "Archived"
			}
			b.WriteString(fmt.Sprintf("%d,%s,%s,%s,%s,Day %d,%s,%s,%s\n",
				t.ID,
				csvEscape(t.Topic),
				csvEscape(g.ProjectName),
				csvEscape(g.ProjectDescription),
				csvEscape(t.Notes),
				GetReviewDay(t.ReviewCycle),
				t.CreatedAt.Format("2006-01-02"),
				t.NextReviewAt.Format("2006-01-02"),
				status,
			))
		}
	}
	return b.String()
}

// csvEscape wraps a value in quotes if it contains commas, quotes, or newlines.
func csvEscape(s string) string {
	if strings.ContainsAny(s, ",\"\n\r") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}
