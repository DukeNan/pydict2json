package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// ============================================================
// Python Dict Parser
// ============================================================

type Parser struct {
	input []rune
	pos   int
}

func NewParser(s string) *Parser {
	return &Parser{input: []rune(s), pos: 0}
}

func (p *Parser) peek() (rune, bool) {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return 0, false
	}
	return p.input[p.pos], true
}

func (p *Parser) current() rune {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *Parser) skipWhitespace() {
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			p.pos++
		} else {
			break
		}
	}
}

func (p *Parser) consume(expected rune) error {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return fmt.Errorf("unexpected end of input, expected '%c'", expected)
	}
	if p.input[p.pos] != expected {
		ctx := p.context()
		return fmt.Errorf("expected '%c', got '%c' at position %d: ...%s...", expected, p.input[p.pos], p.pos, ctx)
	}
	p.pos++
	return nil
}

func (p *Parser) context() string {
	start := p.pos - 10
	if start < 0 {
		start = 0
	}
	end := p.pos + 10
	if end > len(p.input) {
		end = len(p.input)
	}
	return string(p.input[start:end])
}

// Parse dispatches to the right sub-parser based on the next character.
func (p *Parser) Parse() (interface{}, error) {
	c, ok := p.peek()
	if !ok {
		return nil, fmt.Errorf("empty input")
	}
	switch {
	case c == '{':
		return p.parseDict()
	case c == '[':
		return p.parseList()
	case c == '(':
		return p.parseTuple()
	case c == '\'' || c == '"':
		return p.parseString()
	case c == 'T':
		return p.parseTrue()
	case c == 'F':
		return p.parseFalse()
	case c == 'N':
		return p.parseNone()
	case c == '-' || (c >= '0' && c <= '9'):
		return p.parseNumber()
	default:
		return nil, fmt.Errorf("unexpected character '%c' at position %d: ...%s...", c, p.pos, p.context())
	}
}

func (p *Parser) parseDict() (interface{}, error) {
	if err := p.consume('{'); err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	// ordered keys to preserve insertion order in output
	order := []string{}

	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '}' {
		p.pos++
		return result, nil
	}

	for {
		p.skipWhitespace()
		// Parse key (Python dict keys can be strings, ints, bools)
		keyVal, err := p.Parse()
		if err != nil {
			return nil, fmt.Errorf("dict key: %w", err)
		}
		key := fmt.Sprintf("%v", keyVal)

		if err := p.consume(':'); err != nil {
			return nil, fmt.Errorf("dict colon: %w", err)
		}

		val, err := p.Parse()
		if err != nil {
			return nil, fmt.Errorf("dict value for key '%s': %w", key, err)
		}
		result[key] = val
		order = append(order, key)

		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unterminated dict")
		}
		if p.input[p.pos] == '}' {
			p.pos++
			break
		}
		if p.input[p.pos] == ',' {
			p.pos++
			p.skipWhitespace()
			// trailing comma support
			if p.pos < len(p.input) && p.input[p.pos] == '}' {
				p.pos++
				break
			}
		}
	}
	// Return as ordered map representation using a slice of pairs,
	// but since json.Marshal on map is unordered, we use a helper.
	return &orderedMap{keys: order, data: result}, nil
}

func (p *Parser) parseList() (interface{}, error) {
	if err := p.consume('['); err != nil {
		return nil, err
	}
	var result []interface{}

	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == ']' {
		p.pos++
		return result, nil
	}

	for {
		val, err := p.Parse()
		if err != nil {
			return nil, fmt.Errorf("list element: %w", err)
		}
		result = append(result, val)

		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unterminated list")
		}
		if p.input[p.pos] == ']' {
			p.pos++
			break
		}
		if p.input[p.pos] == ',' {
			p.pos++
			p.skipWhitespace()
			if p.pos < len(p.input) && p.input[p.pos] == ']' {
				p.pos++
				break
			}
		}
	}
	return result, nil
}

