# T-003: Topic Tagging

## Status
**TODO**

## Summary
Add cross-cutting tags to topics, orthogonal to the existing project grouping, so users can filter across projects by subject or theme.

## Context
Projects are containers: a topic belongs to exactly one project. There is no way to express cross-cutting concerns (e.g. `#math`, `#interview-prep`, `#work`) that span multiple projects. Tags would allow filtering like "show all interview-prep topics regardless of project," which is not possible today.

## Implementation

### Files changed
| File | Change |
|------|--------|
| `database/database.go` | Add `tags` column to topics table (comma-separated string or junction table) |
| `database/filters.go` | Add `Tags []string` to `TopicFilter`; extend query to filter by tag |
| `database/*.go` | Update CRUD functions to read/write tags |
| `web/handlers.go` | Expose tags in topic JSON; add tag filter query param |
| `web/static/index.html` | Tag chips on topic cards/rows; tag input in Add/Edit modal; tag filter in Topics tab |
| `cmd/add.go`, `cmd/modify.go` | Add `--tags` flag |

### Key decisions
- TBD: comma-separated column (simpler, harder to query) vs. junction table (cleaner, more migration work)
- TBD: whether tags appear in the Review Queue cards

### Technical details
TBD

## Validation
- Add topic with tags → verify tags persist and display
- Filter by tag → verify only tagged topics shown
- CLI `--tags` flag → verify tags stored

## Dependencies
None
