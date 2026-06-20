package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"spaced/database"
)

// ── DTOs ─────────────────────────────────────────────────────────────────────

type topicDTO struct {
	ID             int64     `json:"id"`
	Topic          string    `json:"topic"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"created_at"`
	NextReviewAt   time.Time `json:"next_review_at"`
	ReviewCycle    int64     `json:"review_cycle"`
	ReviewDay      int       `json:"review_day"`
	Completed      bool      `json:"completed"`
	Archived       bool      `json:"archived"`
	Parked         bool      `json:"parked"`
	EasinessFactor float64   `json:"easiness_factor"`
	IntervalDays   int64     `json:"interval_days"`
	ProjectID      *int64    `json:"project_id"`
	ProjectName    string    `json:"project_name"`
}

type projectDTO struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type calendarDay struct {
	Date      string     `json:"date"`
	Pending   int        `json:"pending"`
	Completed int        `json:"completed"`
	Topics    []topicDTO `json:"topics,omitempty"`
}

func toDTO(t database.Topic) topicDTO {
	return topicDTO{
		ID:             t.ID,
		Topic:          t.Topic,
		Notes:          t.Notes,
		CreatedAt:      t.CreatedAt,
		NextReviewAt:   t.NextReviewAt,
		ReviewCycle:    t.ReviewCycle,
		ReviewDay:      database.GetReviewDay(t.ReviewCycle),
		Completed:      t.Completed,
		Archived:       t.Archived,
		Parked:         t.Parked,
		EasinessFactor: t.EasinessFactor,
		IntervalDays:   t.IntervalDays,
		ProjectID:      t.ProjectID,
		ProjectName:    t.ProjectName,
	}
}

