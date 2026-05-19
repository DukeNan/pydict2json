# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
make                              # Build for current platform
make run                          # Build and run with args
make linux/amd64                  # Cross-compile single target
make dist                         # Cross-compile all targets
make clean                        # Remove build artifacts

# Or use go directly
go build -o pydict2json .
go run . "{'key': 'val'}"
```

Cross-compilation outputs go to `dist/<os>-<arch>/`. Windows binaries get `.exe` suffix.

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
