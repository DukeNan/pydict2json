# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o pydict2json .        # Build
go run . "{'key': 'val'}"        # Run directly

# Usage
echo "{'a': 1}" | ./pydict2json
pydict2json -f input.py -o output.json
pydict2json "{'a': [1, True], 'b': None}"
pydict2json -c "{'a': 1}"       # Compact output
```

No test files, no lint config.

## Architecture

Single-file Go program (`main.go`) that converts Python dict literals to JSON. The core is a recursive-descent parser (`Parser` struct) that dispatches to sub-parsers based on the leading character:

- `parseDict` / `parseList` / `parseTuple` — container types (tuple serializes as JSON array)
- `parseString` — supports single-quoted, double-quoted, and triple-quoted strings with escape handling
- `parseNumber` — integers / floats (including scientific notation)
- `parseTrue` / `parseFalse` / `parseNone` — Python booleans and None
- `parseDatetime` — `datetime.datetime(...)`, `datetime.date(...)`, `datetime.time(...)` converted to ISO 8601 strings; keyword args (e.g. `tzinfo=`) are skipped

`orderedMap` preserves Python dict insertion order with a custom `MarshalJSON` implementation. `jsonMarshalIndent` first produces compact JSON, then re-indents via the standard library.

Input source priority: CLI argument > `-f` file > stdin.
