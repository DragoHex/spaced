# T-002: Import via Web UI

## Status
**TODO**

## Summary
Add a file-upload import flow to the web UI so users can migrate or onboard topics without dropping to the terminal.

## Context
Export works in the browser (Markdown and CSV download), but import is CLI-only (`spd import <file>`). Users who manage their data entirely in the web UI are blocked when they want to restore from an export or migrate from another tool. A file-upload widget on an Import tab (or alongside the Export section) would make the tool self-contained.

## Implementation

### Files changed
| File | Change |
|------|--------|
| `web/static/index.html` | Add Import tab or section; file input; JS upload handler |
| `web/handlers.go` | Add `POST /api/import` endpoint accepting multipart form-data |
| `web/server.go` | Register the new route |
| `database/import.go` | No change expected (existing import logic is reusable) |

### Key decisions
- TBD: new "Import" tab vs. combined "Export / Import" tab
- TBD: whether to show a dry-run preview before committing

### Technical details
TBD — existing `database.ImportFromCSV` / `database.ImportFromMarkdown` functions handle the parsing; the web layer just needs to accept the upload and route to the right importer.

## Validation
- Upload a previously exported `.csv` → verify topics appear
- Upload a `.md` export → verify topics appear
- Upload an invalid file → verify user-friendly error message

## Dependencies
None
