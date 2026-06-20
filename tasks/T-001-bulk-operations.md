# T-001: Bulk Operations in Web UI

## Status
**TODO**

## Summary
Allow users to select multiple topics and apply a single action (snooze, archive, assign to project) to all of them at once.

## Context
Every action in the Topics tab is per-topic: snooze, archive, and project-assign all require individual clicks. When doing housekeeping on a large topic set — e.g. archiving everything in a completed project or snoozing a batch of low-priority topics — this becomes tedious. Bulk operations would reduce this to a two-click workflow.

## Implementation

### Files changed
| File | Change |
|------|--------|
| `web/static/index.html` | Add row checkboxes, bulk action toolbar, JS handlers |
| `web/handlers.go` | Add bulk endpoints (or reuse existing per-topic endpoints in a loop) |

### Key decisions
- TBD: server-side bulk endpoint vs. client-side sequential API calls (simpler, acceptable for <100 items)
- TBD: scope — which actions to support (snooze N days, archive, project-assign)

### Technical details
TBD

## Validation
- Select N topics → bulk action → verify all N affected
- Verify partial failure (one topic already archived) doesn't block others

## Dependencies
None
