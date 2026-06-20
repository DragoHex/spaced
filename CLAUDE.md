# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
make build          # produces ./spd binary
go build -o spd     # equivalent

# Test
go test ./...                    # run all tests
go test ./database/...           # run only database tests
go test ./database/ -run TestXxx # run a single test

# Run CLI
./spd --help

# Run web UI (opens browser at http://localhost:7331)
./spd serve
make serve          # build + serve shorthand
./spd serve --port 8080 --no-browser   # custom port, headless
```

No lint configuration is present; use `go vet ./...` for basic checks.

## Architecture

**Spaced** is a CLI spaced-repetition tool written in Go. The binary is `spd`.

### Packages

**`cmd/`** — Cobra command handlers, one file per subcommand (`add`, `pop`, `done`, `list`, `archive`, `delete`, `modify`, `snooze`, `stats`, `export`, `import`, `project`, `serve`). `root.go` initializes the DB and wires up the root command.

**`web/`** — HTTP server and REST API handlers for the web UI. `server.go` sets up routing and embeds `web/static/` into the binary via `//go:embed`. `handlers.go` contains all REST endpoints (see REST API section). The SPA lives in `web/static/index.html` (vanilla HTML/CSS/JS, no build step).

**`database/`** — All business logic and SQLite persistence. The DB lives at `~/.spaced/spaced.db`. Key files:
- `database.go` — schema creation, session init, core topic CRUD
- `projects.go` — project table and project/topic relationships
- `sm2.go` — SM-2 spaced repetition algorithm (computes next interval and easiness factor from recall quality 0–5)
- `filters.go` — unified filtered topic query (`TopicFilter` struct)
- `snooze.go` — push `next_review_at` without advancing the review cycle
- `stats.go`, `export.go`, `import.go`, `notes.go` — analytics, CSV/Markdown export/import, notes updates

### Data model

```go
type Topic struct {
    ID, ReviewCycle, IntervalDays int64
    Topic, Notes                  string
    CreatedAt, NextReviewAt       time.Time
    Completed, Archived           bool
    EasinessFactor                float64
    ProjectID                     *int64
    ProjectName                   string
}
```

Topics belong to a Project (default project: `UNASSIGNED`). The schema uses a single SQLite file with a Topics ↔ Projects foreign key.

### Review scheduling

Two modes, selected per topic at `done` time:

| Mode | Behaviour |
|------|-----------|
| Fixed | Deterministic cycle: day 1 → 3 → 8 → 15 → 30, then completed |
| SM-2  | Adaptive: quality < 3 resets to 1-day interval; quality ≥ 3 grows interval by easiness factor |

### Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/mattn/go-sqlite3` | SQLite driver (CGO required) |
| `github.com/olekukonko/tablewriter` | Formatted table output |
| `github.com/fatih/color` | Coloured console output |
