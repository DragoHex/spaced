# Spaced

A simple CLI tool for spaced repetition.

## Commands

### `add [topic]`

Adds a new topic to your spaced repetition list.

**Example:**
```
spaced add "Learn Go Generics"
```

### `pop`

Shows the topics that are due for review today.

**Example:**
```
spaced pop
```

### `done [topic_id]`

Marks a topic as reviewed. This will push the topic to the next review cycle.

**Example:**
```
spaced done 1
```

### `list`

Lists all the topics in your list, including their review status.

**Example:**
```
spaced list
```

### `archive [topic_id]`

Archives a completed topic. This will hide it from the main list.

**Example:**
```
spaced archive 1
```

### `unarchive [topic_id]`

Unarchives a topic, making it visible in the main list again.

**Example:**
```
spaced unarchive 1
```

### `modify [topic_id] --topic "new_topic_text" --review-cycle <cycle_number>`

Modifies an existing topic's text or its review cycle. You must provide at least one of the flags.

Review cycles are:
- 0: Day 1 (initial addition)
- 1: Day 3
- 2: Day 8
- 3: Day 15
- 4: Day 30

**Examples:**
```
spaced modify 1 --topic "Learn advanced Go Generics and Concurrency"
spaced modify 1 --review-cycle 2
spaced modify 1 --topic "Learn advanced Go Generics and Concurrency" --review-cycle 2
```

### `delete [topic_id]`

Deletes a topic from your list.

**Example:**
```
spaced delete 1
```
