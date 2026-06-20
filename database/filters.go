package database

import (
	"fmt"
	"strings"
)

// TopicFilter controls which topics GetTopicsFiltered returns.
type TopicFilter struct {
	ProjectID       *int64 // if non-nil, restrict to this project
	Overdue         bool   // only topics past their next_review_at
	Completed       bool   // only completed topics
	IncludeArchived bool   // also include archived topics (default: exclude)
	ArchivedOnly    bool   // only archived topics
	ParkedOnly      bool   // only parked topics
}

// GetTopicsFiltered returns topics matching the given filter.
// Without any flags set it behaves like GetAllTopics minus archived rows.
func GetTopicsFiltered(f TopicFilter) ([]Topic, error) {
	var conds []string
	var args []interface{}

	if f.ProjectID != nil {
		conds = append(conds, "t.project_id = ?")
		args = append(args, *f.ProjectID)
	}
	if f.Overdue {
		conds = append(conds, "t.next_review_at <= ? AND t.completed = false")
		args = append(args, nowFn())
	}
	if f.Completed {
		conds = append(conds, "t.completed = true")
	}
	if f.ArchivedOnly {
		conds = append(conds, "t.archived = true")
	} else if !f.IncludeArchived {
		conds = append(conds, "t.archived = false")
	}
	if f.ParkedOnly {
		conds = append(conds, "t.parked = true")
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.topic, t.notes, t.created_at, t.next_review_at, t.review_cycle,
		       t.completed, t.archived, t.parked, t.easiness_factor, t.interval_days, t.project_id, p.name
		FROM topics t
		LEFT JOIN projects p ON t.project_id = p.id
		%s
		ORDER BY t.completed ASC, t.next_review_at DESC`, where)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTopics(rows)
}