func toDTOs(topics []database.Topic) []topicDTO {
	out := make([]topicDTO, 0, len(topics))
	for _, t := range topics {
		out = append(out, toDTO(t))
	}
	return out
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func parseID(segment string) (int64, bool) {
	id, err := strconv.ParseInt(segment, 10, 64)
	return id, err == nil
}

// splitPath splits the URL path after a prefix and returns the remaining segments.
// e.g. "/api/topics/42/done" with prefix "/api/topics/" → ["42","done"]
func splitPath(path, prefix string) []string {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

// ── Stats ─────────────────────────────────────────────────────────────────────

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	stats, err := database.GetStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// ── Topics collection: GET /api/topics, POST /api/topics ─────────────────────

func handleTopics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listTopics(w, r)
	case http.MethodPost:
		createTopic(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func listTopics(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := database.TopicFilter{}

	if pid := q.Get("project_id"); pid != "" {
		id, ok := parseID(pid)
		if !ok {
			writeError(w, http.StatusBadRequest, "invalid project_id")
			return
		}
		f.ProjectID = &id
	}
	f.Overdue = q.Get("overdue") == "true"
	f.Completed = q.Get("completed") == "true"
	if q.Get("archived") == "true" {
		f.ArchivedOnly = true
	} else {
		f.IncludeArchived = q.Get("include_archived") == "true"
	}

	topics, err := database.GetTopicsFiltered(f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toDTOs(topics))
}

func createTopic(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Topic       string `json:"topic"`
		Notes       string `json:"notes"`
		ProjectName string `json:"project_name"`
		Parked      *bool  `json:"parked"`
		ReviewDate  string `json:"review_date"` // optional, YYYY-MM-DD
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Topic) == "" {
		writeError(w, http.StatusBadRequest, "topic is required")
		return
	}

	projectName := body.ProjectName
	if projectName == "" {
		projectName = "UNASSIGNED"
	}
	pid, err := database.GetOrCreateProject(projectName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var newID int64
	if body.ReviewDate != "" {
		reviewAt, err := time.Parse("2006-01-02", body.ReviewDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "review_date must be YYYY-MM-DD")
			return
		}
		newID, err = database.AddTopicFullWithDate(body.Topic, body.Notes, pid, reviewAt)
	} else {
		newID, err = database.AddTopicFull(body.Topic, body.Notes, pid)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// parked defaults to true (UI default); caller must explicitly send false to onboard.
	shouldPark := body.Parked == nil || *body.Parked
	if shouldPark {
		if err := database.ParkTopic(newID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusCreated, map[string]string{"ok": "true"})
}

// ── Topics/review: GET /api/topics/review ────────────────────────────────────

func handleTopicsReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	topics, err := database.GetTopicsToReview()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toDTOs(topics))
}

// ── Individual topic: /api/topics/:id/* ──────────────────────────────────────

func handleTopic(w http.ResponseWriter, r *http.Request) {
	segs := splitPath(r.URL.Path, "/api/topics/")
	if len(segs) < 1 {
		writeError(w, http.StatusBadRequest, "missing topic id")
		return
	}
	id, ok := parseID(segs[0])
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid topic id")
		return
	}

	action := ""
	if len(segs) > 1 {
		action = segs[1]
	}

	switch {
	case r.Method == http.MethodDelete && action == "":
		deleteTopic(w, r, id)
	case r.Method == http.MethodPut && action == "done":
		markDone(w, r, id)
	case r.Method == http.MethodPut && action == "modify":
		modifyTopic(w, r, id)
	case r.Method == http.MethodPut && action == "snooze":
		snoozeTopic(w, r, id)
	case r.Method == http.MethodPut && action == "archive":
		archiveTopic(w, r, id)
	case r.Method == http.MethodPut && action == "unarchive":
		unarchiveTopic(w, r, id)
	case r.Method == http.MethodPut && action == "park":
		parkTopicHandler(w, id)
	case r.Method == http.MethodPut && action == "onboard":
		onboardTopicHandler(w, r, id)
	default:
		writeError(w, http.StatusNotFound, fmt.Sprintf("unknown action: %s %s", r.Method, action))
	}
}

func deleteTopic(w http.ResponseWriter, _ *http.Request, id int64) {
	if err := database.DeleteTopic(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func markDone(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Quality *int `json:"quality"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	var (
		next      time.Time
		completed bool
		err       error
	)
	if body.Quality != nil {
		next, err = database.MarkTopicDoneWithQuality(id, *body.Quality)
	} else {
		next, err = database.MarkTopicDone(id)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = database.LogReview(id)

	completed, _, err = database.GetTopicStatus(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"next_review_at": next,
		"completed":      completed,
	})
}

func modifyTopic(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Topic       *string `json:"topic"`
		Notes       *string `json:"notes"`
		ReviewCycle *int64  `json:"review_cycle"`
		ProjectID   *int64  `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if body.Topic != nil {
		if err := database.ModifyTopic(id, *body.Topic); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.Notes != nil {
		if err := database.UpdateNotes(id, *body.Notes); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.ReviewCycle != nil {
		if err := database.UpdateTopicReviewCycle(id, *body.ReviewCycle); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if body.ProjectID != nil {
		if err := database.AssignTopicToProject(id, *body.ProjectID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func snoozeTopic(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Days int `json:"days"`
	}
	body.Days = 1
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Days < 1 {
		body.Days = 1
	}
	if err := database.SnoozeTopic(id, body.Days); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func archiveTopic(w http.ResponseWriter, _ *http.Request, id int64) {
	if err := database.ArchiveTopic(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func unarchiveTopic(w http.ResponseWriter, _ *http.Request, id int64) {
	if err := database.UnarchiveTopic(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func parkTopicHandler(w http.ResponseWriter, id int64) {
	if err := database.ParkTopic(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func onboardTopicHandler(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Cycle *int64 `json:"cycle"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := database.OnboardTopic(id, body.Cycle); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

// ── Calendar: GET /api/calendar?year=&month= ─────────────────────────────────

func handleCalendar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := r.URL.Query()
	yearStr := q.Get("year")
	monthStr := q.Get("month")

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if yearStr != "" {
		if v, err := strconv.Atoi(yearStr); err == nil {
			year = v
		}
	}
	if monthStr != "" {
		if v, err := strconv.Atoi(monthStr); err == nil && v >= 1 && v <= 12 {
			month = v
		}
	}

	// Fetch all non-archived topics; group by next_review_at date within the month.
	topics, err := database.GetTopicsFiltered(database.TopicFilter{IncludeArchived: false})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	days := map[string]*calendarDay{}
	for _, t := range topics {
		if t.NextReviewAt.Year() != year || int(t.NextReviewAt.Month()) != month {
			continue
		}
		dateKey := t.NextReviewAt.Format("2006-01-02")
		day, ok := days[dateKey]
		if !ok {
			day = &calendarDay{Date: dateKey}
			days[dateKey] = day
		}
		dto := toDTO(t)
		if t.Completed {
			day.Completed++
		} else {
			day.Pending++
		}
		day.Topics = append(day.Topics, dto)
	}

	result := make([]calendarDay, 0, len(days))
	for _, d := range days {
		result = append(result, *d)
	}
	writeJSON(w, http.StatusOK, result)
}

// ── Projects collection: GET /api/projects, POST /api/projects ───────────────

func handleProjects(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listProjects(w, r)
	case http.MethodPost:
		createProject(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func listProjects(w http.ResponseWriter, _ *http.Request) {
	projects, err := database.GetAllProjects()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]projectDTO, 0, len(projects))
	for _, p := range projects {
		out = append(out, projectDTO{ID: p.ID, Name: p.Name, Description: p.Description})
	}
	writeJSON(w, http.StatusOK, out)
}

func createProject(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := database.AddProjectWithDescription(body.Name, body.Description); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"ok": "true"})
}

// ── Individual project: /api/projects/:id/* ───────────────────────────────────

func handleProject(w http.ResponseWriter, r *http.Request) {
	segs := splitPath(r.URL.Path, "/api/projects/")
	if len(segs) < 1 {
		writeError(w, http.StatusBadRequest, "missing project id")
		return
	}
	id, ok := parseID(segs[0])
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid project id")
		return
	}

	action := ""
	if len(segs) > 1 {
		action = segs[1]
	}

	switch {
	case r.Method == http.MethodDelete && action == "":
		deleteProject(w, id)
	case r.Method == http.MethodPut && action == "rename":
		renameProject(w, r, id)
	case r.Method == http.MethodPut && action == "describe":
		describeProject(w, r, id)
	default:
		writeError(w, http.StatusNotFound, fmt.Sprintf("unknown action: %s %s", r.Method, action))
	}
}

func deleteProject(w http.ResponseWriter, id int64) {
	if err := database.DeleteProject(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func renameProject(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Name) == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := database.RenameProject(id, body.Name); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func describeProject(w http.ResponseWriter, r *http.Request, id int64) {
	var body struct {
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := database.UpdateProjectDescription(id, body.Description); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

// ── Stats history: GET /api/stats/history?days=30 ────────────────────────────

func handleStatsHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if v, err := strconv.Atoi(d); err == nil && v > 0 && v <= 365 {
			days = v
		}
	}
	activity, err := database.GetTopicActivity(days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, activity)
}

// ── Export: GET /api/export?format=markdown|csv ───────────────────────────────

func handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "markdown"
	}

	groups, err := database.GetTopicsGroupedByProject()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var content, mime, filename string
	switch format {
	case "csv":
		content = database.RenderCSV(groups)
		mime = "text/csv"
		filename = "spaced-export.csv"
	default:
		content = database.RenderMarkdown(groups)
		mime = "text/plain; charset=utf-8"
		filename = "spaced-export.md"
	}

	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(content))
}
