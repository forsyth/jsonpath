package JSONPath

const eof = -1

// rd tracks lookahead state for the lexer.
// Since no character in the grammar can be non-ASCII, we can just ints at this level.
type rd struct {
	s string
	i int
}

func (r *rd) get() int {
	if r.i >= len(r.s) {
		return eof
	}
	r.i++
	return int(r.s[r.i-1])
}

func (r *rd) unget() {
	if r.i > 0 {
		r.i--
	}
}

func (r *rd) look() int {
	if r.i >= len(r.s) {
		return eof
	}
	return int(r.s[r.i])
}
