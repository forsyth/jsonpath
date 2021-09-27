package JSONPath

// parser represents the state of the path and/or expression parser
// (all of which is currently in the lexer), and provides a scope for the
// parsing methods.
type parser struct {
	*lexer // source of tokens
}

// newParser initialises and returns a parser
func newParser(s string) *parser {
	return &parser{newLexer(&rd{s: s})}
}
