# Spaced

A CLI spaced-repetition tool written in Go. Binary: `spd`.

## Installation

```bash
make build   # produces ./spd
```

## Web UI

```bash
spd serve                          # opens http://localhost:7331 in your browser
spd serve --port 8080 --no-browser # custom port, no auto-open
```

## Commands

### `add <topic>`

Add a topic to your review list.

```bash
spd add "Learn Go Generics"
spd add "Dijkstra's Algorithm" --notes "BFS with priority queue" --project Algorithms
spd add "Low-priority idea" --park   # add without entering the review cycle
```

**Flags:** `--notes`, `--project <name>`, `--project-id <id>`, `--park`

---

### `pop`

Show topics due for review today.

```bash
spd pop
```

---

### `done <id>`

Mark a topic as reviewed and advance it to the next cycle.

```bash
spd done 1              # fixed schedule
spd done 1 --quality 4  # SM-2 adaptive (0 = blackout, 5 = perfect)
```

**Fixed schedule** (days from creation): Day 0 → 1 → 4 → 11 → 25 → 55 → 115, then completed.

**SM-2:** quality < 3 resets interval to 1 day; quality ≥ 3 grows interval by the easiness factor.

---

### `snooze <id>`

Postpone a topic without advancing its cycle.

```bash
spd snooze 1           # push back 1 day (default)
spd snooze 1 --days 7  # push back 7 days
```

---

### `park <id>`

Suspend a topic from the review cycle without losing its progress.

```bash
spd park 1
```

### `onboard <id>`

Return a parked topic to the review cycle (scheduled as immediately due).

```bash
spd onboard 1            # resume from its current cycle stage
spd onboard 1 --cycle 2  # restart from a specific stage (0–6)
```

---

### `list`

List all active topics.

```bash
spd list
```

### `archive <id>` / `unarchive <id>`

```bash
spd archive 1
spd unarchive 1
```

### `delete <id>`

```bash
spd delete 1
```

---

### `modify <id>`

Update a topic's text, notes, project, or review cycle.

```bash
spd modify 1 --topic "Updated topic text"
spd modify 1 --review-cycle 3   # 0–6
spd modify 1 --notes "New context" --project Algorithms
```

---

### `stats`

Show a summary of your study progress.

```bash
spd stats
```

---

### `project`

Manage projects.

```bash
spd project add "Algorithms" --description "DSA prep"
spd project list
spd project rename 2 "Data Structures"
spd project describe 2 --text "Trees, graphs, heaps"
spd project delete 2   # topics are unassigned, not deleted
```

---

### `export`

Export all topics to Markdown or CSV.

```bash
spd export                                  # Markdown to stdout
spd export --format csv --output backup.csv
spd export --format markdown --template     # generate a mock template
```

### `import <file>`

Import topics from a Markdown or CSV file. Format is auto-detected from extension.

```bash
spd import backup.md
spd import backup.csv
spd import data.txt --format markdown
```

Only topic text, notes, project name, and description are imported; IDs, dates, and cycles are recalculated fresh.

---

## Data storage

Data is stored in a SQLite database at `~/.spaced/spaced.db`.
