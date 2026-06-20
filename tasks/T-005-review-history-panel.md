# T-005: Per-Topic Review History Panel

## Status
**TODO**

## Summary
Add a history drawer in the web UI showing when a topic was reviewed and how its interval evolved, so users can see the SM-2 progression and build trust in the algorithm.

## Context
There is no way to see a topic's review history. The `review_logs` table records every review timestamp, but it is only used for the aggregate activity chart. Users who want to know "when did I last review this?" or "is the interval actually growing?" have no way to find out. A history panel (slide-in drawer or expandable row) would surface this data per topic.

## Implementation

### Files changed
| File | Change |
|------|--------|
| `web/handlers.go` | Add `GET /api/topics/{id}/history` returning review_logs rows for the topic |
| `web/server.go` | Register the new route |
| `web/static/index.html` | "History" button on topic rows; drawer/modal rendering the timeline |

### Key decisions
- TBD: drawer vs. modal vs. expandable table row
- TBD: whether to show interval_days snapshots (requires storing them per review) or just timestamps

### Technical details
TBD — `review_logs` currently stores only `(topic_id, reviewed_at)`; interval snapshots would require a schema change to add `interval_days` to that table.

## Validation
- Review a topic 3 times via CLI and web → open history → verify 3 entries with correct timestamps
- Topic with no reviews → history shows empty state

## Dependencies
None (read-only; does not require T-004)
