# pydict2json

A command-line tool that converts Python dict literals to JSON.

## Features

- Parses Python dict syntax and outputs valid JSON
- Supported types: dict, list, tuple, str (single/double/triple-quoted), int, float (with scientific notation), True, False, None
- Parses `datetime.datetime(...)`, `datetime.date(...)`, `datetime.time(...)` as ISO 8601 strings
- Preserves Python dict insertion order
- Supports nested structures
- Supports trailing commas

## Installation

```bash
go build -o pydict2json .
```

## Usage

```bash
# Read from stdin
echo "{'key': 'val', 'n': 42}" | pydict2json

# Read from file
pydict2json -f data.py -o data.json

# Pass inline
pydict2json "{'a': [1, 2, True], 'b': None}"

# Compact output
pydict2json -c "{'a': 1, 'b': 2}"

# Datetime types
pydict2json "{'ts': datetime.datetime(2024, 1, 15, 10, 30, 0)}"
pydict2json "{'d': datetime.date(2024, 1, 15)}"
pydict2json "{'t': datetime.time(10, 30, 0)}"
```

## Options

| Option | Default | Description |
|--------|---------|-------------|
| `-f <file>` | stdin | Input file |
| `-o <file>` | stdout | Output file |
| `-p` | true | Pretty-print output (indented) |
| `-c` | false | Compact output (overrides `-p`) |
| `-h` | — | Show help |

## Examples

Input:

```python
{'name': 'Alice', 'scores': [95.5, 88, 92], 'active': True, 'note': None}
```

Output:

```json
{
  "name": "Alice",
  "scores": [95.5, 88, 92],
  "active": true,
  "note": null
}
```

Datetime input:

```python
{'created': datetime.datetime(2024, 1, 15, 10, 30, 0), 'birthday': datetime.date(2000, 6, 1), 'alarm': datetime.time(7, 0, 0)}
```

Output:

```json
{
  "created": "2024-01-15T10:30:00",
  "birthday": "2000-06-01",
  "alarm": "07:00:00"
}
```

## Type Mapping

| Python | JSON |
|--------|------|
| dict | object |
| list | array |
| tuple | array |
| str | string |
| int | number |
| float | number |
| True | true |
| False | false |
| None | null |
| datetime.datetime(...) | string (ISO 8601) |
| datetime.date(...) | string (ISO 8601) |
| datetime.time(...) | string (ISO 8601) |
