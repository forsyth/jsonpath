package jsonpath

import "fmt"

const eof = -1

// Loc is a location (line and byte offset)
type Loc struct {
	line int // current line, origin 0, usually 0
	pos  int // byte position for next byte in whole input
}

// rd returns input units (currently bytes) from a string, allowing backing up for lexer lookahead (only 1 unit actually used).
// Since non-ASCII characters form identifiers, we can use bytes (as int) at this level.
// TO DO: Perhaps it should just use bufio.Reader, since it provides the 1 unit backing up needed.
type rd struct {
	s string
	Loc
}

// get returns the next input unit and advances the input stream.
func (r *rd) get() int {
	if r.pos >= len(r.s) {
		return eof
	}
	c := int(r.s[r.pos])
	r.pos++
	if c == '\n' {
		r.line++
	}
	return c
}

// unget backs up one unit in the input stream.
func (r *rd) unget() {
	if r.pos > 0 {
		r.pos--
		if r.s[r.pos] == '\n' {
			r.line--
		}
	}
}

// look returns the next input unit without advancing the input.
func (r *rd) look() int {
	c := r.get()
	if c != eof {
		r.unget()
	}
	return c
}

// Loc returns the current line number (origin 1) and byte offset in the input.
func (r *rd) loc() Loc {
	return Loc{r.line + 1, r.pos}
}

func (r *rd) offset() string {
	o := r.pos
	if o != 0 {
		o--
	}
	return fmt.Sprintf("offset %d", o)
}
