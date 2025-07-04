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

### `modify [topic_id] [new_topic]`

Modifies the text of an existing topic.

**Example:**
```
spaced modify 1 "Learn advanced Go Generics"
```

### `delete [topic_id]`

Deletes a topic from your list.

**Example:**
```
spaced delete 1
```
