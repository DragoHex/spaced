# T-004: Interactive Review Session (CLI)

## Status
**TODO**

## Summary
Add a `spd session` command that walks through all due topics interactively — display topic, wait for recall rating, advance to the next — without requiring manual ID lookup and copy-paste.

## Context
The current CLI review flow is: `spd pop` (see due topics) → copy ID → `spd done <id>` → repeat. This is friction-heavy. The web UI's Review Queue eliminates this by showing cards one at a time, but a terminal-native version would let keyboard-focused users study without opening a browser. The flow mirrors Anki's "study now" mode.

## Implementation

### Files changed
| File | Change |
|------|--------|
| `cmd/session.go` | New file: `session` Cobra command with interactive loop |
| `cmd/root.go` | Register `sessionCmd` |

### Key decisions
- TBD: ordering (overdue-first vs. original pop ordering)
- TBD: whether to support `--quality` inline (SM-2 mode) or fixed-only
- TBD: exit behaviour on Ctrl-C (save progress up to interruption)

### Technical details
TBD — reuse `database.GetTopicsForReview()` for the topic list, `database.MarkTopicDone()` / `database.MarkTopicDoneWithQuality()` for advancement, and `database.LogReview()` for the activity log.

## Validation
- `spd session` with 3 due topics → walk through all 3 → verify next_review_at updated for each
- Ctrl-C mid-session → verify already-reviewed topics are saved
- `spd session` with no due topics → verify graceful empty message

## Dependencies
None