func (p *Parser) parseTuple() (interface{}, error) {
	// Tuples become JSON arrays
	if err := p.consume('('); err != nil {
		return nil, err
	}
	var result []interface{}

	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == ')' {
		p.pos++
		return result, nil
	}

	for {
		val, err := p.Parse()
		if err != nil {
			return nil, fmt.Errorf("tuple element: %w", err)
		}
		result = append(result, val)

		p.skipWhitespace()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("unterminated tuple")
		}
		if p.input[p.pos] == ')' {
			p.pos++
			break
		}
		if p.input[p.pos] == ',' {
			p.pos++
			p.skipWhitespace()
			if p.pos < len(p.input) && p.input[p.pos] == ')' {
				p.pos++
				break
			}
		}
	}
	return result, nil
}

func (p *Parser) parseString() (interface{}, error) {
	p.skipWhitespace()
	quote := p.input[p.pos]
	p.pos++

	// Detect triple-quoted strings
	triple := false
	if p.pos+1 < len(p.input) && p.input[p.pos] == quote && p.input[p.pos+1] == quote {
		triple = true
		p.pos += 2
	}

	var sb strings.Builder
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if triple {
			if c == quote && p.pos+2 < len(p.input) && p.input[p.pos+1] == quote && p.input[p.pos+2] == quote {
				p.pos += 3
				return sb.String(), nil
			}
		} else {
			if c == quote {
				p.pos++
				return sb.String(), nil
			}
		}
		if c == '\\' && p.pos+1 < len(p.input) {
			p.pos++
			esc := p.input[p.pos]
			switch esc {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			case 'b':
				sb.WriteRune('\b')
			case 'f':
				sb.WriteRune('\f')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(esc)
			}
			p.pos++
			continue
		}
		sb.WriteRune(c)
		p.pos++
	}
	return nil, fmt.Errorf("unterminated string")
}

func (p *Parser) parseTrue() (interface{}, error) {
	if p.pos+4 <= len(p.input) && string(p.input[p.pos:p.pos+4]) == "True" {
		p.pos += 4
		return true, nil
	}
	return nil, fmt.Errorf("expected 'True' at position %d", p.pos)
}

func (p *Parser) parseFalse() (interface{}, error) {
	if p.pos+5 <= len(p.input) && string(p.input[p.pos:p.pos+5]) == "False" {
		p.pos += 5
		return false, nil
	}
	return nil, fmt.Errorf("expected 'False' at position %d", p.pos)
}

func (p *Parser) parseNone() (interface{}, error) {
	if p.pos+4 <= len(p.input) && string(p.input[p.pos:p.pos+4]) == "None" {
		p.pos += 4
		return nil, nil
	}
	return nil, fmt.Errorf("expected 'None' at position %d", p.pos)
}

func (p *Parser) parseNumber() (interface{}, error) {
	p.skipWhitespace()
	start := p.pos
	isFloat := false

	if p.pos < len(p.input) && p.input[p.pos] == '-' {
		p.pos++
	}
	for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
		p.pos++
	}
	if p.pos < len(p.input) && p.input[p.pos] == '.' {
		isFloat = true
		p.pos++
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
			p.pos++
		}
	}
	if p.pos < len(p.input) && (p.input[p.pos] == 'e' || p.input[p.pos] == 'E') {
		isFloat = true
		p.pos++
		if p.pos < len(p.input) && (p.input[p.pos] == '+' || p.input[p.pos] == '-') {
			p.pos++
		}
		for p.pos < len(p.input) && p.input[p.pos] >= '0' && p.input[p.pos] <= '9' {
			p.pos++
		}
	}

	numStr := string(p.input[start:p.pos])
	if isFloat {
		var f float64
		if _, err := fmt.Sscanf(numStr, "%f", &f); err != nil {
			return nil, fmt.Errorf("invalid float: %s", numStr)
		}
		return f, nil
	}
	var i int64
	if _, err := fmt.Sscanf(numStr, "%d", &i); err != nil {
		return nil, fmt.Errorf("invalid integer: %s", numStr)
	}
	return i, nil
}

// ============================================================
// orderedMap: preserves Python dict insertion order in JSON output
// ============================================================

type orderedMap struct {
	keys []string
	data map[string]interface{}
}

