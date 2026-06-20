# T-007: Calendar Difficulty Heatmap

## Status
**TODO**

## Summary
Colour-code calendar days by average SM-2 quality score rather than just pending/completed dots, giving a visual sense of which periods were hard and whether retention is improving over time.

## Context
The calendar currently shows red (pending) and green (completed) dots per day. This tells you *whether* you reviewed but nothing about *how well*. Users on SM-2 scheduling could benefit from seeing effort over time — consistently low quality scores signal a topic needs more attention or should be restructured. A heatmap layer on the calendar would make this visible at a glance.

## Implementation

### Files changed
| File | Change |
|------|--------|
| `database/review_logs.go` | Add `quality` column to `review_logs` (schema migration); update `LogReview` to accept quality |
| `web/handlers.go` | Update calendar data endpoint to include avg quality per day |
| `web/static/index.html` | Render day cells with a colour gradient based on avg quality (green=easy, red=hard) |

### Key decisions
- TBD: whether to store quality in `review_logs` now (schema migration) or derive it from a separate table
- TBD: colour scale — continuous gradient vs. 3 buckets (easy/medium/hard)
- TBD: whether to show the heatmap alongside existing dots or replace them

### Technical details
TBD — requires `review_logs` schema change to add `quality INT` (nullable for fixed-schedule reviews). The `LogReview` signature would expand to `LogReview(topicID int64, quality *int) error`, propagating through the CLI `done` and web `markDone` callers.

## Validation
- Review 5 topics with varying qualities → calendar shows colour gradient reflecting avg quality
- Fixed-schedule reviews (no quality) → day renders neutral colour, not skewed

## Dependencies
Depends on existing `review_logs` table. Quality storage would be a prerequisite for accurate heatmap data.
