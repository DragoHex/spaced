# T-006: Notifications / Daily Reminders

## Status
**TODO**

## Summary
Add a mechanism to notify the user when topics are due, so reviews happen without needing to remember to open the tool.

## Context
The tool currently has no push mechanism. Reviews only happen when the user actively opens the CLI or web UI. A daily reminder (desktop notification, OS notification, or a `spd notify` command suitable for cron) would close the loop and make the tool behave more like a habit rather than a task.

## Implementation

### Files changed
| File | Change |
|------|--------|
| `cmd/notify.go` | New file: `spd notify` command — prints due count or sends OS notification |

### Key decisions
- TBD: delivery mechanism — stdout summary (cron-friendly), macOS `osascript` notification, cross-platform approach
- TBD: whether to embed this in `serve` as a scheduled background job vs. a standalone command
- TBD: configurable notification time (morning digest vs. real-time)

### Technical details
TBD — `database.GetStats().DueToday` gives the count; `database.GetTopicsForReview()` gives the list for a richer message.

## Validation
- `spd notify` with 0 due → exits cleanly with no-op or "nothing due" message
- `spd notify` with N due → emits correct count
- Cron entry `0 9 * * * spd notify` works without interactive shell

## Dependencies
None