func (o *orderedMap) MarshalJSON() ([]byte, error) {
	var sb strings.Builder
	sb.WriteRune('{')
	for i, k := range o.keys {
		if i > 0 {
			sb.WriteRune(',')
		}
		keyBytes, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		sb.Write(keyBytes)
		sb.WriteRune(':')
		valBytes, err := marshalValue(o.data[k])
		if err != nil {
			return nil, err
		}
		sb.Write(valBytes)
	}
	sb.WriteRune('}')
	return []byte(sb.String()), nil
}

func marshalValue(v interface{}) ([]byte, error) {
	if om, ok := v.(*orderedMap); ok {
		return om.MarshalJSON()
	}
	if arr, ok := v.([]interface{}); ok {
		var sb strings.Builder
		sb.WriteRune('[')
		for i, item := range arr {
			if i > 0 {
				sb.WriteRune(',')
			}
			b, err := marshalValue(item)
			if err != nil {
				return nil, err
			}
			sb.Write(b)
		}
		sb.WriteRune(']')
		return []byte(sb.String()), nil
	}
	return json.Marshal(v)
}

// ============================================================
// Entry point
// ============================================================

func convert(input string, indent bool) (string, error) {
	parser := NewParser(strings.TrimSpace(input))
	parsed, err := parser.Parse()
	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	var jsonBytes []byte
	if indent {
		jsonBytes, err = jsonMarshalIndent(parsed, "", "  ")
	} else {
		jsonBytes, err = marshalValue(parsed)
	}
	if err != nil {
		return "", fmt.Errorf("JSON marshal error: %w", err)
	}
	return string(jsonBytes), nil
}

// jsonMarshalIndent handles orderedMap by first producing compact JSON then re-indenting.
func jsonMarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	compact, err := marshalValue(v)
	if err != nil {
		return nil, err
	}
	var buf strings.Builder
	dec := json.NewDecoder(strings.NewReader(string(compact)))
	dec.UseNumber()
	var raw interface{}
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	enc := json.NewEncoder(&buf)
	enc.SetIndent(prefix, indent)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(raw); err != nil {
		return nil, err
	}
	// Encode adds a trailing newline; trim it
	return []byte(strings.TrimRight(buf.String(), "\n")), nil
}

func main() {
	var (
		inputFile  = flag.String("f", "", "Input file containing Python dict (default: stdin)")
		outputFile = flag.String("o", "", "Output file for JSON (default: stdout)")
		pretty     = flag.Bool("p", true, "Pretty-print JSON with indentation")
		compact    = flag.Bool("c", false, "Compact JSON output (overrides -p)")
		help       = flag.Bool("h", false, "Show help")
	)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `pydict2json — Convert Python dict syntax to JSON

USAGE:
  pydict2json [options] [inline-python-dict]

OPTIONS:
  -f <file>   Read Python dict from file (default: stdin)
  -o <file>   Write JSON to file (default: stdout)
  -p          Pretty-print output (default: true)
  -c          Compact output (overrides -p)
  -h          Show this help

EXAMPLES:
  echo "{'key': 'val', 'n': 42}" | pydict2json
  pydict2json -f data.py -o data.json
  pydict2json "{'a': [1, 2, True], 'b': None}"

SUPPORTED PYTHON TYPES:
  dict  list  tuple  str (single/double/triple-quoted)
  int   float  True  False  None  nested structures`)
	}
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	indent := *pretty && !*compact

	// Determine input source
	var inputData string

	if flag.NArg() > 0 {
		// Inline argument
		inputData = strings.Join(flag.Args(), " ")
	} else if *inputFile != "" {
		bytes, err := os.ReadFile(*inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", *inputFile, err)
			os.Exit(1)
		}
		inputData = string(bytes)
	} else {
		// Read from stdin
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 512)
		for {
			n, err := os.Stdin.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
			}
			if err != nil {
				break
			}
		}
		inputData = string(buf)
	}

	if strings.TrimSpace(inputData) == "" {
		fmt.Fprintln(os.Stderr, "Error: no input provided. Use -h for help.")
		os.Exit(1)
	}

	result, err := convert(inputData, indent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Conversion failed: %v\n", err)
		os.Exit(1)
	}

	if *outputFile != "" {
		if err := os.WriteFile(*outputFile, []byte(result+"\n"), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to '%s': %v\n", *outputFile, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "JSON written to %s\n", *outputFile)
	} else {
		fmt.Println(result)
	}
}
